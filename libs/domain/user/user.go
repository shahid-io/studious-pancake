package user

import (
	"time"

	"gorm.io/gorm"
)

// User represents the core user entity for the booking platform
type User struct {
	gorm.Model
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"not null" json:"-"` // Hidden from JSON responses
	FirstName string    `gorm:"not null" json:"first_name"`
	LastName  string    `gorm:"not null" json:"last_name"`
	Phone     string    `gorm:"size:20" json:"phone,omitempty"`
	Role      string    `gorm:"not null;default:'customer'" json:"role"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	LastLogin time.Time `json:"last_login,omitempty"`
	Timezone  string    `gorm:"default:'UTC'" json:"timezone"`
}

// UserProfile represents additional user details and preferences
type UserProfile struct {
	gorm.Model
	UserID      string    `gorm:"not null;uniqueIndex" json:"user_id"`
	AvatarURL   string    `json:"avatar_url,omitempty"`
	DateOfBirth time.Time `json:"date_of_birth,omitempty"`
	Address     string    `json:"address,omitempty"`
	City        string    `json:"city,omitempty"`
	State       string    `json:"state,omitempty"`
	ZipCode     string    `json:"zip_code,omitempty"`
	Country     string    `json:"country,omitempty"`

	// Preferences
	EmailNotifications bool `gorm:"default:true" json:"email_notifications"`
	SMSNotifications   bool `gorm:"default:false" json:"sms_notifications"`
	PushNotifications  bool `gorm:"default:true" json:"push_notifications"`

	// Service preferences
	PreferredLanguage string `gorm:"default:'en'" json:"preferred_language"`
	Currency          string `gorm:"default:'USD'" json:"currency"`
}

// UserSession represents user login sessions
type UserSession struct {
	gorm.Model
	UserID    string    `gorm:"not null;index" json:"user_id"`
	Token     string    `gorm:"not null;uniqueIndex" json:"token"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	IPAddress string    `json:"ip_address,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
}

// UserVerification represents email/phone verification status
type UserVerification struct {
	gorm.Model
	UserID              string    `gorm:"not null;uniqueIndex" json:"user_id"`
	EmailVerified       bool      `gorm:"default:false" json:"email_verified"`
	PhoneVerified       bool      `gorm:"default:false" json:"phone_verified"`
	EmailToken          string    `json:"email_token,omitempty"`
	PhoneToken          string    `json:"phone_token,omitempty"`
	PasswordResetToken  string    `json:"password_reset_token,omitempty"`
	PasswordResetExpiry time.Time `json:"password_reset_expiry,omitempty"`
	VerifiedAt          time.Time `json:"verified_at,omitempty"`
}

// UserActivity represents user activity tracking
type UserActivity struct {
	gorm.Model
	UserID    string `gorm:"not null;index" json:"user_id"`
	Activity  string `gorm:"not null" json:"activity"` // login, logout, booking_created, etc.
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
	Metadata  string `gorm:"type:json" json:"metadata,omitempty"` // Additional activity data
}

// Role constants for consistent role usage across services
const (
	RoleCustomer      = "customer"
	RoleBusinessOwner = "business_owner"
	RoleStaff         = "staff"
	RoleAdmin         = "admin"
)

// User status constants
const (
	StatusActive    = "active"
	StatusInactive  = "inactive"
	StatusSuspended = "suspended"
	StatusPending   = "pending"
)

// Helper methods for User struct

// GetFullName returns the user's full name
func (u *User) GetFullName() string {
	return u.FirstName + " " + u.LastName
}

// HasRole checks if user has a specific role
func (u *User) HasRole(role string) bool {
	return u.Role == role
}

// IsVerified checks if user has verified their email
func (uv *UserVerification) IsVerified() bool {
	return uv.EmailVerified
}

// IsSessionValid checks if session is still valid
func (us *UserSession) IsSessionValid() bool {
	return us.IsActive && time.Now().Before(us.ExpiresAt)
}

// TableName overrides the table name for UserVerification
func (UserVerification) TableName() string {
	return "user_verifications"
}

// TableName overrides the table name for UserActivity
func (UserActivity) TableName() string {
	return "user_activities"
}
