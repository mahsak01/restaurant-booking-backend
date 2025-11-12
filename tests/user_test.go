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

func TestGetAllUsers(t *testing.T) {
	SetupTestEnvironment(t)
	defer CleanupTestEnvironment(t)

	// Create test users
	CreateTestUser("09111111111", "password123", "Admin User", models.RoleAdmin)
	CreateTestUser("09222222222", "password123", "Customer User", models.RoleCustomer)

	adminToken := getAuthToken(t, "09111111111", "password123")

	t.Run("Get all users as admin", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/admin/users", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, response["success"].(bool))
	})

	t.Run("Get users without auth", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/admin/users", nil)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestUpdateUserRole(t *testing.T) {
	SetupTestEnvironment(t)
	defer CleanupTestEnvironment(t)

	admin, _ := CreateTestUser("09111111111", "password123", "Admin", models.RoleAdmin)
	customer, _ := CreateTestUser("09222222222", "password123", "Customer", models.RoleCustomer)
	adminToken := getAuthToken(t, "09111111111", "password123")

	t.Run("Update user role as admin", func(t *testing.T) {
		payload := map[string]interface{}{
			"role": "admin",
		}
		jsonValue, _ := json.Marshal(payload)

		req, _ := http.NewRequest("PUT", "/api/v1/admin/users/2/role", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, response["success"].(bool))
	})
}

func TestGetProfile(t *testing.T) {
	SetupTestEnvironment(t)
	defer CleanupTestEnvironment(t)

	user, _ := CreateTestUser("09123456789", "password123", "Test User", models.RoleCustomer)
	userToken := getAuthToken(t, "09123456789", "password123")

	t.Run("Get profile", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/profile", nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, response["success"].(bool))
		data := response["data"].(map[string]interface{})
		assert.Equal(t, float64(user.ID), data["user_id"])
	})
}

