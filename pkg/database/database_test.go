package database

import (
	"testing"
	"time"
)

var testDatabaseInstance *TestInstance

func TestMain(m *testing.M) {
	testDatabaseInstance = MustTestInstance()
	defer testDatabaseInstance.MustClose()
	m.Run()
}

func TestNullableTime(t *testing.T) {
	t.Parallel()

	t.Run("zero", func(t *testing.T) {
		t.Parallel()

		if got, want := NullableTime(time.Time{}), (*time.Time)(nil); got != want {
			t.Errorf("expected %q to be %q", got, want)
		}
	})

	t.Run("not_nil", func(t *testing.T) {
		t.Parallel()

		now := time.Now().UTC()
		if got, want := NullableTime(now), &now; !got.Equal(now) {
			t.Errorf("expected %q to be %q", got, want)
		}
	})
}
