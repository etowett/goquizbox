// Package model is a model abstraction of authorized apps.
package model

import (
	"regexp"
	"strings"

	null "gopkg.in/guregu/null.v4"
)

var (
	emailRe              = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,10}$`)
	invalidUsernameRe    = regexp.MustCompile("[^A-Za-z0-9]")
	minPasswordLength    = 8
	invalidPasswordRegex = regexp.MustCompile("[[:^graph:]]")
)

// User represents the configuration for a single user
type User struct {
	SequentialIdentifier
	Username                string      `json:"username"`
	FirstName               string      `json:"first_name"`
	LastName                string      `json:"last_name"`
	Email                   string      `json:"email"`
	EmailVerified           bool        `json:"email_verified"`
	Password                string      `json:"-"`
	PasswordConfirmation    string      `json:"-"`
	PasswordHash            string      `json:"-"`
	Phone                   string      `json:"phone"`
	PhoneVerified           bool        `json:"phone_verified"`
	Status                  UserStatus  `json:"status"`
	PasswordResetKey        null.String `json:"-"`
	PasswordResetKeyExpiry  null.Time   `json:"-"`
	PhoneActivationKey      null.String `json:"-"`
	PhoneKeyExpiresAt       null.Time   `json:"-"`
	PhoneActivationAttempts null.Int    `json:"phone_activation_attempts"`
	EmailActivationKey      null.String `json:"-"`
	EmailKeyExpiresAt       null.Time   `json:"-"`
	EmailActivationAttempts null.Int    `json:"email_activation_attempts"`
	LastRefreshedAt         null.Time   `json:"last_refreshed_at"`
	PreferredTwoFactorAuth  string      `json:"preferred_2fa"`
	Timestamps
}

// NewUser initializes a User structure including
// pre-allocating all included maps.
func NewUser() *User {
	return &User{}
}

// Validate checks a user before a save operation.
func (c *User) Validate() []string {
	errors := make([]string, 0)
	if c.FirstName == "" || c.LastName == "" {
		errors = append(errors, "Name cannot be empty")
	}

	if len(c.Email) < 1 || len(c.Email) > 30 {
		errors = append(errors, "Email cannot be too short or too long")
	}

	if !emailRe.MatchString(c.Email) {
		errors = append(errors, "invalid email provided")
	}

	username := strings.TrimSpace(c.Username)

	if len(username) < 4 || len(username) > 12 {
		errors = append(errors, "invalid username length")
	}

	usernameLoc := invalidUsernameRe.FindStringIndex(username)
	if usernameLoc != nil {
		errors = append(errors, "invalid username provided")
	}

	if len(strings.TrimSpace(c.Password)) < minPasswordLength {
		errors = append(errors, "invalid password length")
	}

	passwordLoc := invalidPasswordRegex.FindStringIndex(c.Password)
	if passwordLoc != nil {
		errors = append(errors, "invalid password provided")
	}

	if c.Password != c.PasswordConfirmation {
		errors = append(errors, "Password and Password confirm must be the same")
	}

	return errors
}
