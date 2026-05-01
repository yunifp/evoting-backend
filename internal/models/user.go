package models

import (
	"time"
)

type User struct {
	ID         string    `gorm:"type:char(36);primaryKey" json:"id"` 
	Name       string    `gorm:"type:varchar(100);not null" json:"name"`
	Email      string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	Password   string    `gorm:"type:varchar(255);not null" json:"-"`
	
	IsApproved bool      `gorm:"default:false" json:"is_approved"` 
	IsActive   bool      `gorm:"default:true" json:"is_active"`
	
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Roles []Role `gorm:"many2many:user_roles;" json:"roles,omitempty"`
}

type UserRole struct {
	UserID    string    `gorm:"type:char(36);primaryKey" json:"user_id"`
	RoleID    uint      `gorm:"primaryKey" json:"role_id"`
	CreatedAt time.Time `json:"created_at"`
}