package middleware

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"file-sharing-platform/internal/auth"
)

// RequestLogger logs details of each request
func RequestLogger(c *gin.Context) {
	start := time.Now()
	c.Next()
	log.Printf("[%d] %s %s %v", c.Writer.Status(), c.Request.Method, c.Request.URL.Path, time.Since(start))
}
func AuthMiddleware(jwtAuth *auth.JWTAuth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			_, err := jwtAuth.ValidateToken(tokenString)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
