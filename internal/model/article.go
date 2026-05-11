package model

import (
	"time"

	"gorm.io/gorm"
)

type Article struct {
	ID         uint           `gorm:"primaryKey"`
	Title      string         `gorm:"not null"`
	Slug       string         `gorm:"uniqueIndex;not null"`
	Content    string         `gorm:"not null"`
	HTML       string         `gorm:"not null"`
	Summary    string
	Status     int            `gorm:"default:0"` // 0=草稿, 1=已发布
	Pinned     int            `gorm:"default:0"`
	CategoryID *uint
	Category   Category
	Tags       []Tag          `gorm:"many2many:article_tags;"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

func (Article) TableName() string {
	return "articles"
}
