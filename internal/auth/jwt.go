package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const TokenTTL = 12 * time.Hour

func GenerateToken(secret, subject string) (string, time.Time, error) {
	exp := time.Now().Add(TokenTTL)
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   subject,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(exp),
	}).SignedString([]byte(secret))
	return token, exp, err
}

func ValidateToken(secret, raw string) error {
	_, err := jwt.ParseWithClaims(raw, &jwt.RegisteredClaims{},
		func(*jwt.Token) (any, error) { return []byte(secret), nil },
		jwt.WithValidMethods([]string{"HS256"}))
	return err
}
