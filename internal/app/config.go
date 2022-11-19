package app

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"

	"goquizbox/internal/setup"
	"goquizbox/pkg/database"
)

var (
	//go:embed templates/*
	templatesFS embed.FS

	//go:embed assets/*
	assetsFolderFS embed.FS
	assetsFS, _    = fs.Sub(assetsFolderFS, "assets")
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

func (c *Config) TemplateRenderer() (*template.Template, error) {
	tmpl, err := template.New("").
		Option("missingkey=zero").
		Funcs(TemplateFuncMap).
		ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates from fs: %w", err)
	}
	return tmpl, nil
}
