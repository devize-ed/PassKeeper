// Package main is the entry point for the PassKeeper CLI client.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"passKeper/internal/client/app"
	"passKeper/internal/client/grpcclient"
	"passKeper/internal/client/tokenStore"
	"passKeper/internal/cmd"
	"passKeper/internal/config"
	"passKeper/internal/logger"
	"syscall"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// load the client configuration
	cfg, err := config.LoadClientConfig()
	if err != nil {
		return fmt.Errorf("failed to load client config: %w", err)
	}
	// initialize the logger
	if err := logger.Initialize(cfg.LogLevel); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	// initialize token store
	store := tokenStore.NewTokenStore(cfg.TokenStorePath)
	// create a new gRPC client
	grpcClient, err := grpcclient.NewClient(cfg.Address, store)
	if err != nil {
		return fmt.Errorf("failed to create gRPC client: %w", err)
	}
	// create a new app instance
	appInstance, err := app.NewApp(grpcClient, store)
	if err != nil {
		return fmt.Errorf("failed to create app instance: %w", err)
	}
	// set app in cmd
	cmd.SetApp(appInstance)
	defer func() { _ = appInstance.Close() }()
	// listen for OS signals for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()
	// execute the cmd with context so it can respond to signals
	err = cmd.Execute(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute command: %w", err)
	}
	return nil
}
