package handlers

import (
	"evoting-backend/internal/config"
	"evoting-backend/internal/models"
	"net/http"
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
)

// DTO untuk validasi input
type LayananInput struct {
	Name     string  `json:"name" binding:"required"`
	LimitDPT int     `json:"limit_dpt" binding:"required,min=1"`
	Price    float64 `json:"price" binding:"required,min=0"`
	Features string  `json:"features"`
	IsActive *bool   `json:"is_active"` // Menggunakan pointer agar bisa deteksi boolean false
}

// 1. GET ALL LAYANAN
func GetLayanan(c *gin.Context) {
	// Ambil query parameter dengan nilai default
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5")) // Tampilkan 5 data per halaman
	search := c.Query("search")

	// Hitung offset
	offset := (page - 1) * limit

	var layanans []models.Layanan
	var totalItems int64

	// Mulai query ke tabel layanans
	query := config.DB.Model(&models.Layanan{})

	// Jika ada pencarian berdasarkan nama
	if search != "" {
		query = query.Where("name LIKE ?", "%"+search+"%")
	}

	// Hitung total data (setelah difilter search)
	query.Count(&totalItems)

	// Ambil data dengan limit dan offset
	if err := query.Limit(limit).Offset(offset).Order("created_at desc").Find(&layanans).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data layanan"})
		return
	}

	// Hitung total halaman
	totalPages := int(math.Ceil(float64(totalItems) / float64(limit)))

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengambil data layanan",
		"data":    layanans,
		"meta": gin.H{
			"current_page": page,
			"total_pages":  totalPages,
			"total_items":  totalItems,
			"limit":        limit,
		},
	})
}

// 2. CREATE LAYANAN
func CreateLayanan(c *gin.Context) {
	var input LayananInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	newLayanan := models.Layanan{
		Name:     input.Name,
		LimitDPT: input.LimitDPT,
		Price:    input.Price,
		Features: input.Features,
		IsActive: isActive,
	}

	if err := config.DB.Create(&newLayanan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat layanan"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Berhasil membuat layanan baru",
		"data":    newLayanan,
	})
}

// 3. UPDATE LAYANAN
func UpdateLayanan(c *gin.Context) {
	id := c.Param("id")
	var input LayananInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var layanan models.Layanan
	if err := config.DB.Where("id = ?", id).First(&layanan).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Layanan tidak ditemukan"})
		return
	}

	layanan.Name = input.Name
	layanan.LimitDPT = input.LimitDPT
	layanan.Price = input.Price
	layanan.Features = input.Features
	if input.IsActive != nil {
		layanan.IsActive = *input.IsActive
	}

	if err := config.DB.Save(&layanan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengupdate layanan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengupdate layanan",
		"data":    layanan,
	})
}

// 4. DELETE LAYANAN
func DeleteLayanan(c *gin.Context) {
	id := c.Param("id")

	var layanan models.Layanan
	if err := config.DB.Where("id = ?", id).First(&layanan).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Layanan tidak ditemukan"})
		return
	}

	if err := config.DB.Delete(&layanan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus layanan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil menghapus layanan",
	})
}