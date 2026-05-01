package handlers

import (
	"evoting-backend/internal/config"
	"evoting-backend/internal/models"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetMyMenus(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak valid"})
		return
	}

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
		if err := config.DB.Preload("SubMenus", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_active = ?", true).Order("sort_order asc")
		}).Where("is_active = ? AND parent_id IS NULL", true).Order("sort_order asc").Find(&menus).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data menu"})
			return
		}
	} else {
		var flatMenus []models.Menu
		err := config.DB.Distinct("menus.*").
			Joins("JOIN permissions ON permissions.menu_id = menus.id").
			Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
			Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
			Where("user_roles.user_id = ?", userID).
			Where("permissions.action = ?", "read").
			Where("menus.is_active = ?", true).
			Order("menus.sort_order asc").
			Find(&flatMenus).Error

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data menu"})
			return
		}

		menuMap := make(map[uint]*models.Menu)
		for i := range flatMenus {
			flatMenus[i].SubMenus = []models.Menu{}
			menuMap[flatMenus[i].ID] = &flatMenus[i]
		}

		for _, menu := range flatMenus {
			if menu.ParentID != nil {
				if parent, exists := menuMap[*menu.ParentID]; exists {
					parent.SubMenus = append(parent.SubMenus, *menuMap[menu.ID])
				}
			}
		}

		for _, menu := range flatMenus {
			if menu.ParentID == nil {
				menus = append(menus, *menuMap[menu.ID])
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengambil menu",
		"data":    menus,
	})
}

type MenuInput struct {
	Name      string `json:"name" binding:"required"`
	Path      string `json:"path"`
	Icon      string `json:"icon"`
	ParentID  *uint  `json:"parent_id"`
	SortOrder int    `json:"sort_order"`
	IsActive  *bool  `json:"is_active"`
}

func GetAllMenus(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))
	search := c.Query("search")

	offset := (page - 1) * limit
	var menus []models.Menu
	var totalItems int64

	query := config.DB.Model(&models.Menu{}).Where("parent_id IS NULL")

	if search != "" {
		query = query.Where("name LIKE ?", "%"+search+"%")
	}

	query.Count(&totalItems)

	if err := query.Preload("SubMenus", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order asc")
	}).Order("sort_order asc").Limit(limit).Offset(offset).Find(&menus).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data menu"})
		return
	}

	totalPages := int(math.Ceil(float64(totalItems) / float64(limit)))

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengambil seluruh daftar menu",
		"data":    menus,
		"meta": gin.H{
			"current_page": page,
			"total_pages":  totalPages,
			"total_items":  totalItems,
			"limit":        limit,
		},
	})
}

func CreateMenu(c *gin.Context) {
	var input MenuInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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

func DeleteMenu(c *gin.Context) {
	menuID := c.Param("id")

	var menu models.Menu
	if err := config.DB.Where("id = ?", menuID).First(&menu).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Menu tidak ditemukan"})
		return
	}

	tx := config.DB.Begin()

	if err := tx.Exec("DELETE FROM role_permissions WHERE permission_id IN (SELECT id FROM permissions WHERE menu_id = ?)", menu.ID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membersihkan pivot role_permissions"})
		return
	}

	if err := tx.Where("menu_id = ?", menu.ID).Delete(&models.Permission{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membersihkan hak akses (permissions)"})
		return
	}

	if err := tx.Where("parent_id = ?", menu.ID).Delete(&models.Menu{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membersihkan sub-menus"})
		return
	}

	if err := tx.Delete(&menu).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus menu"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil menghapus menu beserta seluruh hak akses terkait",
	})
}