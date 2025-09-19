# Database Schema - Entity Relationship Diagram

## Core Tables

### Users

| Column Name | Data Type | Constraints |
| --- | --- | --- |
| id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() |
| email | VARCHAR(255) | UNIQUE, NOT NULL |
| password_hash | VARCHAR(255) | NOT NULL |
| role | VARCHAR(20) | CHECK (role IN ('customer', 'business_owner', 'staff', 'admin')) |
| is_active | BOOLEAN | DEFAULT true |
| created_at | TIMESTAMP | DEFAULT NOW() |
| updated_at | TIMESTAMP | DEFAULT NOW() |

### Businesses

| Column Name | Data Type | Constraints |
| --- | --- | --- |
| id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() |
| owner_id | UUID | REFERENCES users(id) |
| name | VARCHAR(255) | NOT NULL |
| type | VARCHAR(50) | CHECK (type IN ('medical', 'beauty', 'wellness', 'fitness', 'other')) |
| description | TEXT |  |
| address | JSONB |  |
| contact_info | JSONB |  |
| is_verified | BOOLEAN | DEFAULT false |
| is_active | BOOLEAN | DEFAULT true |
| created_at | TIMESTAMP | DEFAULT NOW() |

### Services

| Column Name | Data Type | Constraints |
| --- | --- | --- |
| id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() |
| business_id | UUID | REFERENCES businesses(id) |
| name | VARCHAR(255) | NOT NULL |
| description | TEXT |  |
| duration_minutes | INTEGER | NOT NULL |
| price | DECIMAL(10,2) |  |
| currency | VARCHAR(3) | DEFAULT 'USD' |
| category | VARCHAR(100) |  |
| is_active | BOOLEAN | DEFAULT true |
| created_at | TIMESTAMP | DEFAULT NOW() |

### Bookings

| Column Name | Data Type | Constraints |
| --- | --- | --- |
| id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() |
| user_id | UUID | REFERENCES users(id) |
| service_id | UUID | REFERENCES services(id) |
| business_id | UUID | REFERENCES businesses(id) |
| staff_id | UUID | REFERENCES users(id) |
| booking_date | DATE | NOT NULL |
| start_time | TIME | NOT NULL |
| end_time | TIME | NOT NULL |
| status | VARCHAR(20) | CHECK (status IN ('pending', 'confirmed', 'completed', 'cancelled', 'no_show')) |
| notes | TEXT |  |
| created_at | TIMESTAMP | DEFAULT NOW() |
| updated_at | TIMESTAMP | DEFAULT NOW() |

Relationships
Users 1→M Businesses (owner relationship)
Businesses 1→M Services
Businesses 1→M Staff members
Users M→M Bookings (through services)
Bookings M→1 Services
