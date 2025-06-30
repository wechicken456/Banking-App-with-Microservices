package utils

import (
	"auth/model"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

func RandomAccessToken(userID uuid.UUID) (*model.AccessToken, error) {
	// Get JWT Secret Key from environment variable
	tmp := os.Getenv("JWT_SECRET_KEY")
	if tmp == "" {
		return nil, errors.New("RandomAccessToken: JWT secret key is not set in environment variables")
	}
	jwtSecretKey, err := base64.StdEncoding.DecodeString(tmp)
	if err != nil {
		return nil, errors.Errorf("RandomAccessToken: failed to decode JWT secret key: %v", err)
	}

	// Generate fingerprint for JWT
	fingerprintValue, err := GenerateSecureRandomString(32) // 32 bytes gives 43 URL-safe characters
	if err != nil {
		return nil, errors.Errorf("RandomAccessToken: failed to generate fingerprint string: %v", err)
	}

	// Hash Fingerprint for JWT Claim
	fingerprintHash := HashSha256(fingerprintValue)

	// Prepare JWT Claims
	expirationTime := time.Now().Add(model.TokenShortDuration)
	claim := &model.JWTClaim{
		FingerprintHash: fingerprintHash,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "auth-service",
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Generate JWT
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	signedAccessToken, err := accessToken.SignedString(jwtSecretKey)
	if err != nil {
		return nil, errors.Errorf("RandomAccessToken: failed to sign JWT token: %v", err)
	}

	return &model.AccessToken{
		Token:       signedAccessToken,
		Fingerprint: fingerprintValue,
		Duration:    int(model.TokenShortDuration),
	}, nil
}

func RandomRefreshToken() (*model.RefreshToken, error) {
	// generate refresh token
	refreshToken, err := GenerateSecureRandomString(32) // 32 bytes gives 43 URL-safe characters
	if err != nil {
		return nil, errors.Errorf("RandomRefreshToken: failed to generate refresh string: %v", err)
	}

	return &model.RefreshToken{
		Token:    refreshToken,
		Duration: int(model.RefreshTokenDuration),
	}, nil
}

// set the options for the cookie
func GetTokenAsCookie(name string, val string, maxAgeSeconds int64) string {
	return fmt.Sprintf("__Secure-%s=%s; SameSite=Strict; HttpOnly; Secure; Max-Age: %d", name, val, maxAgeSeconds)
}
