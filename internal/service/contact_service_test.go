package service

import (
	"context"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dto"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupContactServiceForTest is a helper to setup ContactService with an in-memory DB
func setupContactServiceForTest() (*ContactService, *gorm.DB) {
	// Setup in-memory SQLite DB
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// SQLite 不支持 bigint(20)，直接手动建表满足测试需求
	db.Exec(`CREATE TABLE IF NOT EXISTS customers (
        id INTEGER PRIMARY KEY,
        name TEXT,
        phone TEXT,
        email TEXT,
        gender TEXT,
        birthday DATE,
        level TEXT,
        tags TEXT,
        note TEXT,
        source TEXT,
        assigned_to INTEGER,
        deleted_at DATETIME,
        created_at DATETIME,
        updated_at DATETIME
    );`)

	db.Exec(`CREATE TABLE IF NOT EXISTS contacts (
        id INTEGER PRIMARY KEY,
        customer_id INTEGER,
        name TEXT,
        phone TEXT,
        email TEXT,
        position TEXT,
        is_primary BOOLEAN,
        note TEXT,
        deleted_at DATETIME,
        created_at DATETIME,
        updated_at DATETIME
    );`)

	// Create a resource manager and register DB resource
	resManager := resource.NewManager()
	dbRes := resource.NewDBResource(db)
	resManager.Register(resource.DBServiceKey, dbRes)

	contactService := NewContactService(resManager)
	return contactService, db
}

func TestCreateContact_Success(t *testing.T) {
	// Setup
	contactService, db := setupContactServiceForTest()
	ctx := context.Background()

	// Create a test customer
	customer := &model.Customer{ID: 1}
	db.Create(customer)

	// Execute
	req := &dto.ContactCreateRequest{
		Name:      "John Doe",
		Phone:     "1112223333",
		Email:     "john@example.com",
		IsPrimary: true,
	}
	resp, err := contactService.Create(ctx, 1, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "John Doe", resp.Name)
	assert.True(t, resp.IsPrimary)
}

func TestListContacts_Success(t *testing.T) {
	// Setup
	contactService, db := setupContactServiceForTest()
	ctx := context.Background()

	// Create a test customer and contacts
	customer := &model.Customer{ID: 1}
	db.Create(customer)
	db.Create(&model.Contact{CustomerID: 1, Name: "Contact 1"})
	db.Create(&model.Contact{CustomerID: 1, Name: "Contact 2"})

	// Execute
	contacts, err := contactService.List(ctx, 1)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, contacts, 2)
}

func TestGetContactByID_Success(t *testing.T) {
	// Setup
	contactService, db := setupContactServiceForTest()
	ctx := context.Background()

	// Create a test contact
	contact := &model.Contact{CustomerID: 1, Name: "Test Contact"}
	db.Create(contact)

	// Execute
	resp, err := contactService.GetContactByID(ctx, contact.ID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Test Contact", resp.Name)
}

func TestUpdateContact_Success(t *testing.T) {
	// Setup
	contactService, db := setupContactServiceForTest()
	ctx := context.Background()

	// Create a test contact
	contact := &model.Contact{CustomerID: 1, Name: "Old Name"}
	db.Create(contact)

	// Execute
	req := &dto.ContactUpdateRequest{Name: "New Name"}
	err := contactService.Update(ctx, contact.ID, req)

	// Assert
	assert.NoError(t, err)

	// Verify in DB
	updatedContact, _ := contactService.GetContactByID(ctx, contact.ID)
	assert.Equal(t, "New Name", updatedContact.Name)
}

func TestDeleteContact_Success(t *testing.T) {
	// Setup
	contactService, db := setupContactServiceForTest()
	ctx := context.Background()

	// Create a test contact
	contact := &model.Contact{CustomerID: 1, Name: "Contact to delete"}
	db.Create(contact)

	// Execute
	err := contactService.Delete(ctx, contact.ID)

	// Assert
	assert.NoError(t, err)

	// Verify in DB
	_, findErr := contactService.GetContactByID(ctx, contact.ID)
	assert.Error(t, findErr)
	assert.True(t, errors.Is(findErr, ErrContactNotFound))
}
