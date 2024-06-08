package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	jose "github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/lkendrickd/jwtauthorizor/internal/database"
	"github.com/lkendrickd/jwtauthorizor/models"
	"golang.org/x/crypto/bcrypt"
)

// HealthHandler is the health check handler
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(fmt.Sprintf(`{"healthy":true}` + "\n"))); err != nil {
		return
	}
}

// LoginHandler handles user login and token generation
func LoginHandler(w http.ResponseWriter, r *http.Request) {
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

	token, err := generateToken(req.Username)
	if err != nil {
		http.Error(w, "Could not generate token", http.StatusInternalServerError)
		log.Printf("Error generating token: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"token": "%s"}`, token)))
}

// ProtectedHandler handles requests to protected routes
func ProtectedHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(string)
	w.Write([]byte(fmt.Sprintf("Welcome %s!", user)))
}

// checkPassword compares a hashed password with a plain password
func checkPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// generateToken generates a JWT token
func generateToken(username string) (string, error) {
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.HS256, Key: hmacKey}, nil)
	if err != nil {
		return "", err
	}

	claims := models.Claims{
		Username: username,
		Claims: jwt.Claims{
			Issuer:    tokenIssuer,
			Subject:   "user_token",
			Audience:  jwt.Audience{"jwtauthorizor"},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Expiry:    jwt.NewNumericDate(time.Now().Add(time.Duration(tokenExpirationMin) * time.Minute)),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	raw, err := jwt.Signed(signer).Claims(claims).CompactSerialize()
	if err != nil {
		return "", err
	}
	return raw, nil
}
