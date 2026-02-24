// Package config provides configuration management for client and server using viper.
package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

// ClientConfig holds the client configuration (address, token path, log level).
type ClientConfig struct {
	Address        string `mapstructure:"address" env:"SERVER_ADDRESS"`
	TokenStorePath string `mapstructure:"token_store_path" env:"TOKEN_STORE_PATH"`
	LogLevel       string `mapstructure:"log_level" env:"LOG_LEVEL"`
}

// Default client configuration values.
const (
	DefaultServerAddress  = "localhost:50051"     // DefaultServerAddress is the default gRPC server address.
	DefaultTokenStorePath = "~/.passkeeper/token" // DefaultTokenStorePath is the default token file path.
)

// LoadClientConfig loads the client configuration from the viper.
func LoadClientConfig() (*ClientConfig, error) {
	// Set the config name and type.
	viper.SetConfigName("client")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./cfg")
	viper.AddConfigPath("../cfg")
	viper.AddConfigPath("../../cfg")
	// Set the default values.
	viper.SetDefault("address", DefaultServerAddress)
	viper.SetDefault("token_store_path", DefaultTokenStorePath)
	viper.SetDefault("log_level", DefaultLogLevel)
	// Read the config file.
	if err := viper.ReadInConfig(); err != nil {
		_, ok := err.(viper.ConfigFileNotFoundError)
		if !ok {
			return nil, fmt.Errorf("error reading client config: %w", err)
		}
		log.Println("WARN: client config file not found, using default values")
	}
	// set the environment variable binding.
	viper.AutomaticEnv()
	err := viper.BindEnv("token_store_path", "TOKEN_STORE_PATH")
	if err != nil {
		return nil, fmt.Errorf("error binding environment variable: %w", err)
	}
	err = viper.BindEnv("address", "SERVER_ADDRESS")
	if err != nil {
		return nil, fmt.Errorf("error binding environment variable: %w", err)
	}

	// Update the config with the values from the viper.
	var cfg ClientConfig
	err = viper.Unmarshal(&cfg)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling client config: %w", err)
	}
	return &cfg, nil
}
