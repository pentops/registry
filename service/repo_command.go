package service

import (
	"context"
	"time"

	sq "github.com/elgris/sqrl"
	"github.com/google/uuid"
	"github.com/pentops/registry/gen/o5/registry/github/v1/github_pb"
	"github.com/pentops/registry/gen/o5/registry/github/v1/github_spb"
	"github.com/pentops/registry/state"
	"github.com/pentops/sqrlx.go/sqrlx"
)

type RepoCommandService struct {
	db *sqrlx.Wrapper

	stateMachines *state.StateMachines
	*github_spb.UnimplementedGithubCommandServiceServer
}

func NewRepoCommandService(conn sqrlx.Connection, sm *state.StateMachines) (*RepoCommandService, error) {
	db, err := sqrlx.New(conn, sq.Dollar)
	if err != nil {
		return nil, err
	}

	return &RepoCommandService{
		db:            db,
		stateMachines: sm,
	}, nil
}

func (ss *RepoCommandService) ConfigureRepo(ctx context.Context, req *github_spb.ConfigureRepoRequest) (*github_spb.ConfigureRepoResponse, error) {

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
