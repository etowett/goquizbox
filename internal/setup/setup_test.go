package setup_test

import (
	"os"
	"testing"

	"goquizbox/internal/project"
	"goquizbox/internal/setup"
	"goquizbox/pkg/database"
	"goquizbox/pkg/observability"

	envconfig "github.com/sethvargo/go-envconfig"
)

var testDatabaseInstance *database.TestInstance

func TestMain(m *testing.M) {
	testDatabaseInstance = database.MustTestInstance()
	defer testDatabaseInstance.MustClose()
	m.Run()
}

var (
	_ setup.DatabaseConfigProvider              = (*testConfig)(nil)
	_ setup.ObservabilityExporterConfigProvider = (*testConfig)(nil)
)

type testConfig struct {
	Database *database.Config
}

func (t *testConfig) DatabaseConfig() *database.Config {
	return t.Database
}

func (t *testConfig) ObservabilityExporterConfig() *observability.Config {
	return &observability.Config{
		ExporterType: observability.ExporterType("NOOP"),
	}
}

func TestSetupWith(t *testing.T) {
	t.Parallel()

	tmp, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	lookuper := envconfig.MapLookuper(map[string]string{})

	t.Run("default", func(t *testing.T) {
		t.Parallel()

		ctx := project.TestContext(t)
		_, dbconfig := testDatabaseInstance.NewDatabase(t)

		config := &testConfig{Database: dbconfig}
		env, err := setup.SetupWith(ctx, config, lookuper)
		if err != nil {
			t.Fatal(err)
		}
		defer env.Close(ctx)
	})

	t.Run("database", func(t *testing.T) {
		t.Parallel()

		ctx := project.TestContext(t)
		_, dbconfig := testDatabaseInstance.NewDatabase(t)

		config := &testConfig{Database: dbconfig}
		env, err := setup.SetupWith(ctx, config, lookuper)
		if err != nil {
			t.Fatal(err)
		}
		defer env.Close(ctx)

		db := env.Database()
		if db == nil {
			t.Errorf("expected db to exist")
		}
	})

	t.Run("observability_exporter", func(t *testing.T) {
		t.Parallel()

		ctx := project.TestContext(t)
		_, dbconfig := testDatabaseInstance.NewDatabase(t)

		config := &testConfig{Database: dbconfig}
		env, err := setup.SetupWith(ctx, config, lookuper)
		if err != nil {
			t.Fatal(err)
		}
		defer env.Close(ctx)

		oe := env.ObservabilityExporter()
		if oe == nil {
			t.Errorf("expected observability exporter to exist")
		}
		defer func() {
			if err := oe.Close(); err != nil {
				t.Fatal(err)
			}
		}()
	})
}
