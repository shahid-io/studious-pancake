package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
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
		public.POST("/register", registerHandler)
		public.POST("/login", loginHandler)
		public.POST("/refresh", refreshTokenHandler)
		public.POST("/forgot-password", forgotPasswordHandler)
		public.POST("/reset-password", resetPasswordHandler)
		public.GET("/health", healthHandler)
	}

	// Protected routes (require authentication)
	protected := router.Group("/api/v1/auth")
	protected.Use(AuthMiddleware())
	{
		protected.GET("/profile", profileHandler)
		protected.POST("/logout", logoutHandler)
		protected.POST("/change-password", changePasswordHandler)
		protected.POST("/verify-email", verifyEmailHandler)
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
	c.JSON(http.StatusInternalServerError, gin.H{
		"success": true,
		"error":   "Hey, Welcom to Pumpkin",
	})
}

// Handlers
func registerHandler(c *gin.Context) {
	var req auth.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	// Check if user already exists
	var existingUser user.User
	if err := db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error":   "User already exists with this email",
		})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to hash password",
		})
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create user",
			"details": err.Error(),
		})
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to generate token",
		})
		return
	}

	// Log activity
	logUserActivity(newUser.ID, "register", c)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": auth.LoginResponse{
			AccessToken: token,
			TokenType:   "Bearer",
			ExpiresIn:   int64(time.Until(expiresAt).Seconds()),
			ExpiresAt:   expiresAt,
			User:        newUser,
		},
		"message": "User registered successfully",
	})
}

func loginHandler(c *gin.Context) {
	var req auth.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	// Find user by email
	var user user.User
	if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Invalid email or password",
		})
		return
	}

	// Check if user is active
	if !user.IsActive {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Account is deactivated",
		})
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Invalid email or password",
		})
		return
	}

	// Generate JWT token
	token, expiresAt, err := generateJWTToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to generate token",
		})
		return
	}

	// Update last login
	user.LastLogin = time.Now()
	db.Save(&user)

	// Log activity
	logUserActivity(user.ID, "login", c)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": auth.LoginResponse{
			AccessToken: token,
			TokenType:   "Bearer",
			ExpiresIn:   int64(time.Until(expiresAt).Seconds()),
			ExpiresAt:   expiresAt,
			User:        user,
		},
		"message": "Login successful",
	})
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
	// Implement refresh token logic
	c.JSON(http.StatusNotImplemented, gin.H{
		"success": false,
		"error":   "Refresh token endpoint not implemented yet",
	})
}

func logoutHandler(c *gin.Context) {
	// Implement logout logic (token invalidation)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logout successful",
	})
}

func forgotPasswordHandler(c *gin.Context) {
	// Implement forgot password logic
	c.JSON(http.StatusNotImplemented, gin.H{
		"success": false,
		"error":   "Forgot password endpoint not implemented yet",
	})
}

func resetPasswordHandler(c *gin.Context) {
	// Implement reset password logic
	c.JSON(http.StatusNotImplemented, gin.H{
		"success": false,
		"error":   "Reset password endpoint not implemented yet",
	})
}

func changePasswordHandler(c *gin.Context) {
	// Implement change password logic
	c.JSON(http.StatusNotImplemented, gin.H{
		"success": false,
		"error":   "Change password endpoint not implemented yet",
	})
}

func verifyEmailHandler(c *gin.Context) {
	// Implement email verification logic
	c.JSON(http.StatusNotImplemented, gin.H{
		"success": false,
		"error":   "Email verification endpoint not implemented yet",
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
	expiresAt := time.Now().Add(24 * time.Hour) // Token expires in 24 hours

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

func generateRandomToken() string {
	// Implement proper random token generation
	return fmt.Sprintf("%d", time.Now().UnixNano())
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
