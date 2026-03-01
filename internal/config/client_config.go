// Package config provides configuration management for client and server using viper.
package config

import (
	"errors"
	"fmt"
	"log"

	"github.com/spf13/viper"
)

// ClientConfig holds the client configuration (address, token path, log level).
type ClientConfig struct {
	Address        string `mapstructure:"address" env:"SERVER_ADDRESS"`
	TokenStorePath string `mapstructure:"token_store_path" env:"TOKEN_STORE_PATH"`
	LogLevel       string `mapstructure:"log_level" env:"LOG_LEVEL"`
	TLS            bool   `mapstructure:"tls_enabled" env:"TLS_ENABLED"`
	CertFile       string `mapstructure:"tls_cert_file" env:"TLS_CERT_FILE"`
	ServerName     string `mapstructure:"tls_server_name" env:"TLS_SERVER_NAME"`
}

// Default client configuration values.
const (
	DefaultServerAddress  = "localhost:50051"     // DefaultServerAddress is the default gRPC server address.
	DefaultTokenStorePath = "~/.passkeeper/token" // DefaultTokenStorePath is the default token file path.
	DefaultClientTLS      = false                 // DefaultTLS is the default TLS flag.
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

	// Validate the client configuration.
	err = validateClientConfig(&cfg)
	if err != nil {
		return nil, fmt.Errorf("error validating client config: %w", err)
	}
	// Return the client configuration.
	return &cfg, nil
}

// validateClientConfig validates the client configuration.
func validateClientConfig(cfg *ClientConfig) error {
	if cfg.TLS && (cfg.CertFile == "" || cfg.ServerName == "") {
		return errors.New("cert file and server name are required when TLS is enabled")
	}
	return nil
}
