package middlewares

import (
	"evoting-backend/internal/config"
	"evoting-backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequirePermission(menuName string, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User tidak terautentikasi"})
			return
		}

		var user models.User
		if err := config.DB.Preload("Roles").Where("id = ?", userID).First(&user).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User tidak ditemukan"})
			return
		}

		for _, role := range user.Roles {
			if role.Name == "Superadmin" {
				c.Next()
				return
			}
		}

		var count int64
		err := config.DB.Table("permissions").
			Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
			Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
			Joins("JOIN menus ON menus.id = permissions.menu_id").
			Where("user_roles.user_id = ?", userID).
			Where("menus.name = ?", menuName).
			Where("permissions.action = ?", action).
			Count(&count).Error

		if err != nil || count == 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Akses ditolak: Anda tidak memiliki izin untuk melakukan tindakan ini",
			})
			return
		}

		c.Next() 
	}
}