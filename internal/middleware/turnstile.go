package middleware

import (
	"bytes"
	"crm_lite/internal/captcha"
	"crm_lite/internal/core/config"
	"crm_lite/pkg/resp"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// TurnstileMiddleware 校验 Cloudflare Turnstile token。
// 在开发/测试模式自动跳过，便于接口调试。
func TurnstileMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		opts := config.GetInstance()
		if opts.Server.Mode == config.DebugMode || opts.Server.Mode == config.TestMode {
			c.Next()
			return
		}

		// 仅针对写操作进行校验：POST/PUT/PATCH/DELETE
		method := c.Request.Method
		if method != http.MethodPost && method != http.MethodPut && method != http.MethodPatch && method != http.MethodDelete {
			c.Next()
			return
		}

		// 允许从多来源获取 token：Header、Query、Form、JSON Body
		token := c.GetHeader("X-Turnstile-Token")
		if token == "" {
			token = c.Query("captcha_token")
		}
		if token == "" {
			token = c.PostForm("captcha_token")
		}
		if token == "" {
			var bodyBytes []byte
			if c.Request.Body != nil {
				bodyBytes, _ = io.ReadAll(c.Request.Body)
			}
			// 恢复 Body，供后续绑定使用
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			// 尝试从 JSON 读取
			var payload struct {
				CaptchaToken string `json:"captcha_token"`
			}
			if len(bodyBytes) > 0 {
				_ = json.Unmarshal(bodyBytes, &payload)
				token = payload.CaptchaToken
			}
		}

		if token == "" {
			resp.Error(c, resp.CodeInvalidParam, "captcha token is required")
			c.Abort()
			return
		}

		ok, err := captcha.VerifyTurnstile(c.Request.Context(), token, c.ClientIP())
		if err != nil {
			resp.SystemError(c, err)
			c.Abort()
			return
		}
		if !ok {
			resp.Error(c, resp.CodeInvalidParam, "captcha verification failed")
			c.Abort()
			return
		}

		c.Next()
	}
}
