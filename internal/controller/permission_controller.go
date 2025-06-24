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

func (pc *PermissionController) ListPermissionsByRole(c *gin.Context) {
	role := c.Param("role")
	permissions, err := pc.permService.ListPermissionsByRole(c.Request.Context(), role)
	if err != nil {
		resp.Error(c, resp.CodeInternalError, "failed to list permissions")
		return
	}
	resp.Success(c, permissions)
}
