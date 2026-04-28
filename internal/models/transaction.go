package models

import (
	"time"
)

type Transaction struct {
	ID            string     `gorm:"type:char(36);primaryKey" json:"id"`
	UserID        string     `gorm:"type:char(36);not null" json:"user_id"`
	LayananID     uint       `gorm:"not null" json:"layanan_id"`
	Amount        float64    `gorm:"type:decimal(15,2);not null" json:"amount"`
	Status        string     `gorm:"type:varchar(20);default:'pending'" json:"status"`
	PaymentMethod string     `gorm:"type:varchar(50)" json:"payment_method"`
	SnapToken     string     `gorm:"type:varchar(255)" json:"snap_token"`
	PaidAt        *time.Time `json:"paid_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	User          User       `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user,omitempty"`
	Layanan       Layanan    `gorm:"foreignKey:LayananID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"layanan,omitempty"`
}