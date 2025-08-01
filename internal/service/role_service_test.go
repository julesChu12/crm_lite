package service

import (
	"context"
	"crm_lite/internal/core/config"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dao/query"
	"crm_lite/internal/dto"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)
type RoleServiceTestSuite struct {
	suite.Suite
	db         *gorm.DB
	resManager *resource.Manager
	service    *RoleService
}

func (s *RoleServiceTestSuite) SetupSuite() {
	// Create in-memory SQLite database with unique name
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	s.Require().NoError(err)
	s.db = db

	// Create roles table
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS roles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name VARCHAR(50) NOT NULL UNIQUE,
			display_name VARCHAR(100),
			description TEXT,
			is_active TINYINT(1) DEFAULT 1,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		);
	`).Error
	s.Require().NoError(err)

	// Setup resource manager
	resManager := resource.NewManager()
	dbResource := resource.NewDBResource(config.DBOptions{})
	dbResource.DB = db
	resManager.Register(resource.DBServiceKey, dbResource)
	s.resManager = resManager

	// Setup config
	opts := &config.Options{
		Auth: config.AuthOptions{
			DefaultRole: "default-role",
		},
	}
	config.SetInstanceForTest(opts)

	// Create service with our test resource manager
	s.service = &RoleService{
		q:        query.Use(db),
		resource: resManager,
	}
}

func (s *RoleServiceTestSuite) TearDownSuite() {
	// Close database connection
	sqlDB, err := s.db.DB()
	s.Require().NoError(err)
	sqlDB.Close()
}

func (s *RoleServiceTestSuite) TearDownTest() {
	// Clean up data after each test
	s.db.Exec("DELETE FROM roles")
}

func TestRoleService(t *testing.T) {
	suite.Run(t, new(RoleServiceTestSuite))
}

func (s *RoleServiceTestSuite) TestCreateRole_Success() {
	req := &dto.RoleCreateRequest{
		Name:        "test_role",
		DisplayName: "Test Role",
		Description: "A test role",
	}

	resp, err := s.service.CreateRole(context.Background(), req)

	s.NoError(err)
	s.NotNil(resp)
	s.Equal("test_role", resp.Name)
	s.Equal("Test Role", resp.DisplayName)
	s.Equal("A test role", resp.Description)
	s.True(resp.IsActive)
	s.NotZero(resp.ID)
}

func (s *RoleServiceTestSuite) TestCreateRole_DuplicateName() {
	// Create first role using service's query
	role := &model.Role{
		Name:        "duplicate_role",
		DisplayName: "Duplicate Role",
		IsActive:    true,
	}
	err := s.service.q.Role.WithContext(context.Background()).Create(role)
	s.Require().NoError(err)

	// Try to create role with same name
	req := &dto.RoleCreateRequest{
		Name:        "duplicate_role",
		DisplayName: "Another Role",
		Description: "Another test role",
	}

	resp, err := s.service.CreateRole(context.Background(), req)

	s.Error(err)
	s.Nil(resp)
	s.Equal(ErrRoleNameAlreadyExists, err)
}

func (s *RoleServiceTestSuite) TestListRoles_Success() {
	// Create test roles using service
	roles := []*dto.RoleCreateRequest{
		{Name: "role1", DisplayName: "Role 1"},
		{Name: "role2", DisplayName: "Role 2"},
		{Name: "role3", DisplayName: "Role 3"},
	}
	for _, role := range roles {
		_, err := s.service.CreateRole(context.Background(), role)
		s.Require().NoError(err)
	}

	resp, err := s.service.ListRoles(context.Background())

	s.NoError(err)
	s.Len(resp, 3)
	s.Equal("role1", resp[0].Name)
	s.Equal("role2", resp[1].Name)
	s.Equal("role3", resp[2].Name)
}

func (s *RoleServiceTestSuite) TestListRoles_Empty() {
	resp, err := s.service.ListRoles(context.Background())

	s.NoError(err)
	s.Len(resp, 0)
}

func (s *RoleServiceTestSuite) TestGetRoleByID_Success() {
	// Create test role using service
	req := &dto.RoleCreateRequest{
		Name:        "test_role",
		DisplayName: "Test Role",
		Description: "A test role",
		IsActive:    true,
	}
	createdRole, err := s.service.CreateRole(context.Background(), req)
	s.Require().NoError(err)

	resp, err := s.service.GetRoleByID(context.Background(), fmt.Sprintf("%d", createdRole.ID))

	s.NoError(err)
	s.NotNil(resp)
	s.Equal("test_role", resp.Name)
	s.Equal("Test Role", resp.DisplayName)
	s.Equal("A test role", resp.Description)
	s.True(resp.IsActive)
}

func (s *RoleServiceTestSuite) TestGetRoleByID_NotFound() {
	resp, err := s.service.GetRoleByID(context.Background(), "999")

	s.Error(err)
	s.Nil(resp)
	s.Equal(ErrRoleNotFound, err)
}

func (s *RoleServiceTestSuite) TestGetRoleByID_InvalidID() {
	resp, err := s.service.GetRoleByID(context.Background(), "invalid")

	s.Error(err)
	s.Nil(resp)
	s.Equal(ErrRoleNotFound, err)
}

func (s *RoleServiceTestSuite) TestUpdateRole_Success() {
	// Create test role using service
	req := &dto.RoleCreateRequest{
		Name:        "test_role",
		DisplayName: "Test Role",
		Description: "Original description",
	}
	createdRole, err := s.service.CreateRole(context.Background(), req)
	s.Require().NoError(err)

	isActive := false
	updateReq := &dto.RoleUpdateRequest{
		DisplayName: "Updated Role",
		Description: "Updated description",
		IsActive:    &isActive,
	}

	resp, err := s.service.UpdateRole(context.Background(), fmt.Sprintf("%d", createdRole.ID), updateReq)

	s.NoError(err)
	s.NotNil(resp)
	s.Equal("test_role", resp.Name) // Name should not change
	s.Equal("Updated Role", resp.DisplayName)
	s.Equal("Updated description", resp.Description)
	s.False(resp.IsActive)
}

func (s *RoleServiceTestSuite) TestUpdateRole_NotFound() {
	req := &dto.RoleUpdateRequest{
		DisplayName: "Updated Role",
	}

	resp, err := s.service.UpdateRole(context.Background(), "999", req)

	s.Error(err)
	s.Nil(resp)
	s.Equal(ErrRoleNotFound, err)
}

func (s *RoleServiceTestSuite) TestUpdateRole_InvalidID() {
	req := &dto.RoleUpdateRequest{
		DisplayName: "Updated Role",
	}

	resp, err := s.service.UpdateRole(context.Background(), "invalid", req)

	s.Error(err)
	s.Nil(resp)
	s.Equal(ErrRoleNotFound, err)
}

func (s *RoleServiceTestSuite) TestUpdateRole_DuplicateDisplayName() {
	// Create two test roles using service
	role1Req := &dto.RoleCreateRequest{
		Name:        "role1",
		DisplayName: "Role 1",
	}
	role2Req := &dto.RoleCreateRequest{
		Name:        "role2",
		DisplayName: "Role 2",
	}
	_, err := s.service.CreateRole(context.Background(), role1Req)
	s.Require().NoError(err)
	role2, err := s.service.CreateRole(context.Background(), role2Req)
	s.Require().NoError(err)

	// Try to update role2 with role1's display name
	req := &dto.RoleUpdateRequest{
		DisplayName: "Role 1",
	}

	resp, err := s.service.UpdateRole(context.Background(), fmt.Sprintf("%d", role2.ID), req)

	s.Error(err)
	s.Nil(resp)
	s.Equal(ErrRoleNameAlreadyExists, err)
}

func (s *RoleServiceTestSuite) TestDeleteRole_Success() {
	// Create test role using service
	req := &dto.RoleCreateRequest{
		Name:        "test_role",
		DisplayName: "Test Role",
		Description: "A test role",
	}
	createdRole, err := s.service.CreateRole(context.Background(), req)
	s.Require().NoError(err)

	err = s.service.DeleteRole(context.Background(), fmt.Sprintf("%d", createdRole.ID))

	s.NoError(err)

	// Verify role is deleted by trying to get it
	_, err = s.service.GetRoleByID(context.Background(), fmt.Sprintf("%d", createdRole.ID))
	s.Error(err)
}

func (s *RoleServiceTestSuite) TestDeleteRole_NotFound() {
	err := s.service.DeleteRole(context.Background(), "999")

	s.Error(err)
	s.Equal(ErrRoleNotFound, err)
}

func (s *RoleServiceTestSuite) TestDeleteRole_InvalidID() {
	err := s.service.DeleteRole(context.Background(), "invalid")

	s.Error(err)
	s.Equal(ErrRoleNotFound, err)
}

func (s *RoleServiceTestSuite) TestToDTO() {
	role := &model.Role{
		ID:          1,
		Name:        "test_role",
		DisplayName: "Test Role",
		Description: "A test role",
		IsActive:    true,
	}

	dto := s.service.toDTO(role)

	s.Equal(int64(1), dto.ID)
	s.Equal("test_role", dto.Name)
	s.Equal("Test Role", dto.DisplayName)
	s.Equal("A test role", dto.Description)
	s.True(dto.IsActive)
}