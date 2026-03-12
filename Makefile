# Run all tests
test:
	go test -v ./...

# Run coverage and generate HTML report
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out

# Clean up coverage files
clean:
	rm coverage.out