package controllers

import (
	"net/http"
	"strconv"

	"restaurant-booking-backend/config"
	"restaurant-booking-backend/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UserController user controller
type UserController struct {
	BaseController
}

// GetAllUsers gets all users (admin only)
func (uc *UserController) GetAllUsers(c *gin.Context) {
	var users []models.User
	query := config.DB

	// Filter by role if provided
	role := c.Query("role")
	if role != "" {
		query = query.Where("role = ?", role)
	}

	// Search by name or email if provided
	search := c.Query("search")
	if search != "" {
		query = query.Where("name ILIKE ? OR email ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Select("id, email, name, role, created_at, updated_at").Order("created_at DESC").Find(&users).Error; err != nil {
		uc.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch users")
		return
	}

	uc.SuccessResponse(c, users, "Users retrieved successfully")
}

// GetUserByID gets a single user by ID (admin only)
func (uc *UserController) GetUserByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		uc.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var user models.User
	if err := config.DB.Select("id, email, name, role, created_at, updated_at").First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			uc.ErrorResponse(c, http.StatusNotFound, "User not found")
			return
		}
		uc.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch user")
		return
	}

	uc.SuccessResponse(c, user, "User retrieved successfully")
}

// UpdateUserRole updates user role (admin only)
func (uc *UserController) UpdateUserRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		uc.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			uc.ErrorResponse(c, http.StatusNotFound, "User not found")
			return
		}
		uc.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch user")
		return
	}

	var req struct {
		Role models.UserRole `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		uc.ValidationErrorResponse(c, err.Error())
		return
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
		uc.ErrorResponse(c, http.StatusBadRequest, "Invalid role")
		return
	}

	user.Role = req.Role
	if err := config.DB.Save(&user).Error; err != nil {
		uc.ErrorResponse(c, http.StatusInternalServerError, "Failed to update user role")
		return
	}

	// Return user without password
	user.Password = ""
	uc.SuccessResponse(c, user, "User role updated successfully")
}

// DeleteUser deletes a user (admin only)
func (uc *UserController) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		uc.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			uc.ErrorResponse(c, http.StatusNotFound, "User not found")
			return
		}
		uc.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch user")
		return
	}

	// Check if user has active reservations
	var activeReservations int64
	config.DB.Model(&models.Reservation{}).
		Where("user_id = ? AND status IN ?", id, []models.ReservationStatus{
			models.ReservationStatusPending,
			models.ReservationStatusConfirmed,
		}).Count(&activeReservations)

	if activeReservations > 0 {
		uc.ErrorResponse(c, http.StatusBadRequest, "Cannot delete user with active reservations")
		return
	}

	if err := config.DB.Delete(&user).Error; err != nil {
		uc.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	uc.SuccessResponse(c, nil, "User deleted successfully")
}

