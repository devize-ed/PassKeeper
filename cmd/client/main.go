package client

import (
	"fmt"
	"log"
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
	return nil
}
