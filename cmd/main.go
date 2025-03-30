package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"file-sharing-platform/internal/api"
	"file-sharing-platform/internal/auth"

	//"file-sharing-platform/internal/auth"
	"file-sharing-platform/internal/config"
	"file-sharing-platform/internal/db"
	"file-sharing-platform/internal/middleware"
	"file-sharing-platform/internal/service"
	"file-sharing-platform/internal/websocket"
	"file-sharing-platform/internal/worker"
	"file-sharing-platform/pkg/cache"
	"file-sharing-platform/pkg/storage"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	database, err := db.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run migrations
	if err := database.Init(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize repositories
	userRepo := db.NewUserRepository(database)
	fileRepo := db.NewFileRepository(database)

	// Initialize cache
	var cacheClient cache.Cache
	if cfg.RedisURL != "" {
		cacheClient, err = cache.NewRedisCache(cfg.RedisURL)
		if err != nil {
			log.Fatalf("Failed to initialize Redis cache: %v", err)
		}
	} else {
		cacheClient = cache.NewMemoryCache()
	}

	// Corrected fileCache initialization
	fileCache := cache.NewFileCache(cacheClient, cfg.CacheTTL)
	// Initialize storage
	var storageProvider storage.FileStorage
	if cfg.S3Bucket != "" {
		storageProvider, err = storage.NewS3Storage(cfg.S3Region, cfg.S3Bucket, cfg.S3Endpoint, cfg.S3AccessKey, cfg.S3SecretKey)
	} else {
		storageProvider, err = storage.NewLocalStorage(cfg.LocalStoragePath, cfg.LocalStorageBaseURL)
	}
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Initialize WebSocket hub
	notificationHub := websocket.NewNotificationHub()

	// Initialize file service
	fileService := service.NewFileService(fileRepo, storageProvider, fileCache, cfg.BaseShareURL)

	// Initialize background workers
	fileCleanupWorker := worker.NewFileCleanupWorker(fileService, time.Duration(cfg.CacheTTL)*time.Second, 10)

	go fileCleanupWorker.Start()

	// Initialize JWT authentication
	jwtAuth := auth.NewJWTAuth(cfg.JWTSecret, time.Hour*24)

	if err != nil {
		log.Fatalf("Failed to initialize JWT authentication: %v", err)
	}

	// Initialize API handlers
	authHandler := api.NewAuthHandler(userRepo, jwtAuth)
	fileHandler := api.NewFileHandler(fileService)

	// Initialize router
	router := gin.Default()

	// Apply middleware
	router.Use(middleware.RequestLogger)
	router.Use(middleware.AuthMiddleware(jwtAuth))

	// Auth routes
	router.POST("/api/register", authHandler.Register)
	router.POST("/api/login", authHandler.Login)

	// WebSocket route
	router.GET("/ws/notifications", func(c *gin.Context) {
		notificationHub.HandleWebSocket(c.Writer, c.Request)
	})

	// Public file share route
	router.GET("/share/:share_token", fileHandler.GetSharedFile)

	// Protected routes (require authentication)
	authRoutes := router.Group("/api")
	authRoutes.Use(middleware.AuthMiddleware(jwtAuth))

	authRoutes.POST("/upload", fileHandler.UploadFile)
	authRoutes.GET("/files", fileHandler.GetUserFiles)
	authRoutes.DELETE("/files/:file_id", fileHandler.DeleteFile)
	authRoutes.GET("/share/:file_id", fileHandler.ShareFile)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.ServerPort),
		Handler:      router,
		ReadTimeout:  cfg.CacheTTL,
		WriteTimeout: cfg.CacheTTL,
		IdleTimeout:  cfg.CacheTTL,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on port %s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Shutdown server
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	// Stop background workers
	fileCleanupWorker.Stop()

	log.Println("Server stopped gracefully")
}
