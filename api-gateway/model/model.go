package model

import "github.com/golang-jwt/jwt/v5"

type JWTClaim struct {
	jwt.RegisteredClaims
	FingerprintHash string `json:"fp_hash"`
}

type LoginCreds struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateUserResponse struct {
	UserID string `json:"user_id"`
}

type LoginUserResponse struct {
	UserID       string `json:"user_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Fingerprint  string `json:"fingerprint"`
}

type Account struct{}

var (
	FingerprintCookieName  string = "fingerprint"
	RefreshTokenCookieName string = "refresh_token"
	AccessTokenCookieName  string = "access_token"
)
