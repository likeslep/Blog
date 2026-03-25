// models/attachment.go
package models

import (
	"gorm.io/gorm"
)

// models/attachment.go
type Attachment struct {
	gorm.Model
	UserID       uint   `gorm:"index;not null" json:"user_id"`
	Filename     string `gorm:"type:varchar(255);not null" json:"filename"`
	OriginalName string `gorm:"type:varchar(255);not null" json:"original_name"`
	Path         string `gorm:"type:varchar(500);not null" json:"path"`
	Size         int64  `json:"size"`
	MimeType     string `gorm:"type:varchar(100)" json:"mime_type"`
	Type         string `gorm:"type:varchar(20)" json:"type"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	ThumbPath    string `gorm:"type:varchar(500)" json:"thumb_path"`

	// 非数据库字段，用于API响应
	URL      string `gorm:"-" json:"url,omitempty"`
	ThumbURL string `gorm:"-" json:"thumb_url,omitempty"`
}
