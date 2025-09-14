package controller

import "github.com/gin-gonic/gin"

// GetIdempotencyKey 从请求头读取幂等键（预留给订单支付/退款收口使用）
func GetIdempotencyKey(c *gin.Context) string {
	return c.GetHeader("Idempotency-Key")
}
