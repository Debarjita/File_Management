package db

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Database represents the database connection
type Database struct {
	DB *sqlx.DB
}

// New creates a new database connection
func New(connectionString string) (*Database, error) {
	db, err := sqlx.Connect("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return &Database{DB: db}, nil
}

// Init initializes the database schema
func (d *Database) Init() error {
	// Create users table
	_, err := d.DB.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		email VARCHAR(255) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Create files table
	_, err = d.DB.Exec(`
	CREATE TABLE IF NOT EXISTS files (
		id VARCHAR(36) PRIMARY KEY,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		name VARCHAR(255) NOT NULL,
		size BIGINT NOT NULL,
		content_type VARCHAR(255) NOT NULL,
		storage_path VARCHAR(512) NOT NULL,
		public_url VARCHAR(512) NOT NULL,
		is_public BOOLEAN DEFAULT FALSE,
		expires_at TIMESTAMP WITH TIME ZONE,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return fmt.Errorf("failed to create files table: %w", err)
	}

	// Create shared_files table
	_, err = d.DB.Exec(`
	CREATE TABLE IF NOT EXISTS shared_files (
		id VARCHAR(36) PRIMARY KEY,
		file_id VARCHAR(36) NOT NULL REFERENCES files(id) ON DELETE CASCADE,
		share_url VARCHAR(512) NOT NULL,
		expires_at TIMESTAMP WITH TIME ZONE,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return fmt.Errorf("failed to create shared_files table: %w", err)
	}

	// Create indexes for performance
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_files_user_id ON files(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_files_name ON files(name)",
		"CREATE INDEX IF NOT EXISTS idx_files_content_type ON files(content_type)",
		"CREATE INDEX IF NOT EXISTS idx_files_created_at ON files(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_shared_files_file_id ON shared_files(file_id)",
		"CREATE INDEX IF NOT EXISTS idx_files_expires_at ON files(expires_at)",
	}

	for _, idx := range indexes {
		_, err = d.DB.Exec(idx)
		if err != nil {
			log.Printf("Failed to create index: %v", err)
		}
	}

	return nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.DB.Close()
}
