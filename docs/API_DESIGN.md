# API Design Specification

## Base URL

`https://api.studious-pancake.com/v1`

## Authentication

Bearer Token authentication required for all endpoints except:

- User registration
- User login
- Public business listings
- Public service listings

## Core Endpoints

### Auth Service

```http
POST /auth/register
Content-Type: application/json
{
  "email": "user@example.com",
  "password": "securepassword123",
  "first_name": "John",
  "last_name": "Doe",
  "phone": "+1234567890",
  "role": "customer" // customer, business_owner, staff, admin
}

POST /auth/login
Content-Type: application/json
{
  "email": "user@example.com",
  "password": "securepassword123"
}

GET /auth/profile
Authorization: Bearer <token>

POST /auth/refresh
Authorization: Bearer <token>

POST /auth/logout
Authorization: Bearer <token>
```

### User Service

```http
GET /users/{user_id}
Authorization: Bearer <token>

PUT /users/{user_id}
Authorization: Bearer <token>
Content-Type: application/json
{
  "first_name": "John",
  "last_name": "Updated",
  "phone": "+1234567890",
  "preferences": {
    "notifications": true,
    "marketing_emails": false
  }
}

GET /users?role=business_owner&page=1&limit=20
Authorization: Bearer <token>
```

### Business Service

```http
POST /businesses
Authorization: Bearer <token>
Content-Type: application/json
{
  "name": "City Medical Center",
  "type": "medical", // medical, beauty, wellness, fitness, other
  "description": "Premium healthcare services",
  "address": {
    "street": "123 Main St",
    "city": "New York",
    "state": "NY",
    "zip_code": "10001",
    "country": "USA"
  },
  "contact_info": {
    "phone": "+1234567890",
    "email": "contact@citymedical.com"
  }
}

GET /businesses/{business_id}
GET /businesses?type=medical&location=ny&page=1&limit=10

PUT /businesses/{business_id}
Authorization: Bearer <token>

DELETE /businesses/{business_id}
Authorization: Bearer <token>
```

### Service Catalog

```http
POST /businesses/{business_id}/services
Authorization: Bearer <token>
Content-Type: application/json
{
  "name": "Haircut & Styling",
  "description": "Professional haircut with styling",
  "duration_minutes": 60,
  "price": 45.00,
  "currency": "USD",
  "category": "hair"
}

GET /services?business_id=xxx&category=hair
GET /services/{service_id}

PUT /services/{service_id}
Authorization: Bearer <token>

DELETE /services/{service_id}
Authorization: Bearer <token>
```

### Availability Service

```http
GET /availability?business_id=xxx&service_id=yyy&date=2024-01-15
Authorization: Bearer <token>

POST /availability
Authorization: Bearer <token>
Content-Type: application/json
{
  "staff_id": "uuid",
  "business_id": "uuid",
  "date": "2024-01-15",
  "slots": [
    {
      "start_time": "09:00:00",
      "end_time": "10:00:00",
      "is_available": true
    },
    {
      "start_time": "10:00:00", 
      "end_time": "11:00:00",
      "is_available": true
    }
  ]
}

PUT /availability/{availability_id}
Authorization: Bearer <token>
```

### Booking Service

```http
POST /bookings
Authorization: Bearer <token>
Content-Type: application/json
{
  "service_id": "uuid",
  "business_id": "uuid",
  "staff_id": "uuid", // optional
  "booking_date": "2024-01-15",
  "start_time": "14:00:00",
  "notes": "First-time customer, sensitive scalp"
}

GET /bookings?user_id=xxx&status=confirmed&page=1&limit=10
Authorization: Bearer <token>

GET /bookings/{booking_id}
Authorization: Bearer <token>

PUT /bookings/{booking_id}/cancel
Authorization: Bearer <token>

PUT /bookings/{booking_id}/reschedule
Authorization: Bearer <token>
Content-Type: application/json
{
  "new_date": "2024-01-16",
  "new_time": "15:00:00"
}
```

### Payment Service

```http
POST /payments/intent
Authorization: Bearer <token>
Content-Type: application/json
{
  "booking_id": "uuid",
  "amount": 100.00,
  "currency": "USD",
  "payment_method": "card" // card, paypal, etc.
}

POST /payments/confirm
Authorization: Bearer <token>
Content-Type: application/json
{
  "payment_intent_id": "pi_xxx",
  "booking_id": "uuid"
}

POST /payments/refund
Authorization: Bearer <token>
Content-Type: application/json
{
  "payment_id": "uuid",
  "amount": 100.00,
  "reason": "Customer cancellation"
}

GET /payments?booking_id=xxx
Authorization: Bearer <token>
```

### Review Service

```http
POST /reviews
Authorization: Bearer <token>
Content-Type: application/json
{
  "booking_id": "uuid",
  "rating": 5,
  "comment": "Excellent service! Very professional.",
  "review_type": "business" // business, staff, service
}

GET /reviews?business_id=xxx&rating=4&page=1&limit=10
GET /reviews/{review_id}

PUT /reviews/{review_id}
Authorization: Bearer <token>

DELETE /reviews/{review_id}
Authorization: Bearer <token>
```

### Notification Service

```http
POST /notifications
Authorization: Bearer <token>
Content-Type: application/json
{
  "type": "booking_confirmation", // booking_reminder, payment_receipt, etc.
  "user_id": "uuid",
  "booking_id": "uuid",
  "channel": "email" // email, sms, push
}

GET /notifications?user_id=xxx&read=false
Authorization: Bearer <token>

PUT /notifications/{notification_id}/read
Authorization: Bearer <token>
```

## Response Format

```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "name": "Service Name",
    "created_at": "2024-01-15T10:30:00Z"
  },
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "total_pages": 5
  },
  "message": "Operation successful",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## Error Format

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input data",
    "details": [
      {
        "field": "email",
        "message": "Email is required"
      },
      {
        "field": "password", 
        "message": "Password must be at least 8 characters"
      }
    ]
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## Common Error Codes

- `400` - Bad Request (validation errors)
- `401` - Unauthorized (invalid token)
- `403` - Forbidden (insufficient permissions)
- `404` - Not Found (resource doesn't exist)
- `409` - Conflict (resource already exists)
- `429` - Too Many Requests (rate limiting)
- `500` - Internal Server Error

## Rate Limiting

- **Public endpoints**: 100 requests/minute per IP
- **Authenticated endpoints**: 1000 requests/minute per user
- **Payment endpoints**: 50 requests/minute per user

## Pagination

All list endpoints support pagination:

```http
GET /bookings?page=2&limit=20&sort=created_at&order=desc
```

## Filtering

Most list endpoints support filtering:

```http
GET /businesses?type=medical&location=ny&min_rating=4
GET /bookings?status=confirmed&from_date=2024-01-01&to_date=2024-01-31
```

## Versioning

API versioned through URL path:

- `https://api.studious-pancake.com/v1/booking`
- `https://api.studious-pancake.com/v2/booking` (future)

This API design provides a comprehensive foundation for your booking platform with clear endpoints for all core functionality!
