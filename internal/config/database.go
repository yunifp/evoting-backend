// internal/config/database.go
package config

import (
	"evoting-backend/internal/models"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file, using system environment variables")
	}

	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPass, dbHost, dbPort, dbName)

	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Gagal terkoneksi ke database! Error: ", err)
	}

	log.Println("Koneksi ke database MySQL berhasil!")
	DB = database

	err = DB.AutoMigrate(
		&models.Menu{},
		&models.Permission{},
		&models.Role{},
		&models.RolePermission{},
		&models.User{},
		&models.UserRole{},
		&models.Layanan{},
		&models.Transaction{},
		&models.Pemilu{},
		&models.Kandidat{},
		&models.DPT{},
		&models.Setting{},
	)
	if err != nil {
		log.Fatal("Gagal melakukan migrasi database: ", err)
	}
	log.Println("Migrasi database berhasil!")
}