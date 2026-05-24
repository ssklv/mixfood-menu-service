package infrastructure

import (
	"errors"
	"fmt"
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
	fmt.Printf("DEBUG: Parsing token: '%s'\n", tokenString) // Видим, что именно пришло

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		// Проверка метода подписи
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return p.signingKey, nil
	})

	if err != nil {
		fmt.Printf("DEBUG: JWT Parse error: %v\n", err)
		return 0, "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Безопасное получение sub
		sub, ok := claims["sub"].(float64)
		if !ok {
			return 0, "", errors.New("sub claim is not a number")
		}

		role, _ := claims["role"].(string)
		return int64(sub), role, nil
	}
	return 0, "", errors.New("invalid token claims")
}
