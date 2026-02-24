/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"passKeper/internal/logger"

	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Retrieve an information by items ID",
	Long: `Retrieve an information by items ID from the passKeeper database.
	This can be done using the following command:
	passKeper get --item-id <item-id> or 
	with interactive mode.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Log.Debugf("get called")
		// get the item id from the flags
		itemID, err := cmd.Flags().GetString("item-id")
		if err != nil {
			return fmt.Errorf("failed to get item id: %w", err)
		}
		logger.Log.Debugf("item id: %s", itemID)
		// call the get service
		_, err = appInstance.GetItem(context.Background(), itemID)
		if err != nil {
			return fmt.Errorf("failed to get item: %w", err)
		}
		logger.Log.Debugf("item retrieved successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)

	getCmd.Flags().StringP("item-id", "i", "", "Item ID to retrieve")
}
