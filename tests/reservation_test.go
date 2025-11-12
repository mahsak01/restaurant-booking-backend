package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"restaurant-booking-backend/models"

	"github.com/stretchr/testify/assert"
)

func TestCreateReservation(t *testing.T) {
	SetupTestEnvironment(t)
	defer CleanupTestEnvironment(t)

	// Create test user and table
	user, _ := CreateTestUser("09123456789", "password123", "Test User", models.RoleCustomer)
	table, _ := CreateTestTable(1, 4, "Window", models.TableStatusAvailable)
	userToken := getAuthToken(t, "09123456789", "password123")

	t.Run("Create reservation", func(t *testing.T) {
		futureDate := time.Now().Add(24 * time.Hour)
		payload := map[string]interface{}{
			"table_id": table.ID,
			"date":     futureDate.Format("2006-01-02"),
			"time":     "19:00",
		}
		jsonValue, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/api/v1/reservations", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, response["success"].(bool))
	})

	t.Run("Create reservation with past date", func(t *testing.T) {
		pastDate := time.Now().Add(-24 * time.Hour)
		payload := map[string]interface{}{
			"table_id": table.ID,
			"date":     pastDate.Format("2006-01-02"),
			"time":     "19:00",
		}
		jsonValue, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/api/v1/reservations", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Create reservation without auth", func(t *testing.T) {
		futureDate := time.Now().Add(24 * time.Hour)
		payload := map[string]interface{}{
			"table_id": table.ID,
			"date":     futureDate.Format("2006-01-02"),
			"time":     "19:00",
		}
		jsonValue, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/api/v1/reservations", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestGetUserReservations(t *testing.T) {
	SetupTestEnvironment(t)
	defer CleanupTestEnvironment(t)

	user, _ := CreateTestUser("09123456789", "password123", "Test User", models.RoleCustomer)
	userToken := getAuthToken(t, "09123456789", "password123")

	t.Run("Get user reservations", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/reservations", nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, response["success"].(bool))
	})
}

func TestCancelReservation(t *testing.T) {
	SetupTestEnvironment(t)
	defer CleanupTestEnvironment(t)

	user, _ := CreateTestUser("09123456789", "password123", "Test User", models.RoleCustomer)
	table, _ := CreateTestTable(1, 4, "Window", models.TableStatusAvailable)
	userToken := getAuthToken(t, "09123456789", "password123")

	// Create a reservation
	futureDate := time.Now().Add(24 * time.Hour)
	reservation := models.Reservation{
		UserID:  user.ID,
		TableID: table.ID,
		Date:    futureDate,
		Time:    time.Date(0, 0, 0, 19, 0, 0, 0, time.UTC),
		Status:  models.ReservationStatusPending,
	}
	testDB.Create(&reservation)

	t.Run("Cancel reservation", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/api/v1/reservations/1", nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, response["success"].(bool))
	})
}

