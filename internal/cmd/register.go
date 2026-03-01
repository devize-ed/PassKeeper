package cmd

import (
	"fmt"
	"passKeper/internal/logger"

	"github.com/spf13/cobra"
)

// registerCmd represents the register command
// register --username <username> --password <password>
var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new user",
	Long: `Register a new user with the given username and password.
	This can be done using the following command:
	passKeper register --username <username> --password <password> or 
	with interactive mode.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Log.Debugf("register called")
		// get the username and password from the flags
		username, err := cmd.Flags().GetString("username")
		if err != nil {
			return fmt.Errorf("failed to get username: %w", err)
		}
		password, err := cmd.Flags().GetString("password")
		if err != nil {
			return fmt.Errorf("failed to get password: %w", err)
		}
		logger.Log.Debugf("Credentials provided for user: %s", username)
		// call the register service
		err = appInstance.Register(cmd.Context(), username, password)
		if err != nil {
			return fmt.Errorf("failed to register: %w", err)
		}
		logger.Log.Debugf("user registered successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(registerCmd)
	registerCmd.Flags().StringP("username", "u", "", "Username for the new account")
	registerCmd.Flags().StringP("password", "p", "", "Password for the new account")
}
