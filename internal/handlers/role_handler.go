package handlers

import (
	"evoting-backend/internal/config"
	"evoting-backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// DTO untuk input form
type RoleInput struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// 1. GET ALL ROLES
func GetRoles(c *gin.Context) {
	var roles []models.Role
	// Preload jika nanti ingin melihat permissions apa saja yang nempel di role ini
	if err := config.DB.Preload("Permissions").Find(&roles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengambil daftar role",
		"data":    roles,
	})
}

// 2. CREATE ROLE
func CreateRole(c *gin.Context) {
	var input RoleInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Cek apakah nama role sudah ada
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

// 3. UPDATE ROLE
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

	// Proteksi: Superadmin tidak boleh diubah namanya
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

// 4. DELETE ROLE
func DeleteRole(c *gin.Context) {
	roleID := c.Param("id")

	var role models.Role
	if err := config.DB.Where("id = ?", roleID).First(&role).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role tidak ditemukan"})
		return
	}

	// Proteksi: Role core (Superadmin, Admin, Client, Voter) sebaiknya tidak bisa dihapus sembarangan, 
	// tapi minimal Superadmin wajib di-lock.
	if role.Name == "Superadmin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Role Superadmin tidak boleh dihapus"})
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

// DTO untuk input form Assign Permission
type AssignPermissionsInput struct {
	PermissionIDs []uint `json:"permission_ids" binding:"required"`
}

// 5. ASSIGN PERMISSIONS TO ROLE
func AssignPermissions(c *gin.Context) {
	roleID := c.Param("id")
	var input AssignPermissionsInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Cari role-nya
	var role models.Role
	if err := config.DB.Where("id = ?", roleID).First(&role).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role tidak ditemukan"})
		return
	}

	// Proteksi: Superadmin sebaiknya tidak bisa di-revoke hak aksesnya lewat endpoint ini 
	// (agar tidak ada admin iseng yang melumpuhkan Superadmin)
	if role.Name == "Superadmin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Hak akses Superadmin bersifat absolut dan tidak dapat diubah"})
		return
	}

	// Cari daftar permission berdasarkan array ID yang dikirim
	var permissions []models.Permission
	if len(input.PermissionIDs) > 0 {
		if err := config.DB.Where("id IN ?", input.PermissionIDs).Find(&permissions).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memvalidasi permissions"})
			return
		}
	}

	// Replace asosiasi (GORM otomatis menghapus yang tidak ada di list, dan menambahkan yang baru)
	if err := config.DB.Model(&role).Association("Permissions").Replace(permissions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menetapkan hak akses ke role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil memperbarui hak akses untuk role " + role.Name,
	})
}