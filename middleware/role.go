package middleware

import (
	"net/http"

	"restaurant-booking-backend/models"

	"github.com/gin-gonic/gin"
)

// RequireRole middleware to check user role
func RequireRole(requiredRole models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "User role not found",
			})
			c.Abort()
			return
		}

		roleStr, ok := userRole.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Invalid role type",
			})
			c.Abort()
			return
		}

		if models.UserRole(roleStr) != requiredRole {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin middleware to check if user is admin
func RequireAdmin() gin.HandlerFunc {
	return RequireRole(models.RoleAdmin)
}

// RequireCustomer middleware to check if user is customer
func RequireCustomer() gin.HandlerFunc {
	return RequireRole(models.RoleCustomer)
}

