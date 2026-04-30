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
    IsFaceRecognition *bool   `json:"is_face_recognition"` // Tambahan
    IsActive          *bool   `json:"is_active"`
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
        IsFaceRecognition: isFaceRec, // Set nilai
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