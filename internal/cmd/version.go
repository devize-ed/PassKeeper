/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version   string
	buildDate string
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long: `Show version information for the passKeper application.	
	This can be set using the following command:
	go build -ldflags "-X passKeper/internal/cmd.version=1.0.0 -X passKeper/internal/cmd.buildDate=$(date -I)" -o c`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s, Build Date: %s\n", version, buildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
