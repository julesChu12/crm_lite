package middleware

import (
	"crm_lite/internal/core/resource"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const CtxKeyCaptchaRequired = "captcha_required"

// SimpleCaptchaGuard 根据登录失败/风控动态要求验证码（使用 Redis 存储标记）。
func SimpleCaptchaGuard(resManager *resource.Manager) gin.HandlerFunc {
	// 获取 Redis 客户端
	cacheRes, err := resource.Get[*resource.CacheResource](resManager, resource.CacheServiceKey)
	if err != nil || cacheRes == nil || cacheRes.Client == nil {
		panic("captcha guard requires redis cache resource")
	}
	client := cacheRes.Client

	return func(c *gin.Context) {
		ip := c.ClientIP()
		cacheKey := fmt.Sprintf("risk:captcha:ip:%s", ip)

		// 1) 读取是否要求验证码
		reqCaptcha := false
		if n, _ := client.Exists(c.Request.Context(), cacheKey).Result(); n > 0 {
			reqCaptcha = true
		}
		c.Set(CtxKeyCaptchaRequired, reqCaptcha)

		c.Next()

		// 2) 若返回 429 或风控提示，则设置标记（10 分钟）
		status := c.Writer.Status()
		if status == http.StatusTooManyRequests {
			_ = client.Set(c.Request.Context(), cacheKey, "1", 10*time.Minute).Err()
			return
		}
		if msg := c.Writer.Header().Get("X-Risk-Message"); msg != "" {
			lower := strings.ToLower(msg)
			if strings.Contains(lower, "captcha") || strings.Contains(lower, "验证") || strings.Contains(lower, "turnstile") {
				_ = client.Set(c.Request.Context(), cacheKey, "1", 10*time.Minute).Err()
			}
		}
	}
}
