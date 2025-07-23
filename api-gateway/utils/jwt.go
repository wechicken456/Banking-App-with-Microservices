package utils

import (
	"api-gateway/model"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func ValidateJWT(jwtToken string, fingerprintCookie string) (*model.JWTClaim, error) {
	tmp := os.Getenv("JWT_SECRET_KEY")
	if tmp == "" {
		return nil, errors.New("ValidateJWT: JWT secret key is not set in environment variables")
	}
	secret, err := base64.StdEncoding.DecodeString(tmp)
	if err != nil {
		return nil, errors.Errorf("ValidateJWT: failed to decode JWT secret key: %v", err)
	}
	token, err := jwt.ParseWithClaims(jwtToken, &model.JWTClaim{},
		func(token *jwt.Token) (any, error) {
			return []byte(secret), nil
		})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*model.JWTClaim)
	if !ok {
		return nil, errors.New("couldn't parse claims")
	}

	// check token expiration time
	if time.Now().After(claims.ExpiresAt.Time) {
		return nil, errors.New("JWT is expired")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, fmt.Errorf("invalid subject (user id) in token: %v", userID)
	}

	// check fingerprint
	fmt.Printf("fingerprintCookie: %v; claims.FingerprintHash: %v\n", fingerprintCookie, claims.FingerprintHash)
	fingerprintHash := HashSha256(fingerprintCookie)
	if claims.FingerprintHash != fingerprintHash {
		return nil, errors.New("fingerprint doesn't match")
	}

	return claims, nil
}

// set the options for the cookie
func GetTokenAsCookie(name string, val string, maxAgeSeconds int64) string {
	return fmt.Sprintf("__Secure-%s=%s; SameSite=Strict; HttpOnly; Secure; Max-Age: %d", name, val, maxAgeSeconds)
}
