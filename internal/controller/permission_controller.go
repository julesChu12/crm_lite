package controller

import (
	"crm_lite/internal/core/resource"
	"crm_lite/internal/domains/identity"
	"crm_lite/internal/domains/identity/impl"
	"crm_lite/internal/dto"
	"crm_lite/internal/service"
	"crm_lite/pkg/resp"

	"github.com/gin-gonic/gin"
)

type PermissionController struct {
	permService     *service.PermissionService
	identityService identity.Service
}

func NewPermissionController(rm *resource.Manager) *PermissionController {
	// 获取必要的资源
	db, err := resource.Get[*resource.DBResource](rm, resource.DBServiceKey)
	if err != nil {
		panic("Failed to get database resource for PermissionController: " + err.Error())
	}
	casbinRes, err := resource.Get[*resource.CasbinResource](rm, resource.CasbinServiceKey)
	if err != nil {
		panic("Failed to get casbin resource for PermissionController: " + err.Error())
	}

	// 创建Identity域服务
	identityService := impl.NewIdentityService(db.DB, casbinRes.GetEnforcer())

	// 创建旧的PermissionService（保持兼容性）
	ps, err := service.NewPermissionService(rm)
	if err != nil {
		panic("failed to initialize permission service: " + err.Error())
	}

	return &PermissionController{
		permService:     ps,
		identityService: identityService,
	}
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

	// 转换为Identity域请求
	permReq := identity.PermissionRequest{
		Role:   req.Role,
		Path:   req.Path,
		Method: req.Method,
	}

	if err := pc.identityService.AddPermission(c.Request.Context(), permReq); err != nil {
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

	// 转换为Identity域请求
	permReq := identity.PermissionRequest{
		Role:   req.Role,
		Path:   req.Path,
		Method: req.Method,
	}

	if err := pc.identityService.RemovePermission(c.Request.Context(), permReq); err != nil {
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
	permissions, err := pc.identityService.ListRolePermissions(c.Request.Context(), role)
	if err != nil {
		resp.Error(c, resp.CodeInternalError, "failed to list permissions")
		return
	}

	// 转换为简单格式
	permissionResponses := make([][]string, len(permissions))
	for i, perm := range permissions {
		permissionResponses[i] = []string{perm.Role, perm.Path, perm.Method, perm.Action}
	}

	resp.Success(c, permissionResponses)
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

	if err := pc.identityService.AssignRoleToUser(c.Request.Context(), req.UserID, []string{req.Role}); err != nil {
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

	if err := pc.identityService.RemoveRoleFromUser(c.Request.Context(), req.UserID, req.Role); err != nil {
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
	roles, err := pc.identityService.GetRolesForUser(c.Request.Context(), userID)
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
	users, err := pc.identityService.GetUsersForRole(c.Request.Context(), role)
	if err != nil {
		resp.Error(c, resp.CodeInternalError, "failed to get users for role")
		return
	}
	resp.Success(c, users)
}
