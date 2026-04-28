package handlers

import (
	"evoting-backend/internal/config"
	"evoting-backend/internal/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type PemiluInput struct {
	Title     string `json:"title" binding:"required"`
	StartDate string `json:"start_date" binding:"required"` 
	EndDate   string `json:"end_date" binding:"required"`
}

func CreatePemilu(c *gin.Context) {
	clientID, _ := c.Get("userID")

	// 1. Cari transaksi LUNAS terakhir
	var latestTrx models.Transaction
	if err := config.DB.Where("user_id = ? AND status = ?", clientID, "paid").Order("paid_at desc").First(&latestTrx).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses Ditolak: Anda harus memiliki paket aktif (Lunas) untuk membuat acara Pemilu."})
		return
	}

	// 2. VALIDASI KRUSIAL: Cek apakah transaksi ini sudah dipakai untuk membuat acara lain
	var checkPemilu models.Pemilu
	if err := config.DB.Where("transaction_id = ?", latestTrx.ID).First(&checkPemilu).Error; err == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Satu paket layanan hanya berlaku untuk satu acara pemilu. Silakan beli paket baru."})
		return
	}

	var input PemiluInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	layoutFormat := "2006-01-02 15:04:05" 
	startTime, err := time.Parse(layoutFormat, input.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format start_date salah"})
		return
	}

	endTime, err := time.Parse(layoutFormat, input.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format end_date salah"})
		return
	}

	newPemilu := models.Pemilu{
		ClientID:      clientID.(string),
		TransactionID: latestTrx.ID, // IKAT TRANSAKSI DI SINI
		Title:         input.Title,
		StartDate:     startTime,
		EndDate:       endTime,
		Status:        "draft",
	}

	if err := config.DB.Create(&newPemilu).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat Event Pemilu"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Berhasil membuat ruang Event Pemilu", "data": newPemilu})
}

func GetMyPemilu(c *gin.Context) {
	clientID, _ := c.Get("userID")
	var pemiluList []models.Pemilu
	if err := config.DB.Where("client_id = ?", clientID).Preload("Kandidats").Order("created_at desc").Find(&pemiluList).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data pemilu"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Berhasil mengambil data Event Pemilu", "data": pemiluList})
}

func GetPemiluDetail(c *gin.Context) {
	clientID, _ := c.Get("userID")
	pemiluID := c.Param("pemiluId")

	var pemilu models.Pemilu
	if err := config.DB.Preload("Kandidats").Where("id = ? AND client_id = ?", pemiluID, clientID).First(&pemilu).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event Pemilu tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Berhasil mengambil detail", "data": pemilu})
}

// FUNGSI BARU: Untuk Mengaktifkan Acara (Publish)
func PublishPemilu(c *gin.Context) {
	clientID, _ := c.Get("userID")
	pemiluID := c.Param("pemiluId")

	var pemilu models.Pemilu
	if err := config.DB.Where("id = ? AND client_id = ?", pemiluID, clientID).First(&pemilu).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event Pemilu tidak ditemukan"})
		return
	}

	if pemilu.Status != "draft" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Event ini sudah aktif atau selesai"})
		return
	}

	pemilu.Status = "active"
	if err := config.DB.Save(&pemilu).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengaktifkan acara"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Acara berhasil diaktifkan! Pemilih sekarang dapat login."})
}