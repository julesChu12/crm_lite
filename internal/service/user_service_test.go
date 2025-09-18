package service

import (
	"context"
	"crm_lite/internal/core/config"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dto"
	"fmt"
    "os"
	"testing"

	"github.com/casbin/casbin/v2"
	casbinmodel "github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type UserServiceTestSuite struct {
	suite.Suite
	db         *gorm.DB
	resManager *resource.Manager
	service    *UserService
	enforcer   *casbin.Enforcer
}

func (s *UserServiceTestSuite) SetupSuite() {
	// Setup in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	s.Require().NoError(err)
	s.db = db

	// Manually create all tables to avoid sqlite compatibility issues
	err = db.Exec(`
		CREATE TABLE admin_users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid VARCHAR(36) NOT NULL,
			username VARCHAR(50) NOT NULL,
			email VARCHAR(100) NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			real_name VARCHAR(50),
			phone VARCHAR(20),
			avatar VARCHAR(255),
			is_active TINYINT(1) DEFAULT 1,
			last_login_at DATETIME,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		);
	`).Error
	s.Require().NoError(err)

	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS roles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name VARCHAR(50) NOT NULL,
			display_name VARCHAR(100) NOT NULL,
			description TEXT,
			is_active TINYINT(1) DEFAULT 1,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		);
	`).Error
	s.Require().NoError(err)

	err = db.Exec(`
		CREATE TABLE admin_user_roles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			admin_user_id BIGINT NOT NULL,
			role_id BIGINT NOT NULL,
			created_at DATETIME
		);
	`).Error
	s.Require().NoError(err)

	// Setup Casbin enforcer with the in-memory database
	// The adapter will automatically create the `casbin_rule` table if it doesn't exist
	adapter, err := gormadapter.NewAdapterByDB(db)
	s.Require().NoError(err)
	m, err := casbinmodel.NewModelFromString(`
	[request_definition]
	r = sub, obj, act
	
	[policy_definition]
	p = sub, obj, act
	
	[role_definition]
	g = _, _
	
	[policy_effect]
	e = some(where (p.eft == allow))
	
	[matchers]
	m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
	`)
	s.Require().NoError(err)
	enforcer, err := casbin.NewEnforcer(m, adapter)
	s.Require().NoError(err)
	s.enforcer = enforcer

	// Setup resource manager
	resManager := resource.NewManager()
	dbResource := resource.NewDBResource(config.DBOptions{})
	dbResource.DB = db
	_ = resManager.Register(resource.DBServiceKey, dbResource)

	casbinResource := resource.NewCasbinResource(resManager, config.RbacOptions{})
	casbinResource.Enforcer = enforcer
	_ = resManager.Register(resource.CasbinServiceKey, casbinResource)
	s.resManager = resManager

	// Setup config
	opts := &config.Options{
		Auth: config.AuthOptions{
			DefaultRole: "default-role",
		},
	}
	config.SetInstanceForTest(opts)

	// Initialize service
	s.service = NewUserService(s.resManager)
}

func (s *UserServiceTestSuite) TearDownTest() {
	// Clean up database after each test
	s.db.Exec("DELETE FROM admin_users")
	s.db.Exec("DELETE FROM roles")
	s.db.Exec("DELETE FROM admin_user_roles")
	s.db.Exec("DELETE FROM casbin_rule")
}

func TestUserService(t *testing.T) {
    if os.Getenv("RUN_DB_TESTS") != "1" {
        t.Skip("Skipping user service suite because RUN_DB_TESTS is not set")
        return
    }
    suite.Run(t, new(UserServiceTestSuite))
}

func (s *UserServiceTestSuite) TestCreateUserByAdmin_NewUser() {
	ctx := context.Background()
	req := &dto.AdminCreateUserRequest{
		Username: "newuser",
		Email:    "newuser@example.com",
		Phone:    "1234567890",
		Password: "password123",
		RoleIDs:  []int64{},
	}

	resp, err := s.service.CreateUserByAdmin(ctx, req)
	s.NoError(err)
	s.NotNil(resp)
	s.Equal("newuser", resp.Username)
	s.Contains(resp.Roles, "default-role", "User should have the default role")

	// Verify user in DB
	var user model.AdminUser
	err = s.db.Where("username = ?", "newuser").First(&user).Error
	s.NoError(err)
	s.Equal("newuser@example.com", user.Email)

	// Verify casbin role
	roles, err := s.enforcer.GetRolesForUser(resp.UUID)
	s.NoError(err)
	s.Contains(roles, "default-role")
}

func (s *UserServiceTestSuite) TestCreateUserByAdmin_UserAlreadyExists() {
	// Pre-create a user
	existingUser := &model.AdminUser{
		UUID:     "uuid-exists",
		Username: "existinguser",
		Email:    "exists@example.com",
		Phone:    "1112223333",
		IsActive: true,
	}
	s.db.Create(existingUser)

	ctx := context.Background()
	req := &dto.AdminCreateUserRequest{
		Username: "existinguser",
		Email:    "another@example.com",
		Phone:    "0000000000",
		Password: "password",
	}

	// Test username exists
	_, err := s.service.CreateUserByAdmin(ctx, req)
	s.Error(err)
	s.ErrorIs(err, ErrUserAlreadyExists, "Should return ErrUserAlreadyExists")

	// Test email exists
	req.Username = "newuser"
	req.Email = "exists@example.com"
	_, err = s.service.CreateUserByAdmin(ctx, req)
	s.Error(err)
	s.ErrorIs(err, ErrEmailAlreadyExists, "Should return ErrEmailAlreadyExists")
}

func (s *UserServiceTestSuite) TestCreateUserByAdmin_RestoreSoftDeletedUser() {
	// Pre-create a soft-deleted user
	deletedUser := &model.AdminUser{
		ID:       99,
		Username: "deleteduser",
		Email:    "deleted@example.com",
		Phone:    "4445556666",
	}
	s.db.Create(deletedUser)
	s.db.Delete(deletedUser)

	ctx := context.Background()
	req := &dto.AdminCreateUserRequest{
		Username: "deleteduser",
		Email:    "deleted@example.com",
		Phone:    "4445556666",
		Password: "newpassword",
		RealName: "Restored User",
	}

	resp, err := s.service.CreateUserByAdmin(ctx, req)
	s.NoError(err)
	s.NotNil(resp)
	s.Equal("Restored User", resp.RealName)
	s.True(resp.IsActive)

	// Verify user is restored in DB
	var user model.AdminUser
	err = s.db.Unscoped().Where("username = ?", "deleteduser").First(&user).Error
	s.NoError(err)
	s.Nil(user.DeletedAt.Time)
}

func (s *UserServiceTestSuite) TestCreateUserByAdmin_WithSpecificRoles() {
	// Create a role
	role1 := model.Role{ID: 101, Name: "editor"}
	s.db.Create(&role1)

	ctx := context.Background()
	req := &dto.AdminCreateUserRequest{
		Username: "roleuser",
		Email:    "role@example.com",
		Phone:    "9876543210",
		Password: "password123",
		RoleIDs:  []int64{role1.ID},
	}

	resp, err := s.service.CreateUserByAdmin(ctx, req)
	s.NoError(err)
	s.NotNil(resp)
	s.Contains(resp.Roles, "editor", "User should have the 'editor' role")
	s.Contains(resp.Roles, "default-role", "User should also have the default role")

	// Verify casbin roles
	roles, err := s.enforcer.GetRolesForUser(resp.UUID)
	s.NoError(err)
	s.ElementsMatch([]string{"editor", "default-role"}, roles)
}

func (s *UserServiceTestSuite) TestGetUserByUUID_Success() {
	// Pre-create a user and role
	role := model.Role{Name: "viewer"}
	s.db.Create(&role)
	user := model.AdminUser{UUID: "uuid-get", Username: "get-user"}
	s.db.Create(&user)
	s.db.Create(&model.AdminUserRole{AdminUserID: user.ID, RoleID: role.ID})

	resp, err := s.service.GetUserByUUID(context.Background(), "uuid-get")

	s.NoError(err)
	s.NotNil(resp)
	s.Equal("get-user", resp.Username)
	s.Contains(resp.Roles, "viewer")
}

func (s *UserServiceTestSuite) TestGetUserByUUID_NotFound() {
	_, err := s.service.GetUserByUUID(context.Background(), "non-existent-uuid")
	s.Error(err)
	s.Equal(ErrUserNotFound, err)
}

func (s *UserServiceTestSuite) TestListUsers_Success() {
	// Pre-create users
	s.db.Create(&model.AdminUser{Username: "user1", Email: "user1@test.com"})
	s.db.Create(&model.AdminUser{Username: "user2", Email: "user2@test.com"})

	req := &dto.UserListRequest{Page: 1, PageSize: 10}
	resp, err := s.service.ListUsers(context.Background(), req)

	s.NoError(err)
	s.EqualValues(2, resp.Total)
	s.Len(resp.Users, 2)
}

func (s *UserServiceTestSuite) TestUpdateUserByAdmin_Success() {
	// Pre-create user and roles
	user := model.AdminUser{UUID: "uuid-update", Username: "update-user"}
	s.db.Create(&user)
	role_to_remove := model.Role{Name: "old_role"}
	s.db.Create(&role_to_remove)
	role_to_add := model.Role{Name: "new_role"}
	s.db.Create(&role_to_add)
	_, _ = s.enforcer.AddRoleForUser(user.UUID, "old_role")
	s.db.Create(&model.AdminUserRole{AdminUserID: user.ID, RoleID: role_to_remove.ID})

	ctx := context.Background()
	req := &dto.AdminUpdateUserRequest{
		Email:   "updated@example.com",
		RoleIDs: []int64{role_to_add.ID},
	}
	resp, err := s.service.UpdateUserByAdmin(ctx, user.UUID, req)

	s.NoError(err)
	s.NotNil(resp)
	s.Equal("updated@example.com", resp.Email)
	s.NotContains(resp.Roles, "old_role")
	s.Contains(resp.Roles, "new_role")

	// Verify casbin
	casbinRoles, err := s.enforcer.GetRolesForUser(user.UUID)
	s.NoError(err)
	s.Contains(casbinRoles, "new_role")
}

func (s *UserServiceTestSuite) TestDeleteUser_Success() {
	// Pre-create user and role
	user := model.AdminUser{UUID: "uuid-delete", Username: "delete-user"}
	s.db.Create(&user)
	_, _ = s.enforcer.AddRoleForUser(user.UUID, "some-role")

	err := s.service.DeleteUser(context.Background(), user.UUID)
	s.NoError(err)

	// Verify user is soft-deleted
	var dbUser model.AdminUser
	err = s.db.Unscoped().Where("uuid = ?", user.UUID).First(&dbUser).Error
	s.NoError(err)
	s.NotNil(dbUser.DeletedAt)

	// Verify casbin roles are deleted
	casbinRoles, err := s.enforcer.GetRolesForUser(user.UUID)
	s.NoError(err)
	s.Empty(casbinRoles)
}

func (s *UserServiceTestSuite) TestDeleteUser_NotFound() {
	err := s.service.DeleteUser(context.Background(), "non-existent-uuid")
	s.Error(err)
	s.Equal(ErrUserNotFound, err)
}

func (s *UserServiceTestSuite) TestUpdateUserByAdmin_NotFound() {
	req := &dto.AdminUpdateUserRequest{
		Email: "updated@example.com",
	}

	resp, err := s.service.UpdateUserByAdmin(context.Background(), "non-existent-uuid", req)
	s.Error(err)
	s.Nil(resp)
	s.Equal(ErrUserNotFound, err)
}

func (s *UserServiceTestSuite) TestUpdateUserByAdmin_EmailAlreadyExists() {
	// Create two users
	user1 := model.AdminUser{UUID: "uuid-1", Username: "user1", Email: "user1@test.com"}
	user2 := model.AdminUser{UUID: "uuid-2", Username: "user2", Email: "user2@test.com"}
	s.db.Create(&user1)
	s.db.Create(&user2)

	// Try to update user2 with user1's email
	req := &dto.AdminUpdateUserRequest{
		Email: "user1@test.com",
	}

	resp, err := s.service.UpdateUserByAdmin(context.Background(), user2.UUID, req)
	s.Error(err)
	s.Nil(resp)
	s.Equal(ErrEmailAlreadyExists, err)
}

func (s *UserServiceTestSuite) TestUpdateUserByAdmin_PhoneAlreadyExists() {
	// Create two users
	user1 := model.AdminUser{UUID: "uuid-1", Username: "user1", Phone: "1111111111"}
	user2 := model.AdminUser{UUID: "uuid-2", Username: "user2", Phone: "2222222222"}
	s.db.Create(&user1)
	s.db.Create(&user2)

	// Try to update user2 with user1's phone
	req := &dto.AdminUpdateUserRequest{
		Phone: "1111111111",
	}

	resp, err := s.service.UpdateUserByAdmin(context.Background(), user2.UUID, req)
	s.Error(err)
	s.Nil(resp)
	s.Equal(ErrPhoneAlreadyExists, err)
}

func (s *UserServiceTestSuite) TestListUsers_WithPagination() {
	// Create multiple users
	for i := 1; i <= 15; i++ {
		user := model.AdminUser{
			Username: fmt.Sprintf("user%d", i),
			Email:    fmt.Sprintf("user%d@test.com", i),
		}
		s.db.Create(&user)
	}

	// Test first page
	req := &dto.UserListRequest{Page: 1, PageSize: 10}
	resp, err := s.service.ListUsers(context.Background(), req)
	s.NoError(err)
	s.EqualValues(15, resp.Total)
	s.Len(resp.Users, 10)

	// Test second page
	req = &dto.UserListRequest{Page: 2, PageSize: 10}
	resp, err = s.service.ListUsers(context.Background(), req)
	s.NoError(err)
	s.EqualValues(15, resp.Total)
	s.Len(resp.Users, 5)
}

func (s *UserServiceTestSuite) TestListUsers_WithFilters() {
	// Create test users
	users := []model.AdminUser{
		{Username: "alice", Email: "alice@test.com", RealName: "Alice Smith", IsActive: true},
		{Username: "bob", Email: "bob@test.com", RealName: "Bob Johnson", IsActive: false},
		{Username: "charlie", Email: "charlie@test.com", RealName: "Charlie Brown", IsActive: true},
	}
	for _, user := range users {
		s.db.Create(&user)
	}

	// Filter by username
	req := &dto.UserListRequest{Page: 1, PageSize: 10, Username: "alice"}
	resp, err := s.service.ListUsers(context.Background(), req)
	s.NoError(err)
	s.EqualValues(1, resp.Total)
	s.Len(resp.Users, 1)
	s.Equal("alice", resp.Users[0].Username)

	// Filter by email
	req = &dto.UserListRequest{Page: 1, PageSize: 10, Email: "bob@test.com"}
	resp, err = s.service.ListUsers(context.Background(), req)
	s.NoError(err)
	s.EqualValues(1, resp.Total)
	s.Equal("bob", resp.Users[0].Username)

	// Filter by active status
	isActive := true
	req = &dto.UserListRequest{Page: 1, PageSize: 10, IsActive: &isActive}
	resp, err = s.service.ListUsers(context.Background(), req)
	s.NoError(err)
	s.EqualValues(2, resp.Total) // alice and charlie
	s.Len(resp.Users, 2)
}

func (s *UserServiceTestSuite) TestCreateUserByAdmin_PhoneAlreadyExists() {
	// Pre-create a user with phone
	existingUser := &model.AdminUser{
		UUID:     "uuid-phone",
		Username: "phoneuser",
		Email:    "phone@example.com",
		Phone:    "1234567890",
		IsActive: true,
	}
	s.db.Create(existingUser)

	ctx := context.Background()
	req := &dto.AdminCreateUserRequest{
		Username: "newuser",
		Email:    "new@example.com",
		Phone:    "1234567890", // Same phone
		Password: "password",
	}

	_, err := s.service.CreateUserByAdmin(ctx, req)
	s.Error(err)
	s.ErrorIs(err, ErrPhoneAlreadyExists)
}

func (s *UserServiceTestSuite) TestCreateUserByAdmin_InvalidRoleID() {
	ctx := context.Background()
	req := &dto.AdminCreateUserRequest{
		Username: "roleuser",
		Email:    "role@example.com",
		Phone:    "9876543210",
		Password: "password123",
		RoleIDs:  []int64{999}, // Non-existent role
	}

	// This should still succeed but only assign default role
	resp, err := s.service.CreateUserByAdmin(ctx, req)
	s.NoError(err)
	s.NotNil(resp)
	s.Contains(resp.Roles, "default-role")
	s.NotContains(resp.Roles, "non-existent")
}
