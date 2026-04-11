package utils

import (
	"errors"
	"time"
	"wchat/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

// Claims 自定义 JWT 载荷
type Claims struct {
	Uuid    string `json:"uuid"`
	IsAdmin int8   `json:"is_admin"`
	jwt.RegisteredClaims
}

// GenToken 签发 Token
func GenToken(uuid string, isAdmin int8) (string, error) {
	cfg := config.GetConfig()

	expireSeconds := cfg.JwtExpireTime
	if expireSeconds <= 0 {
		expireSeconds = 86400 // 默认 1 天
	}
	expireDur := time.Duration(expireSeconds) * time.Second

	claims := Claims{
		Uuid:    uuid,
		IsAdmin: isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expireDur)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    cfg.AppName,
		},
	}
	// use HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JwtSecret))
}

// ParseToken 解析并校验 Token
func ParseToken(tokenString string) (*Claims, error) {
	cfg := config.GetConfig()

	token, err := jwt.ParseWithClaims(
		tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(cfg.JwtSecret), nil
		},
	)
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
