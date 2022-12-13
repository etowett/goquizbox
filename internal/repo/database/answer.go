package database

import (
	"context"
	"errors"
	"fmt"
	"goquizbox/internal/repo/model"
	"goquizbox/internal/util"
	"goquizbox/internal/web/webutils"
	"goquizbox/pkg/database"
	"strings"

	pgx "github.com/jackc/pgx/v4"
)

const (
	createAnswerSQL = `insert into answers (user_id, question_id, body, created_at) values ($1, $2, $3, $4) returning id`
	selectAnswerSQL = `select id, user_id, question_id, body, created_at, updated_at from answers`
	countAnswerSQL  = `select count(id) from answers`
	updateAnswerSQL = `update answers set (body, updated_at) = ($1, $2) where id=$3`
)

type AnswerDB struct {
	db *database.DB
}

func NewAnswerDB(db *database.DB) *AnswerDB {
	return &AnswerDB{
		db: db,
	}
}

func (a *AnswerDB) Save(ctx context.Context, m *model.Answer) error {
	if errors := m.Validate(); len(errors) > 0 {
		return fmt.Errorf("AnswerDB invalid: %v", strings.Join(errors, ", "))
	}

	m.Touch()
	return a.db.InTx(ctx, pgx.ReadCommitted, func(tx pgx.Tx) error {
		if m.IsNew() {
			err := tx.QueryRow(
				ctx, createAnswerSQL, m.UserID, m.QuestionID, m.Body, m.CreatedAt,
			).Scan(&m.ID)
			if err != nil {
				return fmt.Errorf("inserting answer: %w", err)
			}
			return nil
		}

		_, err := tx.Exec(ctx, updateAnswerSQL, m.Body, m.UpdatedAt, m.ID)
		if err != nil {
			return fmt.Errorf("failed to update: %w", err)
		}
		return nil
	})
}

func (r *AnswerDB) ByQuestion(
	ctx context.Context,
	questionID int64,
	filter *webutils.Filter,
) ([]*model.Answer, error) {
	var answers []*model.Answer

	if err := r.db.InTx(ctx, pgx.ReadCommitted, func(tx pgx.Tx) error {
		query, args := r.buildQuery(
			selectAnswerSQL,
			questionID,
			filter,
		)

		rows, err := tx.Query(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("failed to list answers: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			if err := rows.Err(); err != nil {
				return fmt.Errorf("failed to iterate: %w", err)
			}

			answer, err := r.scanOne(rows)
			if err != nil {
				return fmt.Errorf("failed to parse: %w", err)
			}
			answers = append(answers, answer)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("list recipients: %w", err)
	}

	return answers, nil
}

func (a *AnswerDB) CountByQuestion(
	ctx context.Context,
	questionID int64,
	filter *webutils.Filter,
) (*int, error) {
	var count int
	if err := a.db.InTx(ctx, pgx.ReadCommitted, func(tx pgx.Tx) error {
		query, args := a.buildQuery(
			countAnswerSQL,
			questionID,
			&webutils.Filter{
				Term: filter.Term,
			},
		)
		err := tx.QueryRow(ctx, query, args...).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to count answers: %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("count answers: %w", err)
	}
	return &count, nil
}

func (r *AnswerDB) buildQuery(
	query string,
	messageID int64,
	filter *webutils.Filter,
) (string, []interface{}) {
	conditions := make([]string, 0)
	args := make([]interface{}, 0)
	placeholder := util.NewPlaceholder()

	conditions = append(conditions, fmt.Sprintf(" question_id=$%d", placeholder.Touch()))
	args = append(args, messageID)

	if filter.Term != "" {
		likeStmt := make([]string, 0)
		columns := []string{"body"}
		for _, col := range columns {
			search := fmt.Sprintf(" (lower(%s) like '%%' || $%d || '%%')", col, placeholder.Touch())
			likeStmt = append(likeStmt, search)
			args = append(args, filter.Term)
		}
		conditions = append(conditions, fmt.Sprintf("(%s)", strings.Join(likeStmt, " or")))
	}

	if len(conditions) > 0 {
		query += " where" + strings.Join(conditions, " and")
	}

	if filter.Per > 0 && filter.Page > 0 {
		query += fmt.Sprintf(" order by id desc limit $%d offset $%d", placeholder.Touch(), placeholder.Touch())
		args = append(args, filter.Per, (filter.Page-1)*filter.Per)
	}

	return query, args
}

func (r *AnswerDB) scanOne(row pgx.Row) (*model.Answer, error) {
	answer := model.NewAnswer()

	if err := row.Scan(
		&answer.ID, &answer.UserID, &answer.QuestionID, &answer.Body, &answer.CreatedAt, &answer.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return answer, nil
}
