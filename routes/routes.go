package routes

import (
	"restaurant-booking-backend/controllers"
	"restaurant-booking-backend/middleware"

	"github.com/gin-gonic/gin"
)

var (
	authController = controllers.AuthController{}
)

// SetupRoutes sets up API routes
func SetupRoutes(router *gin.Engine) {
	// API group
	api := router.Group("/api/v1")
	{
		// Health check routes
		api.GET("/health", healthCheck)

		// Authentication routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/signup", authController.Signup)
			auth.POST("/login", authController.Login)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			// Example protected route
			protected.GET("/profile", getProfile)

			// Admin only routes
			admin := protected.Group("")
			admin.Use(middleware.RequireAdmin())
			{
				// Example admin route
				admin.GET("/admin/users", getUsers)
			}

			// Customer only routes
			customer := protected.Group("")
			customer.Use(middleware.RequireCustomer())
			{
				// Example customer route
				customer.GET("/customer/bookings", getCustomerBookings)
			}
		}
	}
}

// healthCheck checks system health
func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "ok",
		"message": "Server is running",
	})
}

// getProfile gets current user profile
func getProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userEmail, _ := c.Get("user_email")
	userRole, _ := c.Get("user_role")

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"user_id": userID,
			"email":   userEmail,
			"role":    userRole,
		},
	})
}

// getUsers gets all users (admin only)
func getUsers(c *gin.Context) {
	c.JSON(200, gin.H{
		"success": true,
		"message": "Admin access granted",
		"data":    "List of users will be here",
	})
}

// getCustomerBookings gets customer bookings (customer only)
func getCustomerBookings(c *gin.Context) {
	userID, _ := c.Get("user_id")
	c.JSON(200, gin.H{
		"success": true,
		"message": "Customer access granted",
		"data":    gin.H{"user_id": userID, "bookings": "List of bookings will be here"},
	})
}

