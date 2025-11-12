.PHONY: test test-verbose test-coverage test-html clean

# Run all tests
test:
	go test ./tests/... -v

# Run tests with coverage
test-coverage:
	go test ./tests/... -cover

# Run tests with detailed coverage report
test-html:
	go test ./tests/... -coverprofile=coverage.out
	go tool cover -html=coverage.out

# Run specific test file
test-auth:
	go test ./tests/auth_test.go -v

test-menu:
	go test ./tests/menu_test.go -v

test-reservation:
	go test ./tests/reservation_test.go -v

# Clean test artifacts
clean:
	rm -f coverage.out

# Install dependencies
deps:
	go mod download
	go mod tidy

