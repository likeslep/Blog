// models/post.go
package models

import (
	"gorm.io/gorm"
)

type Post struct {
	gorm.Model
	Title        string `gorm:"type:varchar(200);not null" json:"title"`
	Slug         string `gorm:"type:varchar(200);uniqueIndex;not null" json:"slug"`
	Summary      string `gorm:"type:varchar(500)" json:"summary"`
	Content      string `gorm:"type:longtext;not null" json:"content"`
	CoverImage   string `gorm:"type:varchar(255)" json:"cover_image"`
	ViewCount    int    `gorm:"default:0" json:"view_count"`
	LikeCount    int    `gorm:"default:0" json:"like_count"`
	CommentCount int    `gorm:"default:0" json:"comment_count"`
	Status       int    `gorm:"default:1" json:"status"` // 1:已发布 0:草稿 2:私密
	IsTop        bool   `gorm:"default:false" json:"is_top"`
	UserID       uint   `gorm:"index;not null" json:"user_id"`
	CategoryID   uint   `gorm:"index" json:"category_id"`

	// 关联
	User     User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Category Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Tags     []Tag    `gorm:"many2many:post_tags;" json:"tags,omitempty"`
}
