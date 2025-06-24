package controller

import (
	"crm_lite/internal/dto"
	"crm_lite/internal/service"
	"crm_lite/pkg/resp"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RoleController struct {
	roleService *service.RoleService
}

func NewRoleController(db *gorm.DB) *RoleController {
	return &RoleController{roleService: service.NewRoleService(db)}
}

func (rc *RoleController) CreateRole(c *gin.Context) {
	var req dto.RoleCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}
	role, err := rc.roleService.CreateRole(c.Request.Context(), &req)
	if err != nil {
		resp.Error(c, resp.CodeInternalError, "failed to create role")
		return
	}
	resp.Success(c, role)
}

func (rc *RoleController) ListRoles(c *gin.Context) {
	roles, err := rc.roleService.ListRoles(c.Request.Context())
	if err != nil {
		resp.Error(c, resp.CodeInternalError, "failed to list roles")
		return
	}
	resp.Success(c, roles)
}

func (rc *RoleController) UpdateRole(c *gin.Context) {
	id := c.Param("id")
	var req dto.RoleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}
	if err := rc.roleService.UpdateRole(c.Request.Context(), id, &req); err != nil {
		resp.Error(c, resp.CodeInternalError, "failed to update role")
		return
	}
	resp.Success(c, nil)
}

func (rc *RoleController) DeleteRole(c *gin.Context) {
	id := c.Param("id")
	if err := rc.roleService.DeleteRole(c.Request.Context(), id); err != nil {
		resp.Error(c, resp.CodeInternalError, "failed to delete role")
		return
	}
	resp.Success(c, nil)
}
