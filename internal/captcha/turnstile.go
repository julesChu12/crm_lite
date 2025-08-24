package captcha

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"crm_lite/internal/core/config"
)

// 默认 Cloudflare Turnstile 验证地址，可在测试中覆盖
var turnstileURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

// HTTP 客户端，超时 4 秒
var httpClient = &http.Client{Timeout: 4 * time.Second}

// VerifyTurnstile 调用 Cloudflare Turnstile API 校验 Token。
// 返回值 ok==true 表示验证通过。
func VerifyTurnstile(ctx context.Context, token, remoteIP string) (bool, error) {
	if token == "" {
		return false, errors.New("empty turnstile token")
	}

	opts := config.GetInstance()
	secret := opts.Auth.Captcha.TurnstileSecret
	if secret == "" {
		return false, errors.New("turnstile secret not configured")
	}

	form := url.Values{}
	form.Set("secret", secret)
	form.Set("response", token)
	if remoteIP != "" {
		form.Set("remoteip", remoteIP)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, turnstileURL, strings.NewReader(form.Encode()))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false, fmt.Errorf("turnstile verify returned status %d", resp.StatusCode)
	}

	var result struct {
		Success bool `json:"success"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	return result.Success, nil
}
