package client

import (
	"fmt"
	"log"
	"passKeper/internal/cmd"
)

var (
	version   string
	buildDate string
)

func main() {
	fmt.Printf("version: %s, buildDate: %s\n", version, buildDate)
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	err := cmd.Execute()
	if err != nil {
		return fmt.Errorf("failed to execute command: %w", err)
	}
	return nil
}
