package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/elgris/sqrl"
	"github.com/pentops/registry/gen/o5/registry/github/v1/github_pb"
	"github.com/pentops/registry/gen/o5/registry/github/v1/github_tpb"
	"github.com/pentops/sqrlx.go/sqrlx"
	"google.golang.org/protobuf/encoding/protojson"
)

type RefStore struct {
	db *sqrlx.Wrapper
}

func NewRefStore(conn sqrlx.Connection) (*RefStore, error) {
	db, err := sqrlx.New(conn, sq.Dollar)
	if err != nil {
		return nil, err
	}

	return &RefStore{
		db: db,
	}, nil
}

func (rs *RefStore) GetRepo(ctx context.Context, push *github_tpb.PushMessage) (*github_pb.RepoState, error) {
	qq := sq.
		Select("state").
		From("repo").
		Where(sq.Eq{
			"owner": push.Owner,
			"name":  push.Repo,
		})

	var stateBytes []byte

	err := rs.db.Transact(ctx, &sqrlx.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  true,
		Retryable: true,
	}, func(ctx context.Context, tx sqrlx.Transaction) error {

		err := tx.SelectRow(ctx, qq).Scan(&stateBytes)
		if err != nil {
			return err
		}
		return nil
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("selecting push targets: %w", err)
	}

	repo := &github_pb.RepoState{}
	if err := protojson.Unmarshal(stateBytes, repo); err != nil {
		return nil, fmt.Errorf("unmarshalling repo state: %w", err)
	}

	return repo, nil
}
