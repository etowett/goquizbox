// Package model is a model abstraction of authorized apps.
package model

// Login represents the configuration for a single login request
type Login struct {
	Username string
	Password string
	Remember bool
}

// NewLogin initializes a Login structure including
// pre-allocating all included maps.
func NewLogin() *Login {
	return &Login{}
}

// Validate checks a login before a save operation.
func (c *Login) Validate() []string {
	errors := make([]string, 0)
	if c.Username == "" {
		errors = append(errors, "Username cannot be empty")
	}
	if len(c.Password) == 0 {
		errors = append(errors, "Password cannot be empty")
	}
	return errors
}
