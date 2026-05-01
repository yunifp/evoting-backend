package models

import (
	"time"
)

type DPT struct {
	ID                string     `gorm:"type:char(36);primaryKey" json:"id"`
	UserUUID          string     `gorm:"type:char(36);not null" json:"user_uuid"`
	PemiluID          uint       `gorm:"not null;index" json:"pemilu_id"`
	NIK               string     `gorm:"type:varchar(16);not null" json:"nik"`
	NKK               string     `gorm:"type:varchar(16)" json:"nkk"`
	Nama              string     `gorm:"type:varchar(100);not null" json:"nama"`
	NamaPenduduk      string     `gorm:"type:varchar(100)" json:"nama_penduduk"`
	TempatLahir       string     `gorm:"type:varchar(100)" json:"tempat_lahir"`
	JenisKelamin      string     `gorm:"type:varchar(20)" json:"jenis_kelamin"`
	Alamat            string     `gorm:"type:text" json:"alamat"`
	RT                string     `gorm:"type:varchar(5)" json:"rt"`
	RW                string     `gorm:"type:varchar(5)" json:"rw"`
	KodePro           string     `gorm:"type:varchar(10)" json:"kode_pro"`
	NamaPro           string     `gorm:"type:varchar(100)" json:"nama_pro"`
	KodeKab           string     `gorm:"type:varchar(10)" json:"kode_kab"`
	NamaKab           string     `gorm:"type:varchar(100)" json:"nama_kab"`
	KodeKec           string     `gorm:"type:varchar(10)" json:"kode_kec"`
	NamaKec           string     `gorm:"type:varchar(100)" json:"nama_kec"`
	KodeKel           string     `gorm:"type:varchar(10)" json:"kode_kel"`
	NamaDesa          string     `gorm:"type:varchar(100)" json:"nama_desa"`
	StatusKawin       string     `gorm:"type:varchar(50)" json:"status_kawin"`
	StatusDisabilitas string     `gorm:"type:varchar(50)" json:"status_disabilitas"`
	NoHP              string     `gorm:"type:varchar(15)" json:"no_hp"`
	FaceTemplate      string     `gorm:"type:text" json:"face_template,omitempty"`
	StatusVoted       bool       `gorm:"default:false" json:"status_voted"`
	OTPToken          string     `gorm:"type:varchar(10)" json:"-"`
	OTPExpiredAt      *time.Time `json:"otp_expired_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	Pemilu            Pemilu     `gorm:"foreignKey:PemiluID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"pemilu,omitempty"`
}