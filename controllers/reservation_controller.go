package controllers

import (
	"strconv"
	"time"

	"restaurant-booking-backend/config"
	"restaurant-booking-backend/models"
	"restaurant-booking-backend/services"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// ReservationController reservation controller
type ReservationController struct {
	BaseController
	notificationService *services.NotificationService
}

// NewReservationController creates a new reservation controller
func NewReservationController() *ReservationController {
	return &ReservationController{
		notificationService: &services.NotificationService{},
	}
}

// CreateReservationRequest create reservation request structure
type CreateReservationRequest struct {
	TableID uint   `json:"table_id" binding:"required"`
	Date    string `json:"date" binding:"required"` // Format: "2006-01-02"
	Time    string `json:"time" binding:"required"` // Format: "15:04"
}

// UpdateReservationStatusRequest update reservation status request structure
type UpdateReservationStatusRequest struct {
	Status models.ReservationStatus `json:"status" binding:"required"`
}

// CreateReservation creates a new reservation (customer only)
func (rc *ReservationController) CreateReservation(c *fiber.Ctx) error {
	// Get user ID from context (set by auth middleware)
	userID := c.Locals("user_id")
	if userID == nil {
		return rc.ErrorResponse(c, fiber.StatusUnauthorized, "User not authenticated")
	}

	var req CreateReservationRequest
	if err := c.BodyParser(&req); err != nil {
		return rc.ValidationErrorResponse(c, err.Error())
	}

	// Parse date and time
	reservationDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return rc.ErrorResponse(c, fiber.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD")
	}

	reservationTime, err := time.Parse("15:04", req.Time)
	if err != nil {
		return rc.ErrorResponse(c, fiber.StatusBadRequest, "Invalid time format. Use HH:MM")
	}

	// Combine date and time
	reservationDateTime := time.Date(
		reservationDate.Year(),
		reservationDate.Month(),
		reservationDate.Day(),
		reservationTime.Hour(),
		reservationTime.Minute(),
		0, 0, time.UTC,
	)

	// Check if reservation is in the past
	if reservationDateTime.Before(time.Now()) {
		return rc.ErrorResponse(c, fiber.StatusBadRequest, "Cannot make reservation in the past")
	}

	// Check if table exists
	var table models.Table
	if err := config.DB.First(&table, req.TableID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return rc.ErrorResponse(c, fiber.StatusNotFound, "Table not found")
		}
		return rc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch table")
	}

	// Check if table is available
	if table.Status != models.TableStatusAvailable {
		return rc.ErrorResponse(c, fiber.StatusBadRequest, "Table is not available")
	}

	// Check if table is already reserved at this date and time
	var existingReservation models.Reservation
	if err := config.DB.Where("table_id = ? AND date = ? AND time = ? AND status IN ?",
		req.TableID,
		reservationDate,
		reservationTime,
		[]models.ReservationStatus{
			models.ReservationStatusPending,
			models.ReservationStatusConfirmed,
		}).First(&existingReservation).Error; err == nil {
		return rc.ErrorResponse(c, fiber.StatusConflict, "Table is already reserved at this date and time")
		return
	} else if err != gorm.ErrRecordNotFound {
		return rc.ErrorResponse(c, fiber.StatusInternalServerError, "Database error")
		return
	}

	// Create reservation
	reservation := models.Reservation{
		UserID: userID.(uint),
		TableID: req.TableID,
		Date:   reservationDate,
		Time:   reservationTime,
		Status: models.ReservationStatusPending,
	}

	if err := config.DB.Create(&reservation).Error; err != nil {
		return rc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create reservation")
		return
	}

	// Update table status to reserved
	table.Status = models.TableStatusReserved
	config.DB.Save(&table)

	// Load relationships for response
	config.DB.Preload("User").Preload("Table").First(&reservation, reservation.ID)

	// Send notification
	if rc.notificationService != nil {
		go rc.notificationService.SendReservationCreatedNotification(&reservation)
	}

		return rc.SuccessResponse(c, reservation, "Reservation created successfully")
}

// GetUserReservations gets all reservations for the current user (customer only)
func (rc *ReservationController) GetUserReservations(c *fiber.Ctx) error) {
	userID := c.Locals("user_id")
	if userID == nil {
		return rc.ErrorResponse(c, fiber.StatusUnauthorized, "User not authenticated")
		return
	}

	var reservations []models.Reservation
	query := config.DB.Where("user_id = ?", userID.(uint)).Preload("Table")

	// Filter by status if provided
	status := c.Query("status")
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Filter by date if provided
	date := c.Query("date")
	if date != "" {
		query = query.Where("date = ?", date)
	}

	if err := query.Order("date DESC, time DESC").Find(&reservations).Error; err != nil {
		return rc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch reservations")
		return
	}

		return rc.SuccessResponse(c, reservations, "Reservations retrieved successfully")
}

// GetReservationByID gets a single reservation by ID
func (rc *ReservationController) GetReservationByID(c *fiber.Ctx) error) {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return rc.ErrorResponse(c, fiber.StatusBadRequest, "Invalid reservation ID")
		return
	}

	userID, exists := c.Locals("user_id")
	userRole, _ := c.Locals("user_role")

	var reservation models.Reservation
	query := config.DB.Preload("User").Preload("Table")

	// If user is customer, only show their own reservations
	if userID != nil && userRole == "customer" {
		query = query.Where("id = ? AND user_id = ?", id, userID.(uint))
	} else {
		query = query.Where("id = ?", id)
	}

	if err := query.First(&reservation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return rc.ErrorResponse(c, fiber.StatusNotFound, "Reservation not found")
			return
		}
		return rc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch reservation")
		return
	}

		return rc.SuccessResponse(c, reservation, "Reservation retrieved successfully")
}

// CancelReservation cancels a reservation (customer can cancel their own, admin can cancel any)
func (rc *ReservationController) CancelReservation(c *fiber.Ctx) error) {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return rc.ErrorResponse(c, fiber.StatusBadRequest, "Invalid reservation ID")
		return
	}

	userID, exists := c.Locals("user_id")
	userRole, _ := c.Locals("user_role")

	var reservation models.Reservation
	query := config.DB

	// If user is customer, only allow canceling their own reservations
	if userID != nil && userRole == "customer" {
		query = query.Where("id = ? AND user_id = ?", id, userID.(uint))
	} else {
		query = query.Where("id = ?", id)
	}

	if err := query.First(&reservation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return rc.ErrorResponse(c, fiber.StatusNotFound, "Reservation not found")
			return
		}
		return rc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch reservation")
		return
	}

	// Check if reservation can be cancelled
	if reservation.Status == models.ReservationStatusCancelled {
		return rc.ErrorResponse(c, fiber.StatusBadRequest, "Reservation is already cancelled")
		return
	}

	if reservation.Status == models.ReservationStatusCompleted {
		return rc.ErrorResponse(c, fiber.StatusBadRequest, "Cannot cancel completed reservation")
		return
	}

	// Update reservation status
	reservation.Status = models.ReservationStatusCancelled
	if err := config.DB.Save(&reservation).Error; err != nil {
		return rc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to cancel reservation")
		return
	}

	// Update table status if no other active reservations
	var activeReservations int64
	config.DB.Model(&models.Reservation{}).
		Where("table_id = ? AND status IN ?", reservation.TableID, []models.ReservationStatus{
			models.ReservationStatusPending,
			models.ReservationStatusConfirmed,
		}).Count(&activeReservations)

	if activeReservations == 0 {
		var table models.Table
		config.DB.First(&table, reservation.TableID)
		table.Status = models.TableStatusAvailable
		config.DB.Save(&table)
	}

	config.DB.Preload("User").Preload("Table").First(&reservation, reservation.ID)

	// Send notification
	if rc.notificationService != nil {
		go rc.notificationService.SendReservationCancelledNotification(&reservation)
	}

		return rc.SuccessResponse(c, reservation, "Reservation cancelled successfully")
}

// GetAllReservations gets all reservations (admin only)
func (rc *ReservationController) GetAllReservations(c *fiber.Ctx) error) {
	var reservations []models.Reservation
	query := config.DB.Preload("User").Preload("Table")

	// Filter by status if provided
	status := c.Query("status")
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Filter by date if provided
	date := c.Query("date")
	if date != "" {
		query = query.Where("date = ?", date)
	}

	// Filter by user_id if provided
	userID := c.Query("user_id")
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	// Filter by table_id if provided
	tableID := c.Query("table_id")
	if tableID != "" {
		query = query.Where("table_id = ?", tableID)
	}

	if err := query.Order("date DESC, time DESC").Find(&reservations).Error; err != nil {
		return rc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch reservations")
		return
	}

		return rc.SuccessResponse(c, reservations, "Reservations retrieved successfully")
}

// UpdateReservationStatus updates reservation status (admin only)
func (rc *ReservationController) UpdateReservationStatus(c *fiber.Ctx) error) {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return rc.ErrorResponse(c, fiber.StatusBadRequest, "Invalid reservation ID")
		return
	}

	var reservation models.Reservation
	if err := config.DB.Preload("Table").First(&reservation, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return rc.ErrorResponse(c, fiber.StatusNotFound, "Reservation not found")
			return
		}
		return rc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch reservation")
		return
	}

	var req UpdateReservationStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return rc.ValidationErrorResponse(c, err.Error())
	}

	// Validate status
	validStatus := false
	for _, status := range []models.ReservationStatus{
		models.ReservationStatusPending,
		models.ReservationStatusConfirmed,
		models.ReservationStatusCancelled,
		models.ReservationStatusCompleted,
	} {
		if req.Status == status {
			validStatus = true
			break
		}
	}

	if !validStatus {
		return rc.ErrorResponse(c, fiber.StatusBadRequest, "Invalid reservation status")
		return
	}

	reservation.Status = req.Status

	if err := config.DB.Save(&reservation).Error; err != nil {
		return rc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update reservation status")
		return
	}

	// Update table status based on reservation status
	var table models.Table
	config.DB.First(&table, reservation.TableID)

	if req.Status == models.ReservationStatusConfirmed {
		table.Status = models.TableStatusReserved
	} else if req.Status == models.ReservationStatusCancelled || req.Status == models.ReservationStatusCompleted {
		// Check if there are other active reservations for this table
		var activeReservations int64
		config.DB.Model(&models.Reservation{}).
			Where("table_id = ? AND id != ? AND status IN ?", reservation.TableID, id, []models.ReservationStatus{
				models.ReservationStatusPending,
				models.ReservationStatusConfirmed,
			}).Count(&activeReservations)

		if activeReservations == 0 {
			table.Status = models.TableStatusAvailable
		}
	}

	config.DB.Save(&table)

	config.DB.Preload("User").Preload("Table").First(&reservation, reservation.ID)

	// Send notification
	if rc.notificationService != nil {
		go rc.notificationService.SendReservationStatusUpdatedNotification(&reservation)
	}

		return rc.SuccessResponse(c, reservation, "Reservation status updated successfully")
}

// GetReservationStatuses gets all available reservation statuses
func (rc *ReservationController) GetReservationStatuses(c *fiber.Ctx) error) {
	statuses := []map[string]string{
		{"value": string(models.ReservationStatusPending), "label": "Pending"},
		{"value": string(models.ReservationStatusConfirmed), "label": "Confirmed"},
		{"value": string(models.ReservationStatusCancelled), "label": "Cancelled"},
		{"value": string(models.ReservationStatusCompleted), "label": "Completed"},
	}

		return rc.SuccessResponse(c, statuses, "Reservation statuses retrieved successfully")
}

