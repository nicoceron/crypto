.PHONY: test

test:
	@echo " Linting..."
	@go vet ./...
	@echo " Running tests with race detector..."
	@go test -race -v ./...
	@echo "------------------------------------------------"
	@echo " ✅ All tests passed successfully!"
	@echo "------------------------------------------------" 