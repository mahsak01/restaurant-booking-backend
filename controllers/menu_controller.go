package controllers

import (
	"strconv"

	"restaurant-booking-backend/config"
	"restaurant-booking-backend/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// MenuController menu controller
type MenuController struct {
	BaseController
}

// CreateMenuItemRequest create menu item request structure
type CreateMenuItemRequest struct {
	Name        string              `json:"name" binding:"required"`
	Description string              `json:"description"`
	Price       float64             `json:"price" binding:"required,gt=0"`
	ImageURL    string              `json:"image_url"`
	Category    models.MenuCategory `json:"category" binding:"required"`
}

// UpdateMenuItemRequest update menu item request structure
type UpdateMenuItemRequest struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Price       float64             `json:"price" binding:"omitempty,gt=0"`
	ImageURL    string              `json:"image_url"`
	Category    models.MenuCategory `json:"category"`
}

// GetAllMenuItems gets all menu items (public)
func (mc *MenuController) GetAllMenuItems(c *fiber.Ctx) error {
	var menuItems []models.MenuItem
	query := config.DB

	// Filter by category if provided
	category := c.Query("category")
	if category != "" {
		query = query.Where("category = ?", category)
	}

	// Search by name if provided
	search := c.Query("search")
	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}

	if err := query.Order("created_at DESC").Find(&menuItems).Error; err != nil {
		return mc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch menu items")
	}

	return mc.SuccessResponse(c, menuItems, "Menu items retrieved successfully")
}

// GetMenuItemByID gets a single menu item by ID (public)
func (mc *MenuController) GetMenuItemByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return mc.ErrorResponse(c, fiber.StatusBadRequest, "Invalid menu item ID")
	}

	var menuItem models.MenuItem
	if err := config.DB.First(&menuItem, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return mc.ErrorResponse(c, fiber.StatusNotFound, "Menu item not found")
		}
		return mc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch menu item")
	}

	return mc.SuccessResponse(c, menuItem, "Menu item retrieved successfully")
}

// GetMenuItemsByCategory gets menu items by category (public)
func (mc *MenuController) GetMenuItemsByCategory(c *fiber.Ctx) error {
	category := c.Params("category")

	// Validate category
	validCategory := false
	for _, cat := range []models.MenuCategory{
		models.CategoryAppetizer,
		models.CategoryMain,
		models.CategoryDessert,
		models.CategoryDrink,
	} {
		if string(cat) == category {
			validCategory = true
			break
		}
	}

	if !validCategory {
		return mc.ErrorResponse(c, fiber.StatusBadRequest, "Invalid category")
	}

	var menuItems []models.MenuItem
	if err := config.DB.Where("category = ?", category).Order("created_at DESC").Find(&menuItems).Error; err != nil {
		return mc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch menu items")
	}

	return mc.SuccessResponse(c, menuItems, "Menu items retrieved successfully")
}

// CreateMenuItem creates a new menu item (admin only)
func (mc *MenuController) CreateMenuItem(c *fiber.Ctx) error {
	var req CreateMenuItemRequest
	if err := c.BodyParser(&req); err != nil {
		return mc.ValidationErrorResponse(c, err.Error())
	}

	if req.Name == "" || req.Price <= 0 || req.Category == "" {
		return mc.ValidationErrorResponse(c, "Name, price, and category are required")
	}

	menuItem := models.MenuItem{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		ImageURL:    req.ImageURL,
		Category:    req.Category,
	}

	if err := config.DB.Create(&menuItem).Error; err != nil {
		return mc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create menu item")
	}

	return mc.SuccessResponse(c, menuItem, "Menu item created successfully")
}

// UpdateMenuItem updates an existing menu item (admin only)
func (mc *MenuController) UpdateMenuItem(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return mc.ErrorResponse(c, fiber.StatusBadRequest, "Invalid menu item ID")
	}

	var menuItem models.MenuItem
	if err := config.DB.First(&menuItem, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return mc.ErrorResponse(c, fiber.StatusNotFound, "Menu item not found")
		}
		return mc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch menu item")
	}

	var req UpdateMenuItemRequest
	if err := c.BodyParser(&req); err != nil {
		return mc.ValidationErrorResponse(c, err.Error())
	}

	// Update fields if provided
	if req.Name != "" {
		menuItem.Name = req.Name
	}
	if req.Description != "" {
		menuItem.Description = req.Description
	}
	if req.Price > 0 {
		menuItem.Price = req.Price
	}
	if req.ImageURL != "" {
		menuItem.ImageURL = req.ImageURL
	}
	if req.Category != "" {
		menuItem.Category = req.Category
	}

	if err := config.DB.Save(&menuItem).Error; err != nil {
		return mc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update menu item")
	}

	return mc.SuccessResponse(c, menuItem, "Menu item updated successfully")
}

// DeleteMenuItem deletes a menu item (admin only)
func (mc *MenuController) DeleteMenuItem(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return mc.ErrorResponse(c, fiber.StatusBadRequest, "Invalid menu item ID")
	}

	var menuItem models.MenuItem
	if err := config.DB.First(&menuItem, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return mc.ErrorResponse(c, fiber.StatusNotFound, "Menu item not found")
		}
		return mc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch menu item")
	}

	if err := config.DB.Delete(&menuItem).Error; err != nil {
		return mc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete menu item")
	}

	return mc.SuccessResponse(c, nil, "Menu item deleted successfully")
}

// GetCategories gets all available menu categories (public)
func (mc *MenuController) GetCategories(c *fiber.Ctx) error {
	categories := []map[string]string{
		{"value": string(models.CategoryAppetizer), "label": "Appetizer"},
		{"value": string(models.CategoryMain), "label": "Main Course"},
		{"value": string(models.CategoryDessert), "label": "Dessert"},
		{"value": string(models.CategoryDrink), "label": "Drink"},
	}

	return mc.SuccessResponse(c, categories, "Categories retrieved successfully")
}
