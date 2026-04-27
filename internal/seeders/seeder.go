package seeders

import (
	"evoting-backend/internal/models"
	"log"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func RunSeeder(db *gorm.DB) {
	// Cek apakah data Role sudah ada, jika ada hentikan seeder agar tidak duplikat
	var count int64
	db.Model(&models.Role{}).Count(&count)
	if count > 0 {
		log.Println("Seeder dilewati: Data dasar sudah ada di database.")
		return
	}

	log.Println("Menjalankan database seeder...")

	// 1. Seed Roles
	superadmin := models.Role{Name: "Superadmin", Description: "Akses penuh sistem"}
	admin := models.Role{Name: "Admin", Description: "Pengelola layanan e-voting"}
	client := models.Role{Name: "Client", Description: "Penyelenggara pemilihan umum"}
	voter := models.Role{Name: "Voter", Description: "Pemilih via aplikasi Android"}

	db.Create(&superadmin)
	db.Create(&admin)
	db.Create(&client)
	db.Create(&voter)

	// 2. Seed Menus
	dashMenu := models.Menu{Name: "Dashboard", Path: "/dashboard", Icon: "layout-dashboard"}
	layananMenu := models.Menu{Name: "Manajemen Layanan", Path: "/dashboard/layanan", Icon: "box"}
	clientMenu := models.Menu{Name: "Manajemen Client", Path: "/dashboard/clients", Icon: "users"}
	
	db.Create(&dashMenu)
	db.Create(&layananMenu)
	db.Create(&clientMenu)

	// 3. Seed Permissions (Aksi Dinamis)
	perms := []models.Permission{
		{MenuID: dashMenu.ID, Name: "Lihat Dashboard", Action: "read"},
		{MenuID: layananMenu.ID, Name: "Lihat Layanan", Action: "read"},
		{MenuID: layananMenu.ID, Name: "Tambah Layanan", Action: "create"},
		{MenuID: layananMenu.ID, Name: "Edit Layanan", Action: "update"},
		{MenuID: layananMenu.ID, Name: "Hapus Layanan", Action: "delete"},
		{MenuID: clientMenu.ID, Name: "Lihat Client", Action: "read"},
		{MenuID: clientMenu.ID, Name: "Approve Client", Action: "approve"},
	}
	db.Create(&perms)

	// 4. Attach Semua Permissions ke Superadmin (Pivot: role_permissions)
	// Kita berikan semua hak akses ke Superadmin
	err := db.Model(&superadmin).Association("Permissions").Append(perms)
	if err != nil {
		log.Println("Gagal attach permissions ke superadmin:", err)
	}

	// 5. Seed Akun Default Superadmin
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	superUser := models.User{
		ID:         uuid.New().String(),
		Name:       "Superadmin Evoting",
		Email:      "superadmin@evoting.com",
		Password:   string(hashedPassword),
		IsApproved: true,
		IsActive:   true,
	}
	db.Create(&superUser)

	// 6. Attach Role 'Superadmin' ke akun Super User (Pivot: user_roles)
	db.Model(&superUser).Association("Roles").Append([]models.Role{superadmin})

	log.Println("Seeder berhasil dieksekusi! Akun superadmin@evoting.com siap digunakan.")
}