package dto

// LoginRequest 定义了用户登录时请求体的数据结构
type LoginRequest struct {
	Username string `json:"username" binding:"required"` // 用户名，必须提供
	Password string `json:"password" binding:"required"` // 密码，必须提供
}

// LoginResponse 定义了用户成功登录后返回的数据结构
type LoginResponse struct {
	AccessToken  string `json:"access_token"`  // 访问令牌
	RefreshToken string `json:"refresh_token"` // 刷新令牌
	TokenType    string `json:"token_type"`    // 令牌类型, 通常是 "Bearer"
	ExpiresIn    int64  `json:"expires_in"`    // 访问令牌的有效期（秒）
}
