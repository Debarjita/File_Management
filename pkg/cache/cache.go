package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"file-sharing-platform/internal/models"

	"github.com/go-redis/redis/v8"
)

// Cache is the interface for caching operations
type Cache interface {
	// Set sets a value in the cache with expiration
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error

	// Get gets a value from the cache
	Get(ctx context.Context, key string, dest interface{}) error

	// Delete deletes a value from the cache
	Delete(ctx context.Context, key string) error

	// Close closes the cache connection
	Close() error
}

// RedisCache implements Cache for Redis
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis cache
func NewRedisCache(redisURL string) (*RedisCache, error) {
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(options)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{client: client}, nil
}

// Set sets a value in Redis
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	err = c.client.Set(ctx, key, data, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set value in cache: %w", err)
	}

	return nil
}

// Get gets a value from Redis
func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return fmt.Errorf("key not found in cache")
		}
		return fmt.Errorf("failed to get value from cache: %w", err)
	}

	err = json.Unmarshal(data, dest)
	if err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return nil
}

// Delete deletes a value from Redis
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete value from cache: %w", err)
	}

	return nil
}

// Close closes the Redis connection
func (c *RedisCache) Close() error {
	return c.client.Close()
}

// MemoryCache implements Cache for in-memory caching
type MemoryCache struct {
	data map[string]cacheItem
}

type cacheItem struct {
	value      []byte
	expiration time.Time
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		data: make(map[string]cacheItem),
	}

	// Start a cleanup goroutine
	go cache.cleanupLoop()

	return cache
}

// Set sets a value in memory cache
func (c *MemoryCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	c.data[key] = cacheItem{
		value:      data,
		expiration: time.Now().Add(expiration),
	}

	return nil
}

// Get gets a value from memory cache
func (c *MemoryCache) Get(ctx context.Context, key string, dest interface{}) error {
	item, ok := c.data[key]
	if !ok {
		return fmt.Errorf("key not found in cache")
	}

	// Check if expired
	if time.Now().After(item.expiration) {
		delete(c.data, key)
		return fmt.Errorf("key not found in cache (expired)")
	}

	err := json.Unmarshal(item.value, dest)
	if err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return nil
}

// Delete deletes a value from memory cache
func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	delete(c.data, key)
	return nil
}

// Close is a no-op for memory cache
func (c *MemoryCache) Close() error {
	return nil
}

// cleanupLoop periodically removes expired items
func (c *MemoryCache) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}
}

// cleanup removes expired items
func (c *MemoryCache) cleanup() {
	now := time.Now()

	for key, item := range c.data {
		if now.After(item.expiration) {
			delete(c.data, key)
		}
	}
}

// FileCache provides caching for file metadata
type FileCache struct {
	cache      Cache
	expiration time.Duration
}

// NewFileCache creates a new file cache
func NewFileCache(cache Cache, expiration time.Duration) *FileCache {
	return &FileCache{
		cache:      cache,
		expiration: expiration,
	}
}

// GetFile gets a file from cache
func (c *FileCache) GetFile(ctx context.Context, fileID string) (*models.File, bool) {
	var file models.File

	key := fmt.Sprintf("file:%s", fileID)
	err := c.cache.Get(ctx, key, &file)
	if err != nil {
		return nil, false
	}

	return &file, true
}

// SetFile sets a file in cache
func (c *FileCache) SetFile(ctx context.Context, file *models.File) error {
	key := fmt.Sprintf("file:%s", file.ID)
	return c.cache.Set(ctx, key, file, c.expiration)
}

// InvalidateFile removes a file from cache
func (c *FileCache) InvalidateFile(ctx context.Context, fileID string) error {
	key := fmt.Sprintf("file:%s", fileID)
	return c.cache.Delete(ctx, key)
}

// GetUserFiles gets user files from cache
func (c *FileCache) GetUserFiles(ctx context.Context, userID int64) ([]models.File, bool) {
	var files []models.File

	key := fmt.Sprintf("user_files:%d", userID)
	err := c.cache.Get(ctx, key, &files)
	if err != nil {
		return nil, false
	}

	return files, true
}

// SetUserFiles sets user files in cache
func (c *FileCache) SetUserFiles(ctx context.Context, userID int64, files []models.File) error {
	key := fmt.Sprintf("user_files:%d", userID)
	return c.cache.Set(ctx, key, files, c.expiration)
}

// InvalidateUserFiles removes user files from cache
func (c *FileCache) InvalidateUserFiles(ctx context.Context, userID int64) error {
	key := fmt.Sprintf("user_files:%d", userID)
	return c.cache.Delete(ctx, key)
}
