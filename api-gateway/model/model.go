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

var (
	FingerprintCookieName  string = "fingerprint"
	RefreshTokenCookieName string = "refresh_token"
)
