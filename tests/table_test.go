package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"restaurant-booking-backend/models"

	"github.com/stretchr/testify/assert"
)

func TestGetAvailableTables(t *testing.T) {
	SetupTestEnvironment(t)
	defer CleanupTestEnvironment(t)

	// Create test tables
	CreateTestTable(1, 4, "Window", models.TableStatusAvailable)
	CreateTestTable(2, 6, "Corner", models.TableStatusReserved)
	CreateTestTable(3, 2, "Window", models.TableStatusAvailable)

	t.Run("Get available tables", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/tables/available", nil)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, response["success"].(bool))
	})

	t.Run("Get available tables with capacity filter", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/tables/available?capacity=4", nil)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestGetTableStatuses(t *testing.T) {
	SetupTestEnvironment(t)
	defer CleanupTestEnvironment(t)

	t.Run("Get table statuses", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/tables/statuses", nil)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, response["success"].(bool))
	})
}

