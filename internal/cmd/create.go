package cmd

import (
	"context"
	"fmt"
	"passKeper/internal/logger"

	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new item",
	Long: `Create and save a new information in the passKeeper database with the given item type and item data.
	This can be done using the following command:
	passKeper create --item-type <item-type> or with interactive mode.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Log.Debugf("create called")
		// get the item type and item data from the flags
		itemType, err := cmd.Flags().GetInt32("item-type")
		if err != nil {
			return fmt.Errorf("failed to get item type: %w", err)
		}
		logger.Log.Debugf("item type: %d", itemType)
		// call the create service
		err = appInstance.CreateItem(context.Background(), itemType)
		if err != nil {
			return fmt.Errorf("failed to create item: %w", err)
		}
		logger.Log.Debugf("item created successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().Int32P("item-type", "t", -1, "Item type to create (1: credential, 2: text, 3: binary, 4: card)")
}
