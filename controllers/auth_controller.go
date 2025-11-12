package controllers

import (
	"net/http"
	"strings"

	"restaurant-booking-backend/config"
	"restaurant-booking-backend/models"
	"restaurant-booking-backend/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthController authentication controller
type AuthController struct {
	BaseController
}

// SignupRequest signup request structure
type SignupRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
}

// LoginRequest login request structure
type LoginRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// SignupResponse signup response structure
type SignupResponse struct {
	User  models.User `json:"user"`
	Token string      `json:"token"`
}

// LoginResponse login response structure
type LoginResponse struct {
	User  models.User `json:"user"`
	Token string      `json:"token"`
}

// Signup handles user registration
func (ac *AuthController) Signup(c *gin.Context) {
	var req SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ac.ValidationErrorResponse(c, err.Error())
		return
	}

	// Validate phone number
	req.Phone = strings.TrimSpace(req.Phone)
	if !utils.ValidatePhoneNumber(req.Phone) {
		ac.ErrorResponse(c, http.StatusBadRequest, "Invalid phone number format")
		return
	}

	// Check if user already exists
	var existingUser models.User
	if err := config.DB.Where("phone = ?", req.Phone).First(&existingUser).Error; err == nil {
		ac.ErrorResponse(c, http.StatusConflict, "User with this phone number already exists")
		return
	} else if err != gorm.ErrRecordNotFound {
		ac.ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	// Create new user
	user := models.User{
		Phone:    req.Phone,
		Password: req.Password, // Will be hashed in BeforeCreate hook
		Name:     req.Name,
		Role:     models.RoleCustomer, // Default role
	}

	if err := config.DB.Create(&user).Error; err != nil {
		ac.ErrorResponse(c, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Generate JWT token
	token, err := utils.GenerateToken(user.ID, user.Phone, string(user.Role))
	if err != nil {
		ac.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// Remove password from response
	user.Password = ""

	ac.SuccessResponse(c, SignupResponse{
		User:  user,
		Token: token,
	}, "User registered successfully")
}

// Login handles user login
func (ac *AuthController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ac.ValidationErrorResponse(c, err.Error())
		return
	}

	// Validate phone number
	req.Phone = strings.TrimSpace(req.Phone)
	if !utils.ValidatePhoneNumber(req.Phone) {
		ac.ErrorResponse(c, http.StatusBadRequest, "Invalid phone number format")
		return
	}

	// Find user by phone
	var user models.User
	if err := config.DB.Where("phone = ?", req.Phone).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ac.ErrorResponse(c, http.StatusUnauthorized, "Invalid phone number or password")
			return
		}
		ac.ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	// Check password
	if !user.CheckPassword(req.Password) {
		ac.ErrorResponse(c, http.StatusUnauthorized, "Invalid phone number or password")
		return
	}

	// Generate JWT token
	token, err := utils.GenerateToken(user.ID, user.Phone, string(user.Role))
	if err != nil {
		ac.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// Remove password from response
	user.Password = ""

	ac.SuccessResponse(c, LoginResponse{
		User:  user,
		Token: token,
	}, "Login successful")
}

