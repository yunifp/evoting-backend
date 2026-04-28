package handlers

import (
	"evoting-backend/internal/config"
	"evoting-backend/internal/models"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RoleInput struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

func GetRoles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))
	search := c.Query("search")

	offset := (page - 1) * limit
	var roles []models.Role
	var totalItems int64

	query := config.DB.Model(&models.Role{})

	if search != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	query.Count(&totalItems)

	if err := query.Preload("Permissions").Limit(limit).Offset(offset).Order("created_at desc").Find(&roles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data role"})
		return
	}

	totalPages := int(math.Ceil(float64(totalItems) / float64(limit)))

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengambil daftar role",
		"data":    roles,
		"meta": gin.H{
			"current_page": page,
			"total_pages":  totalPages,
			"total_items":  totalItems,
			"limit":        limit,
		},
	})
}

func CreateRole(c *gin.Context) {
	var input RoleInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existingRole models.Role
	if err := config.DB.Where("name = ?", input.Name).First(&existingRole).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Nama role sudah digunakan"})
		return
	}

	newRole := models.Role{
		Name:        input.Name,
		Description: input.Description,
	}

	if err := config.DB.Create(&newRole).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat role"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Berhasil membuat role baru",
		"data":    newRole,
	})
}

func UpdateRole(c *gin.Context) {
	roleID := c.Param("id")
	var input RoleInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var role models.Role
	if err := config.DB.Where("id = ?", roleID).First(&role).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role tidak ditemukan"})
		return
	}

	if role.Name == "Superadmin" && input.Name != "Superadmin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Nama role Superadmin tidak boleh diubah"})
		return
	}

	role.Name = input.Name
	role.Description = input.Description

	if err := config.DB.Save(&role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengupdate role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengupdate role",
		"data":    role,
	})
}

func DeleteRole(c *gin.Context) {
	roleID := c.Param("id")

	var role models.Role
	if err := config.DB.Where("id = ?", roleID).First(&role).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role tidak ditemukan"})
		return
	}

	if role.Name == "Superadmin" || role.Name == "Admin" || role.Name == "Client" || role.Name == "Voter" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Role inti sistem tidak boleh dihapus"})
		return
	}

	if err := config.DB.Delete(&role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil menghapus role",
	})
}

type AssignPermissionsInput struct {
	PermissionIDs []uint `json:"permission_ids" binding:"required"`
}

func AssignPermissions(c *gin.Context) {
	roleID := c.Param("id")
	var input AssignPermissionsInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var role models.Role
	if err := config.DB.Where("id = ?", roleID).First(&role).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role tidak ditemukan"})
		return
	}

	if role.Name == "Superadmin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Hak akses Superadmin bersifat absolut dan tidak dapat diubah"})
		return
	}

	var permissions []models.Permission
	if len(input.PermissionIDs) > 0 {
		if err := config.DB.Where("id IN ?", input.PermissionIDs).Find(&permissions).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memvalidasi permissions"})
			return
		}
	}

	if err := config.DB.Model(&role).Association("Permissions").Replace(permissions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menetapkan hak akses ke role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil memperbarui hak akses untuk role " + role.Name,
	})
}