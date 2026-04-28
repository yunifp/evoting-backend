package seeders

import (
	"evoting-backend/internal/models"
	"log"

	"gorm.io/gorm"
)

func SeedLayanan(db *gorm.DB) {
	layanans := []models.Layanan{
		{
			Name:     "Paket FREE (Uji Coba & Komunitas Mikro)",
			LimitDPT: 25,
			Price:    0,
			Features: "Akses Web Panitia, Login Aplikasi, Keamanan Standard, Hasil Real-time, Masa Aktif 12 Jam, Watermark, Support Self-service",
			IsActive: true,
		},
		{
			Name:     "Paket KEBERSAMAAN (Segmen Komunitas/Kecil)",
			LimitDPT: 250,
			Price:    750000,
			Features: "Sistem Web Panitia, Login App via ID & Password, Hasil Real-time, Masa Aktif 24 Jam, Support Chat & Video",
			IsActive: true,
		},
		{
			Name:     "Paket AKADEMIK (Segmen Sekolah/Kampus)",
			LimitDPT: 1500,
			Price:    2500000,
			Features: "Semua fitur Paket Kebersamaan, Custom Branding, Visi-Misi & Foto Kandidat, Export Hasil (Excel/PDF), Masa Aktif 3 Hari",
			IsActive: true,
		},
		{
			Name:     "Paket PROFESIONAL (Segmen Korporasi/Asosiasi)",
			LimitDPT: 5000,
			Price:    7500000,
			Features: "Semua fitur Paket Akademik, Face Recognition & Liveness Detection, Log Audit Keamanan, Notifikasi Email/Push, Sertifikat Hasil QR Code",
			IsActive: true,
		},
		{
			Name:     "Paket INSTITUSI (Segmen Pemerintahan/Besar)",
			LimitDPT: 15000,
			Price:    25000000,
			Features: "Semua fitur Paket Profesional, Server Dedicated, Import Data DPT, Multi-admin, Laporan SPJ, Priority Support 24/7",
			IsActive: true,
		},
	}

	for _, layanan := range layanans {
		var count int64
		db.Model(&models.Layanan{}).Where("name = ?", layanan.Name).Count(&count)
		if count == 0 {
			if err := db.Create(&layanan).Error; err != nil {
				log.Fatalf("Gagal melakukan seeding data layanan: %v", err)
			}
		}
	}
}