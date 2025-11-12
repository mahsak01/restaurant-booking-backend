package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"restaurant-booking-backend/models"

	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	SetupTestEnvironment(t)
	defer CleanupTestEnvironment(t)

	t.Run("Access protected route without token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/profile", nil)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Access protected route with invalid token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/profile", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Access protected route with valid token", func(t *testing.T) {
		user, _ := CreateTestUser("09123456789", "password123", "Test User", models.RoleCustomer)
		userToken := getAuthToken(t, "09123456789", "password123")

		req, _ := http.NewRequest("GET", "/api/v1/profile", nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestRoleMiddleware(t *testing.T) {
	SetupTestEnvironment(t)
	defer CleanupTestEnvironment(t)

	// Create admin and customer users
	admin, _ := CreateTestUser("09111111111", "password123", "Admin", models.RoleAdmin)
	customer, _ := CreateTestUser("09222222222", "password123", "Customer", models.RoleCustomer)

	adminToken := getAuthToken(t, "09111111111", "password123")
	customerToken := getAuthToken(t, "09222222222", "password123")

	t.Run("Admin can access admin routes", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/admin/users", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Customer cannot access admin routes", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/admin/users", nil)
		req.Header.Set("Authorization", "Bearer "+customerToken)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Customer can access customer routes", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/reservations", nil)
		req.Header.Set("Authorization", "Bearer "+customerToken)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

