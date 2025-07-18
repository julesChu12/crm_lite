// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

import (
	"time"

	"gorm.io/gorm"
)

const TableNameAdminUser = "admin_users"

// AdminUser mapped from table <admin_users>
type AdminUser struct {
	ID           int64          `gorm:"column:id;type:bigint(20);primaryKey;autoIncrement:true" json:"id"`
	UUID         string         `gorm:"column:uuid;type:varchar(36);not null;uniqueIndex:uuid,priority:1" json:"uuid"`
	Username     string         `gorm:"column:username;type:varchar(50);not null;uniqueIndex:username,priority:1" json:"username"`
	Email        string         `gorm:"column:email;type:varchar(100);not null;uniqueIndex:email,priority:1" json:"email"`
	PasswordHash string         `gorm:"column:password_hash;type:varchar(255);not null" json:"password_hash"`
	RealName     string         `gorm:"column:real_name;type:varchar(50)" json:"real_name"`
	Phone        string         `gorm:"column:phone;type:varchar(20)" json:"phone"`
	Avatar       string         `gorm:"column:avatar;type:varchar(255)" json:"avatar"`
	IsActive     bool           `gorm:"column:is_active;type:tinyint(1);default:1" json:"is_active"`
	LastLoginAt  time.Time      `gorm:"column:last_login_at;type:datetime(6)" json:"last_login_at"`
	CreatedAt    time.Time      `gorm:"column:created_at;type:timestamp;default:current_timestamp()" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;type:timestamp;default:current_timestamp()" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;type:datetime(6);index:idx_admin_users_deleted_at,priority:1" json:"deleted_at"`
}

// TableName AdminUser's table name
func (*AdminUser) TableName() string {
	return TableNameAdminUser
}
