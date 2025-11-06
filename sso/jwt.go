package sso

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	tokenDuration = 24 * 30 * time.Hour
)

type Claims struct {
	Service string `json:"service"`
	User    string `json:"user"`
	jwt.RegisteredClaims
}

func CreateToken(service Service, user string) (string, error) {
	now := time.Now()
	claims := &Claims{
		Service: service.Name,
		User:    user,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(service.Key))
}
