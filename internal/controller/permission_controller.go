package controller

import (
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dto"
	"crm_lite/internal/service"
	"crm_lite/pkg/resp"

	"github.com/gin-gonic/gin"
)

type PermissionController struct {
	permService *service.PermissionService
}

func NewPermissionController(rm *resource.Manager) *PermissionController {
	ps, err := service.NewPermissionService(rm)
	if err != nil {
		// 在启动时就应该 panic，因为这是一个关键服务的依赖问题
		panic("failed to initialize permission service: " + err.Error())
	}
	return &PermissionController{permService: ps}
}

// AddPermission godoc
// @Summary 添加权限策略
// @Description 添加一条权限策略 (p, role, path, method)
// @Tags Permissions
// @Accept json
// @Produce json
// @Param permission body dto.PermissionRequest true "权限信息"
// @Success 200 {object} resp.Response "成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Security ApiKeyAuth
// @Router /permissions [post]
func (pc *PermissionController) AddPermission(c *gin.Context) {
	var req dto.PermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}
	if err := pc.permService.AddPermission(c.Request.Context(), &req); err != nil {
		resp.Error(c, resp.CodeInternalError, "failed to add permission")
		return
	}
	resp.Success(c, nil)
}

// RemovePermission godoc
// @Summary 移除权限策略
// @Description 移除一条权限策略
// @Tags Permissions
// @Accept json
// @Produce json
// @Param permission body dto.PermissionRequest true "权限信息"
// @Success 200 {object} resp.Response "成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Security ApiKeyAuth
// @Router /permissions [delete]
func (pc *PermissionController) RemovePermission(c *gin.Context) {
	var req dto.PermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}
	if err := pc.permService.RemovePermission(c.Request.Context(), &req); err != nil {
		resp.Error(c, resp.CodeInternalError, "failed to remove permission")
		return
	}
	resp.Success(c, nil)
}

// ListPermissionsByRole godoc
// @Summary 获取角色的所有权限
// @Description 根据角色名获取其拥有的所有权限策略
// @Tags Permissions
// @Produce json
// @Param role path string true "角色名称"
// @Success 200 {object} resp.Response{data=[][]string} "成功"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Security ApiKeyAuth
// @Router /permissions/{role} [get]
func (pc *PermissionController) ListPermissionsByRole(c *gin.Context) {
	role := c.Param("role")
	permissions, err := pc.permService.ListPermissionsByRole(c.Request.Context(), role)
	if err != nil {
		resp.Error(c, resp.CodeInternalError, "failed to list permissions")
		return
	}
	resp.Success(c, permissions)
}

// AssignRoleToUser godoc
// @Summary 给用户分配角色
// @Description 将指定用户添加到一个角色中 (g, user, role)
// @Tags Permissions
// @Accept json
// @Produce json
// @Param user_role body dto.UserRoleRequest true "用户和角色信息"
// @Success 200 {object} resp.Response "成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Security ApiKeyAuth
// @Router /user-roles/assign [post]
func (pc *PermissionController) AssignRoleToUser(c *gin.Context) {
	var req dto.UserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}
	if err := pc.permService.AssignRoleToUser(c.Request.Context(), &req); err != nil {
		resp.Error(c, resp.CodeInternalError, "failed to assign role to user")
		return
	}
	resp.Success(c, nil)
}

// RemoveRoleFromUser godoc
// @Summary 移除用户的角色
// @Description 将用户从指定角色中移除
// @Tags Permissions
// @Accept json
// @Produce json
// @Param user_role body dto.UserRoleRequest true "用户和角色信息"
// @Success 200 {object} resp.Response "成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Security ApiKeyAuth
// @Router /user-roles/remove [post]
func (pc *PermissionController) RemoveRoleFromUser(c *gin.Context) {
	var req dto.UserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}
	if err := pc.permService.RemoveRoleFromUser(c.Request.Context(), &req); err != nil {
		resp.Error(c, resp.CodeInternalError, "failed to remove role from user")
		return
	}
	resp.Success(c, nil)
}

// GetRolesForUser godoc
// @Summary 获取用户的所有角色
// @Description 根据用户ID获取其拥有的所有角色列表
// @Tags Permissions
// @Produce json
// @Param user_id path string true "用户ID"
// @Success 200 {object} resp.Response{data=[]string} "成功"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Security ApiKeyAuth
// @Router /user-roles/roles/{user_id} [get]
func (pc *PermissionController) GetRolesForUser(c *gin.Context) {
	userID := c.Param("user_id")
	roles, err := pc.permService.GetRolesForUser(c.Request.Context(), userID)
	if err != nil {
		resp.Error(c, resp.CodeInternalError, "failed to get roles for user")
		return
	}
	resp.Success(c, roles)
}

// GetUsersForRole godoc
// @Summary 获取角色的所有用户
// @Description 根据角色名称获取拥有该角色的所有用户列表
// @Tags Permissions
// @Produce json
// @Param role path string true "角色名称"
// @Success 200 {object} resp.Response{data=[]string} "成功"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Security ApiKeyAuth
// @Router /user-roles/users/{role} [get]
func (pc *PermissionController) GetUsersForRole(c *gin.Context) {
	role := c.Param("role")
	users, err := pc.permService.GetUsersForRole(c.Request.Context(), role)
	if err != nil {
		resp.Error(c, resp.CodeInternalError, "failed to get users for role")
		return
	}
	resp.Success(c, users)
}
