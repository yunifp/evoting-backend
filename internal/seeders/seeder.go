package seeders

import (
	"evoting-backend/internal/models"
	"log"

	"gorm.io/gorm"
)

func RunSeeder(db *gorm.DB) {
	log.Println("Memulai proses Seeder...")
	
	SeedRoles(db)
	SeedMenusAndPermissions(db)
	SeedRolePermissions(db) // Tambahkan ini agar hak akses terhubung ke role
	SeedDummyData(db)

	log.Println("Seluruh Seeder Selesai!")
}

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

func SeedMenusAndPermissions(db *gorm.DB) {
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
		if err := db.Where("name = ?", item.Menu.Name).FirstOrCreate(&menu, item.Menu).Error; err != nil {
			log.Printf("Gagal seed menu %s: %v", item.Menu.Name, err)
			continue
		}
		db.Model(&menu).Updates(item.Menu)

		for _, p := range item.Perms {
			p.MenuID = menu.ID
			var perm models.Permission
			db.Where("menu_id = ? AND action = ?", menu.ID, p.Action).FirstOrCreate(&perm, p)
			db.Model(&perm).Updates(p)
		}
	}
}

// SeedRolePermissions menghubungkan Role dengan Hak Aksesnya
func SeedRolePermissions(db *gorm.DB) {
	log.Println("Menghubungkan Role dan Permission...")

	var adminRole models.Role
	db.Where("name = ?", "Admin").First(&adminRole)

	// Kita beri role Admin semua permission yang ada di database
	// agar dia bisa mengelola sistem (kecuali jika Anda ingin membatasi tertentu)
	var allPermissions []models.Permission
	db.Find(&allPermissions)

	if adminRole.ID != 0 {
		// Gunakan Association Replace agar data tidak duplikat jika di-run berkali-kali
		db.Model(&adminRole).Association("Permissions").Replace(allPermissions)
	}

	// Untuk Superadmin tidak wajib di-seed di pivot karena sudah ada bypass hardcode 
	// di rbac_middleware.go, tapi bagus untuk integritas data jika ingin dimasukkan juga.
}

func ClearDatabase(db *gorm.DB) {
	log.Println("Membersihkan tabel database...")
	db.Exec("SET FOREIGN_KEY_CHECKS = 0;")
	tables := []string{
		"dpts", "kandidats", "pemilus", "transactions", "layanans",
		"user_roles", "users", "role_permissions", "roles", "permissions", "menus",
	}
	for _, table := range tables {
		db.Exec("TRUNCATE TABLE " + table)
	}
	db.Exec("SET FOREIGN_KEY_CHECKS = 1;")
	log.Println("Database berhasil dibersihkan!")
}