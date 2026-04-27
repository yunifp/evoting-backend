package handlers

import (
	"evoting-backend/internal/config"
	"evoting-backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// DTO untuk input form Permission
type PermissionInput struct {
	MenuID      uint   `json:"menu_id" binding:"required"`
	Name        string `json:"name" binding:"required"`   // cth: "Tambah Layanan"
	Action      string `json:"action" binding:"required"` // cth: "create"
	Description string `json:"description"`
}

// 1. GET ALL PERMISSIONS
func GetPermissions(c *gin.Context) {
	var permissions []models.Permission
	
	// Preload "Menu" agar di response ketahuan permission ini milik menu apa
	if err := config.DB.Preload("Menu").Find(&permissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data permission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengambil daftar permission",
		"data":    permissions,
	})
}

// 2. CREATE PERMISSION
func CreatePermission(c *gin.Context) {
	var input PermissionInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Cek apakah menu-nya ada
	var menu models.Menu
	if err := config.DB.Where("id = ?", input.MenuID).First(&menu).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Menu ID tidak valid atau tidak ditemukan"})
		return
	}

	// Cek apakah kombinasi MenuID dan Action sudah ada (mencegah duplikat action di menu yang sama)
	var existingPerm models.Permission
	if err := config.DB.Where("menu_id = ? AND action = ?", input.MenuID, input.Action).First(&existingPerm).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Permission dengan aksi tersebut sudah ada pada menu ini"})
		return
	}

	newPermission := models.Permission{
		MenuID:      input.MenuID,
		Name:        input.Name,
		Action:      input.Action,
		Description: input.Description,
	}

	if err := config.DB.Create(&newPermission).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat permission"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Berhasil membuat permission baru",
		"data":    newPermission,
	})
}

// 3. UPDATE PERMISSION
func UpdatePermission(c *gin.Context) {
	permID := c.Param("id")
	var input PermissionInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var permission models.Permission
	if err := config.DB.Where("id = ?", permID).First(&permission).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Permission tidak ditemukan"})
		return
	}

	// Cek menu validitas jika diubah
	if permission.MenuID != input.MenuID {
		var menu models.Menu
		if err := config.DB.Where("id = ?", input.MenuID).First(&menu).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Menu ID baru tidak ditemukan"})
			return
		}
	}

	permission.MenuID = input.MenuID
	permission.Name = input.Name
	permission.Action = input.Action
	permission.Description = input.Description

	if err := config.DB.Save(&permission).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengupdate permission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengupdate permission",
		"data":    permission,
	})
}

// 4. DELETE PERMISSION
func DeletePermission(c *gin.Context) {
	permID := c.Param("id")

	var permission models.Permission
	if err := config.DB.Where("id = ?", permID).First(&permission).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Permission tidak ditemukan"})
		return
	}

	if err := config.DB.Delete(&permission).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus permission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil menghapus permission",
	})
}