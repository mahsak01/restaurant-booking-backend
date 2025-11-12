package services

import (
	"fmt"
	"log"

	"restaurant-booking-backend/config"
	"restaurant-booking-backend/models"
)

// NotificationService notification service
type NotificationService struct{}

// SendNotification sends a notification to a user
func (ns *NotificationService) SendNotification(userID uint, message string, notificationType models.NotificationType) error {
	// Create notification in database
	notification := models.Notification{
		UserID:  userID,
		Message: message,
		Type:    notificationType,
	}

	if err := config.DB.Create(&notification).Error; err != nil {
		log.Printf("Failed to create notification: %v", err)
		return err
	}

	// For MVP: Log to console
	// TODO: Later can be replaced with email or WebSocket
	log.Printf("[NOTIFICATION] User ID: %d | Type: %s | Message: %s", userID, notificationType, message)

	return nil
}

// SendReservationCreatedNotification sends notification when reservation is created
func (ns *NotificationService) SendReservationCreatedNotification(reservation *models.Reservation) error {
	// Load relationships if not loaded
	if reservation.User.ID == 0 {
		config.DB.Preload("User").Preload("Table").First(reservation, reservation.ID)
	}

	// Notification to customer
	customerMessage := fmt.Sprintf(
		"Your reservation for table #%d on %s at %s has been created successfully. Status: %s",
		reservation.Table.Number,
		reservation.Date.Format("2006-01-02"),
		reservation.Time.Format("15:04"),
		reservation.Status,
	)

	if err := ns.SendNotification(reservation.UserID, customerMessage, models.NotificationTypeReservation); err != nil {
		return err
	}

	// Notification to all admins
	var admins []models.User
	if err := config.DB.Where("role = ?", models.RoleAdmin).Find(&admins).Error; err == nil {
		adminMessage := fmt.Sprintf(
			"New reservation created: User %s (ID: %d) reserved table #%d on %s at %s",
			reservation.User.Name,
			reservation.UserID,
			reservation.Table.Number,
			reservation.Date.Format("2006-01-02"),
			reservation.Time.Format("15:04"),
		)

		for _, admin := range admins {
			ns.SendNotification(admin.ID, adminMessage, models.NotificationTypeReservation)
		}
	}

	return nil
}

// SendReservationCancelledNotification sends notification when reservation is cancelled
func (ns *NotificationService) SendReservationCancelledNotification(reservation *models.Reservation) error {
	// Load relationships if not loaded
	if reservation.User.ID == 0 {
		config.DB.Preload("User").Preload("Table").First(reservation, reservation.ID)
	}

	// Notification to customer
	customerMessage := fmt.Sprintf(
		"Your reservation for table #%d on %s at %s has been cancelled.",
		reservation.Table.Number,
		reservation.Date.Format("2006-01-02"),
		reservation.Time.Format("15:04"),
	)

	if err := ns.SendNotification(reservation.UserID, customerMessage, models.NotificationTypeReservation); err != nil {
		return err
	}

	// Notification to all admins
	var admins []models.User
	if err := config.DB.Where("role = ?", models.RoleAdmin).Find(&admins).Error; err == nil {
		adminMessage := fmt.Sprintf(
			"Reservation cancelled: User %s (ID: %d) cancelled reservation for table #%d on %s at %s",
			reservation.User.Name,
			reservation.UserID,
			reservation.Table.Number,
			reservation.Date.Format("2006-01-02"),
			reservation.Time.Format("15:04"),
		)

		for _, admin := range admins {
			ns.SendNotification(admin.ID, adminMessage, models.NotificationTypeReservation)
		}
	}

	return nil
}

// SendReservationStatusUpdatedNotification sends notification when reservation status is updated
func (ns *NotificationService) SendReservationStatusUpdatedNotification(reservation *models.Reservation) error {
	// Load relationships if not loaded
	if reservation.User.ID == 0 {
		config.DB.Preload("User").Preload("Table").First(reservation, reservation.ID)
	}

	// Notification to customer
	customerMessage := fmt.Sprintf(
		"Your reservation for table #%d on %s at %s has been updated. New status: %s",
		reservation.Table.Number,
		reservation.Date.Format("2006-01-02"),
		reservation.Time.Format("15:04"),
		reservation.Status,
	)

	return ns.SendNotification(reservation.UserID, customerMessage, models.NotificationTypeReservation)
}

