package controllers

import (
	"strconv"
	"sync"
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
	// tableLocks stores mutexes for each table to prevent concurrent reservations
	tableLocks sync.Map // map[uint]*sync.Mutex
	// globalLock for operations that need global synchronization
	globalLock sync.RWMutex
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

// getTableLock gets or creates a mutex for a specific table
func (rc *ReservationController) getTableLock(tableID uint) *sync.Mutex {
	// Try to get existing mutex
	if lock, ok := rc.tableLocks.Load(tableID); ok {
		return lock.(*sync.Mutex)
	}

	// Create new mutex if it doesn't exist
	newLock := &sync.Mutex{}
	lock, _ := rc.tableLocks.LoadOrStore(tableID, newLock)
	return lock.(*sync.Mutex)
}

// CreateReservation creates a new reservation (customer only)
// Uses mutex and database transaction to prevent concurrent reservation conflicts
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

	// Get table-specific lock to prevent concurrent reservations for the same table
	tableLock := rc.getTableLock(req.TableID)
	tableLock.Lock()
	defer tableLock.Unlock()

	// Use database transaction to ensure atomicity
	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Check if table exists and lock the row (SELECT FOR UPDATE)
	var table models.Table
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&table, req.TableID).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return rc.ErrorResponse(c, fiber.StatusNotFound, "Table not found")
		}
		return rc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch table")
	}

	// Check if table is available
	if table.Status != models.TableStatusAvailable {
		tx.Rollback()
		return rc.ErrorResponse(c, fiber.StatusBadRequest, "Table is not available")
	}

	// Check if table is already reserved at this date and time
	var existingReservation models.Reservation
	if err := tx.Where("table_id = ? AND date = ? AND time = ? AND status IN ?",
		req.TableID,
		reservationDate,
		reservationTime,
		[]models.ReservationStatus{
			models.ReservationStatusPending,
			models.ReservationStatusConfirmed,
		}).First(&existingReservation).Error; err == nil {
		tx.Rollback()
		return rc.ErrorResponse(c, fiber.StatusConflict, "Table is already reserved at this date and time")
	} else if err != gorm.ErrRecordNotFound {
		tx.Rollback()
		return rc.ErrorResponse(c, fiber.StatusInternalServerError, "Database error")
	}

	// Create reservation
	reservation := models.Reservation{
		UserID:  userID.(uint),
		TableID: req.TableID,
		Date:    reservationDate,
		Time:    reservationTime,
		Status:  models.ReservationStatusPending,
	}

	if err := tx.Create(&reservation).Error; err != nil {
		tx.Rollback()
		return rc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create reservation")
	}

	// Update table status to reserved
	table.Status = models.TableStatusReserved
	if err := tx.Save(&table).Error; err != nil {
		tx.Rollback()
		return rc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update table status")
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return rc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to commit reservation")
	}

	// Load relationships for response (outside transaction)
	config.DB.Preload("User").Preload("Table").First(&reservation, reservation.ID)

	// Send notification asynchronously
	if rc.notificationService != nil {
		go rc.notificationService.SendReservationCreatedNotification(&reservation)
	}

	return rc.SuccessResponse(c, reservation, "Reservation created successfully")
}

// GetUserReservations gets all reservations for the current user (customer only)
func (rc *ReservationController) GetUserReservations(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return rc.ErrorResponse(c, fiber.StatusUnauthorized, "User not authenticated")
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
	}

	return rc.SuccessResponse(c, reservations, "Reservations retrieved successfully")
}

// GetReservationByID gets a single reservation by ID
func (rc *ReservationController) GetReservationByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return rc.ErrorResponse(c, fiber.StatusBadRequest, "Invalid reservation ID")
	}

	userID := c.Locals("user_id")
	userRole := c.Locals("user_role")

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
		}
		return rc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch reservation")
	}

	return rc.SuccessResponse(c, reservation, "Reservation retrieved successfully")
}

// CancelReservation cancels a reservation (customer can cancel their own, admin can cancel any)
// Uses mutex and database transaction to prevent concurrent conflicts
func (rc *ReservationController) CancelReservation(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return rc.ErrorResponse(c, fiber.StatusBadRequest, "Invalid reservation ID")
	}

	userID := c.Locals("user_id")
	userRole := c.Locals("user_role")

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
		}
		return rc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch reservation")
	}

	// Check if reservation can be cancelled
	if reservation.Status == models.ReservationStatusCancelled {
		return rc.ErrorResponse(c, fiber.StatusBadRequest, "Reservation is already cancelled")
	}

	if reservation.Status == models.ReservationStatusCompleted {
		return rc.ErrorResponse(c, fiber.StatusBadRequest, "Cannot cancel completed reservation")
	}

	// Get table-specific lock to prevent concurrent modifications
	tableLock := rc.getTableLock(reservation.TableID)
	tableLock.Lock()
	defer tableLock.Unlock()

	// Use database transaction to ensure atomicity
	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Reload reservation within transaction with lock
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&reservation, id).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return rc.ErrorResponse(c, fiber.StatusNotFound, "Reservation not found")
		}
		return rc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch reservation")
	}

	// Double-check status within transaction
	if reservation.Status == models.ReservationStatusCancelled {
		tx.Rollback()
		return rc.ErrorResponse(c, fiber.StatusBadRequest, "Reservation is already cancelled")
	}

	// Update reservation status
	reservation.Status = models.ReservationStatusCancelled
	if err := tx.Save(&reservation).Error; err != nil {
		tx.Rollback()
		return rc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to cancel reservation")
	}

	// Update table status if no other active reservations
	var activeReservations int64
	tx.Model(&models.Reservation{}).
		Where("table_id = ? AND status IN ?", reservation.TableID, []models.ReservationStatus{
			models.ReservationStatusPending,
			models.ReservationStatusConfirmed,
		}).Count(&activeReservations)

	if activeReservations == 0 {
		var table models.Table
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&table, reservation.TableID).Error; err == nil {
			table.Status = models.TableStatusAvailable
			if err := tx.Save(&table).Error; err != nil {
				tx.Rollback()
				return rc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update table status")
			}
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return rc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to commit cancellation")
	}

	// Load relationships for response (outside transaction)
	config.DB.Preload("User").Preload("Table").First(&reservation, reservation.ID)

	// Send notification asynchronously
	if rc.notificationService != nil {
		go rc.notificationService.SendReservationCancelledNotification(&reservation)
	}

	return rc.SuccessResponse(c, reservation, "Reservation cancelled successfully")
}

// GetAllReservations gets all reservations (admin only)
func (rc *ReservationController) GetAllReservations(c *fiber.Ctx) error {
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
	}

	return rc.SuccessResponse(c, reservations, "Reservations retrieved successfully")
}

// UpdateReservationStatus updates reservation status (admin only)
func (rc *ReservationController) UpdateReservationStatus(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return rc.ErrorResponse(c, fiber.StatusBadRequest, "Invalid reservation ID")
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
	}

	// First, get reservation to know which table to lock
	var reservation models.Reservation
	if err := config.DB.Preload("Table").First(&reservation, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return rc.ErrorResponse(c, fiber.StatusNotFound, "Reservation not found")
		}
		return rc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch reservation")
	}

	// Get table-specific lock to prevent concurrent modifications
	tableLock := rc.getTableLock(reservation.TableID)
	tableLock.Lock()
	defer tableLock.Unlock()

	// Use database transaction to ensure atomicity
	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Reload reservation within transaction with lock
	if err := tx.Set("gorm:query_option", "FOR UPDATE").Preload("Table").First(&reservation, id).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return rc.ErrorResponse(c, fiber.StatusNotFound, "Reservation not found")
		}
		return rc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch reservation")
	}

	// Update reservation status
	reservation.Status = req.Status
	if err := tx.Save(&reservation).Error; err != nil {
		tx.Rollback()
		return rc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update reservation status")
	}

	// Update table status based on reservation status
	var table models.Table
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&table, reservation.TableID).Error; err != nil {
		tx.Rollback()
		return rc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch table")
	}

	if req.Status == models.ReservationStatusConfirmed {
		table.Status = models.TableStatusReserved
	} else if req.Status == models.ReservationStatusCancelled || req.Status == models.ReservationStatusCompleted {
		// Check if there are other active reservations for this table
		var activeReservations int64
		tx.Model(&models.Reservation{}).
			Where("table_id = ? AND id != ? AND status IN ?", reservation.TableID, id, []models.ReservationStatus{
				models.ReservationStatusPending,
				models.ReservationStatusConfirmed,
			}).Count(&activeReservations)

		if activeReservations == 0 {
			table.Status = models.TableStatusAvailable
		}
	}

	if err := tx.Save(&table).Error; err != nil {
		tx.Rollback()
		return rc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update table status")
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return rc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to commit status update")
	}

	// Load relationships for response (outside transaction)
	config.DB.Preload("User").Preload("Table").First(&reservation, reservation.ID)

	// Send notification asynchronously
	if rc.notificationService != nil {
		go rc.notificationService.SendReservationStatusUpdatedNotification(&reservation)
	}

	return rc.SuccessResponse(c, reservation, "Reservation status updated successfully")
}

// GetReservationStatuses gets all available reservation statuses
func (rc *ReservationController) GetReservationStatuses(c *fiber.Ctx) error {
	statuses := []map[string]string{
		{"value": string(models.ReservationStatusPending), "label": "Pending"},
		{"value": string(models.ReservationStatusConfirmed), "label": "Confirmed"},
		{"value": string(models.ReservationStatusCancelled), "label": "Cancelled"},
		{"value": string(models.ReservationStatusCompleted), "label": "Completed"},
	}

	return rc.SuccessResponse(c, statuses, "Reservation statuses retrieved successfully")
}
