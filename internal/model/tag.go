package model

type Tag struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"uniqueIndex;not null"`
	Slug string `gorm:"uniqueIndex;not null"`
}

func (Tag) TableName() string {
	return "tags"
}
