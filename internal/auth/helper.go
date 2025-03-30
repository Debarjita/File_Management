package auth

import (
	"errors"
	"net/http"
	"strings"
)

// ValidateToken must be properly imported if it's in another package
// If it's in this package, ensure it's implemented
func ValidateToken(token string) (*JWTClaims, error) {
	// Dummy implementation; replace with actual JWT validation logic
	return nil, errors.New("ValidateToken function is not implemented")
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
