package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config represents the application configuration
type Config struct {
	ServerPort          string
	DatabaseURL         string
	RedisURL            string
	JWTSecret           string
	JWTExpiration       time.Duration
	S3Bucket            string
	S3Region            string
	S3Endpoint          string
	S3AccessKey         string
	S3SecretKey         string
	UseLocalStorage     bool
	LocalStoragePath    string
	LocalStorageBaseURL string
	CacheTTL            time.Duration
	BaseShareURL        string
	RateLimit           int
}

// Load loads the configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if present
	_ = godotenv.Load()

	// Server config
	serverPort := getEnv("SERVER_PORT", "8080")

	// Database config
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/filestore?sslmode=disable")

	// Redis config
	redisURL := getEnv("REDIS_URL", "redis://localhost:6379/0")

	// JWT config
	jwtSecret := getEnv("JWT_SECRET", "your-secret-key")
	jwtExpirationHours, _ := strconv.Atoi(getEnv("JWT_EXPIRATION_HOURS", "24"))

	// S3 config
	s3Bucket := getEnv("S3_BUCKET", "filestore")
	s3Region := getEnv("S3_REGION", "us-east-1")
	s3Endpoint := getEnv("S3_ENDPOINT", "")
	s3AccessKey := getEnv("S3_ACCESS_KEY", "")
	s3SecretKey := getEnv("S3_SECRET_KEY", "")

	// Storage config
	useLocalStorage, _ := strconv.ParseBool(getEnv("USE_LOCAL_STORAGE", "false"))
	localStoragePath := getEnv("LOCAL_STORAGE_PATH", "./storage")

	// Cache config
	cacheTTLMinutes, _ := strconv.Atoi(getEnv("CACHE_TTL_MINUTES", "5"))

	// Rate limiting
	rateLimit, _ := strconv.Atoi(getEnv("RATE_LIMIT", "100"))

	//baseshare url
	baseShareURL := getEnv("BASE_SHARE_URL", "http://localhost:8080")

	// Create config
	config := &Config{
		ServerPort:       serverPort,
		DatabaseURL:      dbURL,
		RedisURL:         redisURL,
		JWTSecret:        jwtSecret,
		JWTExpiration:    time.Duration(jwtExpirationHours) * time.Hour,
		S3Bucket:         s3Bucket,
		S3Region:         s3Region,
		S3Endpoint:       s3Endpoint,
		S3AccessKey:      s3AccessKey,
		S3SecretKey:      s3SecretKey,
		UseLocalStorage:  useLocalStorage,
		LocalStoragePath: localStoragePath,
		CacheTTL:         time.Duration(cacheTTLMinutes) * time.Minute,
		RateLimit:        rateLimit,
		BaseShareURL:     baseShareURL,
	}

	// Ensure local storage directory exists if using local storage
	if config.UseLocalStorage {
		if err := os.MkdirAll(config.LocalStoragePath, 0755); err != nil {
			return nil, fmt.Errorf("failed to create local storage directory: %w", err)
		}
	}

	return config, nil
}

// Helper to get environment variable with fallback
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
