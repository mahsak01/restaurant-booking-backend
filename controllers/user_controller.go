package controllers

import (
	"strconv"

	"restaurant-booking-backend/config"
	"restaurant-booking-backend/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// UserController user controller
type UserController struct {
	BaseController
}

// GetAllUsers gets all users (admin only)
func (uc *UserController) GetAllUsers(c *fiber.Ctx) error {
	var users []models.User
	query := config.DB

	// Filter by role if provided
	role := c.Query("role")
	if role != "" {
		query = query.Where("role = ?", role)
	}

	// Search by name or phone if provided
	search := c.Query("search")
	if search != "" {
		query = query.Where("name ILIKE ? OR phone ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Select("id, phone, name, role, created_at, updated_at").Order("created_at DESC").Find(&users).Error; err != nil {
		return uc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch users")
	}

	return uc.SuccessResponse(c, users, "Users retrieved successfully")
}

// GetUserByID gets a single user by ID (admin only)
func (uc *UserController) GetUserByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return uc.ErrorResponse(c, fiber.StatusBadRequest, "Invalid user ID")
	}

	var user models.User
	if err := config.DB.Select("id, phone, name, role, created_at, updated_at").First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return uc.ErrorResponse(c, fiber.StatusNotFound, "User not found")
		}
		return uc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch user")
	}

	return uc.SuccessResponse(c, user, "User retrieved successfully")
}

// UpdateUserRole updates user role (admin only)
func (uc *UserController) UpdateUserRole(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return uc.ErrorResponse(c, fiber.StatusBadRequest, "Invalid user ID")
	}

	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return uc.ErrorResponse(c, fiber.StatusNotFound, "User not found")
		}
		return uc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch user")
	}

	var req struct {
		Role models.UserRole `json:"role"`
	}

	if err := c.BodyParser(&req); err != nil {
		return uc.ValidationErrorResponse(c, err.Error())
	}

	if req.Role == "" {
		return uc.ValidationErrorResponse(c, "Role is required")
	}

	// Validate role
	validRole := false
	for _, role := range []models.UserRole{
		models.RoleAdmin,
		models.RoleCustomer,
	} {
		if req.Role == role {
			validRole = true
			break
		}
	}

	if !validRole {
		return uc.ErrorResponse(c, fiber.StatusBadRequest, "Invalid role")
	}

	user.Role = req.Role
	if err := config.DB.Save(&user).Error; err != nil {
		return uc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update user role")
	}

	// Return user without password
	user.Password = ""
	return uc.SuccessResponse(c, user, "User role updated successfully")
}

// DeleteUser deletes a user (admin only)
func (uc *UserController) DeleteUser(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return uc.ErrorResponse(c, fiber.StatusBadRequest, "Invalid user ID")
	}

	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return uc.ErrorResponse(c, fiber.StatusNotFound, "User not found")
		}
		return uc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch user")
	}

	// Check if user has active reservations
	var activeReservations int64
	config.DB.Model(&models.Reservation{}).
		Where("user_id = ? AND status IN ?", id, []models.ReservationStatus{
			models.ReservationStatusPending,
			models.ReservationStatusConfirmed,
		}).Count(&activeReservations)

	if activeReservations > 0 {
		return uc.ErrorResponse(c, fiber.StatusBadRequest, "Cannot delete user with active reservations")
	}

	if err := config.DB.Delete(&user).Error; err != nil {
		return uc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete user")
	}

	return uc.SuccessResponse(c, nil, "User deleted successfully")
}
