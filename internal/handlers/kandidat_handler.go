package handlers

import (
    "evoting-backend/internal/config"
    "evoting-backend/internal/models"
    "net/http"
    "os"
    "strconv"

    "github.com/gin-gonic/gin"
)

func AddKandidat(c *gin.Context) {
    clientID, _ := c.Get("userID")
    pemiluID := c.Param("pemiluId")

    noUrutStr := c.PostForm("no_urut")
    name := c.PostForm("name")
    visi := c.PostForm("visi")
    misi := c.PostForm("misi")

    if noUrutStr == "" || name == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "no_urut dan name wajib diisi"})
        return
    }

    noUrut, err := strconv.Atoi(noUrutStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "no_urut harus berupa angka"})
        return
    }

    var pemilu models.Pemilu
    if err := config.DB.Where("id = ? AND client_id = ?", pemiluID, clientID).First(&pemilu).Error; err != nil {
        c.JSON(http.StatusForbidden, gin.H{"error": "Akses ditolak: Event Pemilu ini bukan milik Anda atau tidak ditemukan"})
        return
    }

    photoURL, _ := c.Get("photo_url")
    photoURLStr, _ := photoURL.(string)
    
    localPath, _ := c.Get("photo_local_path")
    localPathStr, _ := localPath.(string)

    newKandidat := models.Kandidat{
        PemiluID: pemilu.ID,
        NoUrut:   noUrut,
        Name:     name,
        Visi:     visi,
        Misi:     misi,
        PhotoURL: photoURLStr,
    }

    if err := config.DB.Create(&newKandidat).Error; err != nil {
        // PERBAIKAN: Rollback (hapus) file foto jika gagal simpan ke database
        if localPathStr != "" {
            os.Remove(localPathStr) 
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menambahkan kandidat"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        "message": "Berhasil menambahkan kandidat",
        "data":    newKandidat,
    })
}

// FITUR BARU: Update Kandidat (Termasuk ganti foto)
func UpdateKandidat(c *gin.Context) {
    clientID, _ := c.Get("userID")
    kandidatID := c.Param("id")

    var kandidat models.Kandidat
    if err := config.DB.Preload("Pemilu").Where("id = ?", kandidatID).First(&kandidat).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Kandidat tidak ditemukan"})
        return
    }

    if kandidat.Pemilu.ClientID != clientID.(string) {
        c.JSON(http.StatusForbidden, gin.H{"error": "Akses ditolak"})
        return
    }

    noUrutStr := c.PostForm("no_urut")
    if noUrutStr != "" {
        noUrut, err := strconv.Atoi(noUrutStr)
        if err == nil {
            kandidat.NoUrut = noUrut
        }
    }
    
    if name := c.PostForm("name"); name != "" {
        kandidat.Name = name
    }
    if visi := c.PostForm("visi"); visi != "" {
        kandidat.Visi = visi
    }
    if misi := c.PostForm("misi"); misi != "" {
        kandidat.Misi = misi
    }

    photoURL, _ := c.Get("photo_url")
    photoURLStr, _ := photoURL.(string)

    // Jika ada upload foto baru
    if photoURLStr != "" {
        // Hapus foto lama dari storage fisik
        if kandidat.PhotoURL != "" {
            // Hilangkan slash di depan agar path relatif terhadap root project
            oldPath := "." + kandidat.PhotoURL 
            os.Remove(oldPath)
        }
        kandidat.PhotoURL = photoURLStr
    }

    if err := config.DB.Save(&kandidat).Error; err != nil {
        // Rollback foto baru jika gagal simpan DB
        localPath, _ := c.Get("photo_local_path")
        if localPathStr, ok := localPath.(string); ok && localPathStr != "" {
            os.Remove(localPathStr)
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengupdate data kandidat"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "Berhasil mengupdate kandidat",
        "data":    kandidat,
    })
}

func DeleteKandidat(c *gin.Context) {
    clientID, _ := c.Get("userID")
    kandidatID := c.Param("id")

    var kandidat models.Kandidat
    if err := config.DB.Preload("Pemilu").Where("id = ?", kandidatID).First(&kandidat).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Kandidat tidak ditemukan"})
        return
    }

    if kandidat.Pemilu.ClientID != clientID.(string) {
        c.JSON(http.StatusForbidden, gin.H{"error": "Akses ditolak: Anda tidak berhak menghapus kandidat ini"})
        return
    }

    // PERBAIKAN: Hapus file dari fisik storage jika ada
    if kandidat.PhotoURL != "" {
        filePath := "." + kandidat.PhotoURL
        if err := os.Remove(filePath); err != nil {
            // Print log saja, tidak perlu menggagalkan proses hapus DB jika file fisik sudah hilang
            // fmt.Println("Warning: Gagal menghapus file fisik", err) 
        }
    }

    if err := config.DB.Delete(&kandidat).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus kandidat"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "Berhasil menghapus kandidat beserta fotonya",
    })
}