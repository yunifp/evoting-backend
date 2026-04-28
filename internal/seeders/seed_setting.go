// internal/seeders/seed_setting.go
package seeders

import (
	"evoting-backend/internal/models"
	"log"

	"gorm.io/gorm"
)

func SeedSettings(db *gorm.DB) {
	settings := []models.Setting{
		{Key: "midtrans_server_key", Value: "SB-Mid-server-p6andJUaTu7JFH_cAE-6wLJk"},
		{Key: "midtrans_client_key", Value: "SB-Mid-client-9u0tqdfPq3HrieJw"},
		{Key: "midtrans_is_production", Value: "false"},
	}

	for _, setting := range settings {
		var count int64
		db.Model(&models.Setting{}).Where("`key` = ?", setting.Key).Count(&count)
		if count == 0 {
			if err := db.Create(&setting).Error; err != nil {
				log.Printf("Gagal melakukan seeding setting '%s': %v", setting.Key, err)
			} else {
				log.Printf("Seeding setting '%s' berhasil.", setting.Key)
			}
		} else {
			log.Printf("Setting '%s' sudah ada, dilewati.", setting.Key)
		}
	}
}