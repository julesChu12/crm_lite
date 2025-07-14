package controller

import (
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dto"
	"crm_lite/internal/service"
	"crm_lite/pkg/resp"
	"errors"

	"github.com/gin-gonic/gin"
)

// UserController 负责处理用户管理相关的HTTP请求
type UserController struct {
	userService *service.UserService
}

// NewUserController 创建一个新的 UserController
func NewUserController(resManager *resource.Manager) *UserController {
	return &UserController{
		userService: service.NewUserService(resManager),
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

	user, err := uc.userService.CreateUserByAdmin(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			resp.Error(c, resp.CodeConflict, "username already exists")
			return
		}
		if errors.Is(err, service.ErrEmailAlreadyExists) {
			resp.Error(c, resp.CodeConflict, "email already exists")
			return
		}
		if errors.Is(err, service.ErrPhoneAlreadyExists) {
			resp.Error(c, resp.CodeConflict, "phone number already exists")
			return
		}
		resp.SystemError(c, err)
		return
	}

	resp.SuccessWithCode(c, resp.CodeCreated, user)
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

	result, err := uc.userService.ListUsers(c.Request.Context(), &req)
	if err != nil {
		resp.SystemError(c, err)
		return
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

	// 复用 UserListRequest DTO 来调用现有的服务层方法
	serviceReq := &dto.UserListRequest{
		UUIDs: req.UUIDs,
	}

	result, err := uc.userService.ListUsers(c.Request.Context(), serviceReq)
	if err != nil {
		resp.SystemError(c, err)
		return
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
	user, err := uc.userService.GetUserByUUID(c.Request.Context(), uuid)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			resp.Error(c, resp.CodeNotFound, "user not found")
			return
		}
		resp.SystemError(c, err)
		return
	}
	resp.Success(c, user)
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

	user, err := uc.userService.UpdateUserByAdmin(c.Request.Context(), uuid, &req)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			resp.Error(c, resp.CodeNotFound, "user not found")
			return
		}
		if errors.Is(err, service.ErrEmailAlreadyExists) {
			resp.Error(c, resp.CodeConflict, "email already exists")
			return
		}
		if errors.Is(err, service.ErrPhoneAlreadyExists) {
			resp.Error(c, resp.CodeConflict, "phone number already exists")
			return
		}
		resp.SystemError(c, err)
		return
	}
	resp.Success(c, user)
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
	err := uc.userService.DeleteUser(c.Request.Context(), uuid)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			// 在DELETE操作中，如果资源不存在，也可以认为是成功的（幂等性）
			resp.SuccessWithCode(c, resp.CodeNoContent, nil)
			return
		}
		resp.SystemError(c, err)
		return
	}
	resp.SuccessWithCode(c, resp.CodeNoContent, nil)
}
