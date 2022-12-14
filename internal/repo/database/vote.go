package database

import (
	"context"
	"errors"
	"fmt"
	"goquizbox/internal/repo/model"
	"goquizbox/pkg/database"
	"strings"

	pgx "github.com/jackc/pgx/v4"
)

const (
	createVoteSQL              = `insert into votes (user_id, kind_id, kind, mode, created_at) values ($1, $2, $3, $4, $5) returning id`
	selectVoteSQL              = `select id, user_id, kind_id, kind, mode, created_at from votes`
	selectVoteByUserAndKindSQL = selectVoteSQL + ` where user_id = $1 and kind_id = $2 and kind = $3`
	updateVoteSQL              = `update votes set (mode, updated_at) = ($1, $2) where id = $3`
	countVotesSQL              = `select count(id) from votes where kind_id= $1 and kind = $2 and mode = $3`
)

type VoteDB struct {
	db *database.DB
}

func NewVoteDB(db *database.DB) *VoteDB {
	return &VoteDB{
		db: db,
	}
}

func (v *VoteDB) Save(ctx context.Context, m *model.Vote) error {
	if errors := m.Validate(); len(errors) > 0 {
		return fmt.Errorf("VoteDB invalid: %v", strings.Join(errors, ", "))
	}

	m.Touch()
	return v.db.InTx(ctx, pgx.ReadCommitted, func(tx pgx.Tx) error {
		if m.IsNew() {
			err := tx.QueryRow(
				ctx, createVoteSQL, m.UserID, m.KindID, m.Kind, m.Mode, m.CreatedAt,
			).Scan(&m.ID)
			if err != nil {
				return fmt.Errorf("inserting answer: %w", err)
			}
			return nil
		}
		_, err := tx.Exec(
			ctx, updateVoteSQL, m.Mode, m.UpdatedAt, m.ID,
		)
		if err != nil {
			return fmt.Errorf("failed to update vote: %w", err)
		}
		return nil
	})
}

func (v *VoteDB) ByUserAndKind(
	ctx context.Context,
	userID int64,
	kindID int64,
	kind string,
) (*model.Vote, error) {
	var vote *model.Vote

	if err := v.db.InTx(ctx, pgx.ReadCommitted, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, selectVoteByUserAndKindSQL, userID, kindID, kind)

		var err error
		vote, err = v.scanOne(row)
		if err != nil {
			return fmt.Errorf("failed to parse: %w", err)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("get vote: %w", err)
	}

	return vote, nil
}

func (v *VoteDB) CountVotes(
	ctx context.Context,
	kindID int64,
	kind string,
	mode string,
) (*int, error) {
	var count int
	if err := v.db.InTx(ctx, pgx.ReadCommitted, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx, countVotesSQL, kindID, kind, mode).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to count votes: %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("count votes: %w", err)
	}
	return &count, nil
}

func (v *VoteDB) scanOne(row pgx.Row) (*model.Vote, error) {
	vote := model.NewVote()

	if err := row.Scan(
		&vote.ID, &vote.UserID, &vote.KindID, &vote.Kind, &vote.Mode, &vote.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return vote, nil
}
