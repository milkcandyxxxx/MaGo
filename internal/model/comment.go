package model

import (
	"time"

	"gorm.io/gorm"
)

type Comment struct {
	ID        uint           `gorm:"primaryKey"`
	ArticleID uint           `gorm:"not null;index"`
	ParentID  *uint          `gorm:"index"` // 父评论 ID，用于回复
	Nickname  string         `gorm:"not null"`
	Email     string
	Content   string         `gorm:"not null"`
	Status    int            `gorm:"default:1"` // 0=待审核, 1=已发布
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// 关联
	Article   Article
	Replies   []Comment `gorm:"foreignKey:ParentID"`
}

func (Comment) TableName() string {
	return "comments"
}
