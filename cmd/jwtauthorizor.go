package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"log/slog"

	"github.com/lkendrickd/jwtauthorizor/config"
	"github.com/lkendrickd/jwtauthorizor/internal/server"
)

func main() {
	// Define command-line flags
	port := flag.String("port", "8080", "The port to listen on")
	logLevel := flag.String("logLevel", "info", "The log level")
	hmacKey := flag.String("hmacKey", "", "The HMAC key used to sign the JWT")
	tokenIssuer := flag.String("tokenIssuer", "", "The issuer of the JWT")
	tokenExpirationMin := flag.Int("tokenExpirationMin", 15, "The expiration time of the JWT in minutes")

	// Parse command-line flags
	flag.Parse()

	cfg := config.NewConfig(
		config.WithPort(*port),
		config.WithHmacKey(*hmacKey),
		config.WithTokenIssuer(*tokenIssuer),
		config.WithTokenExpirationMin(*tokenExpirationMin),
	)

	// override the config values with environment variables if they exist
	if envPort, exists := os.LookupEnv("PORT"); exists {
		cfg.Port = envPort
		log.Printf("Using port from environment variable PORT: %s\n", envPort)
	}

	if envHmacKey, exists := os.LookupEnv("HMAC_KEY"); exists {
		cfg.HmacKey = envHmacKey
	}

	if envTokenIssuer, exists := os.LookupEnv("TOKEN_ISSUER"); exists {
		cfg.TokenIssuer = envTokenIssuer
	}

	if envTokenExpirationMin, exists :=
		os.LookupEnv("TOKEN_EXPIRATION_MIN"); exists {
		tokenExpirationMin, err := strconv.Atoi(envTokenExpirationMin)
		if err != nil {
			log.Fatalf("Failed to parse TOKEN_EXPIRATION_MIN: %v", err)
		}

		cfg.TokenExpirationMin = tokenExpirationMin
	}

	// Check for environment variables and override flag values if necessary for log level
	if envLogLevel, exists := os.LookupEnv("LOG_LEVEL"); exists {
		*logLevel = envLogLevel
	}

	// Set the log level based on the provided logLevel string
	slogLevel := setLogLevel(*logLevel)

	// Initialize the logger with the determined log level
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slogLevel}))

	// Initialize the HTTP server mux
	mux := http.NewServeMux()

	// Create and start the server
	s := server.NewServer(logger, mux, cfg)
	if err := s.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// setLogLevel sets the log level based on the provided string
func setLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		fmt.Printf("Unknown log level: %s, defaulting to info\n", level)
		return slog.LevelInfo
	}
}
