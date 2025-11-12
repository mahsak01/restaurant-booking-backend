# Tests

This directory contains comprehensive tests for the Restaurant Booking Backend API.

## Test Structure

- `setup_test.go` - Test setup and helper functions
- `auth_test.go` - Authentication tests (signup, login)
- `menu_test.go` - Menu management tests
- `table_test.go` - Table management tests
- `reservation_test.go` - Reservation tests
- `notification_test.go` - Notification tests
- `user_test.go` - User management tests
- `health_test.go` - Health check tests

## Running Tests

### Run all tests:
```bash
go test ./tests/... -v
```

### Run specific test file:
```bash
go test ./tests/auth_test.go -v
```

### Run with coverage:
```bash
go test ./tests/... -cover
```

### Run with detailed coverage report:
```bash
go test ./tests/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Test Environment

- Uses in-memory SQLite database for fast testing
- Each test sets up and tears down its own environment
- JWT secret is set to "test-secret-key" for testing
- Gin is set to test mode

## Test Coverage

Tests cover:
- ✅ Authentication (signup, login, token validation)
- ✅ Menu CRUD operations
- ✅ Table management
- ✅ Reservation creation and cancellation
- ✅ Notification management
- ✅ User management (admin only)
- ✅ Authorization and role-based access
- ✅ Input validation
- ✅ Error handling

