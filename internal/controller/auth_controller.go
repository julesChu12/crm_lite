package controller

import (
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dto"
	"crm_lite/internal/service"
	"crm_lite/pkg/resp"
	"errors"

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

// Login @Summary 用户登录
// @Description 使用用户名和密码进行登录
// @Tags Auth
// @Accept json
// @Produce json
// @Param login body dto.LoginRequest true "Login Credentials"
// @Success 200 {object} dto.LoginResponse
// @Failure 400 {object} map[string]string "{"error": "Invalid request"}"
// @Failure 401 {object} map[string]string "{"error": "Invalid credentials"}"
// @Failure 500 {object} map[string]string "{"error": "Internal server error"}"
// @Router /api/v1/auth/login [post]
func (c *AuthController) Login(ctx *gin.Context) {
	var req dto.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		resp.Error(ctx, resp.CodeInvalidParam, "Invalid request payload")
		return
	}

	response, err := c.authService.Login(ctx.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) || errors.Is(err, service.ErrInvalidPassword) {
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
