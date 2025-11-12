package controllers

import (
	"net/http"

	"restaurant-booking-backend/config"
	"restaurant-booking-backend/models"

	"github.com/gin-gonic/gin"
)

// NotificationController notification controller
type NotificationController struct {
	BaseController
}

// GetUserNotifications gets all notifications for the current user
func (nc *NotificationController) GetUserNotifications(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		nc.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var notifications []models.Notification
	query := config.DB.Where("user_id = ?", userID.(uint))

	// Filter by type if provided
	notificationType := c.Query("type")
	if notificationType != "" {
		query = query.Where("type = ?", notificationType)
	}

	// Filter by read status (if we add read field later)
	// For now, just get all notifications

	if err := query.Order("created_at DESC").Find(&notifications).Error; err != nil {
		nc.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch notifications")
		return
	}

	nc.SuccessResponse(c, notifications, "Notifications retrieved successfully")
}

// GetUnreadNotificationsCount gets count of unread notifications for the current user
func (nc *NotificationController) GetUnreadNotificationsCount(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		nc.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var count int64
	config.DB.Model(&models.Notification{}).Where("user_id = ?", userID.(uint)).Count(&count)

	nc.SuccessResponse(c, gin.H{"count": count}, "Unread notifications count retrieved successfully")
}

// MarkNotificationAsRead marks a notification as read (for future implementation)
func (nc *NotificationController) MarkNotificationAsRead(c *gin.Context) {
	// This is a placeholder for future implementation
	// When we add a "read" field to the Notification model
	nc.SuccessResponse(c, nil, "Feature coming soon")
}

// DeleteNotification deletes a notification
func (nc *NotificationController) DeleteNotification(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		nc.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var notification models.Notification
	if err := config.DB.Where("id = ? AND user_id = ?", c.Param("id"), userID.(uint)).First(&notification).Error; err != nil {
		nc.ErrorResponse(c, http.StatusNotFound, "Notification not found")
		return
	}

	if err := config.DB.Delete(&notification).Error; err != nil {
		nc.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete notification")
		return
	}

	nc.SuccessResponse(c, nil, "Notification deleted successfully")
}

