package api

import (
	"net/http"
	"strconv"

	"file-sharing-platform/internal/auth"
	"file-sharing-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type FileHandler struct {
	fileService *service.FileService
}

func NewFileHandler(fileService *service.FileService) *FileHandler {
	return &FileHandler{
		fileService: fileService,
	}
}

func (h *FileHandler) UploadFile(c *gin.Context) {
	userID, err := auth.GetUserIDFromContext(c) // Updated to use gin.Context
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	err = c.Request.ParseMultipartForm(10 << 20)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to parse form"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}
	defer file.Close()

	ctx := c.Request.Context()
	fileInfo, err := h.fileService.UploadFile(ctx, userID, header.Filename, header.Size, header.Header.Get("Content-Type"), file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, fileInfo)
}

func (h *FileHandler) GetUserFiles(c *gin.Context) {
	userID, err := auth.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	ctx := c.Request.Context()
	files, err := h.fileService.GetUserFiles(ctx, userID, 0, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving files"})
		return
	}

	c.JSON(http.StatusOK, files)
}

func (h *FileHandler) ShareFile(c *gin.Context) {
	userID, err := auth.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	fileID, err := strconv.ParseInt(c.Param("file_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	expirationHours := 24
	if expStr := c.Query("expires_in"); expStr != "" {
		if exp, err := strconv.Atoi(expStr); err == nil && exp > 0 {
			expirationHours = exp
		}
	}

	ctx := c.Request.Context()
	shareURL, err := h.fileService.ShareFile(ctx, strconv.FormatInt(userID, 10), fileID, strconv.Itoa(expirationHours))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error sharing file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"share_url": shareURL.ShareURL})
}

func (h *FileHandler) GetSharedFile(c *gin.Context) {
	shareToken := c.Param("share_token")

	ctx := c.Request.Context()
	fileInfo, err := h.fileService.GetFile(ctx, shareToken)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found or share expired"})
		return
	}

	c.Redirect(http.StatusFound, fileInfo.PublicURL)
}

func (h *FileHandler) DeleteFile(c *gin.Context) {
	userID, err := auth.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	fileID, err := strconv.ParseInt(c.Param("file_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	ctx := c.Request.Context()
	err = h.fileService.DeleteFile(ctx, strconv.FormatInt(userID, 10), fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting file"})
		return
	}

	c.Status(http.StatusNoContent)
}
