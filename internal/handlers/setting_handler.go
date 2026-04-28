package handlers

import (
	"evoting-backend/internal/config"
	"evoting-backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetMidtransConfig(c *gin.Context) {
	var clientKey models.Setting
	config.DB.Where("`key` = ?", "midtrans_client_key").First(&clientKey)

	var isProduction models.Setting
	config.DB.Where("`key` = ?", "midtrans_is_production").First(&isProduction)

	isProdBool := false
	if isProduction.Value == "true" {
		isProdBool = true
	}

	c.JSON(http.StatusOK, gin.H{
		"client_key":    clientKey.Value,
		"is_production": isProdBool,
	})
}