package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID       int64  `db:"id" json:"id"`
	Email    string `db:"email" json:"email"`
	Password string `db:"password" json:"-"` // Hashed password, not returned in JSON

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// File represents a file stored in the system
type File struct {
	ID          string    `db:"id" json:"id"`
	UserID      int64     `db:"user_id" json:"user_id"`
	Name        string    `db:"name" json:"name"`
	Size        int64     `db:"size" json:"size"`
	ContentType string    `db:"content_type" json:"content_type"`
	StoragePath string    `db:"storage_path" json:"storage_path,omitempty"`
	PublicURL   string    `db:"public_url" json:"public_url"`
	IsPublic    bool      `db:"is_public" json:"is_public"`
	ExpiresAt   time.Time `db:"expires_at" json:"expires_at,omitempty"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// SharedFile represents a file share link
type SharedFile struct {
	ID        string    `db:"id" json:"id"`
	FileID    string    `db:"file_id" json:"file_id"`
	ShareURL  string    `db:"share_url" json:"share_url"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// AuthRequest represents authentication request data
type AuthRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// RegisterRequest represents user registration data
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// AuthResponse represents authentication response data
type AuthResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

// FileUploadResponse represents the response after a file upload
type FileUploadResponse struct {
	FileID    string `json:"file_id"`
	PublicURL string `json:"public_url"`
}

// ShareFileRequest represents a request to share a file
type ShareFileRequest struct {
	ExpiresIn string `json:"expires_in"` // Duration string like "24h", "7d"
}

// ShareFileResponse represents the response for a file share request
type ShareFileResponse struct {
	ShareURL  string    `json:"share_url"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// SearchFilesRequest represents a request to search for files
type SearchFilesRequest struct {
	Query     string `form:"q"`
	FileType  string `form:"type"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
	Limit     int    `form:"limit,default=20"`
	Offset    int    `form:"offset,default=0"`
}
