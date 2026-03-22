// models/category.go
package models

import (
	"gorm.io/gorm"
)

type Category struct {
	gorm.Model
	Name        string `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"`
	Slug        string `gorm:"type:varchar(50);uniqueIndex;not null" json:"slug"`
	Description string `gorm:"type:varchar(200)" json:"description"`
	PostCount   int    `gorm:"default:0" json:"post_count"`
	SortOrder   int    `gorm:"default:0" json:"sort_order"`
}
