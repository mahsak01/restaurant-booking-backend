package controllers

import (
	"strconv"
	"strings"

	"restaurant-booking-backend/config"
	"restaurant-booking-backend/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// CategoryController category controller
type CategoryController struct {
	BaseController
}

// CreateCategoryRequest create category request structure
type CreateCategoryRequest struct {
	Name        string `json:"name"`         // Category name (e.g., "appetizer") - required
	DisplayName string `json:"display_name"` // Display name (e.g., "Appetizer") - required
	Description string `json:"description"`  // Optional description
	IsActive    *bool  `json:"is_active"`    // Optional, defaults to true
	SortOrder   int    `json:"sort_order"`   // Sort order (optional)
}

// UpdateCategoryRequest update category request structure
type UpdateCategoryRequest struct {
	Name        string  `json:"name"`
	DisplayName string  `json:"display_name"`
	Description *string `json:"description"` // Pointer to allow clearing description
	IsActive    *bool   `json:"is_active"`
	SortOrder   *int    `json:"sort_order"`
}

// GetAllCategories gets all categories (public)
func (cc *CategoryController) GetAllCategories(c *fiber.Ctx) error {
	var categories []models.Category
	query := config.DB

	// Filter by active status if provided
	active := c.Query("active")
	if active == "" || active == "true" {
		// Default: show only active categories for public access
		query = query.Where("is_active = ?", true)
	} else if active == "false" {
		query = query.Where("is_active = ?", false)
	}
	// If active is "all", show all categories (no filter applied)

	if err := query.Order("sort_order ASC, display_name ASC").Find(&categories).Error; err != nil {
		return cc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch categories")
	}

	return cc.SuccessResponse(c, categories, "Categories retrieved successfully")
}

// GetCategoryByID gets a single category by ID (public)
func (cc *CategoryController) GetCategoryByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return cc.ErrorResponse(c, fiber.StatusBadRequest, "Invalid category ID")
	}

	var category models.Category
	if err := config.DB.First(&category, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return cc.ErrorResponse(c, fiber.StatusNotFound, "Category not found")
		}
		return cc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch category")
	}

	return cc.SuccessResponse(c, category, "Category retrieved successfully")
}

// CreateCategory creates a new category (admin only)
func (cc *CategoryController) CreateCategory(c *fiber.Ctx) error {
	var req CreateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return cc.ValidationErrorResponse(c, err.Error())
	}

	// Validate required fields
	req.Name = strings.TrimSpace(strings.ToLower(req.Name))
	req.DisplayName = strings.TrimSpace(req.DisplayName)

	if req.Name == "" || req.DisplayName == "" {
		return cc.ValidationErrorResponse(c, "Name and display_name are required")
	}

	// Check if category with same name already exists
	var existingCategory models.Category
	if err := config.DB.Where("name = ?", req.Name).First(&existingCategory).Error; err == nil {
		return cc.ErrorResponse(c, fiber.StatusConflict, "Category with this name already exists")
	} else if err != gorm.ErrRecordNotFound {
		return cc.ErrorResponse(c, fiber.StatusInternalServerError, "Database error")
	}

	// Set default active status to true if not provided
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	// Create category
	category := models.Category{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: strings.TrimSpace(req.Description),
		IsActive:    isActive,
		SortOrder:   req.SortOrder,
	}

	if err := config.DB.Create(&category).Error; err != nil {
		return cc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create category")
	}

	return cc.SuccessResponse(c, category, "Category created successfully")
}

// UpdateCategory updates an existing category (admin only)
func (cc *CategoryController) UpdateCategory(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return cc.ErrorResponse(c, fiber.StatusBadRequest, "Invalid category ID")
	}

	var category models.Category
	if err := config.DB.First(&category, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return cc.ErrorResponse(c, fiber.StatusNotFound, "Category not found")
		}
		return cc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch category")
	}

	var req UpdateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return cc.ValidationErrorResponse(c, err.Error())
	}

	// Update fields if provided
	if req.Name != "" {
		req.Name = strings.TrimSpace(strings.ToLower(req.Name))
		// Check if new name conflicts with existing category
		var existingCategory models.Category
		if err := config.DB.Where("name = ? AND id != ?", req.Name, id).First(&existingCategory).Error; err == nil {
			return cc.ErrorResponse(c, fiber.StatusConflict, "Category with this name already exists")
		} else if err != gorm.ErrRecordNotFound {
			return cc.ErrorResponse(c, fiber.StatusInternalServerError, "Database error")
		}
		category.Name = req.Name
	}

	if req.DisplayName != "" {
		category.DisplayName = strings.TrimSpace(req.DisplayName)
	}

	// Update description if provided (pointer allows us to distinguish between not provided and empty)
	if req.Description != nil {
		category.Description = strings.TrimSpace(*req.Description)
	}

	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}

	if req.SortOrder != nil {
		category.SortOrder = *req.SortOrder
	}

	if err := config.DB.Save(&category).Error; err != nil {
		return cc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update category")
	}

	return cc.SuccessResponse(c, category, "Category updated successfully")
}

// DeleteCategory deletes a category (admin only)
func (cc *CategoryController) DeleteCategory(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return cc.ErrorResponse(c, fiber.StatusBadRequest, "Invalid category ID")
	}

	var category models.Category
	if err := config.DB.First(&category, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return cc.ErrorResponse(c, fiber.StatusNotFound, "Category not found")
		}
		return cc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch category")
	}

	// Check if category is used by any menu items
	var menuItemsCount int64
	config.DB.Model(&models.MenuItem{}).Where("category = ?", category.Name).Count(&menuItemsCount)
	if menuItemsCount > 0 {
		return cc.ErrorResponse(c, fiber.StatusBadRequest, "Cannot delete category that is used by menu items")
	}

	if err := config.DB.Delete(&category).Error; err != nil {
		return cc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete category")
	}

	return cc.SuccessResponse(c, nil, "Category deleted successfully")
}
