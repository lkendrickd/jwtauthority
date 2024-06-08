package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/lkendrickd/jwtauthorizor/internal/database"
	"github.com/lkendrickd/jwtauthorizor/models"
	"golang.org/x/crypto/bcrypt"
)

// HealthHandler is the health check handler
func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(fmt.Sprintf(`{"healthy":true}` + "\n"))); err != nil {
		return
	}
}

// LoginHandler handles user login and token generation
func (s *Server) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	hashedPassword, ok := database.UserStore[req.Username]
	if !ok || !checkPassword(hashedPassword, req.Password) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := s.GenerateToken(req.Username)
	if err != nil {
		http.Error(w, "Could not generate token", http.StatusInternalServerError)
		log.Printf("Error generating token: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"token": "%s"}`, token)))
}

// ProtectedHandler handles requests to protected routes
func (s *Server) ProtectedHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(models.UserContextKey).(string)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		w.Write([]byte(fmt.Sprintf("Welcome %s!", user)))
	}
}

// ValidateToken validates the JWT token
func (s *Server) ValidateToken(tokenString string) (*models.Claims, error) {
	token, err := jwt.ParseSigned(tokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	claims := &models.Claims{}
	if err := token.Claims([]byte(s.config.HmacKey), claims); err != nil {
		return nil, fmt.Errorf("unable to parse claims: %v", err)
	}

	if err := claims.Validate(jwt.Expected{
		Issuer: s.config.TokenIssuer,
		Time:   time.Now(),
	}); err != nil {
		return nil, fmt.Errorf("unable to validate claims: %v", err)
	}

	return claims, nil
}

func (s *Server) WithAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		claims, err := s.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			log.Printf("Unauthorized access: %v", err)
			return
		}

		ctx := context.WithValue(r.Context(), models.UserContextKey, claims.Username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// checkPassword compares a hashed password with a plain password
func checkPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
