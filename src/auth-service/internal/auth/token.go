package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/ed25519"
)

type JWTSigner struct {
	key ed25519.PrivateKey
}

func NewJWTSigner(privateKey ed25519.PrivateKey) *JWTSigner {
	return &JWTSigner{
		key: privateKey,
	}
}

func (JWT *JWTSigner) CreateAccessToken(userID string) (string, error) {
	now := time.Now()

	claims := jwt.RegisteredClaims{
		Subject:   userID,
		Issuer:    "auth-service",
		Audience:  jwt.ClaimStrings{"smarthome"},
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(accessTokenDuration)),
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodEdDSA,
		claims,
	)

	return token.SignedString(JWT.key)
}
