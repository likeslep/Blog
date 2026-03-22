// models/comment.go
package models

import (
	"gorm.io/gorm"
)

type Comment struct {
	gorm.Model
	Content   string `gorm:"type:text;not null" json:"content"`
	UserID    uint   `gorm:"index;not null" json:"user_id"`
	PostID    uint   `gorm:"index;not null" json:"post_id"`
	ParentID  *uint  `gorm:"index" json:"parent_id"`
	Status    int    `gorm:"default:1" json:"status"` // 1:正常 0:审核中 2:已删除
	LikeCount int    `gorm:"default:0" json:"like_count"`

	// 关联
	User    User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Post    Post      `gorm:"foreignKey:PostID" json:"post,omitempty"`
	Replies []Comment `gorm:"foreignKey:ParentID" json:"replies,omitempty"`
}
