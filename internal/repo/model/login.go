package model

type Login struct {
	Email    string
	Password string
	Remember bool
}

func NewLogin() *Login {
	return &Login{}
}

func (c *Login) Validate() []string {
	errors := make([]string, 0)
	if c.Email == "" {
		errors = append(errors, "Email cannot be empty")
	}
	if len(c.Password) == 0 {
		errors = append(errors, "Password cannot be empty")
	}
	return errors
}
