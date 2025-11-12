package routes

import (
	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up API routes
func SetupRoutes(router *gin.Engine) {
	// API group
	api := router.Group("/api/v1")
	{
		// Health check routes
		api.GET("/health", healthCheck)
		
		// Other routes will be added here
		// Example: api.POST("/restaurants", ...)
	}
}

// healthCheck checks system health
func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "ok",
		"message": "Server is running",
	})
}

