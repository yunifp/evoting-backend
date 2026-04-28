package seeders

import (
	"evoting-backend/internal/models"
	"log"

	"gorm.io/gorm"
)

// RunSeeder dipanggil dari main.go untuk memastikan data esensial ada
func RunSeeder(db *gorm.DB) {
	log.Println("Menjalankan Seeder...")
	SeedRoles(db)
	SeedMenusAndPermissions(db)
	log.Println("Seeder Selesai!")
}

// 1. Seed Roles Dasar
func SeedRoles(db *gorm.DB) {
	roles := []models.Role{
		{Name: "Superadmin", Description: "Akses penuh sistem tak terbatas"},
		{Name: "Admin", Description: "Pengelola layanan dan pengguna"},
		{Name: "Client", Description: "Instansi penyelenggara pemilu"},
		{Name: "Voter", Description: "Pemilih akhir"},
		{Name: "Auditor", Description: "Pemantau log transaksi"},
	}

	for _, role := range roles {
		db.Where("name = ?", role.Name).FirstOrCreate(&role)
	}
}

// 2. Seed Menus & Permissions secara Presisi
func SeedMenusAndPermissions(db *gorm.DB) {
	// Definisi struktur menu dan permission yang 100% presisi untuk RBAC kita
	menusData := []struct {
		Menu  models.Menu
		Perms []models.Permission
	}{
		{
			Menu: models.Menu{Name: "Dashboard", Path: "/dashboard", Icon: "layout-dashboard", SortOrder: 1, IsActive: true},
			Perms: []models.Permission{
				{Name: "Lihat Dashboard", Action: "read"},
			},
		},
		{
			Menu: models.Menu{Name: "Manajemen Layanan", Path: "/dashboard/layanan", Icon: "box", SortOrder: 2, IsActive: true},
			Perms: []models.Permission{
				{Name: "Lihat Layanan", Action: "read"},
				{Name: "Tambah Layanan", Action: "create"},
				{Name: "Edit Layanan", Action: "update"},
				{Name: "Hapus Layanan", Action: "delete"},
			},
		},
		{
			Menu: models.Menu{Name: "Manajemen Pengguna", Path: "/dashboard/users", Icon: "users", SortOrder: 3, IsActive: true},
			Perms: []models.Permission{
				{Name: "Lihat Pengguna", Action: "read"},
				{Name: "Tambah Pengguna", Action: "create"},
				{Name: "Edit Pengguna", Action: "update"},
				{Name: "Hapus Pengguna", Action: "delete"},
				{Name: "Setujui/Approve Pengguna", Action: "approve"},
			},
		},
		{
			Menu: models.Menu{Name: "Manajemen Transaksi", Path: "/dashboard/transactions", Icon: "credit-card", SortOrder: 4, IsActive: true},
			Perms: []models.Permission{
				{Name: "Lihat Transaksi", Action: "read"},
				{Name: "Verifikasi Transaksi", Action: "approve"},
			},
		},
		{
			Menu: models.Menu{Name: "Manajemen Role", Path: "/dashboard/roles", Icon: "shield", SortOrder: 5, IsActive: true},
			Perms: []models.Permission{
				{Name: "Lihat Role", Action: "read"},
				{Name: "Tambah Role", Action: "create"},
				{Name: "Edit Role", Action: "update"},
				{Name: "Hapus Role", Action: "delete"},
				{Name: "Kelola Akses Role", Action: "assign_permission"},
			},
		},
		{
			Menu: models.Menu{Name: "Manajemen Hak Akses", Path: "/dashboard/permissions", Icon: "key", SortOrder: 6, IsActive: true},
			Perms: []models.Permission{
				{Name: "Lihat Hak Akses", Action: "read"},
				{Name: "Tambah Hak Akses", Action: "create"},
				{Name: "Edit Hak Akses", Action: "update"},
				{Name: "Hapus Hak Akses", Action: "delete"},
			},
		},
		{
			Menu: models.Menu{Name: "Manajemen Menu", Path: "/dashboard/menus", Icon: "list", SortOrder: 7, IsActive: true},
			Perms: []models.Permission{
				{Name: "Lihat Menu", Action: "read"},
				{Name: "Tambah Menu", Action: "create"},
				{Name: "Edit Menu", Action: "update"},
				{Name: "Hapus Menu", Action: "delete"},
			},
		},
	}

	for _, item := range menusData {
		var menu models.Menu
		// Buat menu jika belum ada berdasarkan namanya
		if err := db.Where("name = ?", item.Menu.Name).FirstOrCreate(&menu, item.Menu).Error; err != nil {
			log.Printf("Gagal seed menu %s: %v", item.Menu.Name, err)
			continue
		}

		// Update path & icon untuk berjaga-jaga jika ada perubahan
		db.Model(&menu).Updates(item.Menu)

		// Buat permission untuk menu tersebut
		for _, p := range item.Perms {
			p.MenuID = menu.ID
			var perm models.Permission
			// Cari berdasarkan menu_id dan action
			db.Where("menu_id = ? AND action = ?", menu.ID, p.Action).FirstOrCreate(&perm, p)
			
			// Update nama permission agar rapi
			db.Model(&perm).Updates(p)
		}
	}
}