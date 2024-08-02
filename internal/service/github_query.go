package service

import (
	"context"

	sq "github.com/elgris/sqrl"
	"github.com/pentops/protostate/psm"
	"github.com/pentops/registry/internal/gen/j5/registry/github/v1/github_spb"
	"github.com/pentops/registry/internal/state"
	"github.com/pentops/sqrlx.go/sqrlx"
)

type GithubQueryService struct {
	db *sqrlx.Wrapper

	querySet *github_spb.RepoPSMQuerySet
	*github_spb.UnimplementedRepoQueryServiceServer
}

func NewGithubQueryService(conn sqrlx.Connection, states *state.StateMachines) (*GithubQueryService, error) {
	db, err := sqrlx.New(conn, sq.Dollar)
	if err != nil {
		return nil, err

	}

	querySpec := github_spb.DefaultRepoPSMQuerySpec(states.Repo.StateTableSpec())
	querySet, err := github_spb.NewRepoPSMQuerySet(querySpec, psm.StateQueryOptions{})
	if err != nil {
		return nil, err
	}

	return &GithubQueryService{
		db:       db,
		querySet: querySet,
	}, nil
}

func (ds *GithubQueryService) ListRepoEvents(ctx context.Context, req *github_spb.ListRepoEventsRequest) (*github_spb.ListRepoEventsResponse, error) {
	res := &github_spb.ListRepoEventsResponse{}

	return res, ds.querySet.ListEvents(ctx, ds.db, req, res)
}

func (ds *GithubQueryService) GetRepo(ctx context.Context, req *github_spb.GetRepoRequest) (*github_spb.GetRepoResponse, error) {
	res := &github_spb.GetRepoResponse{}

	return res, ds.querySet.Get(ctx, ds.db, req, res)
}

func (ds *GithubQueryService) ListRepos(ctx context.Context, req *github_spb.ListReposRequest) (*github_spb.ListReposResponse, error) {
	res := &github_spb.ListReposResponse{}

	return res, ds.querySet.List(ctx, ds.db, req, res)
}
