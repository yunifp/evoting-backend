package handlers

import (
	"evoting-backend/internal/config"
	"evoting-backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type WilayahResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func GetProvinsi(c *gin.Context) {
	var data []WilayahResponse
	config.DB.Model(&models.RefWilayah{}).
		Select("kode_pro as id, nama_wilayah as name").
		Where("tingkat = ?", 1).
		Order("nama_wilayah asc").
		Find(&data)

	c.JSON(http.StatusOK, gin.H{"data": data})
}

func GetKabupaten(c *gin.Context) {
	kodePro := c.Query("kode_pro")
	var data []WilayahResponse
	config.DB.Model(&models.RefWilayah{}).
		Select("kode_kab as id, nama_wilayah as name").
		Where("tingkat = ? AND kode_pro = ?", 2, kodePro).
		Order("nama_wilayah asc").
		Find(&data)

	c.JSON(http.StatusOK, gin.H{"data": data})
}

func GetKecamatan(c *gin.Context) {
	kodeKab := c.Query("kode_kab")
	var data []WilayahResponse
	config.DB.Model(&models.RefWilayah{}).
		Select("kode_kec as id, nama_wilayah as name").
		Where("tingkat = ? AND kode_kab = ?", 3, kodeKab).
		Order("nama_wilayah asc").
		Find(&data)

	c.JSON(http.StatusOK, gin.H{"data": data})
}

func GetKelurahan(c *gin.Context) {
	kodeKec := c.Query("kode_kec")
	var data []WilayahResponse
	config.DB.Model(&models.RefWilayah{}).
		Select("kode_kel as id, nama_wilayah as name").
		Where("tingkat = ? AND kode_kec = ?", 4, kodeKec).
		Order("nama_wilayah asc").
		Find(&data)

	c.JSON(http.StatusOK, gin.H{"data": data})
}

func GetStatusKawin(c *gin.Context) {
	var data []models.StatusKawin
	config.DB.Find(&data)
	c.JSON(http.StatusOK, gin.H{"data": data})
}