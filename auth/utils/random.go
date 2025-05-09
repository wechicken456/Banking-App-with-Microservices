package utils

import (
	"auth/model"
	"math/rand"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func init() {
	rand.Seed(time.Now().UnixNano())
}

// generates and return a random integer between min and max
func RandMinMax(min, max int) int {
	return min + rand.Intn(max-min+1)
}

func RandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
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
