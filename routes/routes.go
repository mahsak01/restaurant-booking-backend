package routes

import (
	"restaurant-booking-backend/controllers"
	"restaurant-booking-backend/middleware"

	"github.com/gin-gonic/gin"
)

var (
	authController  = controllers.AuthController{}
	menuController  = controllers.MenuController{}
	tableController = controllers.TableController{}
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

		// Menu routes (public - for customers)
		menu := api.Group("/menu")
		{
			menu.GET("", menuController.GetAllMenuItems)
			menu.GET("/categories", menuController.GetCategories)
			menu.GET("/category/:category", menuController.GetMenuItemsByCategory)
			menu.GET("/:id", menuController.GetMenuItemByID)
		}

		// Table routes (public - for customers to view available tables)
		tables := api.Group("/tables")
		{
			tables.GET("/available", tableController.GetAvailableTables)
			tables.GET("/statuses", tableController.GetTableStatuses)
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

				// Menu management routes (admin only)
				adminMenu := admin.Group("/admin/menu")
				{
					adminMenu.POST("", menuController.CreateMenuItem)
					adminMenu.PUT("/:id", menuController.UpdateMenuItem)
					adminMenu.DELETE("/:id", menuController.DeleteMenuItem)
				}

				// Table management routes (admin only)
				adminTables := admin.Group("/admin/tables")
				{
					adminTables.GET("", tableController.GetAllTables)
					adminTables.GET("/:id", tableController.GetTableByID)
					adminTables.POST("", tableController.CreateTable)
					adminTables.PUT("/:id", tableController.UpdateTable)
					adminTables.DELETE("/:id", tableController.DeleteTable)
				}
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

