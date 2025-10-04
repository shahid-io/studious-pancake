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
