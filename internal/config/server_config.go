// Package config provides configuration management for server components using viper.
package config

import (
	"errors"
	"fmt"
	"log"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// ServerConfig holds the server configuration (address, DSN, auth key, log level).
type ServerConfig struct {
	Address  string `mapstructure:"address" env:"SERVER_ADDRESS"`
	DSN      string `mapstructure:"dsn" env:"DATABASE_DSN"`
	AuthKey  string `mapstructure:"auth_key" env:"AUTH_SECRET"`
	LogLevel string `mapstructure:"log_level" env:"LOG_LEVEL"`
	TLS      bool   `mapstructure:"tls_enabled" env:"TLS_ENABLED"`
	CertFile string `mapstructure:"tls_cert_file" env:"TLS_CERT_FILE"`
	KeyFile  string `mapstructure:"tls_key_file" env:"TLS_KEY_FILE"`
}

// Default server configuration values.
const (
	DefaultHost     = "localhost:50051" // DefaultHost is the default gRPC listen address.
	DefaultLogLevel = "info"            // DefaultLogLevel is the default log level (used by client too).
	DefaultTLS      = false             // DefaultTLS is the default TLS flag.
)

// LoadServerConfig loads the server configuration from the viper.
func LoadServerConfig() (*ServerConfig, error) {
	// Set the config name and type.
	viper.SetConfigName("server")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./cfg")
	viper.AddConfigPath("../cfg")
	viper.AddConfigPath("../../cfg")
	// Set the default values.
	viper.SetDefault("address", DefaultHost)
	viper.SetDefault("log_level", DefaultLogLevel)
	viper.SetDefault("tls_enabled", DefaultTLS)
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
	pflag.StringP("address", "a", DefaultHost, "server address")
	pflag.StringP("dsn", "d", "", "database dsn")
	pflag.StringP("auth_key", "k", "", "auth key")
	pflag.StringP("log_level", "l", DefaultLogLevel, "log level")
	pflag.BoolP("tls_enabled", "t", DefaultTLS, "enable TLS")
	pflag.StringP("cert_file", "c", "", "TLS certificate file")
	pflag.StringP("key_file", "b", "", "TLS key file")
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
	if cfg.TLS && (cfg.CertFile == "" || cfg.KeyFile == "") {
		return errors.New("cert file and key file are required when TLS is enabled")
	}
	return nil
}
