all: build test

brew install:
	@brew install go@1.26;
	@brew install docker;
	@brew install docker-compose;
	@brew install docker-buildx;
	@brew install colima;
	@brew install --cask bruno;
	@brew install air;
	@brew install libpq;

build:
	@echo "Building..."
	
	@go build -o main cmd/api/main.go

# Run the application
run:
	@go run cmd/api/main.go

# Create DB container
docker-run:
		docker-compose up --build; \

# Shutdown DB container
docker-down:
		docker-compose down; \

# Test the application
test:
	@echo "Testing..."
	@go test ./... -v

# Integrations Tests for the application
itest:
	@echo "Running integration tests..."
	@go test ./internal/database -v

# Run migrations up
migrate-up:
	@go run ./cmd/migrate up

# Run migrations down
migrate-down:
	@go run ./cmd/migrate down

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main

# Live Reload
watch:
	@if command -v air > /dev/null; then \
            air; \
            echo "Watching...";\
        else \
            read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
            if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
                go install github.com/air-verse/air@latest; \
                air; \
                echo "Watching...";\
            else \
                echo "You chose not to install air. Exiting..."; \
                exit 1; \
            fi; \
        fi
