package handlers

import (
	"evoting-backend/internal/config"
	"evoting-backend/internal/models"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PermissionInput struct {
	MenuID      uint   `json:"menu_id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Action      string `json:"action" binding:"required"`
	Description string `json:"description"`
}

func GetPermissions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))
	search := c.Query("search")
	actionFilter := c.Query("action")
	menuIDFilter := c.Query("menu_id") // ← FIX: tambahkan ini

	offset := (page - 1) * limit
	var permissions []models.Permission
	var totalItems int64

	query := config.DB.Model(&models.Permission{})

	if search != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if actionFilter != "" {
		query = query.Where("action = ?", actionFilter)
	}

	// ← FIX: tambahkan blok ini
	if menuIDFilter != "" {
		query = query.Where("menu_id = ?", menuIDFilter)
	}

	query.Count(&totalItems)

	if err := query.Preload("Menu").Limit(limit).Offset(offset).Order("created_at desc").Find(&permissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data permission"})
		return
	}

	totalPages := int(math.Ceil(float64(totalItems) / float64(limit)))

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengambil daftar permission",
		"data":    permissions,
		"meta": gin.H{
			"current_page": page,
			"total_pages":  totalPages,
			"total_items":  totalItems,
			"limit":        limit,
		},
	})
}

func CreatePermission(c *gin.Context) {
	var input PermissionInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var menu models.Menu
	if err := config.DB.Where("id = ?", input.MenuID).First(&menu).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Menu ID tidak valid atau tidak ditemukan"})
		return
	}

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

func DeletePermission(c *gin.Context) {
	permID := c.Param("id")

	var permission models.Permission
	if err := config.DB.Where("id = ?", permID).First(&permission).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Permission tidak ditemukan"})
		return
	}

	tx := config.DB.Begin()

	if err := tx.Exec("DELETE FROM role_permissions WHERE permission_id = ?", permission.ID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal melepaskan permission dari role"})
		return
	}

	if err := tx.Delete(&permission).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus permission"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil menghapus hak akses",
	})
}