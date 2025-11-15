package controllers

import (
	"restaurant-booking-backend/config"
	"restaurant-booking-backend/models"

	"github.com/gofiber/fiber/v2"
)

// NotificationController notification controller
type NotificationController struct {
	BaseController
}

// GetUserNotifications gets all notifications for the current user
func (nc *NotificationController) GetUserNotifications(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return nc.ErrorResponse(c, fiber.StatusUnauthorized, "User not authenticated")
	}

	var notifications []models.Notification
	query := config.DB.Where("user_id = ?", userID.(uint))

	// Filter by type if provided
	notificationType := c.Query("type")
	if notificationType != "" {
		query = query.Where("type = ?", notificationType)
	}

	// Filter by read status if provided
	readStatus := c.Query("read")
	if readStatus == "true" {
		query = query.Where("is_read = ?", true)
	} else if readStatus == "false" {
		query = query.Where("is_read = ?", false)
	}

	if err := query.Order("created_at DESC").Find(&notifications).Error; err != nil {
		return nc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch notifications")
	}

	return nc.SuccessResponse(c, notifications, "Notifications retrieved successfully")
}

// GetUnreadNotificationsCount gets count of unread notifications for the current user
func (nc *NotificationController) GetUnreadNotificationsCount(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return nc.ErrorResponse(c, fiber.StatusUnauthorized, "User not authenticated")
	}

	var count int64
	config.DB.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID.(uint), false).
		Count(&count)

	return nc.SuccessResponse(c, fiber.Map{"count": count}, "Unread notifications count retrieved successfully")
}

// MarkNotificationAsRead marks a notification as read
func (nc *NotificationController) MarkNotificationAsRead(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return nc.ErrorResponse(c, fiber.StatusUnauthorized, "User not authenticated")
	}

	var notification models.Notification
	if err := config.DB.Where("id = ? AND user_id = ?", c.Params("id"), userID.(uint)).First(&notification).Error; err != nil {
		return nc.ErrorResponse(c, fiber.StatusNotFound, "Notification not found")
	}

	if notification.IsRead {
		return nc.SuccessResponse(c, notification, "Notification is already marked as read")
	}

	notification.IsRead = true
	if err := config.DB.Save(&notification).Error; err != nil {
		return nc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to mark notification as read")
	}

	return nc.SuccessResponse(c, notification, "Notification marked as read successfully")
}

// DeleteNotification deletes a notification
func (nc *NotificationController) DeleteNotification(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return nc.ErrorResponse(c, fiber.StatusUnauthorized, "User not authenticated")
	}

	var notification models.Notification
	if err := config.DB.Where("id = ? AND user_id = ?", c.Params("id"), userID.(uint)).First(&notification).Error; err != nil {
		return nc.ErrorResponse(c, fiber.StatusNotFound, "Notification not found")
	}

	if err := config.DB.Delete(&notification).Error; err != nil {
		return nc.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete notification")
	}

	return nc.SuccessResponse(c, nil, "Notification deleted successfully")
}
