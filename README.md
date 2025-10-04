# Studious Pancake - Universal Booking Platform

A scalable microservices-based booking system for healthcare, beauty, wellness, and service industries.

## Features

- Multi-role authentication (Customers, Business Owners, Staff, Admin)
- Real-time availability management
- Online booking with payment integration
- Multi-business type support (Medical, Beauty, Wellness, etc.)
- Review and rating system
- Notification system

## Architecture

Microservices architecture built with Go, PostgreSQL, and modern web technologies.

## Services

- Auth Service - Authentication & Authorization
- User Service - User management
- Business Service - Business/provider management
- Booking Service - Core booking engine
- Payment Service - Payment processing
- Notification Service - Email/SMS notifications

## Tech Stack

- **Backend**: Go 1.24+
- **Database**: PostgreSQL 15+
- **Cache**: Redis
- **Message Queue**: RabbitMQ
- **Frontend**: React + TypeScript
- **Mobile**: React Native

## Quick Start

```bash
# 1. Clone repository
git clone https://github.com/shahid-io/studious-pancake.git
cd studious-pancake

# 2. Setup environment
cp .env.example .env
nano .env  # Set DATABASE_URL, AUTH_SERVICE_PORT, JWT_SECRET, REDIS_URL for Go services
           # Set REDIS_HOST, REDIS_PORT, REDIS_PASSWORD for Docker Compose

# 3. Start dependencies (Postgres, Redis)
docker-compose up -d postgres redis

# 4. Sync Go workspace
go work sync

# 5. Start the auth service (Go)
cd services/auth-service
go run main.go

# The auth service will be available at:
# http://localhost:<AUTH_SERVICE_PORT>
```

## Development Workflow

### 🚀 **Best Development Commands**

#### **Quick Development (Recommended)**

```bash
cd services/auth-service

# Option 1: Hot reload with Air (recommended for active development)
make dev

# Option 2: Simple development without Air
make dev-simple

# Option 3: Just Go (simplest)
go run main.go
```

#### **Development Tools**

```bash
# Install development tools
make install-tools

# Build the application
make build

# Run tests
make test

# Run tests with coverage
make test-coverage

# Clean build artifacts
make clean

# Run linter
make lint

# View all available commands
make help
```

#### **Docker Development**

```bash
# Build Docker image
make docker-build

# Run in Docker container
make docker-run
```

### 📋 **Available Make Commands**

| Command | Description |
|---------|-------------|
| `make dev` | Start development server with Air (hot reload) |
| `make dev-simple` | Start development server with go run |
| `make build` | Build the application binary |
| `make run` | Build and run the application |
| `make test` | Run all tests |
| `make test-coverage` | Run tests with coverage report |
| `make clean` | Clean build artifacts and tmp files |
| `make deps` | Download and tidy dependencies |
| `make lint` | Run golangci-lint |
| `make docker-build` | Build Docker image |
| `make docker-run` | Run Docker container |
| `make help` | Show all available commands |

### 🛠️ **Development Setup**

#### **Prerequisites**

- Go 1.24+
- PostgreSQL 15+
- Make (for build commands)
- Air (for hot reload) - `go install github.com/cosmtrek/air@latest`

#### **Environment Variables**

```bash
# Required for auth-service
DATABASE_URL=host=localhost user=postgres password=secret dbname=mydb port=5432 sslmode=disable
AUTH_SERVICE_PORT=8080
JWT_SECRET=your-secret-key
REDIS_URL=localhost:6379
ENVIRONMENT=development
```

#### **Hot Reload Configuration**

The project uses Air for hot reloading during development:

- Configuration: `services/auth-service/.air.toml`
- Builds to: `tmp/auth-service`
- Watches: `*.go` files and dependencies
- Excludes: `tmp/`, `vendor/`, `*_test.go`

### 🧪 **Testing**

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific test
go test ./path/to/package -v

# Run tests with race detection
go test -race ./...
```

### 🏗️ **Build & Deploy**

```bash
# Build for current platform
make build

# Build for Linux (production)
make build-linux

# Cross-platform build
GOOS=windows GOARCH=amd64 go build -o auth-service.exe .
GOOS=darwin GOARCH=amd64 go build -o auth-service-mac .
```

## API Documentation

### 🔐 **Auth Service Endpoints**

#### **Public Endpoints (No Authentication Required)**

```bash
# Health Check
GET /api/v1/auth/health

# User Registration
POST /api/v1/auth/register
Content-Type: application/json
{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "first_name": "John",
  "last_name": "Doe",
  "phone": "+1234567890",
  "role": "customer"
}

# User Login
POST /api/v1/auth/login
Content-Type: application/json
{
  "email": "user@example.com",
  "password": "SecurePass123!"
}

# Refresh Token
POST /api/v1/auth/refresh
Content-Type: application/json
{
  "refresh_token": "your_refresh_token_here"
}

# Forgot Password
POST /api/v1/auth/forgot-password
Content-Type: application/json
{
  "email": "user@example.com"
}

# Reset Password
POST /api/v1/auth/reset-password
Content-Type: application/json
{
  "token": "reset_token_from_email",
  "new_password": "NewSecurePass123!",
  "confirm_password": "NewSecurePass123!"
}

# Email Verification
POST /api/v1/auth/verify-email
Content-Type: application/json
{
  "token": "verification_token_from_email"
}

# Or via GET (for email links)
GET /api/v1/auth/verify-email?token=verification_token_from_email

# Resend Verification Email
POST /api/v1/auth/resend-verification
Content-Type: application/json
{
  "email": "user@example.com"
}
```

#### **Protected Endpoints (Requires JWT Token)**

```bash
# Get User Profile
GET /api/v1/auth/profile
Authorization: Bearer your_jwt_token_here

# Change Password
POST /api/v1/auth/change-password
Authorization: Bearer your_jwt_token_here
Content-Type: application/json
{
  "current_password": "CurrentPass123!",
  "new_password": "NewSecurePass123!"
}

# Logout
POST /api/v1/auth/logout
Authorization: Bearer your_jwt_token_here
Content-Type: application/json
{
  "refresh_token": "your_refresh_token_here",
  "logout_all": false
}
```

#### **Authentication Flow**

1. **Register** → Get access token + refresh token
2. **Login** → Get access token + refresh token  
3. **Use access token** for API calls (expires in 15 minutes)
4. **Refresh token** when access token expires (expires in 7 days)
5. **Logout** to invalidate tokens

#### **Rate Limiting**

| Endpoint | Limit | Window |
|----------|-------|--------|
| Register | 5 requests | 10 minutes |
| Login | 5 requests | 15 minutes |
| Forgot Password | 3 requests | 1 hour |
| Password Reset | 5 requests | 10 minutes |
| Refresh Token | 10 requests | 5 minutes |
| Other Protected | 10 requests | 5 minutes |

## Project Structure

```text
studious-pancake/
├── docs/                          # Documentation
│   ├── API_DESIGN.md
│   ├── ARCHITECTURE.md
│   ├── ERD.md
│   └── MICROSERVICES.md
├── libs/                          # Shared libraries
│   └── domain/                    # Domain models
│       ├── auth/                  # Auth DTOs and types
│       └── user/                  # User models
├── pkg/                          # Shared packages
│   ├── config/                   # Configuration management
│   └── database/                 # Database utilities
├── services/                     # Microservices
│   ├── auth-service/             # Authentication service
│   │   ├── scripts/              # Development scripts
│   │   ├── tmp/                  # Build artifacts (gitignored)
│   │   ├── .air.toml             # Hot reload config
│   │   ├── Dockerfile            # Container config
│   │   ├── Makefile              # Build commands
│   │   └── main.go               # Service entry point
│   └── feed-service/             # Feed service (planned)
├── migrations/                   # Database migrations
├── docker-compose.yml            # Development environment
├── go.work                       # Go workspace
└── README.md                     # This file
```

## Features Implemented

### 🔐 **Auth Service (Complete)**

- ✅ **User Registration** with email/password
- ✅ **User Login** with JWT tokens
- ✅ **Refresh Token System** with rotation
- ✅ **Password Reset Flow** with email tokens
- ✅ **Email Verification** system
- ✅ **Change Password** for authenticated users
- ✅ **Logout** with session invalidation
- ✅ **Rate Limiting** on all endpoints
- ✅ **Password Strength Validation**
- ✅ **Session Management** with IP tracking
- ✅ **Activity Logging** for security events
- ✅ **Multi-device Support** with session management

### 🚧 **Planned Services**

- User Service - Extended user management
- Business Service - Business/provider management  
- Booking Service - Core booking engine
- Payment Service - Payment processing
- Notification Service - Email/SMS notifications

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Follow the development workflow (`make dev` for testing)
5. Run tests (`make test`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For support, email [support@studious-pancake.com](mailto:support@studious-pancake.com) or join our Slack channel.
