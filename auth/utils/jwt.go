package utils

import (
	"auth/model"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

func RandomAccessToken(userID uuid.UUID, jwtSecretKey string) (*model.AccessToken, error) {
	// Generate fingerprint for JWT
	fingerprintValue, err := GenerateSecureRandomString(32) // 32 bytes gives 43 URL-safe characters
	if err != nil {
		return nil, errors.Errorf("RandomAccessToken: failed to generate fingerprint string: %v", err)
	}

	// Create the fingerprint cookie string
	// see implementaion reference here: https://github.com/OWASP/CheatSheetSeries/blob/master/cheatsheets/JSON_Web_Token_for_Java_Cheat_Sheet.md#implementation-example-1
	fingerprintAsCookie := GetTokenAsCookie(model.FingerprintCookieName, fingerprintValue, int64(model.TokenShortDuration))
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
		TokenString:             signedAccessToken,
		FingerprintCookieString: fingerprintAsCookie,
	}, nil
}

func RandomRefreshToken() (*model.RefreshToken, error) {
	// generate refresh token
	refreshValue, err := GenerateSecureRandomString(32) // 32 bytes gives 43 URL-safe characters
	if err != nil {
		return nil, errors.Errorf("RandomRefreshToken: failed to generate refresh string: %v", err)
	}
	// Create the refresh token cookie string
	// see implementaion reference here: https://github.com/OWASP/CheatSheetSeries/blob/master/cheatsheets/JSON_Web_Token_for_Java_Cheat_Sheet.md#implementation-example-1
	refreshTokenAsCookie := GetTokenAsCookie(model.RefreshTokenName, refreshValue, int64(model.RefreshTokenDuration))

	// Hash refresh token
	refreshHash := HashSha256(refreshValue)

	return &model.RefreshToken{
		TokenAsCookie: refreshTokenAsCookie,
		TokenHash:     refreshHash,
	}, nil
}

// set the options for the cookie
func GetTokenAsCookie(name string, val string, maxAgeSeconds int64) string {
	return fmt.Sprintf("__Secure-%s=%s; SameSite=Strict; HttpOnly; Secure; Max-Age: %d", name, val, maxAgeSeconds)
}
