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

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all items by type",
	Long: `List all items by type from the passKeeper database.
	This can be done using the following command:
	passKeper list --item-type <item-type> or 
	with interactive mode. if item type is not specified, all items will be listed.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Log.Debugf("list called")
		// get the item type from the flags
		itemType, err := cmd.Flags().GetInt32("item-type")
		if err != nil {
			return fmt.Errorf("failed to get item type: %w", err)
		}
		logger.Log.Debugf("item type: %d", itemType)
		// call the list service
		err = app.ListItems(itemType)
		if err != nil {
			return fmt.Errorf("failed to list items: %w", err)
		}
		logger.Log.Debugf("items listed successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().Int32P("item-type", "t", -1, "Item type to list (0: unspecified, 1: credential, 2: text, 3: binary, 4: card)")
}
