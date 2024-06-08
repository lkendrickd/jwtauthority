package config

import (
	"encoding/json"
	"io"
	"os"
	"strconv"
)

// Config holds the configuration for the application
type Config struct {
	// HmacKey is the key used to sign the JWT
	HmacKey string
	// TokenIssuer is the issuer of the JWT
	TokenIssuer string
	// TokenExpirationMin is the expiration time of the JWT in minutes
	TokenExpirationMin int
	// Port is the port the server listens on
	Port string
}

type ConfigOption func(*Config)

// WithHmacKey sets the HmacKey in Config
func WithHmacKey(hmacKey string) ConfigOption {
	return func(c *Config) {
		c.HmacKey = hmacKey
	}
}

// WithTokenIssuer sets the TokenIssuer in Config
func WithTokenIssuer(tokenIssuer string) ConfigOption {
	return func(c *Config) {
		c.TokenIssuer = tokenIssuer
	}
}

// WithTokenExpirationMin sets the TokenExpirationMin in Config
func WithTokenExpirationMin(tokenExpirationMin int) ConfigOption {
	return func(c *Config) {
		c.TokenExpirationMin = tokenExpirationMin
	}
}

// WithPort sets the Port in Config
func WithPort(port string) ConfigOption {
	return func(c *Config) {
		c.Port = port
	}
}

// NewConfig creates a new Config with the given options
func NewConfig(opts ...ConfigOption) *Config {
	c := &Config{}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// LoadFromFile loads configuration from a JSON file
func LoadFromFile(filepath string) ([]ConfigOption, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return parseConfig(file)
}

// parseConfig parses configuration options from an io.Reader
func parseConfig(r io.Reader) ([]ConfigOption, error) {
	var fileConfig Config
	if err := json.NewDecoder(r).Decode(&fileConfig); err != nil {
		return nil, err
	}

	options := []ConfigOption{
		WithHmacKey(fileConfig.HmacKey),
		WithTokenIssuer(fileConfig.TokenIssuer),
		WithTokenExpirationMin(fileConfig.TokenExpirationMin),
		WithPort(fileConfig.Port),
	}

	return options, nil
}

// loadFromEnv loads configuration values from environment variables
func LoadFromEnv(cfg *Config) {
	if hmacKey := os.Getenv("HMAC_KEY"); hmacKey != "" {
		cfg.HmacKey = hmacKey
	}

	if tokenIssuer := os.Getenv("TOKEN_ISSUER"); tokenIssuer != "" {
		cfg.TokenIssuer = tokenIssuer
	}

	if tokenExpirationMin := os.Getenv("TOKEN_EXPIRATION_MIN"); tokenExpirationMin != "" {
		if val, err := strconv.Atoi(tokenExpirationMin); err == nil {
			cfg.TokenExpirationMin = val
		}
	}

	if port := os.Getenv("PORT"); port != "" {
		cfg.Port = port
	}
}
