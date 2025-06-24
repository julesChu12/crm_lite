package controller

import (
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dto"
	"crm_lite/internal/service"
	"crm_lite/pkg/resp"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService *service.AuthService
}

func NewAuthController(resManager *resource.Manager) *AuthController {
	// 从资源管理器获取数据库连接
	dbResource, err := resource.Get[*resource.DBResource](resManager, resource.DBServiceKey)
	if err != nil {
		panic("Failed to get database resource: " + err.Error())
	}

	return &AuthController{
		authService: service.NewAuthService(dbResource.DB),
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

	response, err := c.authService.Login(ctx.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) || errors.Is(err, service.ErrInvalidPassword) {
			fmt.Println(err.Error())
			resp.Error(ctx, resp.CodeUnauthorized, "Invalid username or password")
			return
		}

		resp.SystemError(ctx, err)
		return
	}

	resp.Success(ctx, response)
}

// Register 用户注册
func (c *AuthController) Register(ctx *gin.Context) {
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
	userID, _ := c.Get("user_id")
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
	userID, _ := c.Get("user_id")
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

// --- 以下为待实现功能 ---

func (ac *AuthController) Logout(c *gin.Context) {
	resp.Success(c, "logout success (no-op on server)")
}

func (ac *AuthController) RefreshToken(c *gin.Context) {
	resp.Error(c, http.StatusNotImplemented, "not implemented yet")
}

func (ac *AuthController) ForgotPassword(c *gin.Context) {
	resp.Error(c, http.StatusNotImplemented, "not implemented yet")
}

func (ac *AuthController) ResetPassword(c *gin.Context) {
	resp.Error(c, http.StatusNotImplemented, "not implemented yet")
}
