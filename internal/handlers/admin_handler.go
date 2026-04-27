package handlers

import (
	"evoting-backend/internal/config"
	"evoting-backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ApproveClient(c *gin.Context) {
	clientID := c.Param("id") // Ambil ID dari URL (contoh: /api/admin/clients/:id/approve)

	var user models.User
	if err := config.DB.Where("id = ?", clientID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
		return
	}

	user.IsApproved = true
	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal meng-approve user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil meng-approve client: " + user.Name,
	})
}