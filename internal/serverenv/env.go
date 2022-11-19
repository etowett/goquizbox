// Package serverenv defines common parameters for the sever environment.
package serverenv

import (
	"context"

	"goquizbox/internal/metrics"
	"goquizbox/pkg/database"
	"goquizbox/pkg/observability"
)

// ExporterFunc defines a factory function for creating a context aware metrics exporter.
type ExporterFunc func(context.Context) metrics.Exporter

// ServerEnv represents latent environment configuration for servers in this application.
type ServerEnv struct {
	database              *database.DB
	exporter              metrics.ExporterFromContext
	observabilityExporter observability.Exporter
}

// Option defines function types to modify the ServerEnv on creation.
type Option func(*ServerEnv) *ServerEnv

// New creates a new ServerEnv with the requested options.
func New(ctx context.Context, opts ...Option) *ServerEnv {
	env := &ServerEnv{}
	// A metrics exporter is required, installs the default log based one.
	// Can be overridden by opts.
	env.exporter = func(ctx context.Context) metrics.Exporter {
		return metrics.NewLogsBasedFromContext(ctx)
	}

	for _, f := range opts {
		env = f(env)
	}

	return env
}

// WithDatabase attached a database to the environment.
func WithDatabase(db *database.DB) Option {
	return func(s *ServerEnv) *ServerEnv {
		s.database = db
		return s
	}
}

// WithMetricsExporter creates an Option to install a different metrics exporter.
func WithMetricsExporter(f metrics.ExporterFromContext) Option {
	return func(s *ServerEnv) *ServerEnv {
		s.exporter = f
		return s
	}
}

// WithObservabilityExporter creates an Option to install a specific observability exporter system.
func WithObservabilityExporter(oe observability.Exporter) Option {
	return func(s *ServerEnv) *ServerEnv {
		s.observabilityExporter = oe
		return s
	}
}

func (s *ServerEnv) Database() *database.DB {
	return s.database
}

func (s *ServerEnv) ObservabilityExporter() observability.Exporter {
	return s.observabilityExporter
}

// MetricsExporter returns a context appropriate metrics exporter.
func (s *ServerEnv) MetricsExporter(ctx context.Context) metrics.Exporter {
	if s.exporter == nil {
		return nil
	}
	return s.exporter(ctx)
}

// Close shuts down the server env, closing database connections, etc.
func (s *ServerEnv) Close(ctx context.Context) error {
	if s == nil {
		return nil
	}

	if s.database != nil {
		s.database.Close(ctx)
	}

	if s.observabilityExporter != nil {
		if err := s.observabilityExporter.Close(); err != nil {
			return nil
		}
	}

	return nil
}
