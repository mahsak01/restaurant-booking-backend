package controllers

import (
	"net/http"

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
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
}

// LoginRequest login request structure
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
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

	// Check if user already exists
	var existingUser models.User
	if err := config.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		ac.ErrorResponse(c, http.StatusConflict, "User with this email already exists")
		return
	} else if err != gorm.ErrRecordNotFound {
		ac.ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	// Create new user
	user := models.User{
		Email:    req.Email,
		Password: req.Password, // Will be hashed in BeforeCreate hook
		Name:     req.Name,
		Role:     models.RoleCustomer, // Default role
	}

	if err := config.DB.Create(&user).Error; err != nil {
		ac.ErrorResponse(c, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Generate JWT token
	token, err := utils.GenerateToken(user.ID, user.Email, string(user.Role))
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

	// Find user by email
	var user models.User
	if err := config.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ac.ErrorResponse(c, http.StatusUnauthorized, "Invalid email or password")
			return
		}
		ac.ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	// Check password
	if !user.CheckPassword(req.Password) {
		ac.ErrorResponse(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Generate JWT token
	token, err := utils.GenerateToken(user.ID, user.Email, string(user.Role))
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

