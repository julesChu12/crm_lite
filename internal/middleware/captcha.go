package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const CtxKeyCaptchaRequired = "captcha_required"

// SimpleCaptchaGuard 根据登录失败情况动态要求验证码。
// 规则：
// - 默认不要求
// - 当上一次响应包含风控/429/提示关键词时，要求验证码一段时间
// - 这里用简单的基于 cookie 的标记模拟（生产建议放 Redis）
func SimpleCaptchaGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 默认不要求
		reqCaptcha := false

		// 1) 通过一个短期 cookie 标记需要验证码
		if cookie, err := c.Cookie("need_captcha"); err == nil && cookie == "1" {
			reqCaptcha = true
		}

		c.Set(CtxKeyCaptchaRequired, reqCaptcha)
		c.Next()

		// 如果本次返回状态是 429 或消息包含关键字，则设置标记（10 分钟）
		status := c.Writer.Status()
		if status == http.StatusTooManyRequests {
			setNeedCaptchaCookie(c)
			return
		}
		// 简易从 Header 提示关键词（实际可在业务中统一设置）
		if msg := c.Writer.Header().Get("X-Risk-Message"); msg != "" {
			lower := strings.ToLower(msg)
			if strings.Contains(lower, "captcha") || strings.Contains(lower, "验证") || strings.Contains(lower, "turnstile") {
				setNeedCaptchaCookie(c)
			}
		}
	}
}

func setNeedCaptchaCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "need_captcha",
		Value:    "1",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   int((10 * time.Minute).Seconds()),
	})
}
