// internal/handlers/transaction_handler.go
package handlers

import (
	"evoting-backend/internal/config"
	"evoting-backend/internal/models"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
)

type CreateTransactionInput struct {
	LayananID     uint   `json:"layanan_id" binding:"required"`
	PaymentMethod string `json:"payment_method" binding:"required"`
}

func CreateTransaction(c *gin.Context) {
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

	var layanan models.Layanan
	if err := config.DB.Where("id = ? AND is_active = ?", input.LayananID, true).First(&layanan).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Layanan tidak ditemukan atau tidak aktif"})
		return
	}

	var user models.User
	config.DB.Where("id = ?", userID).First(&user)

	transactionID := uuid.New().String()

	if layanan.Price == 0 {
		now := time.Now()
		newTransaction := models.Transaction{
			ID:            transactionID,
			UserID:        userID.(string),
			LayananID:     input.LayananID,
			Amount:        layanan.Price,
			Status:        "paid",
			PaymentMethod: "Free",
			PaidAt:        &now,
		}

		if err := config.DB.Create(&newTransaction).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat transaksi paket gratis"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Pesanan paket gratis berhasil diaktifkan.",
			"data":    newTransaction,
		})
		return
	}

	var serverKeySetting models.Setting
	if err := config.DB.Where("key = ?", "midtrans_server_key").First(&serverKeySetting).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kredensial Midtrans belum dikonfigurasi"})
		return
	}

	var isProductionSetting models.Setting
	env := midtrans.Sandbox
	if err := config.DB.Where("key = ?", "midtrans_is_production").First(&isProductionSetting).Error; err == nil {
		if isProductionSetting.Value == "true" {
			env = midtrans.Production
		}
	}

	var snapClient snap.Client
	snapClient.New(serverKeySetting.Value, env)

	req := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  transactionID,
			GrossAmt: int64(layanan.Price),
		},
		CustomerDetail: &midtrans.CustomerDetails{
			FName: user.Name,
			Email: user.Email,
		},
		Items: &[]midtrans.ItemDetails{
			{
				ID:    fmt.Sprint(layanan.ID),
				Name:  layanan.Name,
				Price: int64(layanan.Price),
				Qty:   1,
			},
		},
	}

	snapResp, err := snapClient.CreateTransaction(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal generate token pembayaran dari Midtrans"})
		return
	}

	newTransaction := models.Transaction{
		ID:            transactionID,
		UserID:        userID.(string),
		LayananID:     input.LayananID,
		Amount:        layanan.Price,
		Status:        "pending",
		PaymentMethod: input.PaymentMethod,
		SnapToken:     snapResp.Token,
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

func GetAllTransactions(c *gin.Context) {
	var transactions []models.Transaction
	if err := config.DB.Preload("User").Preload("Layanan").Order("created_at desc").Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data transaksi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengambil seluruh transaksi",
		"data":    transactions,
	})
}

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