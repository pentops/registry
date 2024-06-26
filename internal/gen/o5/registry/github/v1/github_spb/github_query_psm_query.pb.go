// Code generated by protoc-gen-go-psm. DO NOT EDIT.

package github_spb

import (
	psm "github.com/pentops/protostate/psm"
)

// State Query Service for %sRepo
// QuerySet is the query set for the Repo service.

type RepoPSMQuerySet = psm.StateQuerySet[
	*GetRepoRequest,
	*GetRepoResponse,
	*ListReposRequest,
	*ListReposResponse,
	*ListRepoEventsRequest,
	*ListRepoEventsResponse,
]

func NewRepoPSMQuerySet(
	smSpec psm.QuerySpec[
		*GetRepoRequest,
		*GetRepoResponse,
		*ListReposRequest,
		*ListReposResponse,
		*ListRepoEventsRequest,
		*ListRepoEventsResponse,
	],
	options psm.StateQueryOptions,
) (*RepoPSMQuerySet, error) {
	return psm.BuildStateQuerySet[
		*GetRepoRequest,
		*GetRepoResponse,
		*ListReposRequest,
		*ListReposResponse,
		*ListRepoEventsRequest,
		*ListRepoEventsResponse,
	](smSpec, options)
}

type RepoPSMQuerySpec = psm.QuerySpec[
	*GetRepoRequest,
	*GetRepoResponse,
	*ListReposRequest,
	*ListReposResponse,
	*ListRepoEventsRequest,
	*ListRepoEventsResponse,
]

func DefaultRepoPSMQuerySpec(tableSpec psm.QueryTableSpec) RepoPSMQuerySpec {
	return psm.QuerySpec[
		*GetRepoRequest,
		*GetRepoResponse,
		*ListReposRequest,
		*ListReposResponse,
		*ListRepoEventsRequest,
		*ListRepoEventsResponse,
	]{
		QueryTableSpec: tableSpec,
		ListRequestFilter: func(req *ListReposRequest) (map[string]interface{}, error) {
			filter := map[string]interface{}{}
			return filter, nil
		},
		ListEventsRequestFilter: func(req *ListRepoEventsRequest) (map[string]interface{}, error) {
			filter := map[string]interface{}{}
			filter["owner"] = req.Owner
			filter["name"] = req.Name
			return filter, nil
		},
	}
}
