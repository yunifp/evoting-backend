package seeders

import (
	"evoting-backend/internal/models"
	"log"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedDummyData mengisi tabel dengan data awal untuk keperluan testing
func SeedDummyData(db *gorm.DB) {
	log.Println("Menjalankan Seeder Dummy Data...")

	seedLayanan(db)
	seedUsersAndPemilu(db)

	log.Println("Seeder Dummy Data Selesai!")
}

func seedLayanan(db *gorm.DB) {
	layanans := []models.Layanan{
		{Name: "Paket RT/RW", LimitDPT: 500, Price: 150000, Features: "Maks 500 DPT, 3 Kandidat, Support WA", IsActive: true},
		{Name: "Paket Desa/Kampus", LimitDPT: 2000, Price: 500000, Features: "Maks 2000 DPT, 5 Kandidat, Support WA & Email, Laporan PDF", IsActive: true},
		{Name: "Paket Custom", LimitDPT: 10000, Price: 2000000, Features: "DPT Unlimited, Kandidat Unlimited, Prioritas Support", IsActive: true},
	}

	for _, l := range layanans {
		db.Where("name = ?", l.Name).FirstOrCreate(&l)
	}
}

func seedUsersAndPemilu(db *gorm.DB) {
	// 1. Ambil Role yang sudah ada
	var superadminRole, adminRole, clientRole models.Role
	db.Where("name = ?", "Superadmin").First(&superadminRole)
	db.Where("name = ?", "Admin").First(&adminRole)
	db.Where("name = ?", "Client").First(&clientRole)

	// Hash password default "rahasia123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("rahasia123"), bcrypt.DefaultCost)

	// 2. Buat User Superadmin
	superadmin := models.User{
		ID:         uuid.New().String(),
		Name:       "Superadmin Utama",
		Email:      "superadmin@evoting.com",
		Password:   string(hashedPassword),
		IsApproved: true,
		IsActive:   true,
	}
	if err := db.Where("email = ?", superadmin.Email).FirstOrCreate(&superadmin).Error; err == nil {
		// Assign role jika baru dibuat
		if db.Model(&superadmin).Association("Roles").Count() == 0 {
			db.Model(&superadmin).Association("Roles").Append(&superadminRole)
		}
	}

	// 3. Buat User Client (Penyelenggara Pemilu)
	client := models.User{
		ID:         uuid.New().String(),
		Name:       "BEM Universitas XYZ",
		Email:      "bemxyz@evoting.com",
		Password:   string(hashedPassword),
		IsApproved: true, // Langsung di-approve agar bisa login
		IsActive:   true,
	}
	if err := db.Where("email = ?", client.Email).FirstOrCreate(&client).Error; err == nil {
		if db.Model(&client).Association("Roles").Count() == 0 {
			db.Model(&client).Association("Roles").Append(&clientRole)
		}
	}

	// === DATA PEMILU (Hanya buat jika Client sudah ada) ===
	
	// Cek apakah pemilu sudah ada untuk client ini
	var countPemilu int64
	db.Model(&models.Pemilu{}).Where("client_id = ?", client.ID).Count(&countPemilu)

	if countPemilu == 0 {
		// 4. Berikan Transaksi Dummy (Client Beli Paket)
		var layanan models.Layanan
		db.First(&layanan) // Ambil layanan pertama
		
		now := time.Now()
		trx := models.Transaction{
			ID:            uuid.New().String(),
			UserID:        client.ID,
			LayananID:     layanan.ID,
			Amount:        layanan.Price,
			Status:        "paid",
			PaymentMethod: "Bank Transfer",
			PaidAt:        &now,
		}
		db.Create(&trx)

		// 5. Buat Event Pemilu Dummy
		pemilu := models.Pemilu{
			ClientID:  client.ID,
			Title:     "Pemilihan Presiden Mahasiswa 2026",
			StartDate: time.Now().Add(-24 * time.Hour), // Dimulai kemarin
			EndDate:   time.Now().Add(48 * time.Hour),  // Berakhir 2 hari lagi
			Status:    "active",
		}
		db.Create(&pemilu)

		// 6. Buat Kandidat Dummy
		kandidats := []models.Kandidat{
			{PemiluID: pemilu.ID, NoUrut: 1, Name: "Andi & Budi", Visi: "Kampus Inovatif", Misi: "1. Digitalisasi 2. Transparansi"},
			{PemiluID: pemilu.ID, NoUrut: 2, Name: "Citra & Dewi", Visi: "Kampus Hijau", Misi: "1. Ekologi 2. Kesejahteraan"},
		}
		db.Create(&kandidats)

		// 7. Buat DPT Dummy
		dpts := []models.DPT{
			{ID: uuid.New().String(), PemiluID: pemilu.ID, NIK: "1234567890123456", Nama: "Voter Satu", NoHP: "081234567890"},
			{ID: uuid.New().String(), PemiluID: pemilu.ID, NIK: "1234567890123457", Nama: "Voter Dua", NoHP: "081234567891"},
			{ID: uuid.New().String(), PemiluID: pemilu.ID, NIK: "1234567890123458", Nama: "Voter Tiga", NoHP: "081234567892"},
		}
		db.Create(&dpts)
	}
}