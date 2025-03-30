package storage

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

// FileStorage is the interface for file storage operations
type FileStorage interface {
	// Upload uploads a file and returns its path and public URL
	Upload(fileContent io.Reader, fileName, contentType string) (string, string, error)

	// Delete deletes a file
	Delete(storagePath string) error

	// GetPublicURL returns the public URL for a file
	GetPublicURL(storagePath string) string
}

// S3Storage implements FileStorage for AWS S3
type S3Storage struct {
	s3Client *s3.S3
	bucket   string
	region   string
}

// NewS3Storage creates a new S3 storage handler
func NewS3Storage(region, bucket, endpoint, accessKey, secretKey string) (*S3Storage, error) {
	config := &aws.Config{
		Region: aws.String(region),
		Credentials: credentials.NewStaticCredentials(
			accessKey,
			secretKey,
			"",
		),
	}

	// Use custom endpoint if provided (for MinIO or other S3-compatible services)
	if endpoint != "" {
		config.Endpoint = aws.String(endpoint)
		config.S3ForcePathStyle = aws.Bool(true)
	}

	sess, err := session.NewSession(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 session: %w", err)
	}

	s3Client := s3.New(sess)

	// Check if bucket exists, create if not
	_, err = s3Client.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})

	if err != nil {
		_, err = s3Client.CreateBucket(&s3.CreateBucketInput{
			Bucket: aws.String(bucket),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create S3 bucket: %w", err)
		}
	}

	return &S3Storage{
		s3Client: s3Client,
		bucket:   bucket,
		region:   region,
	}, nil
}

// Upload uploads a file to S3
func (s *S3Storage) Upload(fileContent io.Reader, fileName, contentType string) (string, string, error) {
	// Read file content into a buffer to support seeking
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, fileContent)
	if err != nil {
		return "", "", fmt.Errorf("failed to read file content: %w", err)
	}
	body := bytes.NewReader(buf.Bytes()) // Convert buffer to io.ReadSeeker

	// Generate a unique file path
	key := fmt.Sprintf("uploads/%s/%s", time.Now().Format("2006/01/02"), uuid.New().String())

	// Add file extension if present
	if ext := filepath.Ext(fileName); ext != "" {
		key += ext
	}

	// Upload to S3
	_, err = s.s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})

	if err != nil {
		return "", "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	return key, s.GetPublicURL(key), nil
}

// Delete deletes a file from S3
func (s *S3Storage) Delete(storagePath string) error {
	_, err := s.s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(storagePath),
	})

	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}

	return nil
}

// GetPublicURL returns the public URL for a file
func (s *S3Storage) GetPublicURL(storagePath string) string {
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.bucket, storagePath)
}

// LocalStorage implements FileStorage for local file system
type LocalStorage struct {
	basePath string
	baseURL  string
}

// NewLocalStorage creates a new local storage handler
func NewLocalStorage(basePath, baseURL string) (*LocalStorage, error) {
	// Create base directory if it doesn't exist
	err := os.MkdirAll(basePath, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &LocalStorage{
		basePath: basePath,
		baseURL:  baseURL,
	}, nil
}

// Upload uploads a file to local storage
func (l *LocalStorage) Upload(fileContent io.Reader, fileName, contentType string) (string, string, error) {
	// Generate a unique file path
	dir := filepath.Join(l.basePath, time.Now().Format("2006/01/02"))
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return "", "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Generate a unique file name
	uniqueName := uuid.New().String()

	// Add file extension if present
	if ext := filepath.Ext(fileName); ext != "" {
		uniqueName = uniqueName + ext
	}

	fullPath := filepath.Join(dir, uniqueName)
	relativePath := filepath.Join(time.Now().Format("2006/01/02"), uniqueName)

	// Create the file
	file, err := os.Create(fullPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy the content
	_, err = io.Copy(file, fileContent)
	if err != nil {
		return "", "", fmt.Errorf("failed to write file content: %w", err)
	}

	return relativePath, l.GetPublicURL(relativePath), nil
}

// Delete deletes a file from local storage
func (l *LocalStorage) Delete(storagePath string) error {
	fullPath := filepath.Join(l.basePath, storagePath)

	err := os.Remove(fullPath)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// GetPublicURL returns the public URL for a file
func (l *LocalStorage) GetPublicURL(storagePath string) string {
	return fmt.Sprintf("%s/%s", l.baseURL, storagePath)
}
