package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"passKeper/internal/auth"
	"passKeper/internal/config"
	"passKeper/internal/logger"
	"passKeper/internal/repository/db"
	"passKeper/internal/server"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	fmt.Printf("Version: %s, Build Date: %s, Build Commit: %s\n", buildVersion, buildDate, buildCommit)

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// get the server configuration from environment variables or command line flags
	cfg, err := config.LoadServerConfig()
	if err != nil {
		return fmt.Errorf("failed to get server config: %w", err)
	}
	// initialize the logger with the specified log level
	err = logger.Initialize(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Initialize the repository based on the configuration
	repository, err := db.NewDB(context.Background(), cfg.DSN)
	if err != nil {
		return fmt.Errorf("failed to initialize repository: %w", err)
	}
	defer repository.Close()

	// create a context that listens for OS signals to shut down the server
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	// create a new JWT manager
	jwtManager := auth.NewJWTManager(cfg.AuthKey)

	// create a new server
	srv := server.NewServer(jwtManager, repository)
	// start the server
	if err := srv.Serve(ctx, cfg.Address, jwtManager); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}
