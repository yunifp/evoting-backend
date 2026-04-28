package models

import "time"

type Setting struct {
	Key       string    `gorm:"type:varchar(50);primaryKey" json:"key"`
	Value     string    `gorm:"type:text;not null" json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}