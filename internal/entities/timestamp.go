package entities

import (
	"time"

	null "gopkg.in/guregu/null.v4"
)

type Timestamps struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt null.Time `json:"updated_at"`
}

func (t *Timestamps) Touch() {

	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
}
