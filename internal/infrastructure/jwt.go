package infrastructure

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ssklv/mixfood-menu-service/internal/usecase"
)

type tokenProvider struct {
	signingKey []byte
	accessTTL  time.Duration
}

func NewTokenProvider(key string, ttl time.Duration) usecase.TokenProvider {
	return &tokenProvider{
		signingKey: []byte(key),
		accessTTL:  ttl,
	}
}

func (p *tokenProvider) ParseToken(tokenString string) (int64, string, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return p.signingKey, nil
	})
	if err != nil {
		return 0, "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := int64(claims["sub"].(float64))
		role, _ := claims["role"].(string)
		return userID, role, nil
	}
	return 0, "", errors.New("invalid token claims")
}
