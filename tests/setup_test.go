package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"restaurant-booking-backend/config"
	"restaurant-booking-backend/models"
	"restaurant-booking-backend/routes"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var testDB *gorm.DB
var testRouter *gin.Engine

// SetupTestDB sets up a test database
func SetupTestDB(t *testing.T) *gorm.DB {
	// Use in-memory SQLite for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto migrate all models
	err = db.AutoMigrate(
		&models.User{},
		&models.Table{},
		&models.MenuItem{},
		&models.Reservation{},
		&models.Notification{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// SetupTestRouter sets up a test router
func SetupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	routes.SetupRoutes(router)
	return router
}

// SetupTestEnvironment sets up the test environment
func SetupTestEnvironment(t *testing.T) {
	// Set test environment variables
	os.Setenv("JWT_SECRET", "test-secret-key")
	os.Setenv("GIN_MODE", "test")

	// Setup test database
	testDB = SetupTestDB(t)
	config.DB = testDB

	// Setup test router
	testRouter = SetupTestRouter()
}

// CleanupTestEnvironment cleans up the test environment
func CleanupTestEnvironment(t *testing.T) {
	if testDB != nil {
		sqlDB, _ := testDB.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
}

// CreateTestUser creates a test user
func CreateTestUser(phone, password, name string, role models.UserRole) (*models.User, error) {
	user := models.User{
		Phone:    phone,
		Password: password,
		Name:     name,
		Role:     role,
	}
	err := testDB.Create(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateTestTable creates a test table
func CreateTestTable(number, capacity int, location string, status models.TableStatus) (*models.Table, error) {
	table := models.Table{
		Number:   number,
		Capacity: capacity,
		Location: location,
		Status:   status,
	}
	err := testDB.Create(&table).Error
	if err != nil {
		return nil, err
	}
	return &table, nil
}

// CreateTestMenuItem creates a test menu item
func CreateTestMenuItem(name, description string, price float64, category models.MenuCategory) (*models.MenuItem, error) {
	item := models.MenuItem{
		Name:        name,
		Description: description,
		Price:       price,
		Category:    category,
	}
	err := testDB.Create(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// getAuthToken gets authentication token for a user
func getAuthToken(t *testing.T, phone, password string) string {
	payload := map[string]interface{}{
		"phone":    phone,
		"password": password,
	}
	jsonValue, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	data := response["data"].(map[string]interface{})
	return data["token"].(string)
}

