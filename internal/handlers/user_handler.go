package handlers

import (
	"evoting-backend/internal/config"
	"evoting-backend/internal/models"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type ToggleStatusInput struct {
	IsActive bool `json:"is_active"`
}

type CreateUserInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	RoleIDs  []uint `json:"role_ids" binding:"required"`
}

type UpdateUserInput struct {
	Name    string `json:"name" binding:"required"`
	Email   string `json:"email" binding:"required,email"`
	RoleIDs []uint `json:"role_ids" binding:"required"`
}

func GetUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))
	search := c.Query("search")
	roleFilter := c.Query("role")

	offset := (page - 1) * limit
	var users []models.User
	var totalItems int64

	query := config.DB.Model(&models.User{})

	if search != "" {
		query = query.Where("name LIKE ? OR email LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if roleFilter != "" {
		query = query.Where("EXISTS (SELECT 1 FROM user_roles ur JOIN roles r ON r.id = ur.role_id WHERE ur.user_id = users.id AND r.name = ?)", roleFilter)
	}

	query.Count(&totalItems)

	if err := query.Preload("Roles").Limit(limit).Offset(offset).Order("created_at desc").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data pengguna"})
		return
	}

	totalPages := int(math.Ceil(float64(totalItems) / float64(limit)))

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengambil data pengguna",
		"data":    users,
		"meta": gin.H{
			"current_page": page,
			"total_pages":  totalPages,
			"total_items":  totalItems,
			"limit":        limit,
		},
	})
}

func ApproveUser(c *gin.Context) {
	userID := c.Param("id")
	var user models.User
	if err := config.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
		return
	}
	user.IsApproved = true
	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyetujui user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User berhasil disetujui"})
}

func ToggleUserStatus(c *gin.Context) {
	userID := c.Param("id")
	var input ToggleStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	loggedInUserID, _ := c.Get("userID")
	if userID == loggedInUserID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak bisa menonaktifkan akun Anda sendiri"})
		return
	}
	if err := config.DB.Model(&models.User{}).Where("id = ?", userID).Update("is_active", input.IsActive).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengubah status pengguna"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Status pengguna berhasil diubah"})
}

func DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	loggedInUserID, _ := c.Get("userID")
	if userID == loggedInUserID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak bisa menghapus akun Anda sendiri"})
		return
	}
	if err := config.DB.Where("id = ?", userID).Delete(&models.User{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus pengguna"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Pengguna berhasil dihapus"})
}

func CreateUser(c *gin.Context) {
	var input CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing models.User
	if err := config.DB.Where("email = ?", input.Email).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email sudah digunakan"})
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	newUser := models.User{
		ID:         uuid.New().String(),
		Name:       input.Name,
		Email:      input.Email,
		Password:   string(hashedPassword),
		IsApproved: true,
		IsActive:   true,
	}

	var roles []models.Role
	config.DB.Where("id IN ?", input.RoleIDs).Find(&roles)

	tx := config.DB.Begin()
	if err := tx.Create(&newUser).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat user"})
		return
	}
	if err := tx.Model(&newUser).Association("Roles").Append(roles); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menetapkan role"})
		return
	}
	tx.Commit()

	c.JSON(http.StatusCreated, gin.H{"message": "Berhasil membuat pengguna baru", "data": newUser})
}

func UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	var input UpdateUserInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := config.DB.Preload("Roles").Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
		return
	}

	isTargetSuperadmin := false
	for _, r := range user.Roles {
		if r.Name == "Superadmin" {
			isTargetSuperadmin = true
		}
	}
	if isTargetSuperadmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akun Superadmin tidak dapat diedit"})
		return
	}

	user.Name = input.Name
	user.Email = input.Email

	var roles []models.Role
	config.DB.Where("id IN ?", input.RoleIDs).Find(&roles)

	tx := config.DB.Begin()
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengupdate user"})
		return
	}

	if err := tx.Model(&user).Association("Roles").Replace(roles); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengupdate role"})
		return
	}
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "Berhasil mengupdate pengguna"})
}