package handlers

import (
	"evoting-backend/internal/config"
	"evoting-backend/internal/models"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type LayananInput struct {
	Name              string  `json:"name" binding:"required"`
	LimitDPT          int     `json:"limit_dpt" binding:"required,min=1"`
	Price             float64 `json:"price" binding:"required,min=0"`
	Features          string  `json:"features"`
	IsFaceRecognition *bool   `json:"is_face_recognition"`
	IsActive          *bool   `json:"is_active"`
}

type LayananItem struct {
	TransactionID string         `json:"transaction_id"`
	Layanan       models.Layanan `json:"layanan"`
	PemiluID      *uint          `json:"pemilu_id,omitempty"`
	PemiluTitle   string         `json:"pemilu_title,omitempty"`
	PemiluStatus  string         `json:"pemilu_status,omitempty"`
}

type StatusLayananResponse struct {
	Tersedia  []LayananItem `json:"tersedia"`
	Digunakan []LayananItem `json:"digunakan"`
	Selesai   []LayananItem `json:"selesai"`
}

func GetLayanan(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))
	search := c.Query("search")
	minDpt := c.Query("min_dpt")
	maxDpt := c.Query("max_dpt")

	offset := (page - 1) * limit
	var layanans []models.Layanan
	var totalItems int64

	query := config.DB.Model(&models.Layanan{})

	if search != "" {
		query = query.Where("name LIKE ?", "%"+search+"%")
	}
	if minDpt != "" {
		query = query.Where("limit_dpt >= ?", minDpt)
	}
	if maxDpt != "" {
		query = query.Where("limit_dpt <= ?", maxDpt)
	}

	query.Count(&totalItems)

	if err := query.Limit(limit).Offset(offset).Order("created_at desc").Find(&layanans).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data layanan"})
		return
	}

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

	isFaceRec := false
	if input.IsFaceRecognition != nil {
		isFaceRec = *input.IsFaceRecognition
	}

	newLayanan := models.Layanan{
		Name:              input.Name,
		LimitDPT:          input.LimitDPT,
		Price:             input.Price,
		Features:          input.Features,
		IsFaceRecognition: isFaceRec,
		IsActive:          isActive,
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

	if input.IsFaceRecognition != nil {
		layanan.IsFaceRecognition = *input.IsFaceRecognition
	}
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

func GetMyLayananStatus(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak valid"})
		return
	}

	var transactions []models.Transaction
	if err := config.DB.Preload("Layanan").Where("user_id = ? AND status = ?", userID, "paid").Order("paid_at desc").Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data transaksi"})
		return
	}

	var pemilus []models.Pemilu
	if err := config.DB.Where("client_id = ?", userID).Find(&pemilus).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data acara pemilu"})
		return
	}

	pemiluMap := make(map[string]models.Pemilu)
	for _, p := range pemilus {
		if p.TransactionID != "" {
			pemiluMap[p.TransactionID] = p
		}
	}

	response := StatusLayananResponse{
		Tersedia:  []LayananItem{},
		Digunakan: []LayananItem{},
		Selesai:   []LayananItem{},
	}

	for _, trx := range transactions {
		item := LayananItem{
			TransactionID: trx.ID,
			Layanan:       trx.Layanan,
		}

		if p, exists := pemiluMap[trx.ID]; exists {
			item.PemiluID = &p.ID
			item.PemiluTitle = p.Title
			item.PemiluStatus = p.Status

			if p.Status == "selesai" {
				response.Selesai = append(response.Selesai, item)
			} else {
				response.Digunakan = append(response.Digunakan, item)
			}
		} else {
			response.Tersedia = append(response.Tersedia, item)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengambil status layanan",
		"data":    response,
	})
}