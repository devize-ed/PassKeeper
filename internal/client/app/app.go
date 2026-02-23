package app

import (
	"context"
	"fmt"
	"os"
	client "passKeper/internal/client/grpcclient"
	"passKeper/internal/logger"
	"strings"

	"github.com/charmbracelet/x/term"
)

var err error

func Register(username, password string) error {
	if username == "" {
		fmt.Println("Enter your username: ")
		usernameBytes, err := term.ReadPassword(uintptr(os.Stdin.Fd()))
		if err != nil {
			return fmt.Errorf("failed to read username: %w", err)
		}
		username = strings.TrimSpace(string(usernameBytes))
	}
	if password == "" {
		fmt.Println("Enter your password: ")
		passwordBytes, err := term.ReadPassword(uintptr(os.Stdin.Fd()))
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
		password = strings.TrimSpace(string(passwordBytes))
	}
	// call the register service
	grpcClient, err := client.NewClient(os.Getenv("PASSKEEPER_HOST"), nil, logger.Log)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer grpcClient.Close()
	err = grpcClient.Register(context.Background(), username, password)
	if err != nil {
		return fmt.Errorf("failed to register: %w", err)
	}
	return nil
}

func Login(username, password string) error {
	return nil
}

func CreateItem(itemType int32) error {
	return nil
}

func EditItem(itemID string) error {
	return nil
}

func DeleteItem(itemID string) error {
	return nil
}

func GetItem(itemID string) error {
	return nil
}

func ListItems(itemType int32) error {
	return nil
}
