# Setup Guide

## Quick Setup (5 Minutes)

### Prerequisites

- **Go 1.24+** - [Download](https://golang.org/dl/)
- **Docker & Docker Compose** - [Install Guide](https://docs.docker.com/get-docker/)
- **Git** - Pre-installed on most systems

### 1. Clone Repository

```bash
git clone https://github.com/shahid-io/studious-pancake.git
cd studious-pancake
```

### 2. Environment Setup

```bash
# Copy environment template
cp .env.example .env

# Edit with your values
nano .env  # or use your favorite editor
```

### 3. Start Dependencies

```bash
# Start PostgreSQL, Redis, and other services
docker-compose up -d postgres redis
```

### 4. Setup Go Workspace

```bash
# Initialize and sync Go workspace
go work sync
```

### 5. Run Your First Service

```bash
# Start the auth service
cd services/auth-service
go run main.go
```

🎉 **You're running!** Access: <http://localhost:8080>

---

## Detailed Setup

### Environment Configuration (.env)

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_secure_password
DB_NAME=studious_pancake

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# JWT Authentication
JWT_SECRET=your-super-secure-jwt-secret-change-in-production
JWT_EXPIRY=24h

# Service Ports
AUTH_SERVICE_PORT=8080
USER_SERVICE_PORT=8081
BUSINESS_SERVICE_PORT=8082
BOOKING_SERVICE_PORT=8083
PAYMENT_SERVICE_PORT=8084
NOTIFICATION_SERVICE_PORT=8085

# Email (for notifications)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password

# Payment Gateway (Stripe)
STRIPE_SECRET_KEY=sk_test_your_stripe_key
STRIPE_WEBHOOK_SECRET=whsec_your_webhook_secret

# External APIs
GOOGLE_MAPS_API_KEY=your_google_maps_key
```

### Database Setup

```bash
# Run database migrations
go run cmd/migrate/main.go

# Or using migrate tool directly
migrate -path ./migrations -database "postgres://postgres:password@localhost:5432/studious_pancake?sslmode=disable" up
```

### Service-Specific Setup

#### Auth Service

```bash
cd services/auth-service
go mod tidy
go run main.go
```

#### User Service

```bash
cd services/user-service
go mod tidy
go run main.go
```

#### Business Service

```bash
cd services/business-service  
go mod tidy
go run main.go
```

### Docker Development

```bash
# Start all services with hot reload
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down

# Rebuild images
docker-compose build
```

### Development with Air (Hot Reload)

```bash
# Install air if not already installed
go install github.com/air-verse/air@latest

# Run with hot reload
cd services/auth-service
air
```

---

## Development Tools

### Recommended IDE Setup

- **VS Code** with Go extension
- **GoLand** (JetBrains)
- **Postman** for API testing
- **TablePlus** for database management

### VS Code Extensions

```json
{
  "recommendations": [
    "golang.go",
    "ms-azuretools.vscode-docker",
    "humao.rest-client",
    "bungcip.better-toml",
    "mongodb.mongodb-vscode"
  ]
}
```

### Database Management

```bash
# Connect to PostgreSQL
psql -h localhost -U postgres -d studious_pancake

# Or use Docker
docker exec -it studious-pancake-postgres psql -U postgres -d studious_pancake
```

### API Testing

Create `api-test.http` file:

```http
### Register User
POST http://localhost:8080/auth/register
Content-Type: application/json

{
  "email": "test@example.com",
  "password": "password123",
  "first_name": "Test",
  "last_name": "User",
  "role": "customer"
}

### Login
POST http://localhost:8080/auth/login
Content-Type: application/json

{
  "email": "test@example.com",
  "password": "password123"
}
```

---

## Testing Setup

### Run Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test ./... -cover

# Run specific service tests
cd services/auth-service
go test -v ./...

# Run integration tests
go test -tags=integration ./...
```

### Test Database

```bash
# Use test database
export DB_NAME=studious_pancake_test
go test ./...
```

---

## Troubleshooting

### Common Issues

**Port already in use**

```bash
# Find process using port
lsof -i :8080

# Kill process
kill -9 <PID>
```

**Database connection issues**

```bash
# Check if PostgreSQL is running
docker ps

# Check logs
docker logs studious-pancake-postgres

# Reset database
docker-compose down -v
docker-compose up -d postgres
```

**Go module issues**

```bash
# Clean module cache
go clean -modcache

# Sync workspace
go work sync

# Tidy modules
go mod tidy
```

**Docker issues**

```bash
# Restart Docker
sudo systemctl restart docker

# Clean up containers
docker system prune -a
```

### Health Checks

```bash
# Check PostgreSQL
curl http://localhost:5432

# Check Redis
redis-cli ping

# Check service health
curl http://localhost:8080/health
```

---

## Production Setup

### Environment Variables for Production

```bash
# Use environment-specific files
cp .env.production.example .env.production

# Set production values
export NODE_ENV=production
export DB_HOST=production-db.example.com
export JWT_SECRET=your-production-jwt-secret
```

### Docker Production

```bash
# Build production images
docker-compose -f docker-compose.prod.yml build

# Deploy
docker-compose -f docker-compose.prod.yml up -d
```

### Kubernetes (Optional)

```bash
# Apply Kubernetes configurations
kubectl apply -f deployments/kubernetes/

# Check status
kubectl get pods
kubectl get services
```

---

## Update Instructions

### Pull Latest Changes

```bash
git pull origin main
go work sync
go mod tidy
docker-compose build
```

### Database Migrations

```bash
# Run new migrations
go run cmd/migrate/main.go

# Or check migration status
migrate -path ./migrations -database "$DATABASE_URL" version
```

---

## Need Help?

1. **Check Logs**: `docker-compose logs -f`
2. **Verify Environment**: `cat .env | grep DB_`
3. **Test Connections**: Use health check endpoints
4. **Check Issues**: [GitHub Issues](https://github.com/shahid-io/studious-pancake/issues)
5. **Ask Community**: [Discussions](https://github.com/shahid-io/studious-pancake/discussions)

---

## Verification Checklist

- [ ] Docker containers running (`docker ps`)
- [ ] Database accessible (`psql -h localhost -U postgres`)
- [ ] Go workspace synced (`go work sync`)
- [ ] Environment variables set (`cat .env`)
- [ ] Services starting without errors
- [ ] API endpoints responding (`curl http://localhost:8080/health`)

Your setup is complete! 🎊 Now you can start developing your booking platform.
