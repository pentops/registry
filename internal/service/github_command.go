package service

import (
	"context"
	"fmt"
	"time"

	sq "github.com/elgris/sqrl"
	"github.com/google/uuid"
	"github.com/pentops/o5-auth/o5auth"
	"github.com/pentops/registry/internal/gen/j5/registry/github/v1/github_pb"
	"github.com/pentops/registry/internal/gen/j5/registry/github/v1/github_spb"
	"github.com/pentops/registry/internal/github"
	"github.com/pentops/registry/internal/state"
	"github.com/pentops/sqrlx.go/sqrlx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GithubCommandService struct {
	db *sqrlx.Wrapper

	stateMachines *state.StateMachines
	*github_spb.UnimplementedRepoCommandServiceServer

	builder targetBuilder
	refs    RefMatcher
}

type targetBuilder interface {
	buildTarget(ctx context.Context, ref github.RepoRef, target *github_pb.DeployTargetType) error
}

func NewGithubCommandService(conn sqrlx.Connection, sm *state.StateMachines, builder targetBuilder) (*GithubCommandService, error) {
	db, err := sqrlx.New(conn, sq.Dollar)
	if err != nil {
		return nil, err
	}

	refs, err := NewRefStore(conn)
	if err != nil {
		return nil, err
	}

	return &GithubCommandService{
		db:            db,
		stateMachines: sm,
		builder:       builder,
		refs:          refs,
	}, nil
}

func (ss *GithubCommandService) RegisterGRPC(srv *grpc.Server) {
	github_spb.RegisterRepoCommandServiceServer(srv, ss)
}

func (ss *GithubCommandService) ConfigureRepo(ctx context.Context, req *github_spb.ConfigureRepoRequest) (*github_spb.ConfigureRepoResponse, error) {

	action, err := o5auth.GetAuthenticatedAction(ctx)
	if err != nil {
		return nil, err
	}

	evt := &github_pb.RepoPSMEventSpec{
		Keys: &github_pb.RepoKeys{
			Owner: req.Owner,
			Name:  req.Name,
		},
		EventID:   uuid.NewString(),
		Timestamp: time.Now(),
		Action:    action,
		Event:     req.Config,
	}

	newState, err := ss.stateMachines.Repo.Transition(ctx, ss.db, evt)
	if err != nil {
		return nil, err
	}

	return &github_spb.ConfigureRepoResponse{
		Repo: newState,
	}, nil

}
func (ss *GithubCommandService) Trigger(ctx context.Context, req *github_spb.TriggerRequest) (*github_spb.TriggerResponse, error) {

	_, err := o5auth.GetAuthenticatedAction(ctx)
	if err != nil {
		return nil, err
	}

	repo, err := ss.refs.GetRepo(ctx, req.Owner, req.Repo)
	if err != nil {
		return nil, fmt.Errorf("get repo: %w", err)
	}
	if repo == nil {
		return nil, status.Error(codes.NotFound, "repo not found")
	}

	ref := github.RepoRef{
		Owner: repo.Keys.Owner,
		Repo:  repo.Keys.Name,
		Ref:   req.Commit,
	}

	err = ss.builder.buildTarget(ctx, ref, req.Target)
	if err != nil {
		return nil, fmt.Errorf("build targets: %w", err)
	}

	return &github_spb.TriggerResponse{}, nil
}
