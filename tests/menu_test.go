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

func TestGetAllMenuItems(t *testing.T) {
	SetupTestEnvironment(t)
	defer CleanupTestEnvironment(t)

	// Create test menu items
	CreateTestMenuItem("Pasta", "Delicious pasta", 25.99, models.CategoryMain)
	CreateTestMenuItem("Salad", "Fresh salad", 12.50, models.CategoryAppetizer)

	t.Run("Get all menu items", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/menu", nil)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, response["success"].(bool))
	})

	t.Run("Get menu items by category", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/menu?category=main", nil)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Search menu items", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/menu?search=pasta", nil)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestGetMenuItemByID(t *testing.T) {
	SetupTestEnvironment(t)
	defer CleanupTestEnvironment(t)

	item, _ := CreateTestMenuItem("Pasta", "Delicious pasta", 25.99, models.CategoryMain)

	t.Run("Get menu item by ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/menu/1", nil)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, response["success"].(bool))
	})

	t.Run("Get non-existent menu item", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/menu/999", nil)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestCreateMenuItem(t *testing.T) {
	SetupTestEnvironment(t)
	defer CleanupTestEnvironment(t)

	// Create admin user and get token
	admin, _ := CreateTestUser("09111111111", "password123", "Admin", models.RoleAdmin)
	adminToken := getAuthToken(t, "09111111111", "password123")

	t.Run("Create menu item as admin", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":        "Burger",
			"description": "Delicious burger",
			"price":       15.99,
			"category":    "main",
		}
		jsonValue, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/api/v1/admin/menu", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, response["success"].(bool))
	})

	t.Run("Create menu item without auth", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":     "Burger",
			"price":    15.99,
			"category": "main",
		}
		jsonValue, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/api/v1/admin/menu", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

