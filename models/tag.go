// models/tag.go
package models

import (
	"gorm.io/gorm"
)

type Tag struct {
	gorm.Model
	Name      string `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"`
	Slug      string `gorm:"type:varchar(50);uniqueIndex;not null" json:"slug"`
	PostCount int    `gorm:"default:0" json:"post_count"`
}

// 文章-标签关联表
type PostTag struct {
	PostID uint `gorm:"primaryKey"`
	TagID  uint `gorm:"primaryKey"`
}
