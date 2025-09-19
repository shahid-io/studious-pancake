package auth

import (
	"time"

	"github.com/shahid-io/studious-pancake/libs/domain/user"
)

// LoginRequest represents user login credentials
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// RegisterRequest represents user registration data
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Phone     string `json:"phone,omitempty"`
	Role      string `json:"role" binding:"required,oneof=customer business_owner staff admin"`
}

// LoginResponse represents successful login response
type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type"` // Bearer
	ExpiresIn    int64     `json:"expires_in"` // Seconds until expiration
	ExpiresAt    time.Time `json:"expires_at"`
	User         user.User `json:"user"`
}

// RefreshTokenRequest represents refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ForgotPasswordRequest represents password reset request
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest represents password reset confirmation
type ResetPasswordRequest struct {
	Token           string `json:"token" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" binding:"required,min=8"`
}

// ChangePasswordRequest represents password change for authenticated users
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// AuthError represents authentication error response
type AuthError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

// AuthError codes
const (
	ErrorInvalidCredentials = "invalid_credentials"
	ErrorUserNotFound       = "user_not_found"
	ErrorUserInactive       = "user_inactive"
	ErrorEmailNotVerified   = "email_not_verified"
	ErrorTokenExpired       = "token_expired"
	ErrorTokenInvalid       = "token_invalid"
	ErrorPasswordMismatch   = "password_mismatch"
)
