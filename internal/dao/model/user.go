package model

// User 定义了用户模型
type User struct {
	BaseModel
	Username string `gorm:"uniqueIndex;not null;comment:用户名"`
	Password string `gorm:"not null;comment:密码哈希"`
	Email    string `gorm:"uniqueIndex;not null;comment:邮箱"`
	FullName string `gorm:"comment:全名"`
	Role     string `gorm:"default:'user';comment:角色 (e.g., user, admin)"`
	IsActive bool   `gorm:"default:true;comment:账户是否激活"`
}
