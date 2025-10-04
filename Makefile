.PHONY: help auth-dev auth-build auth-test auth-clean all-dev all-build all-test all-clean docker-up docker-down

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
NC=\033[0m # No Color

# Default target
help:
	@echo "$(BLUE)Studious Pancake - Universal Booking Platform$(NC)"
	@echo "$(YELLOW)Available commands:$(NC)"
	@echo ""
	@echo "$(GREEN)🚀 Development Commands:$(NC)"
	@echo "  make auth-dev        - Start auth service with hot reload"
	@echo "  make auth-dev-simple - Start auth service with go run"
	@echo "  make all-dev         - Start all services in development mode"
	@echo ""
	@echo "$(GREEN)📦 Build Commands:$(NC)"
	@echo "  make auth-build      - Build auth service"
	@echo "  make all-build       - Build all services"
	@echo ""
	@echo "$(GREEN)🧪 Test Commands:$(NC)"
	@echo "  make auth-test       - Run auth service tests"
	@echo "  make all-test        - Run all tests"
	@echo ""
	@echo "$(GREEN)🧹 Clean Commands:$(NC)"
	@echo "  make auth-clean      - Clean auth service artifacts"
	@echo "  make all-clean       - Clean all build artifacts"
	@echo ""
	@echo "$(GREEN)🐳 Docker Commands:$(NC)"
	@echo "  make docker-up       - Start development environment"
	@echo "  make docker-down     - Stop development environment"
	@echo ""
	@echo "$(GREEN)🔧 Service-Specific Commands:$(NC)"
	@echo "  cd services/auth-service && make help    - See auth service commands"
	@echo ""

# Auth Service Commands
auth-dev:
	@echo "$(GREEN)🚀 Starting auth service with hot reload...$(NC)"
	@cd services/auth-service && make dev

auth-dev-simple:
	@echo "$(GREEN)🚀 Starting auth service (simple mode)...$(NC)"
	@cd services/auth-service && make dev-simple

auth-build:
	@echo "$(GREEN)📦 Building auth service...$(NC)"
	@cd services/auth-service && make build

auth-test:
	@echo "$(GREEN)🧪 Running auth service tests...$(NC)"
	@cd services/auth-service && make test

auth-clean:
	@echo "$(GREEN)🧹 Cleaning auth service artifacts...$(NC)"
	@cd services/auth-service && make clean

auth-lint:
	@echo "$(GREEN)🔍 Linting auth service...$(NC)"
	@cd services/auth-service && make lint

# All Services Commands
all-dev:
	@echo "$(GREEN)🚀 Starting all services in development mode...$(NC)"
	@echo "$(YELLOW)Currently only auth-service is implemented$(NC)"
	@make auth-dev

all-build:
	@echo "$(GREEN)📦 Building all services...$(NC)"
	@make auth-build
	@echo "$(YELLOW)✅ All services built successfully$(NC)"

all-test:
	@echo "$(GREEN)🧪 Running all tests...$(NC)"
	@make auth-test
	@echo "$(YELLOW)✅ All tests completed$(NC)"

all-clean:
	@echo "$(GREEN)🧹 Cleaning all build artifacts...$(NC)"
	@make auth-clean
	@rm -rf tmp/
	@echo "$(YELLOW)✅ All artifacts cleaned$(NC)"

# Docker Commands
docker-up:
	@echo "$(GREEN)🐳 Starting development environment...$(NC)"
	@docker-compose up -d postgres redis
	@echo "$(YELLOW)✅ Development environment started$(NC)"
	@echo "$(BLUE)PostgreSQL: localhost:5432$(NC)"
	@echo "$(BLUE)Redis: localhost:6379$(NC)"

docker-down:
	@echo "$(GREEN)🐳 Stopping development environment...$(NC)"
	@docker-compose down
	@echo "$(YELLOW)✅ Development environment stopped$(NC)"

# Go workspace commands
workspace-sync:
	@echo "$(GREEN)🔧 Syncing Go workspace...$(NC)"
	@go work sync
	@echo "$(YELLOW)✅ Workspace synced$(NC)"

workspace-init:
	@echo "$(GREEN)🔧 Initializing Go workspace...$(NC)"
	@go work init . ./libs/domain ./pkg ./services/auth-service
	@echo "$(YELLOW)✅ Workspace initialized$(NC)"

# Dependencies
deps:
	@echo "$(GREEN)📦 Downloading dependencies for all services...$(NC)"
	@go mod download
	@cd libs/domain && go mod download
	@cd pkg && go mod download
	@cd services/auth-service && go mod download
	@echo "$(YELLOW)✅ All dependencies downloaded$(NC)"

deps-tidy:
	@echo "$(GREEN)📦 Tidying dependencies for all services...$(NC)"
	@go mod tidy
	@cd libs/domain && go mod tidy
	@cd pkg && go mod tidy
	@cd services/auth-service && go mod tidy
	@echo "$(YELLOW)✅ All dependencies tidied$(NC)"

# Install development tools
install-tools:
	@echo "$(GREEN)🔧 Installing development tools...$(NC)"
	@go install github.com/cosmtrek/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "$(YELLOW)✅ Development tools installed$(NC)"

# Quick setup for new developers
setup:
	@echo "$(BLUE)🏗️  Setting up Studious Pancake development environment...$(NC)"
	@make install-tools
	@make workspace-sync
	@make deps
	@make docker-up
	@echo ""
	@echo "$(GREEN)✅ Setup complete! You can now run:$(NC)"
	@echo "$(YELLOW)  make auth-dev        # Start auth service$(NC)"
	@echo "$(YELLOW)  make help           # See all commands$(NC)"