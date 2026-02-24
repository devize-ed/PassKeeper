package client

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

var (
	version   string
	buildDate string
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
	logger.Initialize(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	// initialize token store
	tokenStore := tokenStore.NewTokenStore(cfg.TokenStorePath)
	if err != nil {
		return fmt.Errorf("failed to initialize token store: %w", err)
	}
	// create a new gRPC client
	grpcClient, err := grpcclient.NewClient(cfg.Address, tokenStore)
	if err != nil {
		return fmt.Errorf("failed to create gRPC client: %w", err)
	}
	// create a new app instance
	appInstance, err := app.NewApp(grpcClient, tokenStore)
	if err != nil {
		return fmt.Errorf("failed to create app instance: %w", err)
	}
	// set app in cmd
	cmd.SetApp(appInstance)
	defer appInstance.Close()
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
