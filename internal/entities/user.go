package entities

import (
	"regexp"
	"strings"

	null "gopkg.in/guregu/null.v4"
)

var (
	emailRe              = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,10}$`)
	minPasswordLength    = 8
	invalidPasswordRegex = regexp.MustCompile("[[:^graph:]]")
)

type User struct {
	SequentialIdentifier
	FirstName          string      `json:"first_name"`
	LastName           string      `json:"last_name"`
	Email              string      `json:"email"`
	EmailActivationKey null.String `json:"-"`
	EmailVerified      bool        `json:"email_verified"`
	Password           string      `json:"-"`
	PasswordConfirm    string      `json:"-"`
	PasswordHash       string      `json:"-"`
	Status             UserStatus  `json:"status"`
	Timestamps
}

func NewUser() *User {
	return &User{}
}

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

	if len(strings.TrimSpace(c.Password)) < minPasswordLength {
		errors = append(errors, "invalid password length")
	}

	passwordLoc := invalidPasswordRegex.FindStringIndex(c.Password)
	if passwordLoc != nil {
		errors = append(errors, "invalid password provided")
	}

	if c.Password != c.PasswordConfirm {
		errors = append(errors, "Password and Password confirm must be the same")
	}

	return errors
}
