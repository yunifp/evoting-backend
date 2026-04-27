package handlers

import (
    "evoting-backend/internal/config"
    "evoting-backend/internal/models"
    "evoting-backend/internal/utils"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "golang.org/x/crypto/bcrypt"
)

type RegisterInput struct {
    Name     string `json:"name" binding:"required"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
}

type LoginInput struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

func Register(c *gin.Context) {
    var input RegisterInput

    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    var existingUser models.User
    if err := config.DB.Where("email = ?", input.Email).First(&existingUser).Error; err == nil {
        c.JSON(http.StatusConflict, gin.H{"error": "Email sudah terdaftar"})
        return
    }

    var clientRole models.Role
    if err := config.DB.Where("name = ?", "Client").First(&clientRole).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Role Client tidak ditemukan di database"})
        return
    }

    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengenkripsi password"})
        return
    }

    newUser := models.User{
        ID:         uuid.New().String(),
        Name:       input.Name,
        Email:      input.Email,
        Password:   string(hashedPassword),
        IsApproved: false, 
        IsActive:   true,
    }

    tx := config.DB.Begin()
    if err := tx.Create(&newUser).Error; err != nil {
        tx.Rollback()
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat user"})
        return
    }

    if err := tx.Model(&newUser).Association("Roles").Append([]models.Role{clientRole}); err != nil {
        tx.Rollback()
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menetapkan role"})
        return
    }
    tx.Commit()

    c.JSON(http.StatusCreated, gin.H{
        "message": "Registrasi berhasil! Silakan tunggu approval dari Admin.",
    })
}

func Login(c *gin.Context) {
    var input LoginInput

    // 1. Validasi Input JSON
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 2. Cari User beserta Roles dan Permissions-nya (UPDATE DI SINI)
    var user models.User
    if err := config.DB.Preload("Roles.Permissions").Where("email = ?", input.Email).First(&user).Error; err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Email atau password salah"})
        return
    }

    // 3. Verifikasi Password
    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Email atau password salah"})
        return
    }

    // 4. Cek Status Approval & Active (Khusus Superadmin, bebas hambatan)
    isSuperadmin := false
    for _, role := range user.Roles {
        if role.Name == "Superadmin" {
            isSuperadmin = true
            break
        }
    }

    if !isSuperadmin {
        if !user.IsActive {
            c.JSON(http.StatusForbidden, gin.H{"error": "Akun Anda telah dinonaktifkan."})
            return
        }
        if !user.IsApproved {
            c.JSON(http.StatusForbidden, gin.H{"error": "Akun Anda belum disetujui oleh Admin."})
            return
        }
    }

    // 5. Generate JWT Token
    token, err := utils.GenerateJWT(user.ID, user.Email)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal men-generate token"})
        return
    }

    // 6. Return Data (Sembunyikan Password)
    // Array permissions akan otomatis ikut di dalam user.Roles karena kita sudah Preload
    c.JSON(http.StatusOK, gin.H{
        "message": "Login berhasil",
        "token":   token,
        "user": gin.H{
            "id":    user.ID,
            "name":  user.Name,
            "email": user.Email,
            "roles": user.Roles,
        },
    })
}