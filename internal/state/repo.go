package state

import (
	"github.com/pentops/protostate/psm"
	"github.com/pentops/registry/internal/gen/o5/registry/github/v1/github_pb"
)

func NewRepoPSM() (*github_pb.RepoPSM, error) {
	sm, err := github_pb.RepoPSMBuilder().
		SystemActor(psm.MustSystemActor("216B6C2E-D996-492C-B80C-9AAD0CCFEEC4")).
		BuildStateMachine()
	if err != nil {
		return nil, err
	}

	sm.From(
		github_pb.RepoStatus_UNSPECIFIED,
		github_pb.RepoStatus_ACTIVE,
	).
		OnEvent(github_pb.RepoPSMEventConfigure).
		SetStatus(github_pb.RepoStatus_ACTIVE).
		Mutate(github_pb.RepoPSMMutation(func(
			state *github_pb.RepoStateData,
			event *github_pb.RepoEventType_Configure,
		) error {
			state.ChecksEnabled = event.ChecksEnabled

			if event.Merge {
				for _, branch := range state.Branches {
					state.Branches = mergeBranch(state.Branches, branch)
				}
			} else {
				state.Branches = event.Branches
			}
			return nil
		}))

	sm.From(
		github_pb.RepoStatus_UNSPECIFIED,
		github_pb.RepoStatus_ACTIVE,
	).
		OnEvent(github_pb.RepoPSMEventConfigureBranch).
		SetStatus(github_pb.RepoStatus_ACTIVE).
		Mutate(github_pb.RepoPSMMutation(func(
			state *github_pb.RepoStateData,
			event *github_pb.RepoEventType_ConfigureBranch,
		) error {
			state.Branches = mergeBranch(state.Branches, event.Branch)

			return nil
		}))

	sm.From(
		github_pb.RepoStatus_ACTIVE,
	).OnEvent(github_pb.RepoPSMEventRemoveBranch).
		Mutate(github_pb.RepoPSMMutation(func(
			state *github_pb.RepoStateData,
			event *github_pb.RepoEventType_RemoveBranch,
		) error {
			for i, branch := range state.Branches {
				if branch.BranchName == event.BranchName {
					state.Branches = append(state.Branches[:i], state.Branches[i+1:]...)
					return nil
				}
			}

			return nil
		}))

	return sm, nil
}

func mergeBranch(branches []*github_pb.Branch, newBranch *github_pb.Branch) []*github_pb.Branch {
	for i, branch := range branches {
		if branch.BranchName == newBranch.BranchName {
			branches[i] = newBranch
			return branches
		}
	}

	return append(branches, newBranch)
}
