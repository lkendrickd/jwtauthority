package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/lkendrickd/jwtauthorizor/models"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method"},
	)

	EndpointCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"path", "method"},
	)
)

// MetricsMiddleware is the middleware for capturing metrics
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := r.URL.Path
		method := r.Method

		// Increment the endpoint counter
		EndpointCount.WithLabelValues(route, method).Inc()

		// Start timer for duration metric
		timer := prometheus.NewTimer(RequestDuration.WithLabelValues(route, method))
		defer timer.ObserveDuration()

		next.ServeHTTP(w, r)
	})
}

// WithAuth middleware for JWT authentication
type contextKey string

func WithAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		claims, err := ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			log.Printf("Unauthorized access: %v", err)
			return
		}

		ctx := context.WithValue(r.Context(), contextKey("user"), claims.Username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ValidateToken validates the JWT token
func ValidateToken(tokenString string) (*models.Claims, error) {
	token, err := jwt.ParseSigned(tokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	claims := &models.Claims{}
	if err := token.Claims(hmacKey, claims); err != nil {
		return nil, fmt.Errorf("unable to parse claims: %v", err)
	}

	if err := claims.Validate(jwt.Expected{
		Issuer: tokenIssuer,
		Time:   time.Now(),
	}); err != nil {
		return nil, fmt.Errorf("unable to validate claims: %v", err)
	}

	return claims, nil
}
