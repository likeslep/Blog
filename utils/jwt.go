// utils/jwt.go
package utils

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

// JWT 配置
type JWTConfig struct {
	Secret     string        `json:"secret"`
	ExpireTime time.Duration `json:"expire_time"`
	Issuer     string        `json:"issuer"`
}

// 自定义 Claims
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// 生成 Token
func GenerateToken(userID uint, username, role string, config JWTConfig) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.ExpireTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    config.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.Secret))
}

// 解析 Token
func ParseToken(tokenString string, secret string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			return nil, ErrInvalidToken
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// 刷新 Token
func RefreshToken(tokenString string, config JWTConfig) (string, error) {
	claims, err := ParseToken(tokenString, config.Secret)
	if err != nil {
		return "", err
	}

	// 检查是否在可刷新时间内（比如过期前30分钟内）
	if time.Until(claims.ExpiresAt.Time) > 30*time.Minute {
		return "", ErrTokenNotExpired
	}

	// 生成新 token
	return GenerateToken(claims.UserID, claims.Username, claims.Role, config)
}

// 从 Token 中获取用户信息
func GetUserInfoFromToken(tokenString string, secret string) (uint, string, string, error) {
	claims, err := ParseToken(tokenString, secret)
	if err != nil {
		return 0, "", "", err
	}
	return claims.UserID, claims.Username, claims.Role, nil
}
