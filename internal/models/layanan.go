package models

import (
	"time"
)

type Layanan struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"` 
	LimitDPT  int       `gorm:"not null" json:"limit_dpt"`              
	Price     float64   `gorm:"type:decimal(15,2)" json:"price"`        
	Features  string    `gorm:"type:text" json:"features"`              
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}