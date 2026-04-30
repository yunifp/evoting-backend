package handlers

import (
    "evoting-backend/internal/config"
    "evoting-backend/internal/models"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
)

type PemiluInput struct {
	TransactionID string `json:"transaction_id" binding:"required"`
    Title     string `json:"title" binding:"required"`
    StartDate string `json:"start_date" binding:"required"` 
    EndDate   string `json:"end_date" binding:"required"`
}

func CreatePemilu(c *gin.Context) {
    clientID, _ := c.Get("userID")

    var input PemiluInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Data input tidak lengkap. Pastikan Anda memilih paket."})
        return
    }

    // 1. Validasi transaksi yang dipilih user (harus milik user tsb dan sudah Lunas)
    var trx models.Transaction
    if err := config.DB.Where("id = ? AND user_id = ? AND status = ?", input.TransactionID, clientID, "paid").First(&trx).Error; err != nil {
        c.JSON(http.StatusForbidden, gin.H{"error": "Akses Ditolak: Paket tidak valid atau belum lunas."})
        return
    }

    // 2. Validasi Krusial: Cek apakah transaksi yang dipilih benar-benar belum terpakai
    var checkPemilu models.Pemilu
    if err := config.DB.Where("transaction_id = ?", trx.ID).First(&checkPemilu).Error; err == nil {
        c.JSON(http.StatusForbidden, gin.H{"error": "Paket ini sudah digunakan untuk acara lain. Silakan pilih paket yang masih kosong."})
        return
    }

    layoutFormat := "2006-01-02 15:04:05" 
    startTime, err := time.Parse(layoutFormat, input.StartDate)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Format start_date salah"})
        return
    }

    endTime, err := time.Parse(layoutFormat, input.EndDate)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Format end_date salah"})
        return
    }

    newPemilu := models.Pemilu{
        ClientID:      clientID.(string),
        TransactionID: trx.ID, // Ikat menggunakan transaksi pilihan Client
        Title:         input.Title,
        StartDate:     startTime,
        EndDate:       endTime,
        Status:        "draft",
    }

    if err := config.DB.Create(&newPemilu).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat Event Pemilu"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"message": "Berhasil membuat ruang Event Pemilu", "data": newPemilu})
}
func GetMyPemilu(c *gin.Context) {
    clientID, _ := c.Get("userID")
    var pemiluList []models.Pemilu
    if err := config.DB.Where("client_id = ?", clientID).Preload("Kandidats").Order("created_at desc").Find(&pemiluList).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data pemilu"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Berhasil mengambil data Event Pemilu", "data": pemiluList})
}

func GetPemiluDetail(c *gin.Context) {
    clientID, _ := c.Get("userID")
    pemiluID := c.Param("pemiluId")

    var pemilu models.Pemilu
    if err := config.DB.Preload("Kandidats").Where("id = ? AND client_id = ?", pemiluID, clientID).First(&pemilu).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Event Pemilu tidak ditemukan"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Berhasil mengambil detail", "data": pemilu})
}

func PublishPemilu(c *gin.Context) {
    clientID, _ := c.Get("userID")
    pemiluID := c.Param("pemiluId")

    var pemilu models.Pemilu
    if err := config.DB.Where("id = ? AND client_id = ?", pemiluID, clientID).First(&pemilu).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Event Pemilu tidak ditemukan"})
        return
    }

    if pemilu.Status != "draft" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Event ini sudah aktif atau selesai"})
        return
    }

    pemilu.Status = "active"
    if err := config.DB.Save(&pemilu).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengaktifkan acara"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Acara berhasil diaktifkan! Pemilih sekarang dapat login."})
}

func UpdatePemilu(c *gin.Context) {
    clientID, _ := c.Get("userID")
    pemiluID := c.Param("pemiluId")

    var pemilu models.Pemilu
    if err := config.DB.Where("id = ? AND client_id = ?", pemiluID, clientID).First(&pemilu).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Event Pemilu tidak ditemukan"})
        return
    }

    if pemilu.Status != "draft" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Hanya acara berstatus draft yang dapat diedit"})
        return
    }

    var input PemiluInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    layoutFormat := "2006-01-02 15:04:05" 
    startTime, err := time.Parse(layoutFormat, input.StartDate)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Format start_date salah"})
        return
    }

    endTime, err := time.Parse(layoutFormat, input.EndDate)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Format end_date salah"})
        return
    }

    pemilu.Title = input.Title
    pemilu.StartDate = startTime
    pemilu.EndDate = endTime

    if err := config.DB.Save(&pemilu).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengupdate Event Pemilu"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Berhasil mengupdate acara", "data": pemilu})
}

func DeletePemilu(c *gin.Context) {
    clientID, _ := c.Get("userID")
    pemiluID := c.Param("pemiluId")

    var pemilu models.Pemilu
    if err := config.DB.Where("id = ? AND client_id = ?", pemiluID, clientID).First(&pemilu).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Event Pemilu tidak ditemukan"})
        return
    }

    if pemilu.Status != "draft" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Hanya acara berstatus draft yang dapat dihapus"})
        return
    }

    if err := config.DB.Delete(&pemilu).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus Event Pemilu"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Berhasil menghapus acara. Paket layanan Anda dapat digunakan kembali."})
}

func ClosePemilu(c *gin.Context) {
    clientID, _ := c.Get("userID")
    pemiluID := c.Param("pemiluId")

    var pemilu models.Pemilu
    if err := config.DB.Where("id = ? AND client_id = ?", pemiluID, clientID).First(&pemilu).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Event Pemilu tidak ditemukan"})
        return
    }

    if pemilu.Status != "active" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Hanya acara yang sedang aktif yang dapat ditutup"})
        return
    }

    pemilu.Status = "selesai" // Ubah status menjadi selesai
    if err := config.DB.Save(&pemilu).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menutup acara"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Acara berhasil ditutup! Hasil akhir sekarang telah dikunci."})
}

func GetAvailablePackages(c *gin.Context) {
    clientID, _ := c.Get("userID")

    subQuery := config.DB.Model(&models.Pemilu{}).Select("transaction_id").Where("transaction_id IS NOT NULL")
    
    var transactions []models.Transaction
    // Gunakan Preload("Layanan") sesuai nama field di struct models.Transaction
    if err := config.DB.Preload("Layanan").Where("user_id = ? AND status = ? AND id NOT IN (?)", clientID, "paid", subQuery).
        Order("paid_at asc").
        Find(&transactions).Error; err != nil {
        
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data paket tersedia"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Berhasil mengambil paket tersedia", "data": transactions})
}