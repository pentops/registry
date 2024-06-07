package service

import (
	"context"

	sq "github.com/elgris/sqrl"
	"github.com/pentops/protostate/psm"
	"github.com/pentops/registry/gen/o5/registry/github/v1/github_spb"
	"github.com/pentops/registry/state"
	"github.com/pentops/sqrlx.go/sqrlx"
)

type RepoQueryService struct {
	db *sqrlx.Wrapper

	querySet *github_spb.RepoPSMQuerySet
	*github_spb.UnimplementedRepoQueryServiceServer
}

func NewRepoQueryService(conn sqrlx.Connection, states *state.StateMachines) (*RepoQueryService, error) {
	db, err := sqrlx.New(conn, sq.Dollar)
	if err != nil {
		return nil, err

	}

	querySpec := github_spb.DefaultRepoPSMQuerySpec(states.Repo.StateTableSpec())
	querySet, err := github_spb.NewRepoPSMQuerySet(querySpec, psm.StateQueryOptions{})
	if err != nil {
		return nil, err
	}

	return &RepoQueryService{
		db:       db,
		querySet: querySet,
	}, nil
}

func (ds *RepoQueryService) ListRepoEvents(ctx context.Context, req *github_spb.ListRepoEventsRequest) (*github_spb.ListRepoEventsResponse, error) {
	res := &github_spb.ListRepoEventsResponse{}

	return res, ds.querySet.ListEvents(ctx, ds.db, req, res)
}

func (ds *RepoQueryService) GetRepo(ctx context.Context, req *github_spb.GetRepoRequest) (*github_spb.GetRepoResponse, error) {
	res := &github_spb.GetRepoResponse{}

	return res, ds.querySet.Get(ctx, ds.db, req, res)
}

func (ds *RepoQueryService) ListRepos(ctx context.Context, req *github_spb.ListReposRequest) (*github_spb.ListReposResponse, error) {
	res := &github_spb.ListReposResponse{}

	return res, ds.querySet.List(ctx, ds.db, req, res)
}
