package controllers

import (
	"net/http"
	"strconv"

	"restaurant-booking-backend/config"
	"restaurant-booking-backend/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TableController table controller
type TableController struct {
	BaseController
}

// CreateTableRequest create table request structure
type CreateTableRequest struct {
	Number   int                `json:"number" binding:"required,gt=0"`
	Capacity int                `json:"capacity" binding:"required,gt=0"`
	Location string             `json:"location" binding:"required"`
	Status   models.TableStatus `json:"status"`
}

// UpdateTableRequest update table request structure
type UpdateTableRequest struct {
	Number   int                `json:"number" binding:"omitempty,gt=0"`
	Capacity int                `json:"capacity" binding:"omitempty,gt=0"`
	Location string             `json:"location"`
	Status   models.TableStatus `json:"status"`
}

// GetAllTables gets all tables with filtering (admin only)
func (tc *TableController) GetAllTables(c *gin.Context) {
	var tables []models.Table
	query := config.DB

	// Filter by status if provided
	status := c.Query("status")
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Filter by capacity if provided
	capacity := c.Query("capacity")
	if capacity != "" {
		capacityInt, err := strconv.Atoi(capacity)
		if err == nil {
			query = query.Where("capacity >= ?", capacityInt)
		}
	}

	// Filter by location if provided
	location := c.Query("location")
	if location != "" {
		query = query.Where("location ILIKE ?", "%"+location+"%")
	}

	if err := query.Order("number ASC").Find(&tables).Error; err != nil {
		tc.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch tables")
		return
	}

	tc.SuccessResponse(c, tables, "Tables retrieved successfully")
}

// GetAvailableTables gets available tables (public - for customers)
func (tc *TableController) GetAvailableTables(c *gin.Context) {
	var tables []models.Table
	query := config.DB.Where("status = ?", models.TableStatusAvailable)

	// Filter by minimum capacity if provided
	capacity := c.Query("capacity")
	if capacity != "" {
		capacityInt, err := strconv.Atoi(capacity)
		if err == nil && capacityInt > 0 {
			query = query.Where("capacity >= ?", capacityInt)
		}
	}

	// Filter by location if provided
	location := c.Query("location")
	if location != "" {
		query = query.Where("location ILIKE ?", "%"+location+"%")
	}

	if err := query.Order("number ASC").Find(&tables).Error; err != nil {
		tc.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch available tables")
		return
	}

	tc.SuccessResponse(c, tables, "Available tables retrieved successfully")
}

// GetTableByID gets a single table by ID
func (tc *TableController) GetTableByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		tc.ErrorResponse(c, http.StatusBadRequest, "Invalid table ID")
		return
	}

	var table models.Table
	if err := config.DB.First(&table, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			tc.ErrorResponse(c, http.StatusNotFound, "Table not found")
			return
		}
		tc.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch table")
		return
	}

	tc.SuccessResponse(c, table, "Table retrieved successfully")
}

// CreateTable creates a new table (admin only)
func (tc *TableController) CreateTable(c *gin.Context) {
	var req CreateTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		tc.ValidationErrorResponse(c, err.Error())
		return
	}

	// Check if table number already exists
	var existingTable models.Table
	if err := config.DB.Where("number = ?", req.Number).First(&existingTable).Error; err == nil {
		tc.ErrorResponse(c, http.StatusConflict, "Table with this number already exists")
		return
	} else if err != gorm.ErrRecordNotFound {
		tc.ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	// Set default status if not provided
	if req.Status == "" {
		req.Status = models.TableStatusAvailable
	}

	table := models.Table{
		Number:   req.Number,
		Capacity: req.Capacity,
		Location: req.Location,
		Status:   req.Status,
	}

	if err := config.DB.Create(&table).Error; err != nil {
		tc.ErrorResponse(c, http.StatusInternalServerError, "Failed to create table")
		return
	}

	tc.SuccessResponse(c, table, "Table created successfully")
}

// UpdateTable updates an existing table (admin only)
func (tc *TableController) UpdateTable(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		tc.ErrorResponse(c, http.StatusBadRequest, "Invalid table ID")
		return
	}

	var table models.Table
	if err := config.DB.First(&table, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			tc.ErrorResponse(c, http.StatusNotFound, "Table not found")
			return
		}
		tc.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch table")
		return
	}

	var req UpdateTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		tc.ValidationErrorResponse(c, err.Error())
		return
	}

	// Check if new number conflicts with existing table
	if req.Number > 0 && req.Number != table.Number {
		var existingTable models.Table
		if err := config.DB.Where("number = ? AND id != ?", req.Number, id).First(&existingTable).Error; err == nil {
			tc.ErrorResponse(c, http.StatusConflict, "Table with this number already exists")
			return
		} else if err != gorm.ErrRecordNotFound {
			tc.ErrorResponse(c, http.StatusInternalServerError, "Database error")
			return
		}
		table.Number = req.Number
	}

	// Update fields if provided
	if req.Capacity > 0 {
		table.Capacity = req.Capacity
	}
	if req.Location != "" {
		table.Location = req.Location
	}
	if req.Status != "" {
		table.Status = req.Status
	}

	if err := config.DB.Save(&table).Error; err != nil {
		tc.ErrorResponse(c, http.StatusInternalServerError, "Failed to update table")
		return
	}

	tc.SuccessResponse(c, table, "Table updated successfully")
}

// DeleteTable deletes a table (admin only)
func (tc *TableController) DeleteTable(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		tc.ErrorResponse(c, http.StatusBadRequest, "Invalid table ID")
		return
	}

	var table models.Table
	if err := config.DB.First(&table, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			tc.ErrorResponse(c, http.StatusNotFound, "Table not found")
			return
		}
		tc.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch table")
		return
	}

	// Check if table has active reservations
	var activeReservations int64
	config.DB.Model(&models.Reservation{}).
		Where("table_id = ? AND status IN ?", id, []models.ReservationStatus{
			models.ReservationStatusPending,
			models.ReservationStatusConfirmed,
		}).Count(&activeReservations)

	if activeReservations > 0 {
		tc.ErrorResponse(c, http.StatusBadRequest, "Cannot delete table with active reservations")
		return
	}

	if err := config.DB.Delete(&table).Error; err != nil {
		tc.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete table")
		return
	}

	tc.SuccessResponse(c, nil, "Table deleted successfully")
}

// GetTableStatuses gets all available table statuses
func (tc *TableController) GetTableStatuses(c *gin.Context) {
	statuses := []map[string]string{
		{"value": string(models.TableStatusAvailable), "label": "Available"},
		{"value": string(models.TableStatusReserved), "label": "Reserved"},
		{"value": string(models.TableStatusOccupied), "label": "Occupied"},
		{"value": string(models.TableStatusMaintenance), "label": "Maintenance"},
	}

	tc.SuccessResponse(c, statuses, "Table statuses retrieved successfully")
}

