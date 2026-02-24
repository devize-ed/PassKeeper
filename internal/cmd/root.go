package cmd

import (
	"fmt"
	"passKeper/internal/client/app"
	"passKeper/internal/logger"

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
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Execute  new app instance
			appInstance, err = app.NewApp(serverAddress, tokenStorePath, logger.Log)
			if err != nil {
				return fmt.Errorf("failed to create app instance: %w", err)
			}

			return nil
		},
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	err := rootCmd.Execute()
	if err != nil {
		return fmt.Errorf("failed to execute command: %w", err)
	}
	return nil
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "./config/passKeper.yaml", "config file (default is ./config/passKeper.yaml)")
	rootCmd.PersistentFlags().StringVar(&serverAddress, "address", "localhost:50051", "server address")
	rootCmd.PersistentFlags().StringVar(&tokenStorePath, "token-store-path", "~/.passkeeper/token", "token store path NOTE: default is ~/.passkeeper/token")
}
