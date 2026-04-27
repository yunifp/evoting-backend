package handlers

import (
	"evoting-backend/internal/config"
	"evoting-backend/internal/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateTransactionInput struct {
	LayananID     uint   `json:"layanan_id" binding:"required"`
	PaymentMethod string `json:"payment_method" binding:"required"`
}

// ==========================================
// BAGIAN CLIENT (Penyelenggara)
// ==========================================

// 1. CLIENT Beli Paket
func CreateTransaction(c *gin.Context) {
	// Ambil ID Client yang sedang login dari JWT
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak valid"})
		return
	}

	var input CreateTransactionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Cek apakah Layanan/Paket tersedia dan aktif
	var layanan models.Layanan
	if err := config.DB.Where("id = ? AND is_active = ?", input.LayananID, true).First(&layanan).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Layanan tidak ditemukan atau tidak aktif"})
		return
	}

	// Buat Transaksi Baru
	// Catatan: Harga (Amount) diambil langsung dari database Layanan, BUKAN dari input user
	// untuk mencegah manipulasi harga dari sisi frontend/postman.
	newTransaction := models.Transaction{
		ID:            uuid.New().String(),
		UserID:        userID.(string),
		LayananID:     input.LayananID,
		Amount:        layanan.Price,
		Status:        "pending",
		PaymentMethod: input.PaymentMethod,
	}

	if err := config.DB.Create(&newTransaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat transaksi"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Pesanan berhasil dibuat. Silakan lakukan pembayaran.",
		"data":    newTransaction,
	})
}

// 2. CLIENT Lihat Riwayat Transaksinya Sendiri
func GetMyTransactions(c *gin.Context) {
	userID, _ := c.Get("userID")

	var transactions []models.Transaction
	if err := config.DB.Where("user_id = ?", userID).Preload("Layanan").Order("created_at desc").Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil riwayat transaksi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengambil riwayat transaksi",
		"data":    transactions,
	})
}

// ==========================================
// BAGIAN ADMIN / SUPERADMIN
// ==========================================

// 3. ADMIN Lihat Semua Transaksi
func GetAllTransactions(c *gin.Context) {
	var transactions []models.Transaction
	// Preload data User pembeli dan Layanannya
	if err := config.DB.Preload("User").Preload("Layanan").Order("created_at desc").Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data transaksi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengambil seluruh transaksi",
		"data":    transactions,
	})
}

// 4. ADMIN Approve Pembayaran Manual
func ApproveTransaction(c *gin.Context) {
	trxID := c.Param("id")

	var transaction models.Transaction
	if err := config.DB.Where("id = ?", trxID).First(&transaction).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaksi tidak ditemukan"})
		return
	}

	if transaction.Status == "paid" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transaksi ini sudah dibayar"})
		return
	}

	now := time.Now()
	transaction.Status = "paid"
	transaction.PaidAt = &now

	if err := config.DB.Save(&transaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memverifikasi pembayaran"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Pembayaran berhasil diverifikasi. Paket kini aktif untuk client.",
		"data":    transaction,
	})
}