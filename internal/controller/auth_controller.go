package controller

import (
	"crm_lite/internal/captcha"
	"crm_lite/internal/core/config"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/domains/identity"
	"crm_lite/internal/domains/identity/impl"
	"crm_lite/internal/dto"
	"crm_lite/internal/middleware"
	"crm_lite/pkg/resp"
	"crm_lite/pkg/utils"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	identityService identity.Service
	cache           *resource.CacheResource
}

func NewAuthController(resManager *resource.Manager) *AuthController {
	// 1. 获取必要的资源
	db, err := resource.Get[*resource.DBResource](resManager, resource.DBServiceKey)
	if err != nil {
		panic("Failed to get database resource for AuthController: " + err.Error())
	}
	cache, err := resource.Get[*resource.CacheResource](resManager, resource.CacheServiceKey)
	if err != nil {
		panic("Failed to get cache resource for AuthController: " + err.Error())
	}
	casbinRes, err := resource.Get[*resource.CasbinResource](resManager, resource.CasbinServiceKey)
	if err != nil {
		panic("Failed to get casbin resource for AuthController: " + err.Error())
	}
	// 2. 创建Identity域服务
	identityService := impl.NewIdentityService(db.DB, casbinRes.GetEnforcer(), resManager)

	return &AuthController{
		identityService: identityService,
		cache:           cache,
	}
}

// Login godoc
// @Summary      用户登录
// @Description  使用用户名和密码进行登录
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        credentials  body      dto.LoginRequest   true  "登录凭证"
// @Success      200          {object}  resp.Response{data=dto.LoginResponse}
// @Failure      400          {object}  resp.Response
// @Failure      401          {object}  resp.Response
// @Router       /auth/login [post]
func (c *AuthController) Login(ctx *gin.Context) {
	var req dto.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		resp.Error(ctx, resp.CodeInvalidParam, err.Error())
		return
	}

	// 1. 仅在需要时校验 Turnstile Token（首次不要求）
	if need, ok := ctx.Get(middleware.CtxKeyCaptchaRequired); ok && need.(bool) {
		ok2, err := captcha.VerifyTurnstile(ctx.Request.Context(), req.CaptchaToken, ctx.ClientIP())
		if err != nil {
			resp.SystemError(ctx, err)
			return
		}
		if !ok2 {
			resp.Error(ctx, resp.CodeInvalidParam, "captcha verification failed")
			return
		}
	}

	// 使用Identity域服务进行登录
	loginReq := &identity.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	}

	response, err := c.identityService.Login(ctx.Request.Context(), loginReq)
	if err != nil {
		// 登录失败后，设置需要验证码标记（短期，存入 Redis，以 IP 维度）
		if c.cache != nil && c.cache.Client != nil {
			key := fmt.Sprintf("risk:captcha:ip:%s", ctx.ClientIP())
			_ = c.cache.Client.Set(ctx.Request.Context(), key, "1", 10*time.Minute).Err()
		}
		resp.Error(ctx, resp.CodeUnauthorized, "Invalid username or password")
		return
	}

	// 将 refreshToken 写入安全 Cookie
	opts := config.GetInstance()
	secure := opts.Server.Mode == config.ReleaseMode
	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     "refreshToken",
		Value:    response.RefreshToken,
		Path:     "/api/v1/auth",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(opts.Auth.JWTOptions.RefreshTokenExpire.Seconds()),
	})

	// 构造响应体，不包含 refresh_token
	resp.Success(ctx, gin.H{
		"access_token": response.AccessToken,
		"token_type":   "Bearer",
		"expires_in":   response.ExpiresIn,
	})
}

// Register 用户注册
// @Summary      用户注册
// @Description  使用用户名和密码进行注册
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        credentials  body      dto.RegisterRequest   true  "注册凭证"
// @Success      200          {object}  resp.Response
func (c *AuthController) Register(ctx *gin.Context) {
	// todo: 防止脚本批量撞库，在注册时，需要限制注册频率，比如1分钟内只能注册10次
	var req dto.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		resp.Error(ctx, resp.CodeInvalidParam, "Invalid request payload")
		return
	}

	// 使用Identity域服务进行注册
	registerReq := &identity.RegisterRequest{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
		RealName: req.RealName,
	}

	if err := c.identityService.Register(ctx.Request.Context(), registerReq); err != nil {
		resp.Error(ctx, resp.CodeInvalidParam, "User already exists")
		return
	}

	resp.Success(ctx, gin.H{"message": "register success"})
}

// UpdateProfile 修改用户信息（仅示例修改 RealName）
func (c *AuthController) UpdateProfile(ctx *gin.Context) {
	var req dto.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		resp.Error(ctx, resp.CodeInvalidParam, "Invalid request payload")
		return
	}

	// 从 JWT 中解析当前用户 ID，这里仅示例，实际应有鉴权中间件
	claimsCtx := ctx.Request.Context()

	// 使用Identity域服务更新资料
	updateReq := &identity.UpdateProfileRequest{
		RealName: req.RealName,
		Avatar:   req.Avatar,
	}

	// 从JWT中获取用户ID（简化实现）
	userID := int64(1) // 实际应该从JWT claims中获取

	if err := c.identityService.UpdateProfile(claimsCtx, userID, updateReq); err != nil {
		resp.Error(ctx, resp.CodeNotFound, "User not found")
		return
	}

	resp.Success(ctx, gin.H{"message": "update success"})
}

// GetProfile godoc
// @Summary      获取个人资料
// @Description  获取当前登录用户的详细信息
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Success      200  {object}  resp.Response{data=dto.UserResponse}
// @Failure      401  {object}  resp.Response
// @Failure      404  {object}  resp.Response
// @Security     ApiKeyAuth
// @Router       /auth/profile [get]
func (ac *AuthController) GetProfile(c *gin.Context) {
	userID, _ := c.Get(middleware.ContextKeyUserID)
	user, err := ac.identityService.GetUserByUUID(c.Request.Context(), userID.(string))
	if err != nil {
		resp.Error(c, resp.CodeNotFound, "user profile not found")
		return
	}

	// 转换为DTO格式
	userResponse := &dto.UserResponse{
		UUID:      user.UUID,
		Username:  user.Username,
		Email:     user.Email,
		RealName:  user.Name,
		Phone:     user.Phone,
		IsActive:  user.Status == "active",
		CreatedAt: utils.FormatTime(time.Unix(user.CreatedAt, 0)),
	}

	resp.Success(c, userResponse)
}

// ChangePassword 修改密码
func (ac *AuthController) ChangePassword(c *gin.Context) {
	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}
	userID, _ := c.Get(middleware.ContextKeyUserID)
	err := ac.identityService.ChangePassword(c.Request.Context(), userID.(string), req.OldPassword, req.NewPassword)
	if err != nil {
		resp.Error(c, resp.CodeInvalidParam, "invalid old password")
		return
	}
	resp.Success(c, nil)
}

// Logout godoc
// @Summary      用户登出
// @Description  将当前用户的JWT加入黑名单以实现登出
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Success      200  {object}  resp.Response
// @Failure      401  {object}  resp.Response
// @Security     ApiKeyAuth
// @Router       /auth/logout [post]
func (ac *AuthController) Logout(c *gin.Context) {
	refreshToken, errCookie := c.Cookie("refreshToken")
	if errCookie != nil || refreshToken == "" {
		resp.Error(c, resp.CodeUnauthorized, "refresh token not found in cookie")
		return
	}

	// 解析 refreshToken 以获取 claims，进而可以将其 jti 加入黑名单
	opts := config.GetInstance()
	_, err := utils.ParseToken(refreshToken, opts.Auth.JWTOptions)
	if err != nil {
		resp.Error(c, resp.CodeUnauthorized, "invalid or expired refresh token")
		return
	}

	// 调用Identity服务进行登出
	if err := ac.identityService.Logout(c.Request.Context(), refreshToken); err != nil {
		resp.SystemError(c, err)
		return
	}

	// 清除 Cookie
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refreshToken",
		Value:    "",
		Path:     "/api/v1/auth",
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})

	resp.Success(c, gin.H{"message": "logout success"})
}

// RefreshToken godoc
// @Summary      刷新令牌
// @Description  使用有效的刷新令牌获取新的访问令牌
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        refresh_token  body      dto.RefreshTokenRequest   true  "刷新令牌"
// @Success      200            {object}  resp.Response{data=dto.LoginResponse}
// @Failure      400            {object}  resp.Response
// @Failure      401            {object}  resp.Response
// @Router       /auth/refresh [post]
func (ac *AuthController) RefreshToken(c *gin.Context) {
	refreshToken, errCookie := c.Cookie("refreshToken")
	if errCookie != nil || refreshToken == "" {
		resp.Error(c, resp.CodeUnauthorized, "refresh token not found in cookie")
		return
	}

	response, err := ac.identityService.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		resp.Error(c, resp.CodeUnauthorized, "invalid or expired refresh token")
		return
	}
	// 更新新的 refreshToken Cookie
	opts := config.GetInstance()
	secure := opts.Server.Mode == config.ReleaseMode
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refreshToken",
		Value:    response.RefreshToken,
		Path:     "/api/v1/auth",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(opts.Auth.JWTOptions.RefreshTokenExpire.Seconds()),
	})

	resp.Success(c, gin.H{
		"access_token": response.Token,
		"token_type":   "Bearer",
		"expires_in":   response.ExpiresAt,
	})
}

// ForgotPassword godoc
// @Summary      忘记密码
// @Description  用户提交邮箱，系统发送重置密码链接/令牌
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        email  body      dto.ForgotPasswordRequest  true  "用户邮箱"
// @Success      200    {object}  resp.Response
// @Failure      400    {object}  resp.Response
// @Failure      404    {object}  resp.Response
// @Router       /auth/forgot-password [post]
func (ac *AuthController) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}
	err := ac.identityService.ResetPassword(c.Request.Context(), req.Email)
	if err != nil {
		// 为安全起见，不明确提示用户是否存在，统一返回成功
		resp.Success(c, gin.H{"message": "if the user exists, a password reset link has been sent to the email"})
		return
	}
	resp.Success(c, gin.H{"message": "if the user exists, a password reset link has been sent to the email"})
}

// ResetPassword godoc
// @Summary      重置密码
// @Description  使用令牌和新密码来重置用户密码
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        credentials  body      dto.ResetPasswordRequest   true  "重置密码凭证"
// @Success      200          {object}  resp.Response
// @Failure      400          {object}  resp.Response
// @Failure      401          {object}  resp.Response
// @Router       /auth/reset-password [post]
func (ac *AuthController) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}
	err := ac.identityService.ConfirmPasswordReset(c.Request.Context(), req.Token, req.NewPassword)
	if err != nil {
		resp.Error(c, resp.CodeUnauthorized, "invalid or expired reset token")
		return
	}
	resp.Success(c, gin.H{"message": "password has been reset successfully"})
}
