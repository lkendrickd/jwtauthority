package config

import (
	"encoding/json"
	"io"
	"os"
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

// NewConfig creates a new Config in which the environment variables take precedence
// then the flags and then the values from the config file
func New(hmacKey string, tokenIssuer string, tokenExpiryMins int, port string) (*Config, error) {
1// stat the file
	stat, err := os.Stat("config/config.json")
	if err != nil {
		// log the error
	}

	// check if the file is a regular file
	if stat.Mode().IsRegular() {
		f, err := os.Open("config/config.json")
		if err != nil {
			// log the error
		}

		defer f.Close()

		data, err := io.ReadAll(f)
		if err != nil {
		// log the error
		}

		// unmarshal the data into a Config struct
		var cfg Config
		if err := json.Unmarshal(data, &cfg); err != nil {
		// log the error
	}

	if hmacKey == "" {
		hmacKey = os.Getenv("HMAC_KEY")
	}

	if tokenIssuer == "" {
		tokenIssuer = os.Getenv("TOKEN_ISSUER")
	}

	// token expiration time in minutes

	if port == "" {
		port = os.Getenv("PORT")
	}

	cfg := &Config{
		HmacKey:            hmacKey,
		TokenIssuer:        tokenIssuer,
		TokenExpirationMin: tokenExpiryMins,
		Port:               port,
	}

}
