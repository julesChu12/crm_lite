package model

import "time"

// BaseModel 定义了所有数据库模型共有的字段
type BaseModel struct {
	ID        uint      `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"index;comment:创建时间"`
	UpdatedAt time.Time `gorm:"index;comment:更新时间"`
}
