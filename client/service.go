package sso

import (
	"fmt"
	"net/url"

	"github.com/golang-jwt/jwt/v5"
)

type Config struct {
	SSOServiceName string
	SSOServiceKey  string
	SSOServiceURL  string
}

type Claims struct {
	Service string `json:"service"`
	User    string `json:"user"`
	jwt.RegisteredClaims
}

type Service interface {
	GetLoginURL() string
	ValidateToken(tokenString string) (*Claims, error)
}

type service struct {
	serviceName string
	key         []byte
	ssoURL      *url.URL
}

func NewService(cfg *Config) (Service, error) {
	if cfg.SSOServiceName == "" {
		return nil, fmt.Errorf("SSO_SERVICE_NAME not set")
	}
	if cfg.SSOServiceKey == "" {
		return nil, fmt.Errorf("SSO_SERVICE_KEY not set")
	}
	if cfg.SSOServiceURL == "" {
		return nil, fmt.Errorf("SSO_SERVICE_URL not set")
	}

	ssoURL, err := url.Parse(cfg.SSOServiceURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SSO_SERVICE_URL: %w", err)
	}

	return &service{
		serviceName: cfg.SSOServiceName,
		key:         []byte(cfg.SSOServiceKey),
		ssoURL:      ssoURL,
	}, nil
}

func (s *service) GetLoginURL() string {
	q := s.ssoURL.Query()
	q.Set("service", s.serviceName)
	loginURL := *s.ssoURL
	loginURL.RawQuery = q.Encode()
	return loginURL.String()
}

func (s *service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.key, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		if claims.Service != s.serviceName {
			return nil, fmt.Errorf("token service mismatch")
		}
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
