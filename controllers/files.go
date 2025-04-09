package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"learningfilesharing/config"
	"learningfilesharing/models"

	"github.com/gin-gonic/gin"
)

func UploadFile(c *gin.Context) {
	// Get user ID from context (set by middleware)
	userID := c.GetUint("user_id")

	// Get the uploaded file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upload file"})
		return
	}
	defer file.Close()

	// Save the file locally (you can replace this with S3 logic later)
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), header.Filename) //Adds a timestamp to avoid filename collisions.
	savePath := filepath.Join("uploads", filename)

	//take the file uploaded by the user and saving it in your system.
	out, err := os.Create(savePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to save file"})
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write file"})
		return
	}

	// Set expiration time (e.g., 24 hour from now)
	expiryDuration := 24 * time.Hour
	expiresAt := time.Now().Add(expiryDuration)

	// Save metadata in DB
	fileRecord := models.File{
		FileName:  header.Filename,
		FileType:  header.Header.Get("Content-Type"),
		Size:      header.Size,
		URL:       "/uploads/" + filename,
		UserID:    userID,
		ExpiresAt: &expiresAt,
	}

	if err := config.DB.Create(&fileRecord).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save metadata"})
		return
	}

	// Return public URL and expiration info
	c.JSON(http.StatusOK, gin.H{
		"message":    "File uploaded successfully",
		"url":        fileRecord.URL,
		"size":       strconv.FormatInt(header.Size, 10),
		"expires_at": expiresAt.Format(time.RFC3339), // optional: format it nicely
	})

}

// user can see only their uploaded files
func GetUserFiles(c *gin.Context) {
	userID := c.GetUint("user_id")

	var files []models.File
	if err := config.DB.Where("user_id = ?", userID).Find(&files).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch files"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"files": files})
}

// It handles the request when someone tries to access a file by its shareable link
func ShareFile(c *gin.Context) {
	id := c.Param("id")

	//Look Up File in Database
	var file models.File
	if err := config.DB.First(&file, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	//Check Expiry of File Link
	if file.ExpiresAt != nil && time.Now().After(*file.ExpiresAt) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Link has expired"})
		return
	}

	//Return the File Details
	c.JSON(http.StatusOK, gin.H{
		"file_name":  file.FileName,
		"url":        file.URL,
		"expires_at": file.ExpiresAt,
	})
}

// Search files  from DB, caching and pagination
func SearchFiles(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get query parameters
	fileName := c.Query("name")
	fileType := c.Query("type")
	uploadDate := c.Query("date") // Expected format: YYYY-MM-DD

	// Pagination parameters
	limitParam := c.DefaultQuery("limit", "20")
	offsetParam := c.DefaultQuery("offset", "0")

	limit, err1 := strconv.Atoi(limitParam)
	offset, err2 := strconv.Atoi(offsetParam)

	if err1 != nil || err2 != nil || limit < 1 || offset < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pagination parameters"})
		return
	}

	// Create a cache key
	cacheKey := fmt.Sprintf("files:%v:name=%s:type=%s:date=%s", userID, fileName, fileType, uploadDate)

	// Try to get data from Redis
	cached, err := config.RedisClient.Get(config.Ctx, cacheKey).Result()
	if err == nil {
		// Cache hit
		var cachedFiles []models.File
		if err := json.Unmarshal([]byte(cached), &cachedFiles); err == nil {
			c.JSON(http.StatusOK, cachedFiles)
			return
		}
	}

	//// Cache miss → fetch from DB
	var files []models.File
	query := config.DB.Where("user_id = ?", userID)

	if fileName != "" {
		query = query.Where("file_name ILIKE ?", "%"+fileName+"%")
	}
	if fileType != "" {
		query = query.Where("file_type = ?", fileType)
	}
	if uploadDate != "" {
		parsedDate, err := time.Parse("2006-01-02", uploadDate)
		if err == nil {
			start := parsedDate
			end := parsedDate.Add(24 * time.Hour)
			query = query.Where("created_at >= ? AND created_at < ?", start, end)
		}
	}

	// Apply pagination
	query = query.Limit(limit).Offset(offset)

	// Fetch results
	if err := query.Find(&files).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch files"})
		return
	}

	// Save result to Redis with TTL of 5 minutes
	jsonData, err := json.Marshal(files)
	if err == nil {
		config.RedisClient.Set(config.Ctx, cacheKey, jsonData, 5*time.Minute)
	}

	c.JSON(http.StatusOK, gin.H{
		"results": files,
		"limit":   limit,
		"offset":  offset,
	})
}

// FILE RENAME AND CACHE INVALIDATION
func UpdateFile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	fileID := c.Param("id")
	var request struct {
		NewName string `json:"new_name"`
	}

	if err := c.ShouldBindJSON(&request); err != nil || request.NewName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var file models.File
	if err := config.DB.Where("id = ? AND user_id = ?", fileID, userID).First(&file).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	file.FileName = request.NewName
	if err := config.DB.Save(&file).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update file"})
		return
	}

	// Invalidate Redis Cache - delete all matching keys for that user
	pattern := fmt.Sprintf("files:%v:*", userID)
	keys, err := config.RedisClient.Keys(config.Ctx, pattern).Result()
	if err == nil {
		for _, key := range keys {
			if delErr := config.RedisClient.Del(config.Ctx, key).Err(); delErr != nil {
				log.Println("Failed to delete key:", key, delErr)
			}
		}
	} else {
		log.Println("Failed to fetch Redis keys:", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "File renamed successfully"})
}
