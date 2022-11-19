package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"goquizbox/internal/app"
	"goquizbox/internal/buildinfo"
	"goquizbox/internal/setup"
	"goquizbox/pkg/logging"
	"goquizbox/pkg/server"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	logger := logging.NewLoggerFromEnv().
		With("build_id", buildinfo.BuildID).
		With("build_tag", buildinfo.BuildTag)
	ctx = logging.WithLogger(ctx, logger)

	defer func() {
		done()
		if r := recover(); r != nil {
			logger.Fatalw("application panic", "panic", r)
		}
	}()

	err := realMain(ctx)
	if err != nil {
		logger.Fatal(err)
	}

	done()

	logger.Info("successful shutdown")
}

func realMain(ctx context.Context) error {
	logger := logging.FromContext(ctx)

	var config app.Config
	env, err := setup.Setup(ctx, &config)
	if err != nil {
		return fmt.Errorf("setup.Setup: %w", err)
	}
	defer env.Close(ctx)

	appServer, err := app.NewServer(&config, env)
	if err != nil {
		return fmt.Errorf("goquizbox.NewServer: %w", err)
	}

	srv, err := server.New(config.Port)
	if err != nil {
		return fmt.Errorf("server.New: %w", err)
	}
	logger.Infof("listening on :%s", config.Port)

	return srv.ServeHTTPHandler(ctx, appServer.Routes(ctx))
}
