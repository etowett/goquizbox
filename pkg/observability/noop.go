// Package observability sets up and configures observability tools.
package observability

import "context"

// Compile-time check to verify implements interface.
var _ Exporter = (*noopExporter)(nil)

// noopExporter is an observability exporter that does nothing.
type noopExporter struct{}

func NewNoop(_ context.Context) (Exporter, error) {
	return &noopExporter{}, nil
}

func (g *noopExporter) StartExporter() error {
	return nil
}

func (g *noopExporter) Close() error {
	return nil
}
