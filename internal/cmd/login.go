/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"passKeper/internal/client/app"
	"passKeper/internal/logger"

	"github.com/spf13/cobra"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to the passKeper application",
	Long: `Login to the passKeper application with the given username and password.
	This can be done using the following command:
	passKeper login --username <username> --password <password> or 
	with interactive mode.
	Note: This command will store the token in the local storage and use it for subsequent commands.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Log.Debugf("login called")
		// get the username and password from the flags
		username, err := cmd.Flags().GetString("username")
		if err != nil {
			return fmt.Errorf("failed to get username: %w", err)
		}
		password, err := cmd.Flags().GetString("password")
		if err != nil {
			return fmt.Errorf("failed to get password: %w", err)
		}
		logger.Log.Debugf("username: %s, password: %s", username, password)
		// call the login service
		err = app.Login(username, password)
		if err != nil {
			return fmt.Errorf("failed to login: %w", err)
		}
		logger.Log.Debugf("user logged in successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)

	loginCmd.Flags().StringP("username", "u", "", "Username for the login")
	loginCmd.Flags().StringP("password", "p", "", "Password for the login")
}
