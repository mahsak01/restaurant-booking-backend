package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"restaurant-booking-backend/models"
	"restaurant-booking-backend/utils"

	"github.com/stretchr/testify/assert"
)

func TestSignup(t *testing.T) {
	SetupTestEnvironment(t)
	defer CleanupTestEnvironment(t)

	t.Run("Successful signup", func(t *testing.T) {
		payload := map[string]interface{}{
			"phone":    "09123456789",
			"password": "password123",
			"name":     "Test User",
		}
		jsonValue, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/api/v1/auth/signup", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, response["success"].(bool))
		assert.NotNil(t, response["data"])
	})

	t.Run("Signup with invalid phone", func(t *testing.T) {
		payload := map[string]interface{}{
			"phone":    "invalid",
			"password": "password123",
			"name":     "Test User",
		}
		jsonValue, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/api/v1/auth/signup", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Signup with duplicate phone", func(t *testing.T) {
		// Create first user
		CreateTestUser("09123456789", "password123", "User 1", models.RoleCustomer)

		payload := map[string]interface{}{
			"phone":    "09123456789",
			"password": "password123",
			"name":     "User 2",
		}
		jsonValue, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/api/v1/auth/signup", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("Signup with short password", func(t *testing.T) {
		payload := map[string]interface{}{
			"phone":    "09123456790",
			"password": "12345",
			"name":     "Test User",
		}
		jsonValue, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/api/v1/auth/signup", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestLogin(t *testing.T) {
	SetupTestEnvironment(t)
	defer CleanupTestEnvironment(t)

	// Create test user
	CreateTestUser("09123456789", "password123", "Test User", models.RoleCustomer)

	t.Run("Successful login", func(t *testing.T) {
		payload := map[string]interface{}{
			"phone":    "09123456789",
			"password": "password123",
		}
		jsonValue, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, response["success"].(bool))
		assert.NotNil(t, response["data"])
	})

	t.Run("Login with wrong password", func(t *testing.T) {
		payload := map[string]interface{}{
			"phone":    "09123456789",
			"password": "wrongpassword",
		}
		jsonValue, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Login with non-existent phone", func(t *testing.T) {
		payload := map[string]interface{}{
			"phone":    "09999999999",
			"password": "password123",
		}
		jsonValue, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestGetAuthToken(t *testing.T) {
	SetupTestEnvironment(t)
	defer CleanupTestEnvironment(t)

	// Create test user
	user, _ := CreateTestUser("09123456789", "password123", "Test User", models.RoleCustomer)

	// Login to get token
	payload := map[string]interface{}{
		"phone":    "09123456789",
		"password": "password123",
	}
	jsonValue, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	data := response["data"].(map[string]interface{})
	token := data["token"].(string)

	// Validate token
	claims, err := utils.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Phone, claims.Phone)
	assert.Equal(t, string(user.Role), claims.Role)
}

