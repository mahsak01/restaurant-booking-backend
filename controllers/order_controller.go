package controllers

import (
	"strconv"
	"strings"

	"restaurant-booking-backend/config"
	"restaurant-booking-backend/models"
	"restaurant-booking-backend/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// OrderController order controller
type OrderController struct {
	BaseController
}

// OrderItemRequest order item request structure
type OrderItemRequest struct {
	MenuItemID uint `json:"menu_item_id" validate:"required"`
	Quantity   int  `json:"quantity" validate:"required,min=1"`
}

// CreateOrderRequest create order request structure
type CreateOrderRequest struct {
	Items []OrderItemRequest `json:"items" validate:"required,min=1"`
}

// CreateOrderByAdminRequest create order by admin request structure
type CreateOrderByAdminRequest struct {
	Phone    string             `json:"phone" validate:"required"` // User phone number
	Name     string             `json:"name"`                      // First name (required if user doesn't exist)
	LastName string             `json:"last_name"`                 // Last name (optional)
	Items    []OrderItemRequest `json:"items" validate:"required,min=1"`
}

// UpdateOrderStatusRequest update order status request structure
type UpdateOrderStatusRequest struct {
	Status string `json:"status" validate:"required"`
}

// CreateOrder creates a new order (customer only)
func (oc *OrderController) CreateOrder(c *fiber.Ctx) error {
	// Get user ID from context (set by auth middleware)
	userID := c.Locals("user_id")
	if userID == nil {
		return oc.ErrorResponse(c, fiber.StatusUnauthorized, "User not authenticated")
	}

	var req CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return oc.ValidationErrorResponse(c, err.Error())
	}

	// Validate request
	if len(req.Items) == 0 {
		return oc.ValidationErrorResponse(c, "At least one item is required")
	}

	// Start transaction
	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var totalPrice float64
	var orderItems []models.OrderItem

	// Validate and process each item
	for _, itemReq := range req.Items {
		if itemReq.Quantity <= 0 {
			tx.Rollback()
			return oc.ValidationErrorResponse(c, "Quantity must be greater than 0")
		}

		// Get menu item
		var menuItem models.MenuItem
		if err := tx.First(&menuItem, itemReq.MenuItemID).Error; err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				return oc.ErrorResponse(c, fiber.StatusNotFound, "Menu item not found")
			}
			return oc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch menu item")
		}

		// Check if menu item is available
		if !menuItem.IsAvailable {
			tx.Rollback()
			return oc.ErrorResponse(c, fiber.StatusBadRequest, "Menu item is not available: "+menuItem.Name)
		}

		// Calculate item total
		itemTotal := menuItem.Price * float64(itemReq.Quantity)
		totalPrice += itemTotal

		// Create order item
		orderItem := models.OrderItem{
			MenuItemID: menuItem.ID,
			Quantity:   itemReq.Quantity,
			Price:      menuItem.Price, // Store price at time of order
		}
		orderItems = append(orderItems, orderItem)
	}

	// Create order
	order := models.Order{
		UserID:     userID.(uint),
		Status:     models.OrderStatusPending,
		TotalPrice: totalPrice,
	}

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		return oc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create order")
	}

	// Create order items
	for i := range orderItems {
		orderItems[i].OrderID = order.ID
		if err := tx.Create(&orderItems[i]).Error; err != nil {
			tx.Rollback()
			return oc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create order items")
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return oc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to commit order")
	}

	// Load relationships for response
	config.DB.Preload("User").Preload("OrderItems.MenuItem").First(&order, order.ID)

	return oc.SuccessResponse(c, order, "Order created successfully")
}

// GetUserOrders gets all orders for the current user (customer only)
func (oc *OrderController) GetUserOrders(c *fiber.Ctx) error {
	// Get user ID from context
	userID := c.Locals("user_id")
	if userID == nil {
		return oc.ErrorResponse(c, fiber.StatusUnauthorized, "User not authenticated")
	}

	var orders []models.Order
	query := config.DB.Where("user_id = ?", userID.(uint))

	// Filter by status if provided
	status := c.Query("status")
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Order by created_at descending (newest first)
	if err := query.Preload("OrderItems.MenuItem").Order("created_at DESC").Find(&orders).Error; err != nil {
		return oc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch orders")
	}

	return oc.SuccessResponse(c, orders, "Orders retrieved successfully")
}

// GetAllOrders gets all orders (admin only)
func (oc *OrderController) GetAllOrders(c *fiber.Ctx) error {
	var orders []models.Order
	query := config.DB

	// Filter by user_id if provided
	userIDStr := c.Query("user_id")
	if userIDStr != "" {
		userID, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			return oc.ErrorResponse(c, fiber.StatusBadRequest, "Invalid user_id")
		}
		query = query.Where("user_id = ?", userID)
	}

	// Filter by status if provided
	status := c.Query("status")
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Order by created_at descending (newest first)
	if err := query.Preload("User").Preload("OrderItems.MenuItem").Order("created_at DESC").Find(&orders).Error; err != nil {
		return oc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch orders")
	}

	return oc.SuccessResponse(c, orders, "Orders retrieved successfully")
}

// GetOrderByID gets a single order by ID
func (oc *OrderController) GetOrderByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return oc.ErrorResponse(c, fiber.StatusBadRequest, "Invalid order ID")
	}

	// Get user ID from context
	userID := c.Locals("user_id")
	if userID == nil {
		return oc.ErrorResponse(c, fiber.StatusUnauthorized, "User not authenticated")
	}

	// Get user role
	userRole := c.Locals("user_role")
	isAdmin := userRole == "admin"

	var order models.Order
	query := config.DB.Preload("User").Preload("OrderItems.MenuItem")

	// If not admin, only allow access to own orders
	if !isAdmin {
		query = query.Where("id = ? AND user_id = ?", id, userID.(uint))
	} else {
		query = query.Where("id = ?", id)
	}

	if err := query.First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return oc.ErrorResponse(c, fiber.StatusNotFound, "Order not found")
		}
		return oc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch order")
	}

	return oc.SuccessResponse(c, order, "Order retrieved successfully")
}

// UpdateOrderStatus updates order status (admin only)
func (oc *OrderController) UpdateOrderStatus(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return oc.ErrorResponse(c, fiber.StatusBadRequest, "Invalid order ID")
	}

	var req UpdateOrderStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return oc.ValidationErrorResponse(c, err.Error())
	}

	// Validate status
	validStatuses := map[string]bool{
		string(models.OrderStatusPending):   true,
		string(models.OrderStatusConfirmed): true,
		string(models.OrderStatusPreparing): true,
		string(models.OrderStatusReady):     true,
		string(models.OrderStatusDelivered): true,
		string(models.OrderStatusCancelled): true,
	}

	if !validStatuses[req.Status] {
		return oc.ErrorResponse(c, fiber.StatusBadRequest, "Invalid order status")
	}

	// Get order
	var order models.Order
	if err := config.DB.First(&order, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return oc.ErrorResponse(c, fiber.StatusNotFound, "Order not found")
		}
		return oc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch order")
	}

	// Update status
	order.Status = models.OrderStatus(req.Status)
	if err := config.DB.Save(&order).Error; err != nil {
		return oc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update order status")
	}

	// Load relationships for response
	config.DB.Preload("User").Preload("OrderItems.MenuItem").First(&order, order.ID)

	return oc.SuccessResponse(c, order, "Order status updated successfully")
}

// CreateOrderByAdmin creates an order by admin for a user (searches by phone, creates user if not exists)
func (oc *OrderController) CreateOrderByAdmin(c *fiber.Ctx) error {
	var req CreateOrderByAdminRequest
	if err := c.BodyParser(&req); err != nil {
		return oc.ValidationErrorResponse(c, err.Error())
	}

	// Validate phone number
	req.Phone = strings.TrimSpace(req.Phone)
	if !utils.ValidatePhoneNumber(req.Phone) {
		return oc.ErrorResponse(c, fiber.StatusBadRequest, "Invalid phone number format")
	}

	// Validate name (required if user doesn't exist)
	req.Name = strings.TrimSpace(req.Name)
	req.LastName = strings.TrimSpace(req.LastName)

	// Validate request
	if len(req.Items) == 0 {
		return oc.ValidationErrorResponse(c, "At least one item is required")
	}

	// Get or create user by phone
	var user models.User
	// Check if active user exists
	if err := config.DB.Where("phone = ?", req.Phone).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// User doesn't exist (or was soft-deleted), create new user without password
			if req.Name == "" {
				return oc.ErrorResponse(c, fiber.StatusBadRequest, "Name is required when creating new user")
			}
			user = models.User{
				Phone:    req.Phone,
				Password: "", // Empty password - user must set password to login
				Name:     req.Name,
				LastName: req.LastName,
				Role:     models.RoleCustomer,
			}
			if err := config.DB.Create(&user).Error; err != nil {
				return oc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create user")
			}
		} else {
			return oc.ErrorResponse(c, fiber.StatusInternalServerError, "Database error while searching for user")
		}
	} else {
		// Active user exists, update name and last name if provided and different
		updated := false
		if req.Name != "" && user.Name != req.Name {
			user.Name = req.Name
			updated = true
		}
		if req.LastName != "" && user.LastName != req.LastName {
			user.LastName = req.LastName
			updated = true
		}
		if updated {
			if err := config.DB.Save(&user).Error; err != nil {
				return oc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update user information")
			}
		}
	}

	// Start transaction
	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var totalPrice float64
	var orderItems []models.OrderItem

	// Validate and process each item
	for _, itemReq := range req.Items {
		if itemReq.Quantity <= 0 {
			tx.Rollback()
			return oc.ValidationErrorResponse(c, "Quantity must be greater than 0")
		}

		// Get menu item
		var menuItem models.MenuItem
		if err := tx.First(&menuItem, itemReq.MenuItemID).Error; err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				return oc.ErrorResponse(c, fiber.StatusNotFound, "Menu item not found")
			}
			return oc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch menu item")
		}

		// Check if menu item is available
		if !menuItem.IsAvailable {
			tx.Rollback()
			return oc.ErrorResponse(c, fiber.StatusBadRequest, "Menu item is not available: "+menuItem.Name)
		}

		// Calculate item total
		itemTotal := menuItem.Price * float64(itemReq.Quantity)
		totalPrice += itemTotal

		// Create order item
		orderItem := models.OrderItem{
			MenuItemID: menuItem.ID,
			Quantity:   itemReq.Quantity,
			Price:      menuItem.Price, // Store price at time of order
		}
		orderItems = append(orderItems, orderItem)
	}

	// Create order
	order := models.Order{
		UserID:     user.ID,
		Status:     models.OrderStatusPending,
		TotalPrice: totalPrice,
	}

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		return oc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create order")
	}

	// Create order items
	for i := range orderItems {
		orderItems[i].OrderID = order.ID
		if err := tx.Create(&orderItems[i]).Error; err != nil {
			tx.Rollback()
			return oc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create order items")
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return oc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to commit order")
	}

	// Load relationships for response
	config.DB.Preload("User").Preload("OrderItems.MenuItem").First(&order, order.ID)

	return oc.SuccessResponse(c, order, "Order created successfully by admin")
}

// GetOrderStatuses returns all available order statuses
func (oc *OrderController) GetOrderStatuses(c *fiber.Ctx) error {
	statuses := []string{
		string(models.OrderStatusPending),
		string(models.OrderStatusConfirmed),
		string(models.OrderStatusPreparing),
		string(models.OrderStatusReady),
		string(models.OrderStatusDelivered),
		string(models.OrderStatusCancelled),
	}

	return oc.SuccessResponse(c, statuses, "Order statuses retrieved successfully")
}
