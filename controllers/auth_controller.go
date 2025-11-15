package controllers

import (
	"strings"

	"restaurant-booking-backend/config"
	"restaurant-booking-backend/models"
	"restaurant-booking-backend/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// AuthController authentication controller
type AuthController struct {
	BaseController
}

// SignupRequest signup request structure
type SignupRequest struct {
	Phone    string `json:"phone" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
	Name     string `json:"name" validate:"required"`
}

// LoginRequest login request structure
type LoginRequest struct {
	Phone    string `json:"phone" validate:"required"`
	Password string `json:"password" validate:"required"`
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
func (ac *AuthController) Signup(c *fiber.Ctx) error {
	var req SignupRequest
	if err := c.BodyParser(&req); err != nil {
		return ac.ValidationErrorResponse(c, err.Error())
	}

	// Validate required fields
	if req.Phone == "" || req.Password == "" || req.Name == "" {
		return ac.ValidationErrorResponse(c, "Phone, password, and name are required")
	}

	if len(req.Password) < 6 {
		return ac.ValidationErrorResponse(c, "Password must be at least 6 characters")
	}

	// Validate phone number
	req.Phone = strings.TrimSpace(req.Phone)
	if !utils.ValidatePhoneNumber(req.Phone) {
		return ac.ErrorResponse(c, fiber.StatusBadRequest, "Invalid phone number format")
	}

	// Check if user already exists
	var existingUser models.User
	if err := config.DB.Where("phone = ?", req.Phone).First(&existingUser).Error; err == nil {
		return ac.ErrorResponse(c, fiber.StatusConflict, "User with this phone number already exists")
	} else if err != gorm.ErrRecordNotFound {
		return ac.ErrorResponse(c, fiber.StatusInternalServerError, "Database error")
	}

	// Create new user
	user := models.User{
		Phone:    req.Phone,
		Password: req.Password, // Will be hashed in BeforeCreate hook
		Name:     req.Name,
		Role:     models.RoleCustomer, // Default role
	}

	if err := config.DB.Create(&user).Error; err != nil {
		return ac.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create user")
	}

	// Generate JWT token
	token, err := utils.GenerateToken(user.ID, user.Phone, string(user.Role))
	if err != nil {
		return ac.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to generate token")
	}

	// Remove password from response
	user.Password = ""

	return ac.SuccessResponse(c, SignupResponse{
		User:  user,
		Token: token,
	}, "User registered successfully")
}

// Login handles user login
func (ac *AuthController) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return ac.ValidationErrorResponse(c, err.Error())
	}

	// Validate required fields
	if req.Phone == "" || req.Password == "" {
		return ac.ValidationErrorResponse(c, "Phone and password are required")
	}

	// Validate phone number
	req.Phone = strings.TrimSpace(req.Phone)
	if !utils.ValidatePhoneNumber(req.Phone) {
		return ac.ErrorResponse(c, fiber.StatusBadRequest, "Invalid phone number format")
	}

	// Find user by phone
	var user models.User
	if err := config.DB.Where("phone = ?", req.Phone).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ac.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid phone number or password")
		}
		return ac.ErrorResponse(c, fiber.StatusInternalServerError, "Database error")
	}

	// Check password
	if !user.CheckPassword(req.Password) {
		return ac.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid phone number or password")
	}

	// Generate JWT token
	token, err := utils.GenerateToken(user.ID, user.Phone, string(user.Role))
	if err != nil {
		return ac.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to generate token")
	}

	// Remove password from response
	user.Password = ""

	return ac.SuccessResponse(c, LoginResponse{
		User:  user,
		Token: token,
	}, "Login successful")
}
