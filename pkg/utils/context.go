package utils

import "context"

// contextKey 是一个私有类型，用于避免不同包之间的 key 冲突
// 参考 https://golang.org/pkg/context/#WithValue 的官方建议
// 使用非导出类型以及零值作为 key
// 这样其它包无法获取到 key 的具体类型，避免了覆盖。
type contextKey string

const (
	userIDKey   contextKey = "user_id"
	usernameKey contextKey = "username"
)

// WithUser 将用户 ID 与用户名存入 context 并返回新的 context
func WithUser(ctx context.Context, userID, username string) context.Context {
	ctx = context.WithValue(ctx, userIDKey, userID)
	return context.WithValue(ctx, usernameKey, username)
}

// GetUserID 从 context 中获取用户 ID
func GetUserID(ctx context.Context) (string, bool) {
	v := ctx.Value(userIDKey)
	if v == nil {
		return "", false
	}
	id, ok := v.(string)
	return id, ok
}

// GetUsername 从 context 中获取用户名
func GetUsername(ctx context.Context) (string, bool) {
	v := ctx.Value(usernameKey)
	if v == nil {
		return "", false
	}
	name, ok := v.(string)
	return name, ok
}
