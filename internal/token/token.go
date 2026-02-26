package token

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"auth-as-a-service/internal/redis"
)

func Generate(userID string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	expiryHours := 24
	if h, err := strconv.Atoi(os.Getenv("JWT_EXPIRY_HOURS")); err == nil && h > 0 {
		expiryHours = h
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"sub": userID,
		"jti": uuid.New().String(),
		"exp": now.Add(time.Duration(expiryHours) * time.Hour).Unix(),
		"iat": now.Unix(),
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(secret))
}

func Validate(ctx context.Context, tokenString string, cache redis.Service) (string, error) {
	secret := os.Getenv("JWT_SECRET")

	t, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok || !t.Valid {
		return "", fmt.Errorf("invalid token")
	}

	jti, ok := claims["jti"].(string)
	if !ok || jti == "" {
		return "", fmt.Errorf("missing jti claim")
	}

	val, err := cache.Get(ctx, "blacklist:"+jti)
	if err == nil && val != "" {
		return "", fmt.Errorf("token revoked")
	}

	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return "", fmt.Errorf("missing sub claim")
	}

	return sub, nil
}

func Revoke(ctx context.Context, tokenString string, cache redis.Service) error {
	t, _, err := jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return fmt.Errorf("parse token: %w", err)
	}

	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return fmt.Errorf("invalid claims")
	}

	jti, ok := claims["jti"].(string)
	if !ok || jti == "" {
		return fmt.Errorf("missing jti claim")
	}

	expFloat, ok := claims["exp"].(float64)
	if !ok {
		return fmt.Errorf("missing exp claim")
	}

	ttl := time.Until(time.Unix(int64(expFloat), 0))
	if ttl <= 0 {
		return nil
	}

	return cache.Set(ctx, "blacklist:"+jti, "1", ttl)
}
