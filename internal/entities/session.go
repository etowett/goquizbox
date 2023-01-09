package entities

import (
	"time"

	null "gopkg.in/guregu/null.v4"
)

type (
	Session struct {
		SequentialIdentifier
		DeactivatedAt   null.Time `json:"deactivated_at"`
		ExpiresAt       null.Time `json:"expires_at"`
		IPAddress       string    `json:"ip_address"`
		LastRefreshedAt time.Time `json:"last_refreshed_at"`
		UserAgent       string    `json:"user_agent"`
		UserID          int64     `json:"user_id"`
		Timestamps
	}

	FullSession struct {
		SequentialIdentifier
		ID              int64      `json:"id"`
		DeactivatedAt   null.Time  `json:"deactivated_at"`
		IPAddress       string     `json:"ip_address"`
		LastRefreshedAt time.Time  `json:"last_refreshed_at"`
		UserAgent       string     `json:"user_agent"`
		UserID          int64      `json:"user_id"`
		UserStatus      UserStatus `json:"user_status"`
		Timestamps
	}
)

func NewSession() *Session {
	return &Session{}
}

func (c *Session) Validate() []string {
	errors := make([]string, 0)
	if c.UserID < 1 {
		errors = append(errors, "Email cannot be too short or too long")
	}
	return errors
}
