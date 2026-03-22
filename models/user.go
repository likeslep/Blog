// models/user.go
package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username  string     `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	Email     string     `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	Password  string     `gorm:"type:varchar(255);not null" json:"-"`
	Avatar    string     `gorm:"type:varchar(255)" json:"avatar"`
	Bio       string     `gorm:"type:varchar(500)" json:"bio"`
	Role      string     `gorm:"type:varchar(20);default:'user'" json:"role"`
	Status    int        `gorm:"default:1" json:"status"`
	LastLogin *time.Time `json:"last_login"`
}
