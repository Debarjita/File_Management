package controllers

import (
	"learningfilesharing/config"
	"learningfilesharing/models"
	"learningfilesharing/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

//Register the user

func Register(c *gin.Context) {
	var user models.User               //creating an empty user from models
	if err := c.ShouldBindJSON(&user); // get JSON inputs from user like email and password
	err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return

	}

	//hash the password
	hashedPassword, err := utils.HashPassword(user.Password) // calling hashed password function from util
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Hashing failed"})
		return

	}
	user.Password = hashedPassword // replace the original with hashed password

	// Try to insert the user
	if err := config.DB.Create(&user).Error; err != nil {
		// Check if it's a duplicate key error (basic check for PostgreSQL, SQLite, etc.)
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "UNIQUE") {
			c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user registered successfully"})
}

// login the user

func Login(c *gin.Context) {
	var input models.User // empty input variable for user login vaues
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	var user models.User
	if err := config.DB.Where("email=?", input.Email).First(&user).Error; // search for the email in db
	err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email not found"})
		return
	}

	// match the hashed password with the normal password entered by user
	if !utils.CheckPasswordHash(input.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect password"})
		return
	}

	//generate jwt token for future use
	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
