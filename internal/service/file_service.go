package service

import (
	"context"
	"fmt"
	"io"
	"time"

	"file-sharing-platform/internal/db"
	"file-sharing-platform/internal/models"
	"file-sharing-platform/pkg/cache"
	"file-sharing-platform/pkg/storage"
)

// FileService handles file operations
type FileService struct {
	fileRepo     *db.FileRepository
	storage      storage.FileStorage
	cache        *cache.FileCache
	baseShareURL string
}

// NewFileService creates a new file service
func NewFileService(fileRepo *db.FileRepository, storage storage.FileStorage, cache *cache.FileCache, baseShareURL string) *FileService {
	return &FileService{
		fileRepo:     fileRepo,
		storage:      storage,
		cache:        cache,
		baseShareURL: baseShareURL,
	}
}

// UploadFile uploads a file
func (s *FileService) UploadFile(ctx context.Context, userID int64, fileName string, fileSize int64, contentType string, fileContent io.Reader) (*models.File, error) {
	// Upload the file to storage
	storagePath, publicURL, err := s.storage.Upload(fileContent, fileName, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Create file metadata in database
	file := &models.File{
		UserID:      userID,
		Name:        fileName,
		Size:        fileSize,
		ContentType: contentType,
		StoragePath: storagePath,
		PublicURL:   publicURL,
		IsPublic:    false,
	}

	// Save to database
	err = s.fileRepo.CreateFile(file)
	if err != nil {
		// Try to cleanup the storage if database insertion fails
		_ = s.storage.Delete(storagePath)
		return nil, fmt.Errorf("failed to save file metadata: %w", err)
	}

	// Invalidate user files cache
	_ = s.cache.InvalidateUserFiles(ctx, userID)

	// Cache the new file
	_ = s.cache.SetFile(ctx, file)

	return file, nil
}

// GetUserFiles gets files for a user
func (s *FileService) GetUserFiles(ctx context.Context, userID int64, limit, offset int) ([]models.File, error) {
	// Try to get from cache first
	if limit == 0 {
		limit = 20 // Default limit
	}

	// Only use cache for first page
	if offset == 0 {
		files, found := s.cache.GetUserFiles(ctx, userID)
		if found {
			return files, nil
		}
	}

	// Get from database if not in cache
	files, err := s.fileRepo.GetFilesByUserID(userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user files: %w", err)
	}

	// Cache the files for first page
	if offset == 0 {
		_ = s.cache.SetUserFiles(ctx, userID, files)
	}

	return files, nil
}

// GetFile gets a file by ID
func (s *FileService) GetFile(ctx context.Context, fileID string) (*models.File, error) {
	// Try to get from cache first
	file, found := s.cache.GetFile(ctx, fileID)
	if found {
		return file, nil
	}

	// Get from database if not in cache
	file, err := s.fileRepo.GetFileByID(fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	// Cache the file
	_ = s.cache.SetFile(ctx, file)

	return file, nil
}

// UpdateFile updates a file
func (s *FileService) UpdateFile(ctx context.Context, fileID string, userID int64, updates map[string]interface{}) (*models.File, error) {
	// Get the file
	file, err := s.fileRepo.GetFileByID(fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	// Check ownership
	if file.UserID != userID {
		return nil, fmt.Errorf("not authorized to update this file")
	}

	// Apply updates
	if name, ok := updates["name"].(string); ok {
		file.Name = name
	}

	if isPublic, ok := updates["is_public"].(bool); ok {
		file.IsPublic = isPublic
	}

	if expiresIn, ok := updates["expires_in"].(string); ok {
		duration, err := time.ParseDuration(expiresIn)
		if err == nil {
			file.ExpiresAt = time.Now().Add(duration)
		}
	}

	// Update in database
	err = s.fileRepo.UpdateFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to update file: %w", err)
	}

	// Invalidate caches
	_ = s.cache.InvalidateFile(ctx, fileID)
	_ = s.cache.InvalidateUserFiles(ctx, userID)

	return file, nil
}

// DeleteFile deletes a file
func (s *FileService) DeleteFile(ctx context.Context, fileID string, userID int64) error {
	// Get the file first
	file, err := s.fileRepo.GetFileByID(fileID)
	if err != nil {
		return fmt.Errorf("failed to get file: %w", err)
	}

	// Check ownership
	if file.UserID != userID {
		return fmt.Errorf("not authorized to delete this file")
	}

	// Delete from storage
	err = s.storage.Delete(file.StoragePath)
	if err != nil {
		return fmt.Errorf("failed to delete file from storage: %w", err)
	}

	// Delete from database
	err = s.fileRepo.DeleteFile(fileID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete file metadata: %w", err)
	}

	// Invalidate caches
	_ = s.cache.InvalidateFile(ctx, fileID)
	_ = s.cache.InvalidateUserFiles(ctx, userID)

	return nil
}

// SearchFiles searches for files
func (s *FileService) SearchFiles(ctx context.Context, userID int64, search *models.SearchFilesRequest) ([]models.File, error) {
	// Search is always from DB as it's dynamic
	files, err := s.fileRepo.SearchFiles(userID, search)
	if err != nil {
		return nil, fmt.Errorf("failed to search files: %w", err)
	}

	return files, nil
}

// ShareFile creates a share link for a file
func (s *FileService) ShareFile(ctx context.Context, fileID string, userID int64, expiresIn string) (*models.SharedFile, error) {
	// Get the file
	file, err := s.GetFile(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	// Check ownership
	if file.UserID != userID {
		return nil, fmt.Errorf("not authorized to share this file")
	}

	// Parse expiration duration
	var expiresAt time.Time
	if expiresIn != "" {
		duration, err := time.ParseDuration(expiresIn)
		if err != nil {
			return nil, fmt.Errorf("invalid expiration format: %w", err)
		}
		expiresAt = time.Now().Add(duration)
	}

	// Create share link
	sharedFile, err := s.fileRepo.CreateShareLink(fileID, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create share link: %w", err)
	}

	// Format the complete share URL
	sharedFile.ShareURL = fmt.Sprintf("%s/shared/%s", s.baseShareURL, sharedFile.ShareURL)

	return sharedFile, nil
}

// GetSharedFile gets a file by share URL
func (s *FileService) GetSharedFile(ctx context.Context, shareID string) (*models.File, error) {
	// Get the shared file record
	sharedFile, err := s.fileRepo.GetSharedFile(shareID)
	if err != nil {
		return nil, fmt.Errorf("shared file not found: %w", err)
	}

	// Check if expired
	if !sharedFile.ExpiresAt.IsZero() && time.Now().After(sharedFile.ExpiresAt) {
		return nil, fmt.Errorf("share link has expired")
	}

	// Get the file
	file, err := s.GetFile(ctx, sharedFile.FileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shared file: %w", err)
	}

	return file, nil
}

// CleanupExpiredFiles deletes expired files
func (s *FileService) CleanupExpiredFiles(ctx context.Context, batchSize int) (int, error) {
	// Get expired files
	files, err := s.fileRepo.GetExpiredFiles(batchSize)
	if err != nil {
		return 0, fmt.Errorf("failed to get expired files: %w", err)
	}

	deletedCount := 0

	// Delete each file from storage
	for _, file := range files {
		err := s.storage.Delete(file.StoragePath)
		if err != nil {
			// Log the error but continue with others
			continue
		}

		// Invalidate cache
		_ = s.cache.InvalidateFile(ctx, file.ID)
		_ = s.cache.InvalidateUserFiles(ctx, file.UserID)

		deletedCount++
	}

	// Delete from database
	if len(files) > 0 {
		count, err := s.fileRepo.DeleteExpiredFiles(batchSize)
		if err != nil {
			return deletedCount, fmt.Errorf("failed to delete expired files from database: %w", err)
		}

		return count, nil
	}

	return deletedCount, nil
}
