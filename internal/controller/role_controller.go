package controller

import (
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dto"
	"crm_lite/internal/service"
	"crm_lite/pkg/resp"
	"errors"

	"github.com/gin-gonic/gin"
)

type RoleController struct {
	roleService *service.RoleService
}

func NewRoleController(resManager *resource.Manager) *RoleController {
	return &RoleController{roleService: service.NewRoleService(resManager)}
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
	role, err := rc.roleService.CreateRole(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrRoleNameAlreadyExists) {
			resp.Error(c, resp.CodeConflict, "role name already exists")
			return
		}
		resp.SystemError(c, err)
		return
	}
	resp.SuccessWithCode(c, resp.CodeCreated, role)
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
	roles, err := rc.roleService.ListRoles(c.Request.Context())
	if err != nil {
		resp.SystemError(c, err)
		return
	}
	resp.Success(c, roles)
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
	role, err := rc.roleService.GetRoleByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrRoleNotFound) {
			resp.Error(c, resp.CodeNotFound, "role not found")
			return
		}
		resp.SystemError(c, err)
		return
	}
	resp.Success(c, role)
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
	role, err := rc.roleService.UpdateRole(c.Request.Context(), id, &req)
	if err != nil {
		if errors.Is(err, service.ErrRoleNotFound) {
			resp.Error(c, resp.CodeNotFound, "role not found")
			return
		}
		if errors.Is(err, service.ErrRoleNameAlreadyExists) {
			resp.Error(c, resp.CodeConflict, "role name already exists")
			return
		}
		resp.SystemError(c, err)
		return
	}
	resp.Success(c, role)
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
	if err := rc.roleService.DeleteRole(c.Request.Context(), id); err != nil {
		if errors.Is(err, service.ErrRoleNotFound) {
			resp.SuccessWithCode(c, resp.CodeNoContent, nil)
			return
		}
		resp.SystemError(c, err)
		return
	}
	resp.SuccessWithCode(c, resp.CodeNoContent, nil)
}
