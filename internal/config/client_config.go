// Package config provides configuration management for server components using viper.
package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

// ServerConfig holds the configuration for the server.
type ClientConfig struct {
	Address        string `mapstructure:"address" env:"SERVER_ADDRESS"`
	TokenStorePath string `mapstructure:"token_store_path" env:"TOKEN_STORE_PATH"`
	LogLevel       string `mapstructure:"log_level" env:"LOG_LEVEL"`
}

const (
	DefaultServerAddress  = "localhost:50051"
	DefaultTokenStorePath = "~/.passkeeper/token"
)

// LoadClientConfig loads the client configuration from the viper.
func LoadClientConfig() (*ClientConfig, error) {
	// Set the config name and type.
	viper.SetConfigName("client")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config/client_config.yaml")
	// Set the default values.
	viper.SetDefault("address", DefaultServerAddress)
	viper.SetDefault("token_store_path", DefaultTokenStorePath)
	viper.SetDefault("log_level", DefaultLogLevel)
	// Read the config file.
	if err := viper.ReadInConfig(); err != nil {
		_, ok := err.(viper.ConfigFileNotFoundError)
		if !ok {
			return nil, fmt.Errorf("error reading server config: %w", err)
		}
		log.Println("WARN: server config file not found, using default values")
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
		return nil, fmt.Errorf("error unmarshalling server config: %w", err)
	}
	// Return the server configuration.
	return &cfg, nil
}
