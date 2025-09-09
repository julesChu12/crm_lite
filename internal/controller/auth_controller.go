package controller

import (
	"crm_lite/internal/captcha"
	"crm_lite/internal/core/config"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dto"
	"crm_lite/internal/middleware"
	"crm_lite/internal/service"
	"crm_lite/pkg/resp"
	"crm_lite/pkg/utils"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService *service.AuthService
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
	opts := config.GetInstance()

	// 2. 创建依赖的服务和仓库
	authRepo := service.NewAuthRepo(db.DB)
	authCache := service.NewAuthCache(cache)

	// EmailService 是可选的，如果资源不存在则为 nil
	var emailSvc service.IEmailService
	emailRes, err := resource.Get[*resource.EmailResource](resManager, resource.EmailServiceKey)
	if err == nil && emailRes != nil {
		emailSvc = service.NewEmailService(emailRes.Opts)
	}

	// 3. 注入所有依赖项来创建 AuthService
	authService := service.NewAuthService(authRepo, authCache, emailSvc, casbinRes, opts)

	return &AuthController{
		authService: authService,
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

	response, err := c.authService.Login(ctx.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) || errors.Is(err, service.ErrInvalidPassword) {
			fmt.Println(err.Error())
			// 登录失败后，设置需要验证码标记（短期）
			http.SetCookie(ctx.Writer, &http.Cookie{
				Name:     "need_captcha",
				Value:    "1",
				Path:     "/",
				HttpOnly: true,
				MaxAge:   600, // 10分钟
			})
			resp.Error(ctx, resp.CodeUnauthorized, "Invalid username or password")
			return
		}

		resp.SystemError(ctx, err)
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
		"token_type":   response.TokenType,
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

	if err := c.authService.Register(ctx.Request.Context(), &req); err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			resp.Error(ctx, resp.CodeInvalidParam, "User already exists")
			return
		}
		resp.SystemError(ctx, err)
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

	if err := c.authService.UpdateProfile(claimsCtx, &req); err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			resp.Error(ctx, resp.CodeNotFound, "User not found")
			return
		}
		resp.SystemError(ctx, err)
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
	user, err := ac.authService.GetProfile(c.Request.Context(), userID.(string))
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			resp.Error(c, resp.CodeNotFound, "user profile not found")
			return
		}
		resp.Error(c, resp.CodeInternalError, "failed to get profile")
		return
	}
	resp.Success(c, user)
}

// ChangePassword 修改密码
func (ac *AuthController) ChangePassword(c *gin.Context) {
	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}
	userID, _ := c.Get(middleware.ContextKeyUserID)
	err := ac.authService.ChangePassword(c.Request.Context(), userID.(string), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidPassword) {
			resp.Error(c, resp.CodeInvalidParam, "invalid old password")
		} else {
			resp.Error(c, resp.CodeInternalError, "failed to change password")
		}
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
	claims, err := utils.ParseToken(refreshToken, opts.Auth.JWTOptions)
	if err != nil {
		resp.Error(c, resp.CodeUnauthorized, "invalid or expired refresh token")
		return
	}

	// 调用服务层将 refreshToken 加入黑名单
	if err := ac.authService.Logout(c.Request.Context(), claims); err != nil {
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

	req := dto.RefreshTokenRequest{RefreshToken: refreshToken}
	response, err := ac.authService.RefreshToken(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			resp.Error(c, resp.CodeUnauthorized, "invalid or expired refresh token")
		} else {
			resp.SystemError(c, err)
		}
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
		"access_token": response.AccessToken,
		"token_type":   response.TokenType,
		"expires_in":   response.ExpiresIn,
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
	err := ac.authService.ForgotPassword(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			// 为安全起见，不明确提示用户是否存在，统一返回成功
			resp.Success(c, gin.H{"message": "if the user exists, a password reset link has been sent to the email"})
			return
		}
		resp.SystemError(c, err)
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
	err := ac.authService.ResetPassword(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			resp.Error(c, resp.CodeUnauthorized, "invalid or expired reset token")
		} else {
			resp.SystemError(c, err)
		}
		return
	}
	resp.Success(c, gin.H{"message": "password has been reset successfully"})
}
