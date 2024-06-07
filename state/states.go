package state

import (
	"fmt"

	"github.com/pentops/registry/gen/o5/registry/github/v1/github_pb"
)

type StateMachines struct {
	Repo *github_pb.RepoPSM
}

func NewStateMachines() (*StateMachines, error) {
	repo, err := NewRepoPSM()
	if err != nil {
		return nil, fmt.Errorf("NewDeploymentEventer: %w", err)
	}

	return &StateMachines{
		Repo: repo,
	}, nil
}
