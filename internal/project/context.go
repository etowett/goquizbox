package project

import (
	"context"
	"testing"

	"goquizbox/pkg/logging"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// TestContext returns a context with test values pre-populated.
func TestContext(tb testing.TB) context.Context {
	ctx := context.Background()
	ctx = logging.WithLogger(ctx, TestLogger(tb))
	return ctx
}

// TestLogger returns a logger configured for test. See [zaptest] for more
// information.
//
// [zaptest]: https://pkg.go.dev/go.uber.org/zap/zaptest
func TestLogger(tb testing.TB) *zap.SugaredLogger {
	return zaptest.NewLogger(tb, zaptest.Level(zap.WarnLevel)).Sugar()
}
