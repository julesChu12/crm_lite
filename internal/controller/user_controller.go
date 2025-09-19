package controller

import (
	"crm_lite/internal/common"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/domains/identity"
	"crm_lite/internal/domains/identity/impl"
	"crm_lite/internal/dto"
	"crm_lite/pkg/resp"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
)

// UserController 负责处理用户管理相关的HTTP请求
// 已完全迁移到 identity 域服务
type UserController struct {
	identityService identity.Service
}

// NewUserController 创建一个新的 UserController
func NewUserController(resManager *resource.Manager) *UserController {
	// 获取必要的资源
	db, err := resource.Get[*resource.DBResource](resManager, resource.DBServiceKey)
	if err != nil {
		panic("Failed to get database resource for UserController: " + err.Error())
	}
	casbinRes, err := resource.Get[*resource.CasbinResource](resManager, resource.CasbinServiceKey)
	if err != nil {
		panic("Failed to get casbin resource for UserController: " + err.Error())
	}

	// 创建Identity域服务
	identityService := impl.NewIdentityService(db.DB, casbinRes.GetEnforcer(), resManager)

	return &UserController{
		identityService: identityService,
	}
}

// CreateUser godoc
// @Summary      管理员创建用户
// @Description  由管理员创建一个新的用户账号并可以指定角色
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        user  body      dto.AdminCreateUserRequest  true  "用户信息"
// @Success      201   {object}  resp.Response{data=dto.UserResponse}
// @Failure      400   {object}  resp.Response
// @Failure      409   {object}  resp.Response
// @Security     ApiKeyAuth
// @Router       /users [post]
func (uc *UserController) CreateUser(c *gin.Context) {
	var req dto.AdminCreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}

	// 转换为Identity域请求
	createReq := identity.CreateUserRequest{
		Username:        req.Username,
		Email:           req.Email,
		Phone:           req.Phone,
		Name:            req.RealName,
		Password:        req.Password,
		ConfirmPassword: req.Password, // 简化实现
		Roles:           []string{},   // 简化实现，不处理角色
	}

	user, err := uc.identityService.CreateUser(c.Request.Context(), createReq)
	if err != nil {
		if errors.Is(err, common.NewBusinessError(common.ErrCodeResourceExists, "")) {
			resp.Error(c, resp.CodeConflict, "username or email already exists")
			return
		}
		resp.SystemError(c, err)
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
		Roles:     []string{}, // 简化实现
		CreatedAt: time.Unix(user.CreatedAt, 0).Format("2006-01-02 15:04:05"),
	}

	resp.SuccessWithCode(c, resp.CodeCreated, userResponse)
}

// GetUserList godoc
// @Summary      获取用户列表
// @Description  分页、筛选、搜索用户列表
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        query  query     dto.UserListRequest  false  "查询参数"
// @Success      200    {object}  resp.Response{data=dto.UserListResponse}
// @Failure      400    {object}  resp.Response
// @Security     ApiKeyAuth
// @Router       /users [get]
func (uc *UserController) GetUserList(c *gin.Context) {
	var req dto.UserListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// 转换状态参数
	status := ""
	if req.IsActive != nil {
		if *req.IsActive {
			status = "active"
		} else {
			status = "inactive"
		}
	}

	users, total, err := uc.identityService.ListUsers(c.Request.Context(), req.Page, req.PageSize, status)
	if err != nil {
		resp.SystemError(c, err)
		return
	}

	// 转换为DTO格式
	userResponses := make([]*dto.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = &dto.UserResponse{
			UUID:      user.UUID,
			Username:  user.Username,
			Email:     user.Email,
			RealName:  user.Name,
			Phone:     user.Phone,
			IsActive:  user.Status == "active",
			Roles:     []string{}, // 简化实现
			CreatedAt: time.Unix(user.CreatedAt, 0).Format("2006-01-02 15:04:05"),
		}
	}

	result := &dto.UserListResponse{
		Total: total,
		Users: userResponses,
	}

	resp.Success(c, result)
}

// BatchGetUsers godoc
// @Summary      批量获取用户
// @Description  通过POST请求体中提供的UUID列表，批量获取用户信息
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        query  body      dto.UserBatchGetRequest  true  "UUID列表"
// @Success      200    {object}  resp.Response{data=dto.UserListResponse}
// @Failure      400    {object}  resp.Response
// @Security     ApiKeyAuth
// @Router       /users/batch-get [post]
func (uc *UserController) BatchGetUsers(c *gin.Context) {
	var req dto.UserBatchGetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}

	users, err := uc.identityService.BatchGetUsers(c.Request.Context(), req.UUIDs)
	if err != nil {
		resp.SystemError(c, err)
		return
	}

	// 转换为DTO格式
	userResponses := make([]*dto.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = &dto.UserResponse{
			UUID:      user.UUID,
			Username:  user.Username,
			Email:     user.Email,
			RealName:  user.Name,
			Phone:     user.Phone,
			IsActive:  user.Status == "active",
			Roles:     []string{}, // 简化实现
			CreatedAt: time.Unix(user.CreatedAt, 0).Format("2006-01-02 15:04:05"),
		}
	}

	result := &dto.UserListResponse{
		Total: int64(len(userResponses)),
		Users: userResponses,
	}

	resp.Success(c, result)
}

// GetUserByID godoc
// @Summary      获取单个用户详情
// @Description  根据用户UUID获取详细信息
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        uuid  path      string  true  "User UUID"
// @Success      200   {object}  resp.Response{data=dto.UserResponse}
// @Failure      404   {object}  resp.Response
// @Security     ApiKeyAuth
// @Router       /users/{uuid} [get]
func (uc *UserController) GetUserByID(c *gin.Context) {
	uuid := c.Param("uuid")
	user, err := uc.identityService.GetUserByUUID(c.Request.Context(), uuid)
	if err != nil {
		resp.Error(c, resp.CodeNotFound, "user not found")
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
		Roles:     []string{}, // 简化实现
		CreatedAt: time.Unix(user.CreatedAt, 0).Format("2006-01-02 15:04:05"),
	}

	resp.Success(c, userResponse)
}

// UpdateUser godoc
// @Summary      管理员更新用户
// @Description  管理员更新用户信息，包括角色等
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        uuid  path      string                      true  "User UUID"
// @Param        user  body      dto.AdminUpdateUserRequest  true  "要更新的用户信息"
// @Success      200   {object}  resp.Response{data=dto.UserResponse}
// @Failure      400   {object}  resp.Response
// @Failure      404   {object}  resp.Response
// @Security     ApiKeyAuth
// @Router       /users/{uuid} [put]
func (uc *UserController) UpdateUser(c *gin.Context) {
	uuid := c.Param("uuid")
	var req dto.AdminUpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}

	// 转换为Identity域请求
	updateReq := identity.UpdateUserRequest{
		Email: &req.Email,
		Phone: &req.Phone,
		Name:  &req.RealName,
	}

	// 转换状态
	if req.IsActive != nil {
		status := "inactive"
		if *req.IsActive {
			status = "active"
		}
		updateReq.Status = &status
	}

	err := uc.identityService.UpdateUser(c.Request.Context(), uuid, updateReq)
	if err != nil {
		if errors.Is(err, common.NewBusinessError(common.ErrCodeResourceNotFound, "")) {
			resp.Error(c, resp.CodeNotFound, "user not found")
			return
		}
		if errors.Is(err, common.NewBusinessError(common.ErrCodeResourceExists, "")) {
			resp.Error(c, resp.CodeConflict, "email or phone already exists")
			return
		}
		resp.SystemError(c, err)
		return
	}

	// 获取更新后的用户信息
	user, err := uc.identityService.GetUserByUUID(c.Request.Context(), uuid)
	if err != nil {
		resp.SystemError(c, err)
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
		Roles:     []string{}, // 简化实现
		CreatedAt: time.Unix(user.CreatedAt, 0).Format("2006-01-02 15:04:05"),
	}

	resp.Success(c, userResponse)
}

// DeleteUser godoc
// @Summary      删除用户
// @Description  管理员删除一个用户
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        uuid  path      string  true  "User UUID"
// @Success      204   {object}  nil
// @Failure      404   {object}  resp.Response
// @Security     ApiKeyAuth
// @Router       /users/{uuid} [delete]
func (uc *UserController) DeleteUser(c *gin.Context) {
	uuid := c.Param("uuid")
	err := uc.identityService.DeleteUser(c.Request.Context(), uuid)
	if err != nil {
		if errors.Is(err, common.NewBusinessError(common.ErrCodeResourceNotFound, "")) {
			// 在DELETE操作中，如果资源不存在，也可以认为是成功的（幂等性）
			resp.SuccessWithCode(c, resp.CodeNoContent, nil)
			return
		}
		resp.SystemError(c, err)
		return
	}
	resp.SuccessWithCode(c, resp.CodeNoContent, nil)
}
