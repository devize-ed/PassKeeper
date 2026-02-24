package cmd

import (
	"context"
	"fmt"
	"passKeper/internal/client/app"

	"github.com/spf13/cobra"
)

var (
	cfgFile        string
	serverAddress  string
	tokenStorePath string

	appInstance *app.App
	err         error

	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "passKeper",
		Short: "Passkeeper is a password manager application",
		Long: `Passkeeper allows you to store and manage your passwords securely.
		Tool allows you to create, update, delete and list your passwords/information.
		for more information, use the -h flag.`,
	}
)

// Execute runs the root command with the given context (e.g. from signal.NotifyContext for graceful shutdown).
func Execute(ctx context.Context) error {
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		return fmt.Errorf("failed to execute command: %w", err)
	}
	return nil
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "./config/passKeper.yaml", "config file (default is ./config/passKeper.yaml)")
	rootCmd.PersistentFlags().StringVar(&serverAddress, "address", "localhost:50051", "server address")
	rootCmd.PersistentFlags().StringVar(&tokenStorePath, "token-store-path", "~/.passkeeper/token", "token store path NOTE: default is ~/.passkeeper/token")
}

// SetApp sets the app instance in the cmd
func SetApp(a *app.App) {
	appInstance = a
}
