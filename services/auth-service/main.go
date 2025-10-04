package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/shahid-io/studious-pancake/libs/domain/auth"
	"github.com/shahid-io/studious-pancake/libs/domain/user"
	"github.com/shahid-io/studious-pancake/pkg/config"
	"github.com/shahid-io/studious-pancake/pkg/database"
)

var (
	db  *gorm.DB
	cfg *config.Config
)

// Rate limiting structures
type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
	}
}

func (rl *RateLimiter) IsAllowed(ip string, limit int, window time.Duration) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	windowStart := now.Add(-window)

	// Clean old requests
	if times, exists := rl.requests[ip]; exists {
		var validTimes []time.Time
		for _, t := range times {
			if t.After(windowStart) {
				validTimes = append(validTimes, t)
			}
		}
		rl.requests[ip] = validTimes
	}

	// Check if limit exceeded
	if len(rl.requests[ip]) >= limit {
		return false
	}

	// Add current request
	rl.requests[ip] = append(rl.requests[ip], now)
	return true
}

var authRateLimiter = NewRateLimiter()

func main() {
	// Load configuration
	cfg = config.Load()

	// Connect to database with retry
	db = database.Connect(cfg.DatabaseURL)

	// Auto-migrate models
	if err := db.AutoMigrate(
		&user.User{},
		&user.UserProfile{},
		&user.UserSession{},
		&user.UserVerification{},
		&user.UserActivity{},
	); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize Gin router
	router := gin.Default()
	router.GET("/", getPumpkin)
	// Middleware
	router.Use(CORSMiddleware())
	router.Use(LoggerMiddleware())

	// Public routes
	public := router.Group("/api/v1/auth")
	{
		public.POST("/register", RateLimitMiddleware(5, time.Minute*10), registerHandler)
		public.POST("/login", RateLimitMiddleware(5, time.Minute*15), loginHandler)
		public.POST("/refresh", RateLimitMiddleware(10, time.Minute*5), refreshTokenHandler)
		public.POST("/forgot-password", RateLimitMiddleware(3, time.Hour), forgotPasswordHandler)
		public.POST("/reset-password", RateLimitMiddleware(5, time.Minute*10), resetPasswordHandler)
		public.POST("/verify-email", RateLimitMiddleware(10, time.Minute*5), verifyEmailHandler)
		public.GET("/verify-email", RateLimitMiddleware(10, time.Minute*5), verifyEmailHandler) // Allow GET for email links
		public.POST("/resend-verification", RateLimitMiddleware(3, time.Minute*10), resendVerificationHandler)
		public.GET("/health", healthHandler)
	}

	// Protected routes (require authentication)
	protected := router.Group("/api/v1/auth")
	protected.Use(AuthMiddleware())
	{
		protected.GET("/profile", profileHandler)
		protected.POST("/logout", RateLimitMiddleware(10, time.Minute*5), logoutHandler)
		protected.POST("/change-password", RateLimitMiddleware(3, time.Minute*10), changePasswordHandler)
	}

	// Start HTTP server
	addr := ":" + cfg.AppPort
	log.Printf("Auth-Service running at http://localhost%s", addr)
	log.Printf("API Documentation: http://localhost%s/api/v1/auth/health", addr)

	if err := router.Run(addr); err != nil {
		log.Fatal("HTTP server failed:", err)
	}
}

// Test route public route
func getPumpkin(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Hey, Welcome to Pumpkin Auth Service",
	})
}

// Standardized response helpers
func sendErrorResponse(c *gin.Context, statusCode int, errorMessage string, details ...string) {
	response := gin.H{
		"success": false,
		"error":   errorMessage,
	}
	if len(details) > 0 && details[0] != "" {
		response["details"] = details[0]
	}
	c.JSON(statusCode, response)
}

func sendSuccessResponse(c *gin.Context, statusCode int, data interface{}, message string) {
	response := gin.H{
		"success": true,
		"message": message,
	}
	if data != nil {
		response["data"] = data
	}
	c.JSON(statusCode, response)
}

// Security helper to check if account is locked
func isAccountLocked(userID uint) bool {
	// TODO: Implement account lockout logic based on failed attempts
	// For now, return false (no lockout implemented)
	return false
}

// Helper to log failed login attempts
func logFailedAttempt(email, clientIP string) {
	// TODO: Implement failed attempt tracking
	log.Printf("Failed login attempt for email: %s from IP: %s", email, clientIP)
}

// Password strength validation
func validatePasswordStrength(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case 'a' <= char && char <= 'z':
			hasLower = true
		case '0' <= char && char <= '9':
			hasDigit = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

// Handlers
func registerHandler(c *gin.Context) {
	var req auth.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sendErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	// Validate password strength
	if err := validatePasswordStrength(req.Password); err != nil {
		sendErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Check if user already exists
	var existingUser user.User
	if err := db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		sendErrorResponse(c, http.StatusConflict, "User already exists with this email")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		sendErrorResponse(c, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// Create user
	newUser := user.User{
		Email:     req.Email,
		Password:  string(hashedPassword),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		Role:      req.Role,
		IsActive:  true,
	}

	if err := db.Create(&newUser).Error; err != nil {
		sendErrorResponse(c, http.StatusInternalServerError, "Failed to create user", err.Error())
		return
	}

	// Create user profile
	userProfile := user.UserProfile{
		UserID:             fmt.Sprintf("%d", newUser.ID),
		EmailNotifications: true,
		PushNotifications:  true,
		PreferredLanguage:  "en",
		Currency:           "USD",
	}
	db.Create(&userProfile)

	// Create verification record
	verification := user.UserVerification{
		UserID:     fmt.Sprintf("%d", newUser.ID),
		EmailToken: generateRandomToken(),
		PhoneToken: generateRandomToken(),
	}
	db.Create(&verification)

	// Generate JWT token
	token, expiresAt, err := generateJWTToken(newUser)
	if err != nil {
		sendErrorResponse(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// Generate refresh token and create session
	refreshSession, err := generateRefreshToken(newUser.ID, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		sendErrorResponse(c, http.StatusInternalServerError, "Failed to generate refresh token")
		return
	}

	// Log activity
	logUserActivity(newUser.ID, "register", c)

	// TODO: Send verification email
	log.Printf("Email verification token for %s: %s", newUser.Email, verification.EmailToken)

	loginResponse := auth.LoginResponse{
		AccessToken:  token,
		RefreshToken: refreshSession.Token,
		TokenType:    "Bearer",
		ExpiresIn:    int64(time.Until(expiresAt).Seconds()),
		ExpiresAt:    expiresAt,
		User:         newUser,
	}

	sendSuccessResponse(c, http.StatusCreated, loginResponse, "User registered successfully")
}

func loginHandler(c *gin.Context) {
	var req auth.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sendErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	// Find user by email
	var foundUser user.User
	if err := db.Where("email = ?", req.Email).First(&foundUser).Error; err != nil {
		logFailedAttempt(req.Email, c.ClientIP())
		sendErrorResponse(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Check if user is active
	if !foundUser.IsActive {
		sendErrorResponse(c, http.StatusUnauthorized, "Account is deactivated")
		return
	}

	// Check if account is locked
	if isAccountLocked(foundUser.ID) {
		sendErrorResponse(c, http.StatusUnauthorized, "Account is temporarily locked due to too many failed attempts")
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(req.Password)); err != nil {
		logFailedAttempt(req.Email, c.ClientIP())
		sendErrorResponse(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Generate JWT token
	token, expiresAt, err := generateJWTToken(foundUser)
	if err != nil {
		sendErrorResponse(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// Generate refresh token and create session
	refreshSession, err := generateRefreshToken(foundUser.ID, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		sendErrorResponse(c, http.StatusInternalServerError, "Failed to generate refresh token")
		return
	}

	// Update last login
	foundUser.LastLogin = time.Now()
	db.Save(&foundUser)

	// Log activity
	logUserActivity(foundUser.ID, "login", c)

	loginResponse := auth.LoginResponse{
		AccessToken:  token,
		RefreshToken: refreshSession.Token,
		TokenType:    "Bearer",
		ExpiresIn:    int64(time.Until(expiresAt).Seconds()),
		ExpiresAt:    expiresAt,
		User:         foundUser,
	}

	sendSuccessResponse(c, http.StatusOK, loginResponse, "Login successful")
}

func profileHandler(c *gin.Context) {
	userID, _ := c.Get("userID")

	var user user.User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    user,
	})
}

func refreshTokenHandler(c *gin.Context) {
	var req auth.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	// Find the session with the refresh token
	var session user.UserSession
	if err := db.Where("token = ? AND is_active = ?", req.RefreshToken, true).First(&session).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Invalid refresh token",
		})
		return
	}

	// Check if session is expired
	if !session.IsSessionValid() {
		// Deactivate expired session
		session.IsActive = false
		db.Save(&session)

		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Refresh token expired",
		})
		return
	}

	// Find the user
	var user user.User
	if err := db.First(&user, session.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "User not found",
		})
		return
	}

	// Check if user is still active
	if !user.IsActive {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Account is deactivated",
		})
		return
	}

	// Generate new access token
	accessToken, expiresAt, err := generateJWTToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to generate access token",
		})
		return
	}

	// Rotate refresh token (invalidate old, create new)
	session.IsActive = false
	db.Save(&session)

	newRefreshSession, err := generateRefreshToken(user.ID, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to generate new refresh token",
		})
		return
	}

	// Log activity
	logUserActivity(user.ID, "token_refresh", c)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": auth.LoginResponse{
			AccessToken:  accessToken,
			RefreshToken: newRefreshSession.Token,
			TokenType:    "Bearer",
			ExpiresIn:    int64(time.Until(expiresAt).Seconds()),
			ExpiresAt:    expiresAt,
			User:         user,
		},
		"message": "Token refreshed successfully",
	})
}

func logoutHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	// Get refresh token from request body (optional)
	var req struct {
		RefreshToken string `json:"refresh_token,omitempty"`
		LogoutAll    bool   `json:"logout_all,omitempty"` // Option to logout from all devices
	}
	c.ShouldBindJSON(&req)

	if req.LogoutAll {
		// Invalidate all sessions for this user
		result := db.Model(&user.UserSession{}).
			Where("user_id = ? AND is_active = ?", fmt.Sprintf("%v", userID), true).
			Update("is_active", false)

		log.Printf("Invalidated %d sessions for user %v", result.RowsAffected, userID)
	} else if req.RefreshToken != "" {
		// Invalidate specific session
		db.Model(&user.UserSession{}).
			Where("token = ? AND user_id = ?", req.RefreshToken, fmt.Sprintf("%v", userID)).
			Update("is_active", false)
	} else {
		// If no refresh token provided, try to invalidate sessions from this IP/User-Agent
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		db.Model(&user.UserSession{}).
			Where("user_id = ? AND ip_address = ? AND user_agent = ? AND is_active = ?",
				fmt.Sprintf("%v", userID), clientIP, userAgent, true).
			Update("is_active", false)
	}

	// Convert userID to uint for logging
	userIDStr := fmt.Sprintf("%v", userID)
	if userIDUint, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
		logUserActivity(uint(userIDUint), "logout", c)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logout successful",
	})
}

func forgotPasswordHandler(c *gin.Context) {
	var req auth.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	// Find user by email
	var foundUser user.User
	if err := db.Where("email = ?", req.Email).First(&foundUser).Error; err != nil {
		// Don't reveal if email exists or not for security
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "If the email exists, a password reset link has been sent",
		})
		return
	}

	// Check if user is active
	if !foundUser.IsActive {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "If the email exists, a password reset link has been sent",
		})
		return
	}

	// Generate password reset token
	resetToken := generateRandomToken()
	resetExpiry := time.Now().Add(1 * time.Hour) // Token expires in 1 hour

	// Update or create verification record
	verification := user.UserVerification{
		UserID:              fmt.Sprintf("%d", foundUser.ID),
		PasswordResetToken:  resetToken,
		PasswordResetExpiry: resetExpiry,
	}

	// First try to update existing record
	result := db.Model(&verification).Where("user_id = ?", fmt.Sprintf("%d", foundUser.ID)).
		Updates(map[string]interface{}{
			"password_reset_token":  resetToken,
			"password_reset_expiry": resetExpiry,
		})

	if result.RowsAffected == 0 {
		// Create new verification record if none exists
		db.Create(&verification)
	}

	// Log activity
	logUserActivity(foundUser.ID, "password_reset_requested", c)

	// TODO: Send email with reset link
	// For now, log the token (in production, send email)
	log.Printf("Password reset token for %s: %s", foundUser.Email, resetToken)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "If the email exists, a password reset link has been sent",
	})
}

func resetPasswordHandler(c *gin.Context) {
	var req auth.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	// Validate password confirmation
	if req.NewPassword != req.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Password confirmation does not match",
		})
		return
	}

	// Find verification record with the reset token
	var verification user.UserVerification
	if err := db.Where("password_reset_token = ? AND password_reset_expiry > ?",
		req.Token, time.Now()).First(&verification).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid or expired reset token",
		})
		return
	}

	// Find the user
	var foundUser user.User
	if err := db.First(&foundUser, verification.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "User not found",
		})
		return
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to hash password",
		})
		return
	}

	// Update user password
	foundUser.Password = string(hashedPassword)
	if err := db.Save(&foundUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update password",
		})
		return
	}

	// Clear the reset token
	verification.PasswordResetToken = ""
	verification.PasswordResetExpiry = time.Time{}
	db.Save(&verification)

	// Invalidate all user sessions for security
	db.Model(&user.UserSession{}).Where("user_id = ?", verification.UserID).
		Update("is_active", false)

	// Log activity
	logUserActivity(foundUser.ID, "password_reset_completed", c)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Password reset successfully",
	})
}

func changePasswordHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	var req auth.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	// Find the user
	var foundUser user.User
	if err := db.First(&foundUser, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "User not found",
		})
		return
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(req.CurrentPassword)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Current password is incorrect",
		})
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to hash new password",
		})
		return
	}

	// Update password
	foundUser.Password = string(hashedPassword)
	if err := db.Save(&foundUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update password",
		})
		return
	}

	// Invalidate all sessions except current one for security
	// (Force re-login on other devices)
	db.Model(&user.UserSession{}).
		Where("user_id = ? AND ip_address != ? AND is_active = ?",
			fmt.Sprintf("%v", userID), c.ClientIP(), true).
		Update("is_active", false)

	// Convert userID to uint for logging
	userIDStr := fmt.Sprintf("%v", userID)
	if userIDUint, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
		logUserActivity(uint(userIDUint), "password_changed", c)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Password changed successfully",
	})
}

func verifyEmailHandler(c *gin.Context) {
	// Get token from query parameter or request body
	token := c.Query("token")
	if token == "" {
		var req struct {
			Token string `json:"token" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Email verification token is required",
				"details": err.Error(),
			})
			return
		}
		token = req.Token
	}

	// Find verification record with the email token
	var verification user.UserVerification
	if err := db.Where("email_token = ?", token).First(&verification).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid email verification token",
		})
		return
	}

	// Check if email is already verified
	if verification.EmailVerified {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Email is already verified",
		})
		return
	}

	// Find the user
	var foundUser user.User
	if err := db.First(&foundUser, verification.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "User not found",
		})
		return
	}

	// Mark email as verified
	verification.EmailVerified = true
	verification.VerifiedAt = time.Now()
	verification.EmailToken = "" // Clear the token
	if err := db.Save(&verification).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update verification status",
		})
		return
	}

	// Log activity
	logUserActivity(foundUser.ID, "email_verified", c)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Email verified successfully",
		"data": gin.H{
			"user_id":     foundUser.ID,
			"email":       foundUser.Email,
			"verified":    true,
			"verified_at": verification.VerifiedAt,
		},
	})
}

func resendVerificationHandler(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	// Find user by email
	var foundUser user.User
	if err := db.Where("email = ?", req.Email).First(&foundUser).Error; err != nil {
		// Don't reveal if email exists or not for security
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "If the email exists and is not verified, a verification email has been sent",
		})
		return
	}

	// Find verification record
	var verification user.UserVerification
	if err := db.Where("user_id = ?", fmt.Sprintf("%d", foundUser.ID)).First(&verification).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "If the email exists and is not verified, a verification email has been sent",
		})
		return
	}

	// Check if already verified
	if verification.EmailVerified {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Email is already verified",
		})
		return
	}

	// Generate new verification token
	newToken := generateRandomToken()
	verification.EmailToken = newToken
	db.Save(&verification)

	// Log activity
	logUserActivity(foundUser.ID, "verification_email_resent", c)

	// TODO: Send verification email
	log.Printf("Email verification token for %s: %s", foundUser.Email, newToken)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "If the email exists and is not verified, a verification email has been sent",
	})
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"status":    "healthy",
			"service":   "auth-service",
			"timestamp": time.Now().Format(time.RFC3339),
			"version":   "1.0.0",
			"database":  "connected",
		},
	})
}

// Middleware
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Authorization header required",
			})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Bearer token required",
			})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid or expired token",
			})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid token claims",
			})
			c.Abort()
			return
		}

		c.Set("userID", claims["sub"])
		c.Set("userEmail", claims["email"])
		c.Set("userRole", claims["role"])
		c.Next()
	}
}

func RateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		if !authRateLimiter.IsAllowed(clientIP, limit, window) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success":     false,
				"error":       "Too many requests. Please try again later.",
				"retry_after": int(window.Seconds()),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		log.Printf("%s %s %d %v", c.Request.Method, c.Request.URL.Path, c.Writer.Status(), duration)
	}
}

// Helper functions
func generateJWTToken(user user.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(15 * time.Minute) // Short-lived access token (15 minutes)

	claims := jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"role":  user.Role,
		"exp":   expiresAt.Unix(),
		"iat":   time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.JWTSecret))

	return tokenString, expiresAt, err
}

func generateRefreshToken(userID uint, clientIP, userAgent string) (*user.UserSession, error) {
	// Generate a secure random token
	refreshToken := generateRandomToken()
	expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days

	session := &user.UserSession{
		UserID:    fmt.Sprintf("%d", userID),
		Token:     refreshToken,
		ExpiresAt: expiresAt,
		IPAddress: clientIP,
		UserAgent: userAgent,
		IsActive:  true,
	}

	if err := db.Create(session).Error; err != nil {
		return nil, err
	}

	return session, nil
}

func generateRandomToken() string {
	// Generate 32 bytes of random data
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to time-based token if crypto/rand fails
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

func logUserActivity(userID uint, activity string, c *gin.Context) {
	metadata, _ := json.Marshal(map[string]interface{}{
		"ip_address": c.ClientIP(),
		"user_agent": c.Request.UserAgent(),
	})

	userActivity := user.UserActivity{
		UserID:   fmt.Sprintf("%d", userID),
		Activity: activity,
		Metadata: string(metadata),
	}
	db.Create(&userActivity)
}
