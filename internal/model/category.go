package model

type Category struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"uniqueIndex;not null"`
	Slug string `gorm:"uniqueIndex;not null"`
}

func (Category) TableName() string {
	return "categories"
}
