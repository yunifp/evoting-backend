package handlers

import (
	"evoting-backend/internal/config"
	"evoting-backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DPTInput struct {
	NIK  string `json:"nik" binding:"required,len=16"` // Wajib 16 digit
	Nama string `json:"nama" binding:"required"`
	NoHP string `json:"no_hp" binding:"required"`
}

// 1. CLIENT: Tambah DPT Satuan
func AddDPT(c *gin.Context) {
	clientID, _ := c.Get("userID")
	pemiluID := c.Param("pemiluId")

	var input DPTInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// VALIDASI KEAMANAN: Pastikan Client adalah pemilik event pemilu ini
	var pemilu models.Pemilu
	if err := config.DB.Where("id = ? AND client_id = ?", pemiluID, clientID).First(&pemilu).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Event Pemilu tidak ditemukan atau bukan milik Anda"})
		return
	}

	// Validasi NIK: NIK tidak boleh ganda di dalam 1 Event Pemilu yang sama
	var existingDPT models.DPT
	if err := config.DB.Where("pemilu_id = ? AND nik = ?", pemilu.ID, input.NIK).First(&existingDPT).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "NIK ini sudah terdaftar di daftar pemilih event ini"})
		return
	}

	newDPT := models.DPT{
		ID:       uuid.New().String(),
		PemiluID: pemilu.ID,
		NIK:      input.NIK,
		Nama:     input.Nama,
		NoHP:     input.NoHP,
	}

	if err := config.DB.Create(&newDPT).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menambahkan data pemilih"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Berhasil menambahkan data pemilih (DPT)",
		"data":    newDPT,
	})
}

// 2. CLIENT: Lihat Daftar DPT di Pemilu Tertentu
func GetDPTByPemilu(c *gin.Context) {
	clientID, _ := c.Get("userID")
	pemiluID := c.Param("pemiluId")

	// Validasi akses Client
	var pemilu models.Pemilu
	if err := config.DB.Where("id = ? AND client_id = ?", pemiluID, clientID).First(&pemilu).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Event Pemilu tidak valid"})
		return
	}

	var dptList []models.DPT
	// Kita tidak perlu mem-preload Pemilu di sini agar payload JSON tidak terlalu bengkak
	if err := config.DB.Where("pemilu_id = ?", pemilu.ID).Order("nama asc").Find(&dptList).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data DPT"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengambil data DPT",
		"data":    dptList,
	})
}