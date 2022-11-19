// Package database is a database interface to authorized apps.
package database

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"goquizbox/internal/repo/model"
	"goquizbox/internal/util"
	"goquizbox/internal/web/webutils"
	"goquizbox/pkg/database"

	pgx "github.com/jackc/pgx/v4"
)

const (
	createUserSQL        = `insert into users (username, first_name, last_name, email, email_verified, status, password_hash, created_at) values ($1, $2, $3, $4, $5, $6, $7, $8) returning id`
	updateUserSQL        = `update users set username=$1, first_name=$2, last_name=$3, email=$4, email_activation_key=$5, status=$6, updated_at=$7 where id = $8`
	getUsersSQL          = `select id, username, first_name, last_name, email, email_activation_key, email_verified, status, password_hash, created_at, updated_at from users`
	getUserByIDSQL       = getUsersSQL + ` where id=$1`
	getUserByUsernameSQL = getUsersSQL + ` where lower(username)=lower($1)`
	getUserByEmailSQL    = getUsersSQL + ` where lower(email)=lower($1)`
	getUserByPhoneSQL    = getUsersSQL + ` where phone=$1`
	countUsersSQL        = "select count(id) from users"
	deleteUserSQL        = `delete from users where id=$1`
)

type UserDB struct {
	db *database.DB
}

func NewUserDB(db *database.DB) *UserDB {
	return &UserDB{
		db: db,
	}
}

func (a *UserDB) List(ctx context.Context, filter *webutils.Filter) ([]*model.User, error) {
	var users []*model.User

	if err := a.db.InTx(ctx, pgx.ReadCommitted, func(tx pgx.Tx) error {
		query, args := a.buildQuery(
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

			user, err := scanOneUser(rows)
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

func (a *UserDB) Count(ctx context.Context, filter *webutils.Filter) (*int, error) {
	var count int
	if err := a.db.InTx(ctx, pgx.ReadCommitted, func(tx pgx.Tx) error {
		query, args := a.buildQuery(
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

func (u *UserDB) Save(ctx context.Context, m *model.User) error {
	if errors := m.Validate(); len(errors) > 0 {
		return fmt.Errorf("UserDB invalid: %v", strings.Join(errors, ", "))
	}

	m.Touch()

	return u.db.InTx(ctx, pgx.ReadCommitted, func(tx pgx.Tx) error {
		if m.IsNew() {
			err := tx.QueryRow(
				ctx, createUserSQL, m.Username, m.FirstName, m.LastName, m.Email,
				m.EmailVerified, m.Status, m.PasswordHash, m.CreatedAt,
			).Scan(&m.ID)
			if err != nil {
				return fmt.Errorf("inserting user: %w", err)
			}
			return nil
		}
		_, err := tx.Exec(
			ctx, updateUserSQL, m.Username, m.FirstName, m.LastName, m.Email, m.EmailActivationKey,
			m.Status, m.UpdatedAt, m.ID,
		)
		if err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}
		return nil
	})
}

func (u *UserDB) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user *model.User

	if err := u.db.InTx(ctx, pgx.ReadCommitted, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, getUserByUsernameSQL, username)

		var err error
		user, err = scanOneUser(row)
		if err != nil {
			return fmt.Errorf("failed to parse: %w", err)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("get user by username: %w", err)
	}

	return user, nil
}

func (u *UserDB) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user *model.User

	if err := u.db.InTx(ctx, pgx.ReadCommitted, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, getUserByEmailSQL, email)

		var err error
		user, err = scanOneUser(row)
		if err != nil {
			return fmt.Errorf("failed to parse: %w", err)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return user, nil
}

func (u *UserDB) GetByID(ctx context.Context, id int64) (*model.User, error) {
	var user *model.User

	if err := u.db.InTx(ctx, pgx.ReadCommitted, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, getUserByIDSQL, id)

		var err error
		user, err = scanOneUser(row)
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

func (u *UserDB) buildQuery(
	query string,
	filter *webutils.Filter,
) (string, []interface{}) {

	conditions := make([]string, 0)
	args := make([]interface{}, 0)
	counter := util.NewPlaceholder()

	if filter.Term != "" {
		filterColumns := []string{"first_name", "last_name", "email", "username", "phone"}
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

func scanOneUser(row pgx.Row) (*model.User, error) {
	user := model.NewUser()

	if err := row.Scan(
		&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.Email,
		&user.EmailActivationKey, &user.EmailVerified, &user.Status, &user.PasswordHash,
		&user.Timestamps.CreatedAt, &user.Timestamps.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}
