package routes

import (
	"learningfilesharing/controllers"
	"learningfilesharing/middleware"

	"github.com/gin-gonic/gin" //This is the Gin web framework. It helps us define routes, handle HTTP requests, etc.
)

func SetupRoutes(router *gin.Engine) {

	//When someone sends a POST request to /register, call the Register function from the controllers package

	router.POST("/register", controllers.Register)
	router.POST("/login", controllers.Login)

	// Protected routes
	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.POST("/upload", controllers.UploadFile)
		protected.GET("/files", controllers.GetUserFiles)
		protected.GET("/files/search", controllers.SearchFiles)
		protected.GET("/share/:id", controllers.ShareFile)
		protected.PUT("/files/:id", controllers.UpdateFile)
	}
}
