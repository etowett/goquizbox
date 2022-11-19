package entities

import (
	"time"
)

type TokenInfo struct {
	Exp       time.Time
	Refresh   time.Time
	SessionID int64
	Status    string
	UserID    int64
}

func (ti *TokenInfo) RequiresRefresh() bool {
	return time.Now().After(ti.Refresh)
}
