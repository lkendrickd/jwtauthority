package database

import (
	"log"

	"golang.org/x/crypto/bcrypt"
)

// This package can be replaced with a real database connection
// or a more robust in-memory store.

// UserStore - In-memory user store for demonstration purposes
var UserStore = map[string]string{
	"admin":     hashPassword("adminpassword"),
	"alphauser": hashPassword("alpha"),
}

// hashPassword hashes the given password using bcrypt
func hashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}
	return string(hash)
}
