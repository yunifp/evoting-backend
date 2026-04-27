package models

import (
	"time"
)


type Role struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"`
	Description string    `gorm:"type:varchar(255)" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
	Users       []User       `gorm:"many2many:user_roles;" json:"users,omitempty"`
}


type Menu struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"` 
	Path      string    `gorm:"type:varchar(100);not null" json:"path"` 
	Icon      string    `gorm:"type:varchar(100)" json:"icon"`          
	ParentID  *uint     `json:"parent_id"`                              
	SortOrder int       `gorm:"default:0" json:"sort_order"`            
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	SubMenus    []Menu       `gorm:"foreignKey:ParentID" json:"sub_menus,omitempty"`
	Permissions []Permission `gorm:"foreignKey:MenuID" json:"permissions,omitempty"`
}

type Permission struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	MenuID      uint      `gorm:"not null" json:"menu_id"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`   
	Action      string    `gorm:"type:varchar(50);not null" json:"action"` 
	Description string    `gorm:"type:varchar(255)" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Menu Menu `gorm:"foreignKey:MenuID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"menu,omitempty"`
}

type RolePermission struct {
	RoleID       uint      `gorm:"primaryKey" json:"role_id"`
	PermissionID uint      `gorm:"primaryKey" json:"permission_id"`
	CreatedAt    time.Time `json:"created_at"`
}