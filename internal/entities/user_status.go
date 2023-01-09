package entities

import (
	"database/sql/driver"
)

type UserStatus string

const (
	UserStatusActive     UserStatus = "active"
	UserStatusInactive   UserStatus = "inactive"
	UserStatusUnverified UserStatus = "unverified"
)

// Scan implements the Scanner interface.
func (s *UserStatus) Scan(value interface{}) error {
	*s = UserStatus(string(value.(string)))
	return nil
}

// Value implements the driver Valuer interface.
func (s UserStatus) Value() (driver.Value, error) {
	return s.String(), nil
}

func (s UserStatus) String() string {
	return string(s)
}

func (s UserStatus) IsActive() bool {
	return s == UserStatusActive
}

func (s UserStatus) IsUnverified() bool {
	return s == UserStatusUnverified
}
