// Package database is a database interface to authorized apps.
package database

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"goquizbox/internal/repo/model"
	"goquizbox/pkg/database"

	pgx "github.com/jackc/pgx/v4"
)

const (
	createSessionSQL      = `insert into sessions (user_id, deactivated_at, expires_at, ip_address, last_refreshed_at, user_agent, created_at) values ($1, $2, $3, $4, $5, $6, $7) returning id`
	selectSessionsSQL     = `select id, user_id, deactivated_at, expires_at, ip_address, last_refreshed_at, user_agent, created_at, updated_at from sessions`
	getSessionByIDSQL     = selectSessionsSQL + " where id=$1"
	getFullSessionByIDSQL = `select s.id, s.deactivated_at, s.ip_address, s.last_refreshed_at,
		s.user_agent, s.user_id, u.status AS user_status, s.created_at, s.updated_at from sessions s join users u ON s.user_id = u.id where s.id = $1`
	updateSessionSQL = `UPDATE sessions SET (deactivated_at, ip_address,
		last_refreshed_at, user_agent, user_id, updated_at) =
		($1, $2, $3, $4, $5, $6) WHERE id = $7`
)

type SessionDB struct {
	db *database.DB
}

func NewSessionDB(db *database.DB) *SessionDB {
	return &SessionDB{
		db: db,
	}
}

func (u *SessionDB) Save(ctx context.Context, m *model.Session) error {
	if errors := m.Validate(); len(errors) > 0 {
		return fmt.Errorf("SessionDB invalid: %v", strings.Join(errors, ", "))
	}

	m.Touch()

	return u.db.InTx(ctx, pgx.ReadCommitted, func(tx pgx.Tx) error {
		if m.IsNew() {
			err := tx.QueryRow(
				ctx, createSessionSQL, m.UserID, m.DeactivatedAt, m.ExpiresAt, m.IPAddress,
				m.LastRefreshedAt, m.UserAgent, m.CreatedAt,
			).Scan(&m.ID)
			if err != nil {
				return fmt.Errorf("inserting session: %w", err)
			}
			return nil
		}

		_, err := tx.Exec(
			ctx, updateSessionSQL, m.DeactivatedAt, m.IPAddress, m.LastRefreshedAt,
			m.UserAgent, m.UserID, m.UpdatedAt, m.ID,
		)
		if err != nil {
			return fmt.Errorf("failed to update session: %w", err)
		}
		return nil
	})
}

func (u *SessionDB) GetSessionByID(ctx context.Context, id int64) (*model.Session, error) {
	var user *model.Session

	if err := u.db.InTx(ctx, pgx.ReadCommitted, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, getSessionByIDSQL, id)

		var err error
		user, err = scanOneSession(row)
		if err != nil {
			return fmt.Errorf("failed to parse: %w", err)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("get session by id: %w", err)
	}

	return user, nil
}

func (u *SessionDB) GetFullSessionByID(ctx context.Context, id int64) (*model.FullSession, error) {
	var session model.FullSession

	if err := u.db.InTx(ctx, pgx.ReadCommitted, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, getFullSessionByIDSQL, id)
		err := row.Scan(
			&session.ID, &session.DeactivatedAt, &session.IPAddress, &session.LastRefreshedAt,
			&session.UserAgent, &session.UserID, &session.UserStatus,
			&session.Timestamps.CreatedAt, &session.Timestamps.UpdatedAt,
		)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
			return err
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("get full session by id: %w", err)
	}

	return &session, nil
}

func scanOneSession(row pgx.Row) (*model.Session, error) {
	session := model.NewSession()

	if err := row.Scan(
		&session.ID, &session.UserID, &session.DeactivatedAt, &session.ExpiresAt, &session.IPAddress,
		&session.LastRefreshedAt, &session.UserAgent, &session.CreatedAt, &session.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return session, nil
}
