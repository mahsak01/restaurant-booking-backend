package middleware

import (
	"restaurant-booking-backend/models"

	"github.com/gofiber/fiber/v2"
)

// RequireRole middleware to check user role
func RequireRole(requiredRole models.UserRole) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole := c.Locals("user_role")
		if userRole == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "User role not found",
			})
		}

		roleStr, ok := userRole.(string)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Invalid role type",
			})
		}

		if models.UserRole(roleStr) != requiredRole {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "Insufficient permissions",
			})
		}

		return c.Next()
	}
}

// RequireAdmin middleware to check if user is admin
func RequireAdmin() fiber.Handler {
	return RequireRole(models.RoleAdmin)
}

// RequireCustomer middleware to check if user is customer
func RequireCustomer() fiber.Handler {
	return RequireRole(models.RoleCustomer)
}
