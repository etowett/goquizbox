package repos

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"goquizbox/internal/database"
	"goquizbox/internal/entities"
	"goquizbox/internal/util"
	"goquizbox/internal/web/webutils"

	pgx "github.com/jackc/pgx/v4"
)

const (
	createQuestionSQL  = `insert into questions (user_id, title, body, tags, created_at) values ($1, $2, $3, $4, $5) returning id`
	updateQuestionSQL  = `update questions set title=$1, body=$2, tags=$3, updated_at=$4 where id = $5`
	getQuestionsSQL    = `select id, user_id, title, body, tags, created_at, updated_at from questions`
	getQuestionByIDSQL = getQuestionsSQL + ` where id=$1`
	countCuestionsSQL  = "select count(id) from questions"
	deleteQuestionSQL  = `delete from questions where id=$1`
)

type QuestionDB struct {
	db *database.DB
}

func NewQuestionDB(db *database.DB) *QuestionDB {
	return &QuestionDB{
		db: db,
	}
}

func (r *QuestionDB) List(ctx context.Context, filter *webutils.Filter) ([]*entities.Question, error) {
	questions := make([]*entities.Question, 0)

	if err := r.db.InTx(ctx, pgx.ReadCommitted, func(tx pgx.Tx) error {
		query, args := r.buildQuery(
			getQuestionsSQL,
			filter,
		)

		rows, err := tx.Query(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("failed to list questions: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			if err := rows.Err(); err != nil {
				return fmt.Errorf("failed to iterate: %w", err)
			}

			question, err := r.scan(rows)
			if err != nil {
				return fmt.Errorf("failed to parse: %w", err)
			}
			questions = append(questions, question)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("list questions: %w", err)
	}

	return questions, nil
}

func (q *QuestionDB) Count(ctx context.Context, filter *webutils.Filter) (*int, error) {
	var count int
	if err := q.db.InTx(ctx, pgx.ReadCommitted, func(tx pgx.Tx) error {
		query, args := q.buildQuery(
			countCuestionsSQL,
			&webutils.Filter{
				Term: filter.Term,
			},
		)
		err := tx.QueryRow(ctx, query, args...).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to count questions: %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("count questions: %w", err)
	}
	return &count, nil
}

func (u *QuestionDB) Save(ctx context.Context, m *entities.Question) error {
	if errors := m.Validate(); len(errors) > 0 {
		return fmt.Errorf("QuestionDB invalid: %v", strings.Join(errors, ", "))
	}

	m.Touch()
	return u.db.InTx(ctx, pgx.ReadCommitted, func(tx pgx.Tx) error {
		if m.IsNew() {
			err := tx.QueryRow(
				ctx, createQuestionSQL, m.UserID, m.Title, m.Body, m.Tags, m.CreatedAt,
			).Scan(&m.ID)
			if err != nil {
				return fmt.Errorf("inserting question: %w", err)
			}
			return nil
		}
		_, err := tx.Exec(
			ctx, updateQuestionSQL, m.Title, m.Body, m.Tags, m.UpdatedAt, m.ID,
		)
		if err != nil {
			return fmt.Errorf("failed to update question: %w", err)
		}
		return nil
	})
}

func (r *QuestionDB) ByID(ctx context.Context, id int64) (*entities.Question, error) {
	question := entities.NewQuestion()

	if err := r.db.InTx(ctx, pgx.ReadCommitted, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, getQuestionByIDSQL, id)

		var err error
		question, err = r.scan(row)
		if err != nil {
			return fmt.Errorf("failed to parse: %w", err)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("get question by id: %w", err)
	}

	return question, nil
}

func (q *QuestionDB) Delete(ctx context.Context, id int64) error {
	if err := q.db.InTx(ctx, pgx.ReadCommitted, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, deleteQuestionSQL, id)
		return err
	}); err != nil {
		return fmt.Errorf("delete question by id: %w", err)
	}
	return nil
}

func (q *QuestionDB) buildQuery(
	query string,
	filter *webutils.Filter,
) (string, []interface{}) {

	conditions := make([]string, 0)
	args := make([]interface{}, 0)
	counter := util.NewPlaceholder()

	if filter.Term != "" {
		filterColumns := []string{"title", "body"}
		likeStatements := make([]string, 0)

		args = append(args, strings.ToLower(filter.Term))
		termPlaceholder := counter.Touch()
		for _, col := range filterColumns {
			stmt := fmt.Sprintf("lower(%s) LIKE '%%' || $%d || '%%'", col, termPlaceholder)
			likeStatements = append(likeStatements, stmt)
		}
		condition := fmt.Sprintf(" (%s)", strings.Join(likeStatements, " OR "))
		conditions = append(conditions, condition)
	}

	if filter.FromTime.Valid && filter.ToTime.Valid {
		condition := fmt.Sprintf(
			" (created_at >= $%d and created_at < $%d)",
			counter.Touch(),
			counter.Touch(),
		)
		conditions = append(conditions, condition)
		args = append(args, filter.FromTime.Time, filter.ToTime.Time)
	}

	if len(conditions) > 0 {
		query += " where" + strings.Join(conditions, " and ")
	}

	return query, args
}

func (*QuestionDB) scan(row pgx.Row) (*entities.Question, error) {
	question := entities.NewQuestion()

	if err := row.Scan(
		&question.ID, &question.UserID, &question.Title, &question.Body, &question.Tags,
		&question.Timestamps.CreatedAt, &question.Timestamps.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return question, nil
}
