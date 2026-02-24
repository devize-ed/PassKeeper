package app

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	client "passKeper/internal/client/grpcclient"
	"passKeper/internal/logger"
	pb "passKeper/pkg/api"
	"strconv"
	"strings"

	"github.com/charmbracelet/x/term"
)

// TokenStore interface provides the methods to interact with the token store.
type TokenStore interface {
	Save(token string) error
	Load() (string, error)
}

// App struct provides the methods to interact with the application.
type App struct {
	grpcClient *client.Client
	ts         TokenStore
	in         *bufio.Reader
	out        *bufio.Writer
}

// NewApp creates a new app instance.
func NewApp(grpcClient *client.Client, ts TokenStore) (*App, error) {
	return &App{grpcClient: grpcClient, ts: ts, in: bufio.NewReader(os.Stdin), out: bufio.NewWriter(os.Stdout)}, nil
}

// Close closes the app instance.
func (a *App) Close() error {
	return a.grpcClient.Close()
}

// Register registers a new user.
func (a *App) Register(ctx context.Context, username, password string) error {
	logger.Log.Debugf("Registering user: %s", username)
	// if username and password are not provided, read from stdin
	if username == "" {
		fmt.Println("Enter your username: ")
		usernameStr, err := a.in.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read username: %w", err)
		}
		username = strings.TrimSpace(usernameStr)
	}
	if password == "" {
		fmt.Println("Enter your password: ")
		passwordBytes, err := term.ReadPassword(uintptr(os.Stdin.Fd()))
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
		password = strings.TrimSpace(string(passwordBytes))
	}
	logger.Log.Debugf("Calling register service for user: %s", username)
	// call the register service
	err := a.grpcClient.Register(ctx, username, password)
	if err != nil {
		return fmt.Errorf("failed to register: %w", err)
	}
	return nil
}

// Login logs in a user.
func (a *App) Login(ctx context.Context, username, password string) error {
	logger.Log.Debugf("Logging in user: %s", username)
	// if username and password are not provided, read from stdin
	if username == "" {
		fmt.Println("Enter your username: ")
		usernameStr, err := a.in.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read username: %w", err)
		}
		username = strings.TrimSpace(usernameStr)
	}
	if password == "" {
		fmt.Println("Enter your password: ")
		passwordBytes, err := term.ReadPassword(uintptr(os.Stdin.Fd()))
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
		password = strings.TrimSpace(string(passwordBytes))
	}
	logger.Log.Debugf("Calling login service for user: %s", username)
	// call the login service
	token, err := a.grpcClient.Login(ctx, username, password)
	if err != nil {
		return fmt.Errorf("failed to login: %w", err)
	}
	// save the token to the token store
	err = a.ts.Save(token)
	if err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}
	//
	fmt.Println("Login successful. Token saved to token store.")
	return nil
}

// CreateItem creates a new item.
func (a *App) CreateItem(ctx context.Context, itemType int32) error {
	err := a.requireAuthentication()
	if err != nil {
		return fmt.Errorf("failed to require authentication: %w", err)
	}
	logger.Log.Debugf("Creating item with type: %d", itemType)
	// if item type is not provided, read from stdin
	if itemType < 1 || itemType > 4 {
		fmt.Println("Enter the item type: 1: credential, 2: text, 3: binary, 4: card)")
		itemTypeStr, err := a.in.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read item type: %w", err)
		}
		itemTypeStr = strings.TrimSpace(itemTypeStr)
		itemTypeInt, err := strconv.Atoi(itemTypeStr)
		if err != nil {
			return fmt.Errorf("failed to convert item type to int: %w", err)
		}
		itemType = int32(itemTypeInt)
	}
	// create the item data
	itemData, err := a.formItemData(itemType)
	if err != nil {
		return fmt.Errorf("failed to create item data: %w", err)
	}
	logger.Log.Debugf("Calling create item service for type: %d", itemType)
	err = a.grpcClient.CreateItem(ctx, pb.ItemType(itemType), itemData)
	if err != nil {
		return fmt.Errorf("failed to create item: %w", err)
	}
	return nil
}

// EditItem edits an existing item.
func (a *App) EditItem(ctx context.Context, itemID string) error {
	err := a.requireAuthentication()
	if err != nil {
		return fmt.Errorf("failed to require authentication: %w", err)
	}
	logger.Log.Debugf("Editing item with ID: %s", itemID)
	// if item type is not provided, read from stdin
	if itemID == "" {
		fmt.Println("Enter the item ID: ")
		itemIDStr, err := a.in.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read item ID: %w", err)
		}
		itemID = strings.TrimSpace(itemIDStr)
	}
	logger.Log.Debugf("Calling pre-update get item service for ID: %s", itemID)
	// get the current item data
	itemData, err := a.GetItem(ctx, itemID)
	if err != nil {
		return fmt.Errorf("failed to get item data: %w", err)
	}
	logger.Log.Debugf("Forming new item data for type: %d", itemData.Type)
	// form the new item data
	newItemData, err := a.formItemData(int32(itemData.Type))
	if err != nil {
		return fmt.Errorf("failed to form new item data: %w", err)
	}
	logger.Log.Debugf("Calling update item service for ID: %s", itemID)
	// update the item on the server
	err = a.grpcClient.UpdateItem(ctx, itemID, itemData.Type, newItemData, itemData.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update item: %w", err)
	}
	// print the updated item information
	fmt.Println("Item updated successfully.")
	return nil
}

// GetItem retrieves an information about an existing item from the server.
func (a *App) GetItem(ctx context.Context, itemID string) (*pb.Item, error) {
	err := a.requireAuthentication()
	if err != nil {
		return nil, fmt.Errorf("failed to require authentication: %w", err)
	}
	logger.Log.Debugf("Getting item with ID: %s", itemID)
	// if item ID is not provided, read from stdin
	if itemID == "" {
		fmt.Println("Enter the item ID: ")
		itemIDStr, err := a.in.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read item ID: %w", err)
		}
		itemID = strings.TrimSpace(itemIDStr)
	}
	logger.Log.Debugf("Calling get item service for ID: %s", itemID)
	// get the item from the server
	item, err := a.grpcClient.GetItem(ctx, itemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}
	logger.Log.Debugf("Marshalling item to JSON for ID: %s", itemID)
	// marshal the item to JSON
	itemJSON, err := json.MarshalIndent(item, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal item: %w", err)
	}
	// write the item to the output
	a.out.WriteString("Current item information for item ID: " + itemID + ": \n")
	a.out.WriteString(string(itemJSON) + "\n")
	a.out.Flush()
	return item, nil
}

func (a *App) ListItems(ctx context.Context, itemType int32) error {
	err := a.requireAuthentication()
	if err != nil {
		return fmt.Errorf("failed to require authentication: %w", err)
	}
	logger.Log.Debugf("Listing items with type: %d", itemType)
	// if item type is not provided, read from stdin
	if itemType < 0 || itemType > 4 {
		fmt.Println(`Enter the item type: 0: unspecified, 1: credential, 2: text, 3: binary, 4: card)
		for unspecified all items will be listed.`)
		itemTypeStr, err := a.in.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read item type: %w", err)
		}
		itemTypeStr = strings.TrimSpace(itemTypeStr)
		itemTypeInt, err := strconv.Atoi(itemTypeStr)
		if err != nil {
			return fmt.Errorf("failed to convert item type to int: %w", err)
		}
		itemType = int32(itemTypeInt)
	}
	logger.Log.Debugf("Calling list items service for type: %d", itemType)
	// list the items from the server
	items, err := a.grpcClient.ListItems(ctx, pb.ItemType(itemType))
	if err != nil {
		return fmt.Errorf("failed to list items: %w", err)
	}
	// marshal the items to JSON
	itemsJSON, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal items: %w", err)
	}
	// write the items to the output
	a.out.WriteString("Current items information: \n")
	a.out.WriteString(string(itemsJSON) + "\n")
	a.out.Flush()
	return nil
}

// formItemData forms a item data for the given item type in interactive mode.
func (a *App) formItemData(itemType int32) (*pb.ItemData, error) {
	if itemType == 1 {
		return a.createCredentialItem()
	}
	if itemType == 2 {
		return a.createTextItem()
	}
	if itemType == 3 {
		return a.createBinaryItem()
	}
	if itemType == 4 {
		return a.createCardItem()
	}
	return nil, fmt.Errorf("invalid item type: %d", itemType)
}

// createCredentialItem creates a credential item with interactive mode.
func (a *App) createCredentialItem() (*pb.ItemData, error) {
	fmt.Println("Enter the user login: ")
	userLoginStr, err := a.in.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read user login: %w", err)
	}
	userLogin := strings.TrimSpace(userLoginStr)
	fmt.Println("Enter the user password: ")
	userPasswordStr, err := term.ReadPassword(uintptr(os.Stdin.Fd()))
	if err != nil {
		return nil, fmt.Errorf("failed to read user password: %w", err)
	}
	userPassword := strings.TrimSpace(string(userPasswordStr))
	fmt.Println("Enter the metadata: ")
	metadataStr, err := a.in.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}
	metadata := strings.TrimSpace(metadataStr)
	itemData := &pb.ItemData{
		Data: &pb.ItemData_Credentials{
			Credentials: &pb.Credentials{
				UserLogin:    userLogin,
				UserPassword: userPassword,
				Metadata:     metadata,
			},
		},
	}
	return itemData, nil
}

// createTextItem creates a text item with interactive mode.
func (a *App) createTextItem() (*pb.ItemData, error) {
	fmt.Println("Enter the text: ")
	textStr, err := a.in.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read text: %w", err)
	}
	text := strings.TrimSpace(textStr)
	fmt.Println("Enter the metadata: ")
	metadataStr, err := a.in.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}
	metadata := strings.TrimSpace(metadataStr)
	itemData := &pb.ItemData{
		Data: &pb.ItemData_Text{
			Text: &pb.Text{
				Text:     text,
				Metadata: metadata,
			},
		},
	}
	return itemData, nil
}

// createBinaryItem creates a binary item with interactive mode.
func (a *App) createBinaryItem() (*pb.ItemData, error) {
	fmt.Println("Enter the binary data: ")
	binaryDataStr, err := a.in.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read binary data: %w", err)
	}
	binaryData := strings.TrimSpace(binaryDataStr)
	fmt.Println("Enter the metadata: ")
	metadataStr, err := a.in.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}
	metadata := strings.TrimSpace(metadataStr)
	itemData := &pb.ItemData{
		Data: &pb.ItemData_Binary{
			Binary: &pb.Binary{
				Data:     []byte(binaryData),
				Metadata: metadata,
			},
		},
	}
	return itemData, nil
}

// createCardItem creates a card item with interactive mode.
func (a *App) createCardItem() (*pb.ItemData, error) {
	fmt.Println("Enter the name: ")
	nameStr, err := a.in.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read name: %w", err)
	}
	name := strings.TrimSpace(nameStr)
	fmt.Println("Enter the number: ")
	numberStr, err := term.ReadPassword(uintptr(os.Stdin.Fd()))
	if err != nil {
		return nil, fmt.Errorf("failed to read number: %w", err)
	}
	number := strings.TrimSpace(string(numberStr))
	fmt.Println("Enter the exp: ")
	expStr, err := a.in.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read exp: %w", err)
	}
	exp := strings.TrimSpace(expStr)
	fmt.Println("Enter the cvv: ")
	cvvStr, err := term.ReadPassword(uintptr(os.Stdin.Fd()))
	if err != nil {
		return nil, fmt.Errorf("failed to read cvv: %w", err)
	}
	cvv := strings.TrimSpace(string(cvvStr))
	fmt.Println("Enter the metadata: ")
	metadataStr, err := a.in.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}
	metadata := strings.TrimSpace(metadataStr)
	itemData := &pb.ItemData{
		Data: &pb.ItemData_Card{
			Card: &pb.Card{
				Name:     name,
				Number:   number,
				Exp:      exp,
				Cvv:      cvv,
				Metadata: metadata,
			},
		},
	}
	return itemData, nil
}

func (a *App) requireAuthentication() error {
	// check if the token is present in the token store
	token, err := a.ts.Load()
	if err != nil {
		return fmt.Errorf("authentication required: %w", err)
	}
	if token == "" {
		return fmt.Errorf("authentication required: please login")
	}
	return nil
}
