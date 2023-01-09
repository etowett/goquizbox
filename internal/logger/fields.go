package logger

import "go.uber.org/zap"

type (
	Field = zap.Field
)

var (
	Duration = zap.Duration
	Err      = zap.Error
	Float64  = zap.Float64
	Int      = zap.Int
	Int64    = zap.Int64
	String   = zap.String
	Time     = zap.Time
)
