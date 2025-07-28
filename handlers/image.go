package handler

import (
    "bytes"
    "image"
    "image/color"
    "image/draw"
    "image/png"
    "io/ioutil"
    "net/http"

    "github.com/disintegration/imaging"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"

    "github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/models"
)

// SupplierImageUploadHandler handles supplier image upload, cropping and watermarking
func SupplierImageUploadHandler(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        email := c.Param("email")
        file, _, err := c.Request.FormFile("image")
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Image file is required"})
            return
        }
        defer file.Close()

        // Read image bytes from uploaded file
        imgBytes, err := ioutil.ReadAll(file)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read image file"})
            return
        }

        // Decode image
        img, _, err := image.Decode(bytes.NewReader(imgBytes))
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image format"})
            return
        }

        // Step 1: Crop the image to square (center crop)
        minSide := img.Bounds().Dx()
        if img.Bounds().Dy() < minSide {
            minSide = img.Bounds().Dy()
        }
        cropRect := image.Rect(
            (img.Bounds().Dx()-minSide)/2,
            (img.Bounds().Dy()-minSide)/2,
            (img.Bounds().Dx()+minSide)/2,
            (img.Bounds().Dy()+minSide)/2,
        )
        croppedImg := imaging.Crop(img, cropRect)

        // Optional: Resize cropped image to standard size (e.g. 256x256)
        resizedImg := imaging.Resize(croppedImg, 256, 256, imaging.Lanczos)

        // Step 2: Add watermark text or image
        // Here we add a simple semi-transparent text watermark
        // For better text watermark (with fonts), you may use freetype library.
        // For this example, we'll add a simple rectangle watermark

        watermark := imaging.New(100, 30, color.NRGBA{0, 0, 0, 80}) // semi-transparent black rectangle

        // Draw watermark on bottom-right corner
        offset := image.Pt(resizedImg.Bounds().Dx()-watermark.Bounds().Dx()-10, resizedImg.Bounds().Dy()-watermark.Bounds().Dy()-10)
        result := image.NewNRGBA(resizedImg.Bounds())
        draw.Draw(result, resizedImg.Bounds(), resizedImg, image.Point{}, draw.Src)
        draw.Draw(result, watermark.Bounds().Add(offset), watermark, image.Point{}, draw.Over)

        // Optionally add text on the watermark rectangle (not shown here - requires freetype or other lib)

        // Encode final image to PNG bytes
        var buf bytes.Buffer
        if err := png.Encode(&buf, result); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode image"})
            return
        }

        // Save image bytes to supplier record in DB
        err = models.UpdateSupplierImage(email, buf.Bytes(), db)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update image in database"})
            return
        }

        c.JSON(http.StatusOK, gin.H{"message": "Image uploaded, cropped, and watermarked successfully"})
    }
}
