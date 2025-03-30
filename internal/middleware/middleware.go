package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger logs details of each request
func RequestLogger(c *gin.Context) {
	start := time.Now()
	c.Next()
	log.Printf("[%d] %s %s %v", c.Writer.Status(), c.Request.Method, c.Request.URL.Path, time.Since(start))
}
