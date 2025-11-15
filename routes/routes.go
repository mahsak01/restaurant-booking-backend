package routes

import (
	"restaurant-booking-backend/config"
	"restaurant-booking-backend/controllers"
	"restaurant-booking-backend/middleware"
	"restaurant-booking-backend/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

var (
	authController         = controllers.AuthController{}
	menuController         = controllers.MenuController{}
	tableController        = controllers.TableController{}
	reservationController  = controllers.NewReservationController()
	notificationController = controllers.NotificationController{}
	userController         = controllers.UserController{}
	categoryController     = controllers.CategoryController{}
	orderController        = controllers.OrderController{}
)

// SetupRoutes sets up API routes
func SetupRoutes(app *fiber.App) {
	// API group
	api := app.Group("/api/v1")

	// Health check routes
	api.Get("/health", healthCheck)

	// Authentication routes (public)
	auth := api.Group("/auth")
	{
		auth.Post("/signup", authController.Signup)
		auth.Post("/login", authController.Login)
	}

	// Menu routes (public - for customers)
	menu := api.Group("/menu")
	{
		menu.Get("", menuController.GetAllMenuItems)
		menu.Get("/categories", menuController.GetCategories)
		menu.Get("/category/:category", menuController.GetMenuItemsByCategory)
		menu.Get("/:id", menuController.GetMenuItemByID)
	}

	// Table routes (public - for customers to view available tables)
	tables := api.Group("/tables")
	{
		tables.Get("/available", tableController.GetAvailableTables)
		tables.Get("/statuses", tableController.GetTableStatuses)
	}

	// Category routes (public - for customers to view categories)
	categories := api.Group("/categories")
	{
		categories.Get("", categoryController.GetAllCategories)
		categories.Get("/:id", categoryController.GetCategoryByID)
	}

	// Protected routes
	protected := api.Group("", middleware.AuthMiddleware())
	{
		// Example protected route
		protected.Get("/profile", getProfile)

		// Admin only routes
		admin := protected.Group("", middleware.RequireAdmin())
		{
			// User management routes (admin only)
			adminUsers := admin.Group("/admin/users")
			{
				adminUsers.Get("", userController.GetAllUsers)
				adminUsers.Get("/:id", userController.GetUserByID)
				adminUsers.Put("/:id/role", userController.UpdateUserRole)
				adminUsers.Delete("/:id", userController.DeleteUser)
			}

			// Menu management routes (admin only)
			adminMenu := admin.Group("/admin/menu")
			{
				adminMenu.Post("", menuController.CreateMenuItem)
				adminMenu.Put("/:id", menuController.UpdateMenuItem)
				adminMenu.Delete("/:id", menuController.DeleteMenuItem)
			}

			// Category management routes (admin only)
			adminCategories := admin.Group("/admin/categories")
			{
				adminCategories.Post("", categoryController.CreateCategory)
				adminCategories.Put("/:id", categoryController.UpdateCategory)
				adminCategories.Delete("/:id", categoryController.DeleteCategory)
			}

			// Table management routes (admin only)
			adminTables := admin.Group("/admin/tables")
			{
				adminTables.Get("", tableController.GetAllTables)
				adminTables.Get("/:id", tableController.GetTableByID)
				adminTables.Post("", tableController.CreateTable)
				adminTables.Put("/:id", tableController.UpdateTable)
				adminTables.Delete("/:id", tableController.DeleteTable)
			}

			// Reservation management routes (admin only)
			adminReservations := admin.Group("/admin/reservations")
			{
				adminReservations.Post("", reservationController.CreateReservationByAdmin)
				adminReservations.Get("", reservationController.GetAllReservations)
				adminReservations.Get("/statuses", reservationController.GetReservationStatuses)
				adminReservations.Get("/:id", reservationController.GetReservationByID)
				adminReservations.Put("/:id/status", reservationController.UpdateReservationStatus)
				adminReservations.Delete("/:id", reservationController.CancelReservation)
			}

			// Order management routes (admin only)
			adminOrders := admin.Group("/admin/orders")
			{
				adminOrders.Post("", orderController.CreateOrderByAdmin)
				adminOrders.Get("", orderController.GetAllOrders)
				adminOrders.Get("/statuses", orderController.GetOrderStatuses)
				adminOrders.Get("/:id", orderController.GetOrderByID)
				adminOrders.Put("/:id/status", orderController.UpdateOrderStatus)
			}
		}

		// Customer only routes
		customer := protected.Group("", middleware.RequireCustomer())
		{
			// Reservation routes (customer)
			customerReservations := customer.Group("/reservations")
			{
				customerReservations.Post("", reservationController.CreateReservation)
				customerReservations.Get("", reservationController.GetUserReservations)
				customerReservations.Get("/:id", reservationController.GetReservationByID)
				customerReservations.Delete("/:id", reservationController.CancelReservation)
			}

			// Order routes (customer)
			customerOrders := customer.Group("/orders")
			{
				customerOrders.Post("", orderController.CreateOrder)
				customerOrders.Get("", orderController.GetUserOrders)
				customerOrders.Get("/:id", orderController.GetOrderByID)
			}
		}

		// Notification routes (for all authenticated users)
		notifications := protected.Group("/notifications")
		{
			notifications.Get("", notificationController.GetUserNotifications)
			notifications.Get("/count", notificationController.GetUnreadNotificationsCount)
			notifications.Put("/:id/read", notificationController.MarkNotificationAsRead)
			notifications.Delete("/:id", notificationController.DeleteNotification)
		}
	}
}

// healthCheck checks system health
func healthCheck(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "ok",
		"message": "Server is running",
	})
}

// getProfile gets current user profile with all information
func getProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "User not authenticated",
		})
	}

	var user models.User
	// Get user with all information from database
	if err := config.DB.
		Preload("Reservations").
		Preload("Notifications").
		First(&user, userID.(uint)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to fetch user profile",
		})
	}

	// Remove password from response
	user.Password = ""

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Profile retrieved successfully",
		"data":    user,
	})
}
