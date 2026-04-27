package models

import (
	"time"
)

type DPT struct {
	ID           string     `gorm:"type:char(36);primaryKey" json:"id"` 
	PemiluID     uint       `gorm:"not null;index" json:"pemilu_id"`    
	NIK          string     `gorm:"type:varchar(16);not null" json:"nik"`
	Nama         string     `gorm:"type:varchar(100);not null" json:"nama"`
	NoHP         string     `gorm:"type:varchar(15)" json:"no_hp"`      
	StatusVoted  bool       `gorm:"default:false" json:"status_voted"` 
	OTPToken     string     `gorm:"type:varchar(10)" json:"-"`        
	OTPExpiredAt *time.Time `json:"otp_expired_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Pemilu Pemilu `gorm:"foreignKey:PemiluID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"pemilu,omitempty"`
}