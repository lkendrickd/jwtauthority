package models

import (
	"github.com/go-jose/go-jose/v3/jwt"
)

// LoginRequest struct to hold login request data
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Claims struct to hold custom claims
type Claims struct {
	Username string `json:"username"`
	jwt.Claims
}

type ContextKey string

var UserContextKey = ContextKey("user")
