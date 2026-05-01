package middlewares

import (
    "fmt"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
)

func UploadFile(formField string, uploadDir string) gin.HandlerFunc {
    return func(c *gin.Context) {
        file, err := c.FormFile(formField)
        if err != nil {
            if err == http.ErrMissingFile {
                c.Set(formField+"_url", "")
                c.Set(formField+"_local_path", "")
                c.Next()
                return
            }
            c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Gagal membaca file"})
            return
        }

        if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
            c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat direktori upload"})
            return
        }

        // Ekstensi tetap diambil untuk nama file, tapi pengecekan format (if ext != ...) DIHAPUS.
        ext := strings.ToLower(filepath.Ext(file.Filename))

        // Opsional: Tetap ada batas ukuran file agar server tidak jebol (misal 5MB)
        if file.Size > 5*1024*1024 {
            c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Ukuran file maksimal 5MB"})
            return
        }

        // Menyimpan file dengan nama unik timestamp + ekstensi aslinya
        filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
        savePath := filepath.Join(uploadDir, filename)

        if err := c.SaveUploadedFile(file, savePath); err != nil {
            c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan file"})
            return
        }

        fileUrl := "/" + filepath.ToSlash(savePath)

        c.Set(formField+"_url", fileUrl)
        c.Set(formField+"_local_path", savePath)
        c.Next()
    }
}