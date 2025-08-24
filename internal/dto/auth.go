package dto

// LoginRequest 定义了用户登录时请求体的数据结构
type LoginRequest struct {
	Username     string `json:"username" binding:"required"`      // 用户名，必须提供
	Password     string `json:"password" binding:"required"`      // 密码，必须提供
	CaptchaToken string `json:"captcha_token" binding:"required"` // Turnstile Token
}

// LoginResponse 定义了用户成功登录后返回的数据结构
type LoginResponse struct {
	AccessToken  string `json:"access_token"`  // 访问令牌
	RefreshToken string `json:"refresh_token"` // 刷新令牌
	TokenType    string `json:"token_type"`    // 令牌类型, 通常是 "Bearer"
	ExpiresIn    int    `json:"expires_in"`    // access_token 的有效期（秒）
}

// RefreshTokenRequest 刷新令牌请求体
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ChangePasswordRequest 修改密码请求体
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// ForgotPasswordRequest 忘记密码请求体
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest 重置密码请求体
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}
