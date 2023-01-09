package database

import (
	"testing"
	"time"

	"goquizbox/internal/project"

	"github.com/google/go-cmp/cmp"
)

func TestNewFromEnv(t *testing.T) {
	t.Parallel()

	ctx := project.TestContext(t)

	t.Run("bad_conn", func(t *testing.T) {
		t.Parallel()

		if _, err := NewFromEnv(ctx, &Config{}); err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestDBValues(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		config Config
		want   map[string]string
	}{
		{
			name: "empty configs",
			want: make(map[string]string),
		},
		{
			name: "some config",
			config: Config{
				Name:              "myDatabase",
				User:              "superuser",
				Password:          "notAG00DP@ssword",
				Port:              "1234",
				ConnectionTimeout: 5,
				PoolHealthCheck:   5 * time.Minute,
			},
			want: map[string]string{
				"dbname":                   "myDatabase",
				"password":                 "notAG00DP@ssword",
				"port":                     "1234",
				"user":                     "superuser",
				"connect_timeout":          "5",
				"pool_health_check_period": "5m0s",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := dbValues(&tc.config)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("mismatch (-want, +got):\n%s", diff)
			}
		})
	}
}
