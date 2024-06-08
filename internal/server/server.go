// server/server.go

package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log/slog"

	"github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/lkendrickd/jwtauthorizor/config"
	"github.com/lkendrickd/jwtauthorizor/internal/middleware"
	"github.com/lkendrickd/jwtauthorizor/models"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func init() {
	// Register the prometheus metrics
	prometheus.MustRegister(middleware.RequestDuration)
	prometheus.MustRegister(middleware.EndpointCount)
}

// TokenValidator defines the interface for validating tokens
type TokenValidator interface {
	ValidateToken(tokenString string) (*models.Claims, error)
}

// Server is the HTTP server
type Server struct {
	config *config.Config
	logger *slog.Logger
	muxer  *http.ServeMux
	server *http.Server
}

// NewServer creates a new Server with middleware applied
func NewServer(l *slog.Logger, mux *http.ServeMux, config *config.Config) *Server {
	// Wrap the existing muxer with the metricsMiddleware
	wrappedMux := middleware.MetricsMiddleware(mux)

	// Create a new http.Server using the wrapped muxer
	server := &http.Server{
		Addr:    config.Port,
		Handler: wrappedMux,
	}

	return &Server{
		config: config,
		logger: l,
		muxer:  mux,
		server: server,
	}
}

// Start starts the server and gracefully handles shutdown
func (s *Server) Start() error {
	// Setting up signal capturing
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	// Add routes to the muxer
	s.logger.Debug(`{"message": "setting up routes"}`)
	s.SetupRoutes()

	// Starting server in a goroutine
	go func() {
		s.logger.Log(context.Background(), slog.LevelInfo, `{"message": "starting server"}`)
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error(`{"message": "server failed to start", "error": ` + err.Error() + `}`)
		}
	}()

	// Block until a signal is received
	<-stopChan
	s.logger.Log(context.Background(), slog.LevelInfo, `{"message": "shutting down server"}`)

	// Create a deadline to wait for this is the duration the server will wait for existing connections to finish
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown the server
	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error(`{"message": "server shutdown failed", "error": ` + err.Error() + `}`)
		return err
	}
	s.logger.Log(context.Background(), slog.LevelInfo, `{"message": "server exited properly"}`)
	return nil
}

// SetupRoutes sets up the server routes
func (s *Server) SetupRoutes() {
	path := "/api/v1"
	s.muxer.HandleFunc(
		fmt.Sprintf(
			"%s %s/login",
			http.MethodPost,
			path,
		), s.LoginHandler,
	)

	s.muxer.Handle(
		fmt.Sprintf(
			"%s %s/protected",
			http.MethodGet,
			path,
		), s.WithAuth(s.ProtectedHandler()),
	)

	s.muxer.HandleFunc("GET /health", s.HealthHandler)
	s.muxer.Handle("GET /metrics", promhttp.Handler())
}

// GenerateToken generates a JWT token
func (s *Server) GenerateToken(username string) (string, error) {
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.HS256, Key: s.config.HmacKey}, nil)
	if err != nil {
		return "", err
	}

	claims := models.Claims{
		Username: username,
		Claims: jwt.Claims{
			Issuer:    s.config.TokenIssuer,
			Subject:   "user_token",
			Audience:  jwt.Audience{"jwtauthorizor"},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Expiry:    jwt.NewNumericDate(time.Now().Add(time.Duration(s.config.TokenExpirationMin) * time.Minute)),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	raw, err := jwt.Signed(signer).Claims(claims).CompactSerialize()
	if err != nil {
		return "", err
	}
	return raw, nil
}
