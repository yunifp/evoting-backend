package handlers

import (
    "evoting-backend/internal/config"
    "evoting-backend/internal/models"
    "fmt"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

type DPTInput struct {
    NIK          string `json:"nik" binding:"required,len=16"` 
    Nama         string `json:"nama" binding:"required"`
    NoHP         string `json:"no_hp" binding:"required"`
    FaceTemplate string `json:"face_template"` // Disediakan jika diinput saat pendaftaran (opsional)
}

func AddDPT(c *gin.Context) {
    clientID, _ := c.Get("userID")
    pemiluID := c.Param("pemiluId")

    var input DPTInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 1. Ambil Event Pemilu
    var pemilu models.Pemilu
    if err := config.DB.Where("id = ? AND client_id = ?", pemiluID, clientID).First(&pemilu).Error; err != nil {
        c.JSON(http.StatusForbidden, gin.H{"error": "Event Pemilu tidak ditemukan atau bukan milik Anda"})
        return
    }

    // 2. Ambil Transaksi yang terikat dengan Pemilu INI untuk melihat fitur Layanannya
    var trx models.Transaction
    if err := config.DB.Preload("Layanan").Where("id = ?", pemilu.TransactionID).First(&trx).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memverifikasi paket layanan pemilu ini"})
        return
    }

    limitDPT := trx.Layanan.LimitDPT
    isFaceRec := trx.Layanan.IsFaceRecognition

    // 3. Cek Limit DPT
    var countDPT int64
    config.DB.Model(&models.DPT{}).Where("pemilu_id = ?", pemilu.ID).Count(&countDPT)

    if countDPT >= int64(limitDPT) {
        c.JSON(http.StatusForbidden, gin.H{
            "error": fmt.Sprintf("Kuota DPT penuh! Batas maksimal paket acara ini adalah %d pemilih.", limitDPT),
        })
        return
    }

    // 4. Validasi NIK Ganda
    var existingDPT models.DPT
    if err := config.DB.Where("pemilu_id = ? AND nik = ?", pemilu.ID, input.NIK).First(&existingDPT).Error; err == nil {
        c.JSON(http.StatusConflict, gin.H{"error": "NIK ini sudah terdaftar di daftar pemilih event ini"})
        return
    }

    newDPT := models.DPT{
        ID:       uuid.New().String(),
        PemiluID: pemilu.ID,
        NIK:      input.NIK,
        Nama:     input.Nama,
        NoHP:     input.NoHP,
    }

    // Masukkan data wajah hanya jika paket layanan mendukung Face Recognition
    if isFaceRec {
        newDPT.FaceTemplate = input.FaceTemplate
    }

    if err := config.DB.Create(&newDPT).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menambahkan data pemilih"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        "message": "Berhasil menambahkan data pemilih (DPT)",
        "data": gin.H{
            "dpt": newDPT,
            "is_face_recognition_active": isFaceRec, // Beritahu frontend status fitur ini
        },
    })
}

// 2. CLIENT: Lihat Daftar DPT di Pemilu Tertentu
func GetDPTByPemilu(c *gin.Context) {
    clientID, _ := c.Get("userID")
    pemiluID := c.Param("pemiluId")

    var pemilu models.Pemilu
    if err := config.DB.Where("id = ? AND client_id = ?", pemiluID, clientID).First(&pemilu).Error; err != nil {
        c.JSON(http.StatusForbidden, gin.H{"error": "Event Pemilu tidak valid"})
        return
    }

    var dptList []models.DPT
    if err := config.DB.Where("pemilu_id = ?", pemilu.ID).Order("nama asc").Find(&dptList).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data DPT"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "Berhasil mengambil data DPT",
        "data":    dptList,
    })
}