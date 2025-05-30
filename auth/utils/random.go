package utils

import (
	"auth/model"
	"crypto/rand"
	"encoding/hex"
	mrand "math/rand/v2"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// generates and return a random integer between min and max
func RandMinMax(min, max int) int {
	return min + mrand.IntN(max-min+1)
}

func RandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[mrand.IntN(len(charset))]
	}
	return string(b)
}

func RandomEmail() string {
	return RandomString(10) + "@" + RandomString(5) + "." + RandomString(3)
}

func RandomUser() *model.User {
	return &model.User{
		Email:    RandomEmail(),
		Password: RandomString(10),
	}
}

// GenerateSecureRandomString creates a cryptographically secure random string.
func GenerateSecureRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return hex.EncodeToString(bytes), nil
}
