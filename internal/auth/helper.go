package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

var jwtSecret = []byte("your-secret-key") // Ensure this is loaded from env/config

// GetUserIDFromContext retrieves user ID from gin.Context
func GetUserIDFromContext(c *gin.Context) (int64, error) {
	userID, exists := c.Get("userID")
	if !exists {
		return 0, errors.New("user ID not found in context")
	}

	uid, ok := userID.(int64)
	if !ok {
		return 0, errors.New("user ID is not a valid int64")
	}

	return uid, nil
}

func ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GetTokenFromRequest extracts JWT token from Authorization header
func GetTokenFromRequest(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header is missing")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid authorization header format")
	}

	return parts[1], nil
}

// GetUserIDFromRequest extracts user ID from JWT token in request
func GetUserIDFromRequest(r *http.Request) (int64, error) {
	token, err := GetTokenFromRequest(r)
	if err != nil {
		return 0, err
	}

	claims, err := ValidateToken(token)
	if err != nil {
		return 0, err
	}

	return claims.UserID, nil
}
