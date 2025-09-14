package controller

import (
	"crm_lite/internal/common"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/domains/identity"
	"crm_lite/internal/domains/identity/impl"
	"crm_lite/internal/dto"
	"crm_lite/internal/service"
	"crm_lite/pkg/resp"
	"errors"

	"github.com/gin-gonic/gin"
)

type RoleController struct {
	roleService     *service.RoleService
	identityService identity.Service
}

func NewRoleController(resManager *resource.Manager) *RoleController {
	// 获取必要的资源
	db, err := resource.Get[*resource.DBResource](resManager, resource.DBServiceKey)
	if err != nil {
		panic("Failed to get database resource for RoleController: " + err.Error())
	}
	casbinRes, err := resource.Get[*resource.CasbinResource](resManager, resource.CasbinServiceKey)
	if err != nil {
		panic("Failed to get casbin resource for RoleController: " + err.Error())
	}

	// 创建Identity域服务
	identityService := impl.NewIdentityService(db.DB, casbinRes.GetEnforcer())

	return &RoleController{
		roleService:     service.NewRoleService(resManager),
		identityService: identityService,
	}
}

// CreateRole godoc
// @Summary      创建角色
// @Description  创建一个新的用户角色
// @Tags         Roles
// @Accept       json
// @Produce      json
// @Param        role  body      dto.RoleCreateRequest  true  "角色信息"
// @Success      201   {object}  resp.Response{data=dto.RoleResponse}
// @Failure      400   {object}  resp.Response
// @Security     ApiKeyAuth
// @Router       /roles [post]
func (rc *RoleController) CreateRole(c *gin.Context) {
	var req dto.RoleCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}

	// 转换为Identity域请求
	createReq := identity.CreateRoleRequest{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		IsActive:    true, // 默认激活
	}

	role, err := rc.identityService.CreateRole(c.Request.Context(), createReq)
	if err != nil {
		if errors.Is(err, common.NewBusinessError(common.ErrCodeResourceExists, "")) {
			resp.Error(c, resp.CodeConflict, "role name already exists")
			return
		}
		resp.SystemError(c, err)
		return
	}

	// 转换为DTO格式
	roleResponse := &dto.RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		DisplayName: role.DisplayName,
		Description: role.Description,
		IsActive:    role.IsActive,
	}

	resp.SuccessWithCode(c, resp.CodeCreated, roleResponse)
}

// ListRoles godoc
// @Summary      获取角色列表
// @Description  获取所有可用的用户角色
// @Tags         Roles
// @Accept       json
// @Produce      json
// @Success      200  {object}  resp.Response{data=[]dto.RoleResponse}
// @Security     ApiKeyAuth
// @Router       /roles [get]
func (rc *RoleController) ListRoles(c *gin.Context) {
	roles, err := rc.identityService.ListRoles(c.Request.Context())
	if err != nil {
		resp.SystemError(c, err)
		return
	}

	// 转换为DTO格式
	roleResponses := make([]*dto.RoleResponse, len(roles))
	for i, role := range roles {
		roleResponses[i] = &dto.RoleResponse{
			ID:          role.ID,
			Name:        role.Name,
			DisplayName: role.DisplayName,
			Description: role.Description,
			IsActive:    role.IsActive,
		}
	}

	resp.Success(c, roleResponses)
}

// GetRoleByID godoc
// @Summary      获取单个角色详情
// @Description  根据角色ID获取详细信息
// @Tags         Roles
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Role ID"
// @Success      200  {object}  resp.Response{data=dto.RoleResponse}
// @Failure      404  {object}  resp.Response
// @Security     ApiKeyAuth
// @Router       /roles/{id} [get]
func (rc *RoleController) GetRoleByID(c *gin.Context) {
	id := c.Param("id")
	role, err := rc.identityService.GetRole(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, common.NewBusinessError(common.ErrCodeResourceNotFound, "")) {
			resp.Error(c, resp.CodeNotFound, "role not found")
			return
		}
		resp.SystemError(c, err)
		return
	}

	// 转换为DTO格式
	roleResponse := &dto.RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		DisplayName: role.DisplayName,
		Description: role.Description,
		IsActive:    role.IsActive,
	}

	resp.Success(c, roleResponse)
}

// UpdateRole godoc
// @Summary      更新角色
// @Description  更新一个已存在角色的信息
// @Tags         Roles
// @Accept       json
// @Produce      json
// @Param        id    path      string                 true  "Role ID"
// @Param        role  body      dto.RoleUpdateRequest  true  "要更新的角色信息"
// @Success      200   {object}  resp.Response{data=dto.RoleResponse}
// @Failure      400   {object}  resp.Response
// @Failure      404   {object}  resp.Response
// @Security     ApiKeyAuth
// @Router       /roles/{id} [put]
func (rc *RoleController) UpdateRole(c *gin.Context) {
	id := c.Param("id")
	var req dto.RoleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}

	// 转换为Identity域请求
	updateReq := identity.UpdateRoleRequest{
		DisplayName: &req.DisplayName,
		Description: &req.Description,
		IsActive:    req.IsActive,
	}

	err := rc.identityService.UpdateRole(c.Request.Context(), id, updateReq)
	if err != nil {
		if errors.Is(err, common.NewBusinessError(common.ErrCodeResourceNotFound, "")) {
			resp.Error(c, resp.CodeNotFound, "role not found")
			return
		}
		resp.SystemError(c, err)
		return
	}

	// 获取更新后的角色信息
	role, err := rc.identityService.GetRole(c.Request.Context(), id)
	if err != nil {
		resp.SystemError(c, err)
		return
	}

	// 转换为DTO格式
	roleResponse := &dto.RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		DisplayName: role.DisplayName,
		Description: role.Description,
		IsActive:    role.IsActive,
	}

	resp.Success(c, roleResponse)
}

// DeleteRole godoc
// @Summary      删除角色
// @Description  根据ID删除一个角色
// @Tags         Roles
// @Accept       json
// @Produce      json
// @Param        id  path      string  true  "Role ID"
// @Success      204  {object}  nil
// @Failure      404  {object}  resp.Response
// @Security     ApiKeyAuth
// @Router       /roles/{id} [delete]
func (rc *RoleController) DeleteRole(c *gin.Context) {
	id := c.Param("id")
	err := rc.identityService.DeleteRole(c.Request.Context(), id)
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
