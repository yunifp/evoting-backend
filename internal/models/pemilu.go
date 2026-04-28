package models

import (
	"time"
)

type Pemilu struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ClientID      string    `gorm:"type:char(36);not null" json:"client_id"` 
	TransactionID string    `gorm:"type:char(36)" json:"transaction_id"` // TAMBAHAN: Mengikat 1 transaksi = 1 acara
	Title         string    `gorm:"type:varchar(150);not null" json:"title"` 
	StartDate     time.Time `gorm:"not null" json:"start_date"`              
	EndDate       time.Time `gorm:"not null" json:"end_date"`                
	Status        string    `gorm:"type:varchar(20);default:'draft'" json:"status"` 
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	Client    User       `gorm:"foreignKey:ClientID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"client,omitempty"`
	Kandidats []Kandidat `gorm:"foreignKey:PemiluID" json:"kandidats,omitempty"`
}

type Kandidat struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	PemiluID uint   `gorm:"not null" json:"pemilu_id"`
	NoUrut   int    `gorm:"not null" json:"no_urut"`
	Name     string `gorm:"type:varchar(100);not null" json:"name"`
	Visi     string `gorm:"type:text" json:"visi"`
	Misi     string `gorm:"type:text" json:"misi"`
	PhotoURL string `gorm:"type:varchar(255)" json:"photo_url"` 

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Pemilu Pemilu `gorm:"foreignKey:PemiluID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"pemilu,omitempty"`
}