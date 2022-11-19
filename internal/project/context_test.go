package project

import (
	"testing"

	"smsim/pkg/logging"
)

func TestTestContext(t *testing.T) {
	t.Parallel()

	ctx := TestContext(t)
	logger := logging.FromContext(ctx)

	if logger == nil {
		t.Error("expected a test logger")
	}
}
