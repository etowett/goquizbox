// Package server provides an opinionated http server.
//
// Although exported, this package is non intended for general consumption. It
// is a shared dependency between multiple exposure notifications projects. We
// cannot guarantee that there won't be breaking changes in the future.
package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"goquizbox/pkg/logging"

	"github.com/hashicorp/go-multierror"

	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/tracecontext"
	"google.golang.org/grpc"
)

// Server provides a gracefully-stoppable http server implementation. It is safe
// for concurrent use in goroutines.
type Server struct {
	ip       string
	port     string
	listener net.Listener
}

// New creates a new server listening on the provided address that responds to
// the http.Handler. It starts the listener, but does not start the server. If
// an empty port is given, the server randomly chooses one.
func New(port string) (*Server, error) {
	// Create the net listener first, so the connection ready when we return. This
	// guarantees that it can accept requests.
	addr := fmt.Sprintf(":" + port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to create listener on %s: %w", addr, err)
	}

	return &Server{
		ip:       listener.Addr().(*net.TCPAddr).IP.String(),
		port:     strconv.Itoa(listener.Addr().(*net.TCPAddr).Port),
		listener: listener,
	}, nil
}

// NewFromListener creates a new server on the given listener. This is useful if
// you want to customize the listener type (e.g. udp or tcp) or bind network
// more than `New` allows.
func NewFromListener(listener net.Listener) (*Server, error) {
	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		return nil, fmt.Errorf("listener is not tcp")
	}

	return &Server{
		ip:       addr.IP.String(),
		port:     strconv.Itoa(addr.Port),
		listener: listener,
	}, nil
}

// ServeHTTP starts the server and blocks until the provided context is closed.
// When the provided context is closed, the server is gracefully stopped with a
// timeout of 5 seconds.
//
// Once a server has been stopped, it is NOT safe for reuse.
func (s *Server) ServeHTTP(ctx context.Context, srv *http.Server) error {
	logger := logging.FromContext(ctx)

	// Spawn a goroutine that listens for context closure. When the context is
	// closed, the server is stopped.
	errCh := make(chan error, 1)
	go func() {
		<-ctx.Done()

		logger.Debugf("server.Serve: context closed")
		shutdownCtx, done := context.WithTimeout(context.Background(), 5*time.Second)
		defer done()

		logger.Debugf("server.Serve: shutting down")
		errCh <- srv.Shutdown(shutdownCtx)
	}()

	// Create the prometheus metrics proxy.
	metricsDone, err := ServeMetricsIfPrometheus(ctx)
	if err != nil {
		return fmt.Errorf("failed to serve metrics: %w", err)
	}

	// Run the server. This will block until the provided context is closed.
	if err := srv.Serve(s.listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to serve: %w", err)
	}

	logger.Debugf("server.Serve: serving stopped")

	var merr *multierror.Error

	// Shutdown the prometheus metrics proxy.
	if metricsDone != nil {
		if err := metricsDone(); err != nil {
			merr = multierror.Append(merr, fmt.Errorf("failed to close metrics exporter: %w", err))
		}
	}

	// Return any errors that happened during shutdown.
	if err := <-errCh; err != nil {
		merr = multierror.Append(merr, fmt.Errorf("failed to shutdown server: %w", err))
	}
	return merr.ErrorOrNil()
}

// ServeHTTPHandler is a convenience wrapper around ServeHTTP. It creates an
// HTTP server using the provided handler, wrapped in OpenCensus for
// observability.
func (s *Server) ServeHTTPHandler(ctx context.Context, handler http.Handler) error {
	return s.ServeHTTP(ctx, &http.Server{
		ReadHeaderTimeout: 10 * time.Second,
		Handler: &ochttp.Handler{
			Handler:          handler,
			IsPublicEndpoint: true,
			Propagation:      &tracecontext.HTTPFormat{},
		},
	})
}

// ServeGRPC starts the server and blocks until the provided context is closed.
// When the provided context is closed, the server is gracefully stopped with a
// timeout of 5 seconds.
//
// Once a server has been stopped, it is NOT safe for reuse.
func (s *Server) ServeGRPC(ctx context.Context, srv *grpc.Server) error {
	logger := logging.FromContext(ctx)

	// Spawn a goroutine that listens for context closure. When the context is
	// closed, the server is stopped.
	errCh := make(chan error, 1)
	go func() {
		<-ctx.Done()

		logger.Debugf("server.Serve: context closed")
		logger.Debugf("server.Serve: shutting down")
		srv.GracefulStop()
	}()

	// Run the server. This will block until the provided context is closed.
	if err := srv.Serve(s.listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		return fmt.Errorf("failed to serve: %w", err)
	}

	logger.Debugf("server.Serve: serving stopped")

	// Return any errors that happened during shutdown.
	select {
	case err := <-errCh:
		return fmt.Errorf("failed to shutdown: %w", err)
	default:
		return nil
	}
}

// Addr returns the server's listening address (ip + port).
func (s *Server) Addr() string {
	return net.JoinHostPort(s.ip, s.port)
}

// IP returns the server's listening IP.
func (s *Server) IP() string {
	return s.ip
}

// Port returns the server's listening port.
func (s *Server) Port() string {
	return s.port
}
