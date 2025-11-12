package routes

import (
	"restaurant-booking-backend/controllers"
	"restaurant-booking-backend/middleware"

	"github.com/gin-gonic/gin"
)

var (
	authController        = controllers.AuthController{}
	menuController        = controllers.MenuController{}
	tableController       = controllers.TableController{}
	reservationController = controllers.NewReservationController()
	notificationController = controllers.NotificationController{}
	userController         = controllers.UserController{}
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
				// User management routes (admin only)
				adminUsers := admin.Group("/admin/users")
				{
					adminUsers.GET("", userController.GetAllUsers)
					adminUsers.GET("/:id", userController.GetUserByID)
					adminUsers.PUT("/:id/role", userController.UpdateUserRole)
					adminUsers.DELETE("/:id", userController.DeleteUser)
				}

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

				// Reservation management routes (admin only)
				adminReservations := admin.Group("/admin/reservations")
				{
					adminReservations.GET("", reservationController.GetAllReservations)
					adminReservations.GET("/statuses", reservationController.GetReservationStatuses)
					adminReservations.GET("/:id", reservationController.GetReservationByID)
					adminReservations.PUT("/:id/status", reservationController.UpdateReservationStatus)
					adminReservations.DELETE("/:id", reservationController.CancelReservation)
				}
			}

			// Customer only routes
			customer := protected.Group("")
			customer.Use(middleware.RequireCustomer())
			{
				// Reservation routes (customer)
				customerReservations := customer.Group("/reservations")
				{
					customerReservations.POST("", reservationController.CreateReservation)
					customerReservations.GET("", reservationController.GetUserReservations)
					customerReservations.GET("/:id", reservationController.GetReservationByID)
					customerReservations.DELETE("/:id", reservationController.CancelReservation)
				}

			}

			// Notification routes (for all authenticated users)
			notifications := protected.Group("/notifications")
			{
				notifications.GET("", notificationController.GetUserNotifications)
				notifications.GET("/count", notificationController.GetUnreadNotificationsCount)
				notifications.PUT("/:id/read", notificationController.MarkNotificationAsRead)
				notifications.DELETE("/:id", notificationController.DeleteNotification)
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

// getCustomerBookings gets customer bookings (customer only)
func getCustomerBookings(c *gin.Context) {
	userID, _ := c.Get("user_id")
	c.JSON(200, gin.H{
		"success": true,
		"message": "Customer access granted",
		"data":    gin.H{"user_id": userID, "bookings": "List of bookings will be here"},
	})
}

