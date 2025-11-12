package controllers

import (
	"net/http"
	"strconv"

	"restaurant-booking-backend/config"
	"restaurant-booking-backend/models"

	"github.com/gin-gonic/gin"
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
func (mc *MenuController) GetAllMenuItems(c *gin.Context) {
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
		mc.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch menu items")
		return
	}

	mc.SuccessResponse(c, menuItems, "Menu items retrieved successfully")
}

// GetMenuItemByID gets a single menu item by ID (public)
func (mc *MenuController) GetMenuItemByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		mc.ErrorResponse(c, http.StatusBadRequest, "Invalid menu item ID")
		return
	}

	var menuItem models.MenuItem
	if err := config.DB.First(&menuItem, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			mc.ErrorResponse(c, http.StatusNotFound, "Menu item not found")
			return
		}
		mc.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch menu item")
		return
	}

	mc.SuccessResponse(c, menuItem, "Menu item retrieved successfully")
}

// GetMenuItemsByCategory gets menu items by category (public)
func (mc *MenuController) GetMenuItemsByCategory(c *gin.Context) {
	category := c.Param("category")

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
		mc.ErrorResponse(c, http.StatusBadRequest, "Invalid category")
		return
	}

	var menuItems []models.MenuItem
	if err := config.DB.Where("category = ?", category).Order("created_at DESC").Find(&menuItems).Error; err != nil {
		mc.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch menu items")
		return
	}

	mc.SuccessResponse(c, menuItems, "Menu items retrieved successfully")
}

// CreateMenuItem creates a new menu item (admin only)
func (mc *MenuController) CreateMenuItem(c *gin.Context) {
	var req CreateMenuItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		mc.ValidationErrorResponse(c, err.Error())
		return
	}

	menuItem := models.MenuItem{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		ImageURL:    req.ImageURL,
		Category:    req.Category,
	}

	if err := config.DB.Create(&menuItem).Error; err != nil {
		mc.ErrorResponse(c, http.StatusInternalServerError, "Failed to create menu item")
		return
	}

	mc.SuccessResponse(c, menuItem, "Menu item created successfully")
}

// UpdateMenuItem updates an existing menu item (admin only)
func (mc *MenuController) UpdateMenuItem(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		mc.ErrorResponse(c, http.StatusBadRequest, "Invalid menu item ID")
		return
	}

	var menuItem models.MenuItem
	if err := config.DB.First(&menuItem, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			mc.ErrorResponse(c, http.StatusNotFound, "Menu item not found")
			return
		}
		mc.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch menu item")
		return
	}

	var req UpdateMenuItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		mc.ValidationErrorResponse(c, err.Error())
		return
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
		mc.ErrorResponse(c, http.StatusInternalServerError, "Failed to update menu item")
		return
	}

	mc.SuccessResponse(c, menuItem, "Menu item updated successfully")
}

// DeleteMenuItem deletes a menu item (admin only)
func (mc *MenuController) DeleteMenuItem(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		mc.ErrorResponse(c, http.StatusBadRequest, "Invalid menu item ID")
		return
	}

	var menuItem models.MenuItem
	if err := config.DB.First(&menuItem, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			mc.ErrorResponse(c, http.StatusNotFound, "Menu item not found")
			return
		}
		mc.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch menu item")
		return
	}

	if err := config.DB.Delete(&menuItem).Error; err != nil {
		mc.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete menu item")
		return
	}

	mc.SuccessResponse(c, nil, "Menu item deleted successfully")
}

// GetCategories gets all available menu categories (public)
func (mc *MenuController) GetCategories(c *gin.Context) {
	categories := []map[string]string{
		{"value": string(models.CategoryAppetizer), "label": "Appetizer"},
		{"value": string(models.CategoryMain), "label": "Main Course"},
		{"value": string(models.CategoryDessert), "label": "Dessert"},
		{"value": string(models.CategoryDrink), "label": "Drink"},
	}

	mc.SuccessResponse(c, categories, "Categories retrieved successfully")
}

