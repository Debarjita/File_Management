package db

import (
	"fmt"
	"time"

	"file-sharing-platform/internal/models"

	"github.com/google/uuid"
)

// FileRepository handles file-related database operations
type FileRepository struct {
	db *Database
}

// NewFileRepository creates a new file repository
func NewFileRepository(db *Database) *FileRepository {
	return &FileRepository{db: db}
}

// CreateFile adds a new file to the database
func (r *FileRepository) CreateFile(file *models.File) error {
	if file.ID == "" {
		file.ID = uuid.New().String()
	}

	now := time.Now()
	file.CreatedAt = now
	file.UpdatedAt = now

	query := `
		INSERT INTO files (
			id, user_id, name, size, content_type, storage_path, 
			public_url, is_public, expires_at, created_at, updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
	`

	_, err := r.db.DB.Exec(
		query,
		file.ID,
		file.UserID,
		file.Name,
		file.Size,
		file.ContentType,
		file.StoragePath,
		file.PublicURL,
		file.IsPublic,
		file.ExpiresAt,
		file.CreatedAt,
		file.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	return nil
}

// GetFileByID retrieves a file by ID
func (r *FileRepository) GetFileByID(id string) (*models.File, error) {
	var file models.File
	query := `
		SELECT id, user_id, name, size, content_type, storage_path, 
		       public_url, is_public, expires_at, created_at, updated_at
		FROM files
		WHERE id = $1
	`

	err := r.db.DB.Get(&file, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get file by ID: %w", err)
	}

	return &file, nil
}

// GetFilesByUserID gets all files for a user
func (r *FileRepository) GetFilesByUserID(userID int64, limit, offset int) ([]models.File, error) {
	files := []models.File{}
	query := `
		SELECT id, user_id, name, size, content_type, 
		       public_url, is_public, expires_at, created_at, updated_at
		FROM files
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	err := r.db.DB.Select(&files, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get files by user ID: %w", err)
	}

	return files, nil
}

// UpdateFile updates a file in the database
func (r *FileRepository) UpdateFile(file *models.File) error {
	file.UpdatedAt = time.Now()

	query := `
		UPDATE files
		SET name = $1, is_public = $2, expires_at = $3, updated_at = $4
		WHERE id = $5 AND user_id = $6
	`

	result, err := r.db.DB.Exec(
		query,
		file.Name,
		file.IsPublic,
		file.ExpiresAt,
		file.UpdatedAt,
		file.ID,
		file.UserID,
	)

	if err != nil {
		return fmt.Errorf("failed to update file: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if count == 0 {
		return fmt.Errorf("file not found or not owned by user")
	}

	return nil
}

// DeleteFile deletes a file from the database
func (r *FileRepository) DeleteFile(id string, userID int64) error {
	query := `DELETE FROM files WHERE id = $1 AND user_id = $2`

	result, err := r.db.DB.Exec(query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if count == 0 {
		return fmt.Errorf("file not found or not owned by user")
	}

	return nil
}

// SearchFiles searches for files by various criteria
func (r *FileRepository) SearchFiles(userID int64, search *models.SearchFilesRequest) ([]models.File, error) {
	files := []models.File{}

	// Base query with user filter
	query := `
		SELECT id, user_id, name, size, content_type, 
		       public_url, is_public, expires_at, created_at, updated_at
		FROM files
		WHERE user_id = $1
	`

	// Build dynamic query based on search params
	params := []interface{}{userID}
	paramCount := 2 // Starting from 2 because $1 is already used for userID

	// Add name search if provided
	if search.Query != "" {
		query += fmt.Sprintf(" AND name ILIKE $%d", paramCount)
		params = append(params, "%"+search.Query+"%")
		paramCount++
	}

	// Add file type filter if provided
	if search.FileType != "" {
		query += fmt.Sprintf(" AND content_type ILIKE $%d", paramCount)
		params = append(params, "%"+search.FileType+"%")
		paramCount++
	}

	// Add date range if provided
	if search.StartDate != "" {
		startDate, err := time.Parse("2006-01-02", search.StartDate)
		if err == nil {
			query += fmt.Sprintf(" AND created_at >= $%d", paramCount)
			params = append(params, startDate)
			paramCount++
		}
	}

	if search.EndDate != "" {
		endDate, err := time.Parse("2006-01-02", search.EndDate)
		if err == nil {
			// Add one day to include the end date
			endDate = endDate.AddDate(0, 0, 1)
			query += fmt.Sprintf(" AND created_at < $%d", paramCount)
			params = append(params, endDate)
			paramCount++
		}
	}

	// Add ordering and pagination
	query += " ORDER BY created_at DESC LIMIT $" + fmt.Sprintf("%d", paramCount)
	params = append(params, search.Limit)
	paramCount++

	query += " OFFSET $" + fmt.Sprintf("%d", paramCount)
	params = append(params, search.Offset)

	// Execute query
	err := r.db.DB.Select(&files, query, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to search files: %w", err)
	}

	return files, nil
}

// CreateShareLink creates a share link for a file
func (r *FileRepository) CreateShareLink(fileID string, expiresAt time.Time) (*models.SharedFile, error) {
	sharedFile := models.SharedFile{
		ID:        uuid.New().String(),
		FileID:    fileID,
		ShareURL:  uuid.New().String(), // Use UUID as unique share URL path
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	query := `
		INSERT INTO shared_files (id, file_id, share_url, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.DB.Exec(
		query,
		sharedFile.ID,
		sharedFile.FileID,
		sharedFile.ShareURL,
		sharedFile.ExpiresAt,
		sharedFile.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create share link: %w", err)
	}

	return &sharedFile, nil
}

// GetSharedFile gets a shared file by share URL
func (r *FileRepository) GetSharedFile(shareURL string) (*models.SharedFile, error) {
	var sharedFile models.SharedFile
	query := `
		SELECT id, file_id, share_url, expires_at, created_at
		FROM shared_files
		WHERE share_url = $1
	`

	err := r.db.DB.Get(&sharedFile, query, shareURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get shared file: %w", err)
	}

	return &sharedFile, nil
}

// GetExpiredFiles gets all expired files
func (r *FileRepository) GetExpiredFiles(batchSize int) ([]models.File, error) {
	files := []models.File{}
	query := `
		SELECT id, user_id, name, size, content_type, storage_path, 
		       public_url, is_public, expires_at, created_at, updated_at
		FROM files
		WHERE expires_at IS NOT NULL AND expires_at < NOW()
		LIMIT $1
	`

	err := r.db.DB.Select(&files, query, batchSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get expired files: %w", err)
	}

	return files, nil
}

// DeleteExpiredFiles deletes expired files
func (r *FileRepository) DeleteExpiredFiles(batchSize int) (int, error) {
	query := `
		DELETE FROM files
		WHERE id IN (
			SELECT id FROM files
			WHERE expires_at IS NOT NULL AND expires_at < NOW()
			LIMIT $1
		)
	`

	result, err := r.db.DB.Exec(query, batchSize)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired files: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return int(count), nil
}
