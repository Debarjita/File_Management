package main

import (
	"learningfilesharing/config"
	"learningfilesharing/jobs"
	"learningfilesharing/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	config.InitDB()

	r := gin.Default()
	routes.SetupRoutes(r)
	go jobs.StartCleanupJob()

	r.Run(":8080") // Run on http://localhost:8080
}
