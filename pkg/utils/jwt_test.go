package utils

import (
	"crm_lite/internal/core/config"
	"testing"
	"time"
)

func TestGenerateAndParseTokens(t *testing.T) {
	// 准备测试配置
	opts := &config.Options{}
	opts.Auth.JWTOptions = config.JWTOptions{
		Secret:             "test_secret",
		Issuer:             "crm_test",
		AccessTokenExpire:  time.Minute,
		RefreshTokenExpire: 2 * time.Minute,
	}
	config.SetInstanceForTest(opts)

	access, refresh, err := GenerateTokens("uid123", "tester", []string{"admin", "user"})
	if err != nil {
		t.Fatalf("GenerateTokens error: %v", err)
	}
	if access == "" || refresh == "" {
		t.Fatalf("tokens should not be empty")
	}

	claims, err := ParseToken(access, opts.Auth.JWTOptions)
	if err != nil {
		t.Fatalf("ParseToken error: %v", err)
	}
	if claims.UserID != "uid123" || claims.Username != "tester" {
		t.Errorf("claims mismatch: %+v", claims)
	}

	// 解析非法 token 应该返回错误
	if _, err := ParseToken("abc", opts.Auth.JWTOptions); err == nil {
		t.Errorf("expected error for invalid token")
	}
}
