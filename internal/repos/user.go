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
	createUserSQL     = `insert into users (first_name, last_name, email, email_verified, status, password_hash, created_at) values ($1, $2, $3, $4, $5, $6, $7) returning id`
	updateUserSQL     = `update users set first_name=$1, last_name=$2, email=$3, email_activation_key=$4, status=$5, updated_at=$6 where id = $7`
	getUsersSQL       = `select id, first_name, last_name, email, email_activation_key, email_verified, status, password_hash, created_at, updated_at from users`
	getUserByIDSQL    = getUsersSQL + ` where id=$1`
	getUserByEmailSQL = getUsersSQL + ` where lower(email)=lower($1)`
	getUserByPhoneSQL = getUsersSQL + ` where phone=$1`
	countUsersSQL     = "select count(id) from users"
	deleteUserSQL     = `delete from users where id=$1`
)

type UserDB struct {
	db *database.DB
}

func NewUserDB(db *database.DB) *UserDB {
	return &UserDB{
		db: db,
	}
}

func (r *UserDB) List(ctx context.Context, filter *webutils.Filter) ([]*entities.User, error) {
	users := make([]*entities.User, 0)

	if err := r.db.InTx(ctx, pgx.ReadCommitted, func(tx pgx.Tx) error {
		query, args := r.buildQuery(
			getUsersSQL,
			filter,
		)

		rows, err := tx.Query(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("failed to list users: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			if err := rows.Err(); err != nil {
				return fmt.Errorf("failed to iterate: %w", err)
			}

			user, err := r.scan(rows)
			if err != nil {
				return fmt.Errorf("failed to parse: %w", err)
			}
			users = append(users, user)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	return users, nil
}

func (r *UserDB) Count(ctx context.Context, filter *webutils.Filter) (*int, error) {
	var count int
	if err := r.db.InTx(ctx, pgx.ReadCommitted, func(tx pgx.Tx) error {
		query, args := r.buildQuery(
			countUsersSQL,
			&webutils.Filter{
				Term: filter.Term,
			},
		)
		err := tx.QueryRow(ctx, query, args...).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to count users: %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("count users: %w", err)
	}
	return &count, nil
}

func (r *UserDB) Save(ctx context.Context, m *entities.User) error {
	if errors := m.Validate(); len(errors) > 0 {
		return fmt.Errorf("UserDB invalid: %v", strings.Join(errors, ", "))
	}

	m.Touch()
	return r.db.InTx(ctx, pgx.ReadCommitted, func(tx pgx.Tx) error {
		if m.IsNew() {
			err := tx.QueryRow(
				ctx, createUserSQL, m.FirstName, m.LastName, m.Email,
				m.EmailVerified, m.Status, m.PasswordHash, m.CreatedAt,
			).Scan(&m.ID)
			if err != nil {
				return fmt.Errorf("inserting user: %w", err)
			}
			return nil
		}
		_, err := tx.Exec(
			ctx, updateUserSQL, m.FirstName, m.LastName, m.Email, m.EmailActivationKey,
			m.Status, m.UpdatedAt, m.ID,
		)
		if err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}
		return nil
	})
}

func (r *UserDB) ByEmail(ctx context.Context, email string) (*entities.User, error) {
	user := entities.NewUser()

	if err := r.db.InTx(ctx, pgx.ReadCommitted, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, getUserByEmailSQL, email)

		var err error
		user, err = r.scan(row)
		if err != nil {
			return fmt.Errorf("failed to parse: %w", err)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return user, nil
}

func (r *UserDB) GetByID(ctx context.Context, id int64) (*entities.User, error) {
	user := entities.NewUser()

	if err := r.db.InTx(ctx, pgx.ReadCommitted, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, getUserByIDSQL, id)

		var err error
		user, err = r.scan(row)
		if err != nil {
			return fmt.Errorf("failed to parse: %w", err)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}

func (u *UserDB) Delete(ctx context.Context, id int64) error {
	if err := u.db.InTx(ctx, pgx.ReadCommitted, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, deleteUserSQL, id)
		return err
	}); err != nil {
		return fmt.Errorf("delete user by id: %w", err)
	}
	return nil
}

func (r *UserDB) buildQuery(
	query string,
	filter *webutils.Filter,
) (string, []interface{}) {

	conditions := make([]string, 0)
	args := make([]interface{}, 0)
	counter := util.NewPlaceholder()

	if filter.Term != "" {
		filterColumns := []string{"first_name", "last_name", "email"}
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

func (r *UserDB) scan(row pgx.Row) (*entities.User, error) {
	user := entities.NewUser()

	if err := row.Scan(
		&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.EmailActivationKey,
		&user.EmailVerified, &user.Status, &user.PasswordHash,
		&user.Timestamps.CreatedAt, &user.Timestamps.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}
