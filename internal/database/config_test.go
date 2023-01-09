package database

import (
	"testing"

	"goquizbox/internal/project"

	"github.com/sethvargo/go-envconfig"
)

func TestConfig_DatabaseConfig(t *testing.T) {
	t.Parallel()

	cfg1 := &Config{}
	cfg2 := cfg1.DatabaseConfig()

	if cfg1 != cfg2 {
		t.Errorf("expected %#v to be %#v", cfg1, cfg2)
	}
}

func TestConfig_ConnectionURL(t *testing.T) {
	t.Parallel()

	ctx := project.TestContext(t)

	cases := []struct {
		name   string
		config *Config
		want   string
	}{
		{
			name:   "nil",
			config: nil,
			want:   "",
		},
		{
			name: "host",
			config: &Config{
				Host: "myhost",
			},
			want: "postgres://myhost:5432?sslmode=require",
		},
		{
			name: "host_port",
			config: &Config{
				Host: "myhost",
				Port: "1234",
			},
			want: "postgres://myhost:1234?sslmode=require",
		},
		{
			name: "basic_auth",
			config: &Config{
				User:     "myuser",
				Password: "mypass",
			},
			want: "postgres://myuser:mypass@localhost:5432?sslmode=require",
		},
		{
			name: "connection_timeout",
			config: &Config{
				ConnectionTimeout: 60,
			},
			want: "postgres://localhost:5432?connect_timeout=60&sslmode=require",
		},
		{
			name: "sslmode",
			config: &Config{
				SSLMode: "panda",
			},
			want: "postgres://localhost:5432?sslmode=panda",
		},
		{
			name: "sslcert",
			config: &Config{
				SSLCertPath: "sslcertpath",
			},
			want: "postgres://localhost:5432?sslcert=sslcertpath&sslmode=require",
		},
		{
			name: "sslkey",
			config: &Config{
				SSLKeyPath: "sslkeypath",
			},
			want: "postgres://localhost:5432?sslkey=sslkeypath&sslmode=require",
		},
		{
			name: "sslrootcert",
			config: &Config{
				SSLRootCertPath: "sslrootcertpath",
			},
			want: "postgres://localhost:5432?sslmode=require&sslrootcert=sslrootcertpath",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := tc.config
			if cfg != nil {
				if err := envconfig.ProcessWith(ctx, cfg, envconfig.MapLookuper(nil)); err != nil {
					t.Fatal(err)
				}
			}

			if got, want := cfg.ConnectionURL(), tc.want; got != want {
				t.Errorf("expected %q to be %q", got, want)
			}
		})
	}
}
