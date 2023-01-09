package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"goquizbox/internal/app"
	"goquizbox/internal/logger"
	"goquizbox/internal/server"
	"goquizbox/internal/setup"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	logger.MustInit()
	defer logger.Flush()

	defer func() {
		done()
		if r := recover(); r != nil {
			logger.Fatalf("application panic: %v", r)
		}
	}()

	err := realMain(ctx)
	if err != nil {
		logger.Fatal(err.Error())
	}

	done()

	logger.Info("successful shutdown")
}

func realMain(ctx context.Context) error {
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
