package middleware

// Gin 上下文键常量，避免魔法字符串散落各处。
const (
	ContextKeyUserID   = "user_id"
	ContextKeyUsername = "username"
	ContextKeyRoles    = "roles"
	ContextKeyClaims   = "claims"
)
