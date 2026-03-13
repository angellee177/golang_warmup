# Start the application
run:
	go run main.go

# Run all tests
test:
	go test -v ./...

# Run coverage and generate HTML report
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out

# Tidy up go.mod and go.sum
tidy:
	go mod tidy

# Clean up coverage files
clean:
	rm coverage.out

# only run integration testing
test-integration:
	go test -v ./repository/... -run TestTaskRepository_Integration