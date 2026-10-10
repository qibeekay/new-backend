package service

import (
	"crypto/subtle"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"portfolio-api/internal/auth"
	"portfolio-api/internal/config"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type AuthService struct {
	email, hash, secret string
}

func NewAuthService(cfg *config.Config) *AuthService {
	return &AuthService{email: cfg.AdminEmail, hash: cfg.AdminPasswordHash, secret: cfg.JWTSecret}
}

// Login checks the single admin account from config and returns a signed token.
func (s *AuthService) Login(email, password string) (string, time.Time, error) {
	// Run both checks every time so timing doesn't reveal which one failed.
	emailOK := subtle.ConstantTimeCompare(
		[]byte(strings.ToLower(strings.TrimSpace(email))),
		[]byte(strings.ToLower(s.email))) == 1
	passOK := bcrypt.CompareHashAndPassword([]byte(s.hash), []byte(password)) == nil
	if !emailOK || !passOK {
		return "", time.Time{}, ErrInvalidCredentials
	}
	return auth.GenerateToken(s.secret, s.email)
}
