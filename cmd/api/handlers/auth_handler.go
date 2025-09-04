package handlers

import (
	"net/http"

	"agro-mas-backend/internal/marketplace/users"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	userService *users.Service
}

func NewAuthHandler(userService *users.Service) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req users.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
			"details": err.Error(),
		})
		return
	}

	user, err := h.userService.CreateUser(c.Request.Context(), &req)
	if err != nil {
		status := http.StatusInternalServerError
		code := "REGISTRATION_FAILED"

		// Handle specific errors
		switch err {
		case users.ErrUserExists:
			status = http.StatusConflict
			code = "USER_EXISTS"
		case users.ErrCUITExists:
			status = http.StatusConflict
			code = "CUIT_EXISTS"
		case users.ErrInvalidRole:
			status = http.StatusBadRequest
			code = "INVALID_ROLE"
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
			"code":  code,
		})
		return
	}

	// Generate JWT token for newly registered user
	tokenResponse, err := h.userService.GenerateTokenForUser(c.Request.Context(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate authentication token",
			"code":  "TOKEN_GENERATION_FAILED",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user":    user.ToResponse(),
		"token":   tokenResponse,
	})
}

// Login handles user authentication
func (h *AuthHandler) Login(c *gin.Context) {
	var req users.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
			"details": err.Error(),
		})
		return
	}

	tokenResponse, user, err := h.userService.Authenticate(c.Request.Context(), &req)
	if err != nil {
		status := http.StatusUnauthorized
		code := "AUTHENTICATION_FAILED"

		// Handle specific errors
		switch err {
		case users.ErrUserNotFound:
			code = "USER_NOT_FOUND"
		case users.ErrInvalidPassword:
			code = "INVALID_CREDENTIALS"
		case users.ErrUserNotActive:
			code = "ACCOUNT_INACTIVE"
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
			"code":  code,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user":    user.ToResponse(),
		"token":   tokenResponse,
	})
}

// GetProfile returns the current user's profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
			"code":  "AUTH_REQUIRED",
		})
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user profile",
			"code":  "PROFILE_FETCH_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user.ToResponse(),
	})
}

// UpdateProfile updates the current user's profile
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
			"code":  "AUTH_REQUIRED",
		})
		return
	}

	var req users.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
			"details": err.Error(),
		})
		return
	}

	user, err := h.userService.UpdateUser(c.Request.Context(), userID.(uuid.UUID), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update profile",
			"code":  "PROFILE_UPDATE_FAILED",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user":    user.ToResponse(),
	})
}

// ChangePassword handles password changes
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
			"code":  "AUTH_REQUIRED",
		})
		return
	}

	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
			"details": err.Error(),
		})
		return
	}

	err := h.userService.ChangePassword(c.Request.Context(), userID.(uuid.UUID), req.CurrentPassword, req.NewPassword)
	if err != nil {
		status := http.StatusInternalServerError
		code := "PASSWORD_CHANGE_FAILED"

		if err == users.ErrInvalidPassword {
			status = http.StatusBadRequest
			code = "INVALID_CURRENT_PASSWORD"
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
			"code":  code,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully",
	})
}

// GetPublicProfile returns public information about a user
func (h *AuthHandler) GetPublicProfile(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID format",
			"code":  "INVALID_USER_ID",
		})
		return
	}

	publicUser, err := h.userService.GetPublicUserByID(c.Request.Context(), userID)
	if err != nil {
		status := http.StatusNotFound
		if err != users.ErrUserNotFound {
			status = http.StatusInternalServerError
		}

		c.JSON(status, gin.H{
			"error": "User not found",
			"code":  "USER_NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": publicUser,
	})
}

// Logout handles user logout (client-side token invalidation)
func (h *AuthHandler) Logout(c *gin.Context) {
	// Since we're using stateless JWT, logout is primarily client-side
	// In a production system, you might want to implement token blacklisting
	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// GetSellerProfile returns the current user's seller profile status
func (h *AuthHandler) GetSellerProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
			"code":  "AUTH_REQUIRED",
		})
		return
	}

	profile, err := h.userService.GetSellerProfile(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get seller profile",
			"code":  "SELLER_PROFILE_FETCH_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"profile": profile,
	})
}

// CheckSellerProfileComplete checks if user has complete seller profile
func (h *AuthHandler) CheckSellerProfileComplete(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
			"code":  "AUTH_REQUIRED",
		})
		return
	}

	isComplete, err := h.userService.IsSellerProfileComplete(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to check seller profile status",
			"code":  "SELLER_PROFILE_CHECK_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"is_complete": isComplete,
	})
}

// UpgradeToSeller upgrades a buyer to seller with required information
func (h *AuthHandler) UpgradeToSeller(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
			"code":  "AUTH_REQUIRED",
		})
		return
	}

	var req users.SellerProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
			"details": err.Error(),
		})
		return
	}

	err := h.userService.UpgradeToSeller(c.Request.Context(), userID.(uuid.UUID), &req)
	if err != nil {
		status := http.StatusInternalServerError
		code := "SELLER_UPGRADE_FAILED"

		// Handle specific errors
		if err == users.ErrCUITExists {
			status = http.StatusConflict
			code = "CUIT_EXISTS"
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
			"code":  code,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully upgraded to seller",
	})
}

// RegisterRoutes registers authentication routes
func (h *AuthHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	auth := router.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/logout", h.Logout)
		
		// Protected routes
		protected := auth.Group("/")
		protected.Use(authMiddleware)
		{
			protected.GET("/profile", h.GetProfile)
			protected.PUT("/profile", h.UpdateProfile)
			protected.POST("/change-password", h.ChangePassword)
			
			// Seller profile routes
			protected.GET("/seller-profile", h.GetSellerProfile)
			protected.GET("/seller-profile/check", h.CheckSellerProfileComplete)
			protected.POST("/upgrade-to-seller", h.UpgradeToSeller)
		}
	}

	// Public user routes
	users := router.Group("/users")
	{
		users.GET("/:id", h.GetPublicProfile)
	}
}