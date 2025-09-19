# Studious Pancake - Universal Booking Platform

A scalable microservices-based booking system for healthcare, beauty, wellness, and service industries.

## 🚀 Features

- Multi-role authentication (Customers, Business Owners, Staff, Admin)
- Real-time availability management
- Online booking with payment integration
- Multi-business type support (Medical, Beauty, Wellness, etc.)
- Review and rating system
- Notification system

## 🏗️ Architecture

Microservices architecture built with Go, PostgreSQL, and modern web technologies.

## 📦 Services

- Auth Service - Authentication & Authorization
- User Service - User management
- Business Service - Business/provider management
- Booking Service - Core booking engine
- Payment Service - Payment processing
- Notification Service - Email/SMS notifications

## 🛠️ Tech Stack

- **Backend**: Go 1.24+
- **Database**: PostgreSQL 15+
- **Cache**: Redis
- **Message Queue**: RabbitMQ
- **Frontend**: React + TypeScript
- **Mobile**: React Native

## 🚦 Quick Start

```bash
# Clone repository
git clone https://github.com/shahid-io/studious-pancake.git

# Setup environment
cp .env.example .env

# Start services
docker-compose up -d

# Run application
go work sync
cd services/auth-service
go run main.go
