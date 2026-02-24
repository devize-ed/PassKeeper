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

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit an existing item",
	Long: `Edit an existing information in the passKeeper database with the given item ID and item data.
	This can be done using the following command:
	passKeeper edit --item-id <item-id> or with interactive mode.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Log.Debugf("edit called")
		// get the item id and item data from the flags
		itemID, err := cmd.Flags().GetString("item-id")
		if err != nil {
			return fmt.Errorf("failed to get item id: %w", err)
		}
		itemData, err := cmd.Flags().GetString("item-data")
		if err != nil {
			return fmt.Errorf("failed to get item data: %w", err)
		}
		logger.Log.Debugf("item id: %s, item data: %s", itemID, itemData)
		// call the edit service
		err = appInstance.EditItem(context.Background(), itemID)
		if err != nil {
			return fmt.Errorf("failed to edit item: %w", err)
		}
		logger.Log.Debugf("item updated successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(editCmd)

	editCmd.Flags().StringP("item-id", "i", "", "Item ID to edit")
}
