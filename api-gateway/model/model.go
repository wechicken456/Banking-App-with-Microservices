package model

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaim struct {
	jwt.RegisteredClaims
	FingerprintHash string `json:"fpHash"`
}

type LoginCreds struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateUserResponse struct {
	UserID string `json:"userId"`
}

type LoginResponse struct {
	UserID               string `json:"userId"`
	Email                string `json:"email"`
	AccessToken          string `json:"accessToken"`
	Fingerprint          string `json:"fingerprint"`
	RefreshToken         string `json:"refreshToken"`
	AccessTokenDuration  int32  `json:"accessTokenDuration"`
	RefreshTokenDuration int32  `json:"refreshTokenDuration"`
}

type RenewAccessTokenResponse struct {
	AccessToken         string `json:"accessToken"`
	Fingerprint         string `json:"fingerprint"`
	AccessTokenDuration int32  `json:"accessTokenDuration"`
}

type Account struct{}

var (
	TokenShortDuration     time.Duration = 15 * time.Minute
	TokenAbsoluteDuration  time.Duration = 4 * time.Hour
	RefreshTokenDuration   time.Duration = 24 * time.Hour
	FingerprintCookieName  string        = "fingerprint"
	AccessTokenCookieName  string        = "accessToken"
	RefreshTokenCookieName string        = "refreshToken"
)
