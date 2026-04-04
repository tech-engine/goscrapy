.PHONY: test test-race test-verbose cover cover-html cover-func clean

# Run all tests
test:
	@echo "Running tests..."
	@go test ./...

# Run all tests with race detector
test-race:
	@echo "Running tests with race detector..."
	@go test -race ./...

# Run all tests with verbose output
test-verbose:
	@echo "Running tests with verbose output..."
	@go test -race -v ./...

# Run tests for a specific package (usage: make test-pkg PKG=./pkg/scheduler)
test-pkg:
	@echo "Running tests for package $(PKG)..."
	@go test -race -v $(PKG)

# Generate coverage profile and print total
cover:
	@echo "Running coverage..."
	@go test -coverprofile=coverage.out ./...
	@echo ""
	@echo "=== Coverage Summary ==="
	@go tool cover -func=coverage.out | tail -1

# Generate coverage profile and print per-function breakdown
cover-func:
	@echo "Running coverage per function..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out

# Generate coverage profile and open HTML report in browser
cover-html:
	@echo "Running coverage HTML..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean generated files
clean:
	@echo "Running clean..."
	@rm -f coverage.out coverage.html
