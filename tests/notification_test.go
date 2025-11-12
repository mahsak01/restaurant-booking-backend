package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"restaurant-booking-backend/models"

	"github.com/stretchr/testify/assert"
)

func TestGetUserNotifications(t *testing.T) {
	SetupTestEnvironment(t)
	defer CleanupTestEnvironment(t)

	user, _ := CreateTestUser("09123456789", "password123", "Test User", models.RoleCustomer)
	userToken := getAuthToken(t, "09123456789", "password123")

	// Create test notification
	notification := models.Notification{
		UserID:  user.ID,
		Message: "Test notification",
		Type:    models.NotificationTypeReservation,
		IsRead:  false,
	}
	testDB.Create(&notification)

	t.Run("Get user notifications", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/notifications", nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, response["success"].(bool))
	})

	t.Run("Get unread notifications count", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/notifications/count", nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, response["success"].(bool))
	})
}

func TestMarkNotificationAsRead(t *testing.T) {
	SetupTestEnvironment(t)
	defer CleanupTestEnvironment(t)

	user, _ := CreateTestUser("09123456789", "password123", "Test User", models.RoleCustomer)
	userToken := getAuthToken(t, "09123456789", "password123")

	// Create test notification
	notification := models.Notification{
		UserID:  user.ID,
		Message: "Test notification",
		Type:    models.NotificationTypeReservation,
		IsRead:  false,
	}
	testDB.Create(&notification)

	t.Run("Mark notification as read", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/api/v1/notifications/1/read", nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, response["success"].(bool))
	})
}

func TestDeleteNotification(t *testing.T) {
	SetupTestEnvironment(t)
	defer CleanupTestEnvironment(t)

	user, _ := CreateTestUser("09123456789", "password123", "Test User", models.RoleCustomer)
	userToken := getAuthToken(t, "09123456789", "password123")

	// Create test notification
	notification := models.Notification{
		UserID:  user.ID,
		Message: "Test notification",
		Type:    models.NotificationTypeReservation,
		IsRead:  false,
	}
	testDB.Create(&notification)

	t.Run("Delete notification", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/api/v1/notifications/1", nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, response["success"].(bool))
	})
}

