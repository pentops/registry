package service

import (
	"context"
	"time"

	sq "github.com/elgris/sqrl"
	"github.com/google/uuid"
	"github.com/pentops/registry/internal/gen/o5/registry/github/v1/github_pb"
	"github.com/pentops/registry/internal/gen/o5/registry/github/v1/github_spb"
	"github.com/pentops/registry/internal/state"
	"github.com/pentops/sqrlx.go/sqrlx"
)

type GithubCommandService struct {
	db *sqrlx.Wrapper

	stateMachines *state.StateMachines
	*github_spb.UnimplementedGithubCommandServiceServer
}

func NewGithubCommandService(conn sqrlx.Connection, sm *state.StateMachines) (*GithubCommandService, error) {
	db, err := sqrlx.New(conn, sq.Dollar)
	if err != nil {
		return nil, err
	}

	return &GithubCommandService{
		db:            db,
		stateMachines: sm,
	}, nil
}

func (ss *GithubCommandService) ConfigureRepo(ctx context.Context, req *github_spb.ConfigureRepoRequest) (*github_spb.ConfigureRepoResponse, error) {

	cause, err := CommandCause(ctx)
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
		Cause:     cause,
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
