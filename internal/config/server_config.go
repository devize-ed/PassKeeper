// Package config provides configuration management for server components using viper.
package config

import (
	"errors"
	"fmt"
	"log"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// ServerConfig holds the configuration for the server.
type ServerConfig struct {
	Address  string `mapstructure:"address" env:"SERVER_ADDRESS"`
	DSN      string `mapstructure:"dsn" env:"DATABASE_DSN"`
	AuthKey  string `mapstructure:"auth_key" env:"AUTH_SECRET"`
	LogLevel string `mapstructure:"log_level" env:"LOG_LEVEL"`
}

const (
	DefaultAddress  = ":8080"
	DefaultLogLevel = "info"
)

// LoadServerConfig loads the server configuration from the viper.
func LoadServerConfig() (*ServerConfig, error) {
	// Set the config name and type.
	viper.SetConfigName("server")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config/server_config.yaml")
	// Set the default values.
	viper.SetDefault("address", DefaultAddress)
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
	err := viper.BindEnv("dsn", "DATABASE_DSN")
	if err != nil {
		return nil, fmt.Errorf("error binding environment variable: %w", err)
	}
	err = viper.BindEnv("auth_key", "AUTH_SECRET")
	if err != nil {
		return nil, fmt.Errorf("error binding environment variable: %w", err)
	}
	err = viper.BindEnv("address", "SERVER_ADDRESS")
	if err != nil {
		return nil, fmt.Errorf("error binding environment variable: %w", err)
	}
	//Set flags from the command line.
	pflag.StringP("address", "a", DefaultAddress, "server address")
	pflag.StringP("dsn", "d", "", "database dsn")
	pflag.StringP("auth_key", "k", "", "auth key")
	pflag.StringP("log_level", "l", DefaultLogLevel, "log level")
	// Parse the flags.
	pflag.Parse()
	// Bind the flags to the viper.
	err = viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		return nil, fmt.Errorf("error binding flags to viper: %w", err)
	}

	// Update the config with the values from the viper.
	var cfg ServerConfig
	err = viper.Unmarshal(&cfg)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling server config: %w", err)
	}

	// Validate the server configuration.
	err = validateServerConfig(&cfg)
	if err != nil {
		return nil, fmt.Errorf("error validating server config: %w", err)
	}
	// Return the server configuration.
	return &cfg, nil
}

// validateServerConfig validates the server configuration.
func validateServerConfig(cfg *ServerConfig) error {
	if cfg.AuthKey == "" {
		return errors.New("secret key is required")
	}
	if cfg.DSN == "" {
		return errors.New("dsn is required")
	}
	return nil
}
