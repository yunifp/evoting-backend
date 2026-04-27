package handlers

import (
	"evoting-backend/internal/config"
	"evoting-backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type KandidatInput struct {
	NoUrut int    `json:"no_urut" binding:"required"`
	Name   string `json:"name" binding:"required"`
	Visi   string `json:"visi"`
	Misi   string `json:"misi"`
	PhotoURL string `json:"photo_url"` 
}

func AddKandidat(c *gin.Context) {
	clientID, _ := c.Get("userID")
	pemiluID := c.Param("pemiluId") 

	var input KandidatInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var pemilu models.Pemilu
	if err := config.DB.Where("id = ? AND client_id = ?", pemiluID, clientID).First(&pemilu).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses ditolak: Event Pemilu ini bukan milik Anda atau tidak ditemukan"})
		return
	}

	newKandidat := models.Kandidat{
		PemiluID: pemilu.ID,
		NoUrut:   input.NoUrut,
		Name:     input.Name,
		Visi:     input.Visi,
		Misi:     input.Misi,
		PhotoURL: input.PhotoURL,
	}

	if err := config.DB.Create(&newKandidat).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menambahkan kandidat"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Berhasil menambahkan kandidat",
		"data":    newKandidat,
	})
}

func DeleteKandidat(c *gin.Context) {
	clientID, _ := c.Get("userID")
	kandidatID := c.Param("id")

	var kandidat models.Kandidat
	if err := config.DB.Preload("Pemilu").Where("id = ?", kandidatID).First(&kandidat).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kandidat tidak ditemukan"})
		return
	}

	if kandidat.Pemilu.ClientID != clientID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses ditolak: Anda tidak berhak menghapus kandidat ini"})
		return
	}

	if err := config.DB.Delete(&kandidat).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus kandidat"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil menghapus kandidat",
	})
}