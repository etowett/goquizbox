// Package setup provides common logic for configuring the various services.
package setup

import (
	"context"
	"fmt"

	"goquizbox/internal/database"
	"goquizbox/internal/logger"
	"goquizbox/internal/serverenv"

	envconfig "github.com/sethvargo/go-envconfig"
)

// DatabaseConfigProvider ensures that the environment config can provide a DB config.
// All binaries in this application connect to the database via the same method.
type DatabaseConfigProvider interface {
	DatabaseConfig() *database.Config
}

// Setup runs common initialization code for all servers. See SetupWith.
func Setup(ctx context.Context, config interface{}) (*serverenv.ServerEnv, error) {
	return SetupWith(ctx, config, envconfig.OsLookuper())
}

// SetupWith processes the given configuration using envconfig. It is
// responsible for establishing database connections, resolving secrets, and
// accessing app configs. The provided interface must implement the various
// interfaces.
func SetupWith(
	ctx context.Context,
	config interface{},
	l envconfig.Lookuper,
) (*serverenv.ServerEnv, error) { //nolint:golint

	// Build a list of mutators. This list will grow as we initialize more of the
	// configuration, such as the secret manager.
	var mutatorFuncs []envconfig.MutatorFunc

	// Build a list of options to pass to the server env.
	var serverEnvOpts []serverenv.Option

	// Process first round of environment variables.
	if err := envconfig.ProcessWith(ctx, config, l, mutatorFuncs...); err != nil {
		return nil, fmt.Errorf("error loading environment variables: %w", err)
	}
	logger.Infof("redis config: %v", config)

	// Setup the database connection.
	if provider, ok := config.(DatabaseConfigProvider); ok {
		logger.Info("configuring database")

		dbConfig := provider.DatabaseConfig()
		db, err := database.NewFromEnv(ctx, dbConfig)
		if err != nil {
			return nil, fmt.Errorf("unable to connect to database: %w", err)
		}

		// Update serverEnv setup.
		serverEnvOpts = append(serverEnvOpts, serverenv.WithDatabase(db))

		logger.Infof("database config: %v", dbConfig)
	}

	return serverenv.New(ctx, serverEnvOpts...), nil
}
