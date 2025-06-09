package utils

import (
	"crm_lite/internal/core/config"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// CustomClaims 定义了自定义的 JWT Claims，包含了用户ID和角色等信息
type CustomClaims struct {
	UserID   string `json:"user_id"` // 修改为 string 类型以支持 UUID
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateTokens 生成 Access Token 和 Refresh Token
func GenerateTokens(userID string, username string) (accessToken string, refreshToken string, err error) {
	opts := config.GetInstance().Auth.JWTOptions

	// 生成 Access Token
	accessClaims := CustomClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(opts.AccessTokenExpire)),
			Issuer:    opts.Issuer,
		},
	}
	accessToken, err = createToken(accessClaims, opts.Secret)
	if err != nil {
		return "", "", fmt.Errorf("failed to create access token: %w", err)
	}

	// 生成 Refresh Token
	refreshClaims := CustomClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(opts.RefreshTokenExpire)),
			Issuer:    opts.Issuer,
		},
	}
	refreshToken, err = createToken(refreshClaims, opts.Secret) // 实践中可以使用不同的密钥
	if err != nil {
		return "", "", fmt.Errorf("failed to create refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// createToken 使用指定的 claims 和密钥生成一个 token
func createToken(claims CustomClaims, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken 解析并验证一个JWT
func ParseToken(tokenString string) (*CustomClaims, error) {
	opts := config.GetInstance().Auth.JWTOptions
	secret := opts.Secret

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
