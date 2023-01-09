package repos

import (
	"context"
	"errors"
	"fmt"
	"goquizbox/internal/database"
	"goquizbox/internal/entities"
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

func (r *VoteDB) Save(ctx context.Context, m *entities.Vote) error {
	if errors := m.Validate(); len(errors) > 0 {
		return fmt.Errorf("VoteDB invalid: %v", strings.Join(errors, ", "))
	}

	m.Touch()
	return r.db.InTx(ctx, pgx.ReadCommitted, func(tx pgx.Tx) error {
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

func (r *VoteDB) ByUserAndKind(
	ctx context.Context,
	userID int64,
	kindID int64,
	kind string,
) (*entities.Vote, error) {
	vote := entities.NewVote()

	if err := r.db.InTx(ctx, pgx.ReadCommitted, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, selectVoteByUserAndKindSQL, userID, kindID, kind)

		var err error
		vote, err = r.scan(row)
		if err != nil {
			return fmt.Errorf("failed to parse: %w", err)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("get vote: %w", err)
	}

	return vote, nil
}

func (r *VoteDB) CountVotes(
	ctx context.Context,
	kindID int64,
	kind string,
	mode string,
) (*int, error) {
	var count int
	if err := r.db.InTx(ctx, pgx.ReadCommitted, func(tx pgx.Tx) error {
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

func (*VoteDB) scan(row pgx.Row) (*entities.Vote, error) {
	vote := entities.NewVote()

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
