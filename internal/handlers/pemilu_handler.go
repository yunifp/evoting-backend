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

	var input PemiluInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	layoutFormat := "2006-01-02 15:04:05" 
	startTime, err := time.Parse(layoutFormat, input.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format start_date salah. Gunakan YYYY-MM-DD HH:MM:SS"})
		return
	}

	endTime, err := time.Parse(layoutFormat, input.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format end_date salah. Gunakan YYYY-MM-DD HH:MM:SS"})
		return
	}

	newPemilu := models.Pemilu{
		ClientID:  clientID.(string),
		Title:     input.Title,
		StartDate: startTime,
		EndDate:   endTime,
		Status:    "draft",
	}

	if err := config.DB.Create(&newPemilu).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat Event Pemilu"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Berhasil membuat ruang Event Pemilu",
		"data":    newPemilu,
	})
}

func GetMyPemilu(c *gin.Context) {
	clientID, _ := c.Get("userID")

	var pemiluList []models.Pemilu
	if err := config.DB.Where("client_id = ?", clientID).Preload("Kandidats").Order("created_at desc").Find(&pemiluList).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data pemilu"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengambil data Event Pemilu",
		"data":    pemiluList,
	})
}