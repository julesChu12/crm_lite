package controller

import (
	"crm_lite/internal/dto"
	"crm_lite/internal/service"
	"crm_lite/pkg/resp"
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserController struct {
	userService *service.UserService
}

func NewUserController(db *gorm.DB) *UserController {
	return &UserController{
		userService: service.NewUserService(db),
	}
}

// GetUser godoc
// @Summary      获取单个用户信息
// @Description  根据 UUID 获取用户的详细信息
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        uuid   path      string  true  "User UUID"
// @Success      200  {object}  resp.Response{data=dto.UserResponse}
// @Failure      400  {object}  resp.Response
// @Failure      404  {object}  resp.Response
// @Failure      500  {object}  resp.Response
// @Router       /api/v1/users/{uuid} [get]
func (uc *UserController) GetUser(c *gin.Context) {
	// 1. 绑定并验证 URI 参数
	var req dto.GetUserRequest
	if err := c.ShouldBindUri(&req); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}

	// 2. 调用服务层
	user, err := uc.userService.GetUserByUUID(c.Request.Context(), req.UUID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			resp.Error(c, resp.CodeNotFound, "user not found")
		} else {
			resp.Error(c, resp.CodeInternalError, "failed to get user details")
		}
		return
	}

	// 3. 成功返回
	resp.Success(c, user)
}
