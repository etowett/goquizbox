package app

import (
	"goquizbox/internal/database"
	"goquizbox/internal/setup"
)

var (
	_ setup.DatabaseConfigProvider = (*Config)(nil)
)

type Config struct {
	Database    database.Config
	Environment string `env:"ENV, default=local"`
	Port        string `env:"PORT, default=8090"`
}

func (c *Config) DatabaseConfig() *database.Config {
	return &c.Database
}
