package service

import (
	"context"
	"crm_lite/internal/core/config"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dao/query"
	"crm_lite/internal/dto"
	"crm_lite/pkg/utils"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	q        *query.Query
	resource *resource.Manager // 添加 resource manager
}

func NewUserService(resManager *resource.Manager) *UserService {
	db, err := resource.Get[*resource.DBResource](resManager, resource.DBServiceKey)
	if err != nil {
		panic("Failed to get database resource for UserService: " + err.Error())
	}
	return &UserService{
		q:        query.Use(db.DB),
		resource: resManager, // 初始化 resource manager
	}
}

// GetUserByUUID 根据 UUID 获取单个用户详细信息
func (s *UserService) GetUserByUUID(ctx context.Context, uuid string) (*dto.UserResponse, error) {
	// 1. 查询用户基本信息
	user, err := s.q.AdminUser.WithContext(ctx).Where(s.q.AdminUser.UUID.Eq(uuid)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("database query error: %w", err)
	}

	// 2. 查询用户角色
	var roles []string
	err = s.q.Role.WithContext(ctx).
		Select(s.q.Role.Name).
		LeftJoin(s.q.AdminUserRole, s.q.AdminUserRole.RoleID.EqCol(s.q.Role.ID)).
		Where(s.q.AdminUserRole.AdminUserID.Eq(user.ID)).
		Scan(&roles)
	if err != nil {
		return nil, fmt.Errorf("failed to query user roles: %w", err)
	}

	// 3. 组装成 DTO 返回
	return &dto.UserResponse{
		UUID:      user.UUID,
		Username:  user.Username,
		Email:     user.Email,
		RealName:  user.RealName,
		Phone:     user.Phone,
		Avatar:    user.Avatar,
		IsActive:  user.IsActive,
		Roles:     roles,
		CreatedAt: utils.FormatTime(user.CreatedAt),
	}, nil
}

// CreateUserByAdmin 管理员创建用户
func (s *UserService) CreateUserByAdmin(ctx context.Context, req *dto.AdminCreateUserRequest) (*dto.UserResponse, error) {
	// 使用事务确保原子性
	var response *dto.UserResponse
	var roles_to_add []string

	txErr := s.q.Transaction(func(tx *query.Query) error {
		// 1. 检查用户是否存在（包括软删除的）
		var existingUser *model.AdminUser
		// 使用 Unscoped() 来查找包括软删除在内的记录
		existingUser, _ = tx.AdminUser.WithContext(ctx).Unscoped().
			Where(tx.AdminUser.Username.Eq(req.Username)).
			Or(tx.AdminUser.Email.Eq(req.Email)).
			Or(tx.AdminUser.Phone.Eq(req.Phone)).
			First()

		var newUser *model.AdminUser

		if existingUser != nil {
			// 如果用户存在
			if existingUser.DeletedAt.Valid {
				// a. 如果是软删除状态，则恢复并更新
				// 哈希新密码
				hashed, err_hash := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
				if err_hash != nil {
					return fmt.Errorf("failed to hash password: %w", err_hash)
				}

				// 准备更新
				updates := map[string]interface{}{
					"password_hash": string(hashed),
					"real_name":     req.RealName,
					"phone":         req.Phone,
					"avatar":        req.Avatar,
					"is_active":     true,
					"deleted_at":    nil, // 恢复用户
				}
				if req.IsActive != nil {
					updates["is_active"] = *req.IsActive
				}
				if _, err_update := tx.AdminUser.WithContext(ctx).Unscoped().Where(tx.AdminUser.ID.Eq(existingUser.ID)).Updates(updates); err_update != nil {
					return err_update
				}
				newUser = existingUser // 后续处理角色时使用
				log.Printf("User %s restored and updated.", newUser.Username)

			} else {
				// b. 如果是正常状态，则返回错误
				if existingUser.Username == req.Username {
					return ErrUserAlreadyExists
				}
				if existingUser.Email == req.Email {
					return ErrEmailAlreadyExists
				}
				if existingUser.Phone == req.Phone {
					return ErrPhoneAlreadyExists
				}
			}
		} else {
			// 2. 如果用户不存在，则创建新用户
			hashed, err_hash := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
			if err_hash != nil {
				return fmt.Errorf("failed to hash password: %w", err_hash)
			}
			newUser = &model.AdminUser{
				UUID:         uuid.New().String(),
				Username:     req.Username,
				Email:        req.Email,
				PasswordHash: string(hashed),
				RealName:     req.RealName,
				Phone:        req.Phone,
				Avatar:       req.Avatar,
			}
			if req.IsActive != nil {
				newUser.IsActive = *req.IsActive
			} else {
				newUser.IsActive = true // 默认激活
			}
			if err_create := tx.AdminUser.WithContext(ctx).Create(newUser); err_create != nil {
				return err_create
			}
		}

		// 3. 处理角色关联 (无论是创建还是恢复)
		// a. 获取默认角色并添加到待处理的角色列表中
		defaultRoleName := config.GetInstance().Auth.DefaultRole
		finalRoleIDs := make([]int64, 0, len(req.RoleIDs)+1)
		if len(req.RoleIDs) > 0 {
			finalRoleIDs = append(finalRoleIDs, req.RoleIDs...)
		}

		if defaultRoleName != "" {
			defaultRole, err_role := tx.Role.WithContext(ctx).Where(tx.Role.Name.Eq(defaultRoleName)).First()
			if err_role != nil && !errors.Is(err_role, gorm.ErrRecordNotFound) {
				return fmt.Errorf("failed to find default role: %w", err_role)
			}
			// 如果默认角色存在，则添加其ID
			if defaultRole != nil {
				// 避免重复添加
				isDefaultRolePresent := false
				for _, id := range finalRoleIDs {
					if id == defaultRole.ID {
						isDefaultRolePresent = true
						break
					}
				}
				if !isDefaultRolePresent {
					finalRoleIDs = append(finalRoleIDs, defaultRole.ID)
				}
			}
		}

		// b. 先删除旧的角色关联
		if _, err_del_roles := tx.AdminUserRole.WithContext(ctx).Where(tx.AdminUserRole.AdminUserID.Eq(newUser.ID)).Delete(); err_del_roles != nil {
			return err_del_roles
		}
		// c. 再添加新的角色关联
		if len(finalRoleIDs) > 0 {
			roles := make([]*model.AdminUserRole, 0, len(finalRoleIDs))
			for _, roleID := range finalRoleIDs {
				roles = append(roles, &model.AdminUserRole{
					AdminUserID: newUser.ID,
					RoleID:      roleID,
				})
			}
			if err_create_roles := tx.AdminUserRole.WithContext(ctx).Create(roles...); err_create_roles != nil {
				return err_create_roles
			}
		}

		// Find role names for casbin and response
		if len(finalRoleIDs) > 0 {
			role_models, err_find_roles := tx.Role.WithContext(ctx).Where(tx.Role.ID.In(finalRoleIDs...)).Find()
			if err_find_roles != nil {
				return fmt.Errorf("failed to find roles by ids: %w", err_find_roles)
			}
			if len(role_models) != len(finalRoleIDs) {
				return ErrRoleNotFound
			}
			for _, r_model := range role_models {
				roles_to_add = append(roles_to_add, r_model.Name)
			}
		}

		// 4. 准备返回数据
		response = &dto.UserResponse{
			UUID:      newUser.UUID,
			Username:  newUser.Username,
			Email:     newUser.Email,
			RealName:  newUser.RealName,
			Phone:     newUser.Phone,
			Avatar:    newUser.Avatar,
			IsActive:  newUser.IsActive,
			Roles:     roles_to_add, // Use the list we just built
			CreatedAt: utils.FormatTime(newUser.CreatedAt),
		}
		return nil
	})

	if txErr != nil {
		return nil, txErr
	}

	// 5. 在事务成功后，更新 Casbin
	casbinRes, err := resource.Get[*resource.CasbinResource](s.resource, resource.CasbinServiceKey)
	if err != nil {
		log.Printf("Error: failed to get casbin resource after user transaction: %v", err)
		return response, nil
	}
	enforcer := casbinRes.GetEnforcer()
	if _, err := enforcer.DeleteRolesForUser(response.UUID); err != nil {
		log.Printf("Error: failed to delete old roles from casbin for user %s: %v", response.UUID, err)
	}
	if len(response.Roles) > 0 {
		if _, err := enforcer.AddRolesForUser(response.UUID, response.Roles); err != nil {
			log.Printf("Error: failed to add new roles to casbin for user %s: %v", response.UUID, err)
		}
	}

	return response, nil
}

// ListUsers 获取用户列表（带分页和筛选）
func (s *UserService) ListUsers(ctx context.Context, req *dto.UserListRequest) (*dto.UserListResponse, error) {
	q := s.q.AdminUser.WithContext(ctx)

	// 优先处理批量UUID查询
	if len(req.UUIDs) > 0 {
		q = q.Where(s.q.AdminUser.UUID.In(req.UUIDs...))
	} else {
		// 构建常规查询条件
		if req.Username != "" {
			q = q.Where(s.q.AdminUser.Username.Like("%" + req.Username + "%"))
		}
		if req.Email != "" {
			q = q.Where(s.q.AdminUser.Email.Eq(req.Email))
		}
		if req.IsActive != nil {
			q = q.Where(s.q.AdminUser.IsActive.Is(*req.IsActive))
		}
	}

	// 获取总数
	total, err := q.Count()
	if err != nil {
		return nil, err
	}

	// 分页查询
	var users []*model.AdminUser
	if len(req.UUIDs) == 0 {
		users, err = q.Limit(req.PageSize).Offset((req.Page - 1) * req.PageSize).Find()
	} else {
		// 如果是批量查询，则不使用分页，返回所有匹配的用户
		users, err = q.Find()
	}

	if err != nil {
		return nil, err
	}

	// 组装响应
	userResponses := make([]*dto.UserResponse, 0, len(users))
	if len(users) > 0 {
		userIDs := make([]int64, len(users))
		for i, u := range users {
			userIDs[i] = u.ID
		}

		// 一次性查询所有用户的角色
		type UserRole struct {
			AdminUserID int64
			RoleName    string
		}
		var userRoles []UserRole
		s.q.Role.WithContext(ctx).
			Select(s.q.AdminUserRole.AdminUserID, s.q.Role.Name.As("role_name")).
			LeftJoin(s.q.AdminUserRole, s.q.AdminUserRole.RoleID.EqCol(s.q.Role.ID)).
			Where(s.q.AdminUserRole.AdminUserID.In(userIDs...)).
			Scan(&userRoles)

		// 将角色按用户ID分组
		rolesMap := make(map[int64][]string, len(users))
		for _, ur := range userRoles {
			rolesMap[ur.AdminUserID] = append(rolesMap[ur.AdminUserID], ur.RoleName)
		}

		for _, u := range users {
			userResponses = append(userResponses, &dto.UserResponse{
				UUID:      u.UUID,
				Username:  u.Username,
				Email:     u.Email,
				RealName:  u.RealName,
				Phone:     u.Phone,
				Avatar:    u.Avatar,
				IsActive:  u.IsActive,
				Roles:     rolesMap[u.ID], // 从 map 中获取角色
				CreatedAt: utils.FormatTime(u.CreatedAt),
			})
		}
	}

	return &dto.UserListResponse{
		Total: total,
		Users: userResponses,
	}, nil
}

// UpdateUserByAdmin 管理员更新用户信息
func (s *UserService) UpdateUserByAdmin(ctx context.Context, uuid_str string, req *dto.AdminUpdateUserRequest) (*dto.UserResponse, error) {
	err := s.q.Transaction(func(tx *query.Query) error {
		// 0. 在事务中根据 uuid 获取用户实体，确保数据一致性并获取主键ID
		user, err := tx.AdminUser.WithContext(ctx).Where(tx.AdminUser.UUID.Eq(uuid_str)).First()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrUserNotFound
			}
			return fmt.Errorf("finding user by uuid failed: %w", err)
		}

		// 1. 更新用户信息
		updates := make(map[string]interface{})
		if req.Email != "" {
			updates["email"] = req.Email
		}
		if req.RealName != "" {
			updates["real_name"] = req.RealName
		}
		if req.Phone != "" {
			updates["phone"] = req.Phone
		}
		if req.Avatar != "" {
			updates["avatar"] = req.Avatar
		}
		if req.IsActive != nil {
			updates["is_active"] = *req.IsActive
		}

		// 2. 检查 email 和 phone 的唯一性（排除软删除的记录）
		if req.Email != "" {
			count, err := tx.AdminUser.WithContext(ctx).
				Where(tx.AdminUser.Email.Eq(req.Email), tx.AdminUser.UUID.Neq(uuid_str)).
				Where(tx.AdminUser.DeletedAt.IsNull()).
				Count()
			if err != nil {
				return err
			}
			if count > 0 {
				return ErrEmailAlreadyExists
			}
			updates["email"] = req.Email
		}
		if req.Phone != "" {
			count, err := tx.AdminUser.WithContext(ctx).
				Where(tx.AdminUser.Phone.Eq(req.Phone), tx.AdminUser.UUID.Neq(uuid_str)).
				Where(tx.AdminUser.DeletedAt.IsNull()).
				Count()
			if err != nil {
				return err
			}
			if count > 0 {
				return ErrPhoneAlreadyExists
			}
			updates["phone"] = req.Phone
		}

		if len(updates) > 0 {
			if _, err := tx.AdminUser.WithContext(ctx).Where(tx.AdminUser.UUID.Eq(uuid_str)).Updates(updates); err != nil {
				return err
			}
		}

		// 3. 更新数据库中的角色关联 (如果请求中包含了 RoleIDs)
		if req.RoleIDs != nil {
			// a. 先删除旧的角色关联
			if _, err := tx.AdminUserRole.WithContext(ctx).Where(tx.AdminUserRole.AdminUserID.Eq(user.ID)).Delete(); err != nil {
				return fmt.Errorf("failed to delete old user roles from db: %w", err)
			}
			// b. 再添加新的角色关联
			if len(req.RoleIDs) > 0 {
				newRoles := make([]*model.AdminUserRole, len(req.RoleIDs))
				for i, roleID := range req.RoleIDs {
					newRoles[i] = &model.AdminUserRole{AdminUserID: user.ID, RoleID: roleID}
				}
				if err := tx.AdminUserRole.WithContext(ctx).Create(newRoles...); err != nil {
					return fmt.Errorf("failed to create new user roles in db: %w", err)
				}
			}
		}

		// 4. 更新 Casbin 中的角色策略 (如果请求中包含了 RoleIDs)
		// 注意: Casbin 的操作不是事务性的，如果这里失败，前面的数据库更新会回滚。
		if req.RoleIDs != nil {
			casbinRes, err := resource.Get[*resource.CasbinResource](s.resource, resource.CasbinServiceKey)
			if err != nil {
				return fmt.Errorf("failed to get casbin resource for role update: %w", err)
			}
			enforcer := casbinRes.GetEnforcer()

			// a. 先删除用户的所有旧角色
			if _, err := enforcer.DeleteRolesForUser(uuid_str); err != nil {
				return fmt.Errorf("failed to delete old roles for user from casbin: %w", err)
			}

			// b. 再添加新角色
			if len(req.RoleIDs) > 0 {
				role_models, err := tx.Role.WithContext(ctx).Where(tx.Role.ID.In(req.RoleIDs...)).Find()
				if err != nil {
					return fmt.Errorf("failed to find roles by ids from db: %w", err)
				}
				if len(role_models) != len(req.RoleIDs) {
					return ErrRoleNotFound // 如果提供的某些RoleID无效
				}

				var roles_to_add []string
				for _, r_model := range role_models {
					roles_to_add = append(roles_to_add, r_model.Name)
				}

				if len(roles_to_add) > 0 {
					if _, err := enforcer.AddRolesForUser(uuid_str, roles_to_add); err != nil {
						return fmt.Errorf("failed to add new roles for user to casbin: %w", err)
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update user in transaction: %w", err)
	}

	return s.GetUserByUUID(ctx, uuid_str)
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(ctx context.Context, uuid_str string) error {
	// 1. 先在事务中删除数据库用户
	err := s.q.Transaction(func(tx *query.Query) error {
		_, err := tx.AdminUser.WithContext(ctx).Where(tx.AdminUser.UUID.Eq(uuid_str)).Delete()
		return err
	})
	if err != nil {
		return err
	}

	// 2. 事务成功后，再清理 Casbin
	casbinRes, err := resource.Get[*resource.CasbinResource](s.resource, resource.CasbinServiceKey)
	if err != nil {
		log.Printf("Warn: failed to get casbin resource for user %s deletion. Casbin policies may need manual cleanup. Error: %v", uuid_str, err)
		return nil
	}
	enforcer := casbinRes.GetEnforcer()
	if _, err := enforcer.DeleteRolesForUser(uuid_str); err != nil {
		log.Printf("Warn: failed to delete user roles from casbin for user %s. Casbin policies may need manual cleanup. Error: %v", uuid_str, err)
	}
	return nil
}
