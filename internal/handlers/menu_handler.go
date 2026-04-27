package handlers

import (
	"evoting-backend/internal/config"
	"evoting-backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetMyMenus(c *gin.Context) {
	// Ambil userID dari token JWT yang sudah di-set oleh Middleware
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak valid"})
		return
	}

	// 1. Cek apakah user ini Superadmin (bypass semua menu)
	var user models.User
	config.DB.Preload("Roles").Where("id = ?", userID).First(&user)
	
	isSuperadmin := false
	for _, role := range user.Roles {
		if role.Name == "Superadmin" {
			isSuperadmin = true
			break
		}
	}

	var menus []models.Menu

	if isSuperadmin {
		// Superadmin dapat semua menu
		if err := config.DB.Where("is_active = ?", true).Order("sort_order asc").Find(&menus).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data menu"})
			return
		}
	} else {
		// 2. Query dinamis untuk user biasa (hanya ambil menu yang punya permission "read")
		err := config.DB.Distinct("menus.*").
			Joins("JOIN permissions ON permissions.menu_id = menus.id").
			Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
			Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
			Where("user_roles.user_id = ?", userID).
			Where("permissions.action = ?", "read").
			Where("menus.is_active = ?", true).
			Order("menus.sort_order asc").
			Find(&menus).Error

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data menu"})
			return
		}
	}

	// Return daftar menu ke frontend untuk di-render di Sidebar
	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengambil menu",
		"data":    menus,
	})
}

type MenuInput struct {
	Name      string `json:"name" binding:"required"`
	Path      string `json:"path" binding:"required"`
	Icon      string `json:"icon"`
	ParentID  *uint  `json:"parent_id"` // Menggunakan pointer agar bisa menerima null
	SortOrder int    `json:"sort_order"`
	IsActive  *bool  `json:"is_active"` // Pointer agar bisa mendeteksi nilai false yang dikirim
}

// 1. GET ALL MENUS (Untuk tabel di Dashboard Superadmin)
// Berbeda dengan GetMyMenus, ini akan mengambil semua menu tanpa mempedulikan Role
func GetAllMenus(c *gin.Context) {
	var menus []models.Menu

	// Kita ambil menu parent saja (ParentID == nil), dan kita preload SubMenus-nya
	if err := config.DB.Where("parent_id IS NULL").Preload("SubMenus").Order("sort_order asc").Find(&menus).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data menu"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengambil seluruh daftar menu",
		"data":    menus,
	})
}

// 2. CREATE MENU
func CreateMenu(c *gin.Context) {
	var input MenuInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Default IsActive adalah true jika tidak dikirim dari frontend
	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	newMenu := models.Menu{
		Name:      input.Name,
		Path:      input.Path,
		Icon:      input.Icon,
		ParentID:  input.ParentID,
		SortOrder: input.SortOrder,
		IsActive:  isActive,
	}

	if err := config.DB.Create(&newMenu).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat menu"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Berhasil membuat menu baru",
		"data":    newMenu,
	})
}

// 3. UPDATE MENU
func UpdateMenu(c *gin.Context) {
	menuID := c.Param("id")
	var input MenuInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var menu models.Menu
	if err := config.DB.Where("id = ?", menuID).First(&menu).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Menu tidak ditemukan"})
		return
	}

	menu.Name = input.Name
	menu.Path = input.Path
	menu.Icon = input.Icon
	menu.ParentID = input.ParentID
	menu.SortOrder = input.SortOrder
	
	if input.IsActive != nil {
		menu.IsActive = *input.IsActive
	}

	if err := config.DB.Save(&menu).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengupdate menu"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengupdate menu",
		"data":    menu,
	})
}

// 4. DELETE MENU
func DeleteMenu(c *gin.Context) {
	menuID := c.Param("id")

	var menu models.Menu
	if err := config.DB.Where("id = ?", menuID).First(&menu).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Menu tidak ditemukan"})
		return
	}

	// GORM otomatis akan melakukan cascade delete pada SubMenus dan relasi Permissions
	// karena kita sudah set constraint OnDelete:CASCADE di models
	if err := config.DB.Delete(&menu).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus menu"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil menghapus menu",
	})
}