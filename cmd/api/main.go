package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"agro-mas-backend/cmd/api/handlers"
	"agro-mas-backend/internal/auth"
	"agro-mas-backend/internal/config"
	"agro-mas-backend/internal/marketplace/products"
	"agro-mas-backend/internal/marketplace/transactions"
	"agro-mas-backend/internal/marketplace/users"
	"agro-mas-backend/internal/storage"
	"agro-mas-backend/pkg/gcloud"
	"agro-mas-backend/pkg/middleware"
	"agro-mas-backend/pkg/whatsapp"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set Gin mode
	gin.SetMode(cfg.Server.GinMode)

	// Initialize database
	db, err := storage.NewDatabase(cfg.GetDatabaseURL())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run database migrations
	if err := db.RunMigrations("./migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize Google Cloud Storage (production or emulator)
	ctx := context.Background()
	var storageClient *gcloud.StorageClient
	if cfg.GoogleCloud.ProjectID != "" && cfg.GoogleCloud.StorageBucket != "" {
		var err error
		if cfg.GoogleCloud.UseEmulator && cfg.GoogleCloud.EmulatorHost != "" {
			// Use emulator for development
			log.Printf("🔧 Using Google Cloud Storage Emulator at %s", cfg.GoogleCloud.EmulatorHost)
			storageClient, err = gcloud.NewStorageClientWithEmulator(ctx, cfg.GoogleCloud.ProjectID, cfg.GoogleCloud.StorageBucket, cfg.GoogleCloud.EmulatorHost)
		} else if cfg.IsProduction() {
			// Use real Google Cloud Storage in production
			log.Println("☁️  Using Google Cloud Storage")
			storageClient, err = gcloud.NewStorageClient(ctx, cfg.GoogleCloud.ProjectID, cfg.GoogleCloud.CredentialsFile, cfg.GoogleCloud.StorageBucket)
		}

		if err != nil {
			log.Fatalf("Failed to initialize Google Cloud Storage: %v", err)
		}
		if storageClient != nil {
			defer storageClient.Close()
		}
	} else {
		log.Println("⚠️  Google Cloud Storage disabled - no configuration provided")
	}

	// Initialize WhatsApp client
	whatsappClient := whatsapp.NewClient(cfg.WhatsApp.APIUrl, cfg.WhatsApp.BusinessNumber)

	// Initialize authentication components
	passwordManager := auth.NewPasswordManager(nil)
	jwtManager := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.ExpirationHours)

	// Initialize repositories
	userRepo := users.NewRepository(db.GetDB())
	productRepo := products.NewRepository(db.GetDB())
	transactionRepo := transactions.NewRepository(db.GetDB())

	// Initialize services
	userService := users.NewService(userRepo, passwordManager, jwtManager)
	imageService := products.NewImageService(db.GetDB(), storageClient)
	productService := products.NewService(productRepo, imageService)
	transactionService := transactions.NewService(transactionRepo)
	whatsappService := whatsapp.NewService(whatsappClient, db.GetDB())

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userService)
	productsHandler := handlers.NewProductsHandler(productService, imageService, userService)

	// Initialize Gin router
	router := gin.New()

	// Set multipart form memory limit (default is 32MB)
	router.MaxMultipartMemory = 64 << 20 // 64 MiB

	// Add debug middleware for PUT requests
	router.Use(func(c *gin.Context) {
		if c.Request.Method == "PUT" && strings.Contains(c.Request.URL.Path, "/products/") {
			fmt.Printf("[DEBUG MIDDLEWARE] PUT request intercepted\n")
			fmt.Printf("[DEBUG MIDDLEWARE] Path: %s\n", c.Request.URL.Path)
			fmt.Printf("[DEBUG MIDDLEWARE] Content-Type: %s\n", c.Request.Header.Get("Content-Type"))
			fmt.Printf("[DEBUG MIDDLEWARE] Content-Length: %s\n", c.Request.Header.Get("Content-Length"))

			// Try to access the body
			if c.Request.Body != nil {
				fmt.Printf("[DEBUG MIDDLEWARE] Body is not nil\n")
			}
		}
		c.Next()
	})

	// Add middleware
	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.SecurityHeadersMiddleware())
	router.Use(middleware.APIVersionMiddleware("v1"))
	router.Use(middleware.ContentTypeMiddleware())

	// JWT Test endpoint - REMOVE IN PRODUCTION
	router.GET("/test-jwt", func(c *gin.Context) {
		// Test JWT generation
		testUserID := uuid.New()
		tokenResponse, err := jwtManager.GenerateToken(
			testUserID,
			"test@example.com",
			"seller",
			nil,
			nil,
			1,
			false,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "JWT generation failed",
				"details": err.Error(),
			})
			return
		}

		// Test JWT verification
		claims, err := jwtManager.VerifyToken(tokenResponse.AccessToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "JWT verification failed",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"jwt_generation":   "success",
			"jwt_verification": "success",
			"token_info": gin.H{
				"user_id":    claims.UserID,
				"email":      claims.Email,
				"role":       claims.Role,
				"expires_at": claims.ExpiresAt.Time,
			},
			"config_info": gin.H{
				"jwt_secret_length":    len(cfg.JWT.Secret),
				"jwt_expiration_hours": cfg.JWT.ExpirationHours.Hours(),
			},
		})
	})

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		// Check database connection
		if err := db.HealthCheck(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":   "unhealthy",
				"database": "down",
				"error":    err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":      "healthy",
			"timestamp":   time.Now(),
			"version":     "v1",
			"database":    "up",
			"environment": cfg.Environment,
		})
	})

	// Database Test endpoint - REMOVE IN PRODUCTION
	router.GET("/test-db", func(c *gin.Context) {
		// Test database connection
		if err := db.HealthCheck(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "Database connection failed",
				"details": err.Error(),
			})
			return
		}

		// Test user lookup - create a test user first
		testUser := &users.User{
			ID:                uuid.New(),
			Email:             "debug@test.com",
			PasswordHash:      "test-hash",
			FirstName:         "Debug",
			LastName:          "User",
			Role:              "seller",
			IsActive:          true,
			VerificationLevel: 1,
		}

		// Try to get existing test user first
		existingUser, err := userService.GetUserByID(c.Request.Context(), testUser.ID)
		if err != nil && existingUser == nil {
			// User doesn't exist, this is expected
		}

		c.JSON(http.StatusOK, gin.H{
			"database":     "connected",
			"user_service": "initialized",
			"db_url":       cfg.GetDatabaseURL(),
		})
	})

	// Environment test endpoint
	router.GET("/env", func(c *gin.Context) {
		environment := cfg.Environment
		var message string
		var color string

		switch environment {
		case "production":
			message = "🚀 You are connected to PRODUCTION environment"
			color = "red"
		case "development":
			message = "🛠️ You are connected to DEVELOPMENT environment"
			color = "blue"
		default:
			message = "❓ Unknown environment"
			color = "gray"
		}

		c.JSON(http.StatusOK, gin.H{
			"environment": environment,
			"message":     message,
			"color":       color,
			"timestamp":   time.Now(),
			"project_id":  cfg.GoogleCloud.ProjectID,
			"gin_mode":    cfg.Server.GinMode,
			"db_host":     cfg.Database.Host,
		})
	})

	// Debug Authentication endpoint - REMOVE IN PRODUCTION
	router.POST("/debug-auth", func(c *gin.Context) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request format",
				"details": err.Error(),
			})
			return
		}

		// Step 1: Check if user exists
		_, err := userService.GetUserByID(c.Request.Context(), uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")) // Test UUID
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"step":         "user_lookup_by_id",
				"status":       "failed",
				"error":        err.Error(),
				"user_service": "initialized",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"debug_mode": true,
			"jwt_config": gin.H{
				"secret_length":    len(cfg.JWT.Secret),
				"expiration_hours": cfg.JWT.ExpirationHours.Hours(),
			},
			"database": "connected",
			"services": "initialized",
			"message":  "Debug endpoint working",
		})
	})

	// API routes
	api := router.Group("/api/v1")

	// Initialize middleware for protected routes
	authMiddleware := middleware.AuthMiddleware(jwtManager)
	optionalAuthMiddleware := middleware.OptionalAuthMiddleware(jwtManager)
	sellerMiddleware := middleware.SellerOnly()
	adminMiddleware := middleware.AdminOnly()

	// Register routes
	authHandler.RegisterRoutes(api, authMiddleware)
	productsHandler.RegisterRoutes(api, authMiddleware, optionalAuthMiddleware, sellerMiddleware)

	// Additional API endpoints
	registerAdditionalRoutes(api, authMiddleware, adminMiddleware, userService, transactionService, whatsappService)

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("🌾 Agro Mas API server starting on port %s", cfg.Server.Port)
		log.Printf("📍 Environment: %s", cfg.Environment)
		log.Printf("🔗 Health check: http://localhost:%s/health", cfg.Server.Port)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	// Give outstanding requests a deadline for completion
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("⚠️  Server forced to shutdown: %v", err)
	} else {
		log.Println("✅ Server shutdown complete")
	}
}

// registerAdditionalRoutes adds remaining API routes
func registerAdditionalRoutes(
	api *gin.RouterGroup,
	authMiddleware, adminMiddleware gin.HandlerFunc,
	userService *users.Service,
	transactionService *transactions.Service,
	whatsappService *whatsapp.Service,
) {
	// Transaction routes
	transactions := api.Group("/transactions")
	transactions.Use(authMiddleware)
	{
		transactions.GET("/", getTransactions(transactionService))
		transactions.GET("/:id", getTransaction(transactionService))
		transactions.POST("/", createTransaction(transactionService))
		transactions.PUT("/:id", updateTransaction(transactionService))
		transactions.POST("/:id/review", addTransactionReview(transactionService))
	}

	// Inquiry routes
	inquiries := api.Group("/inquiries")
	inquiries.Use(authMiddleware)
	{
		inquiries.POST("/", createInquiry(transactionService))
		inquiries.POST("/:id/respond", respondToInquiry(transactionService))
	}

	// WhatsApp routes
	whatsappGroup := api.Group("/whatsapp")
	whatsappGroup.Use(authMiddleware)
	{
		whatsappGroup.POST("/link", createWhatsAppLink(whatsappService))
		whatsappGroup.GET("/links", getUserWhatsAppLinks(whatsappService))
		whatsappGroup.POST("/track/:id", trackWhatsAppClick(whatsappService))
	}

	// Admin routes
	admin := api.Group("/admin")
	admin.Use(adminMiddleware)
	{
		admin.GET("/users", getUsers(userService))
		admin.PUT("/users/:id/verification", updateUserVerification(userService))
		admin.GET("/stats", getSystemStats(userService, transactionService))
	}
}

// Transaction handlers
func getTransactions(service *transactions.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		uid := userID.(uuid.UUID)

		req := &transactions.TransactionListRequest{}
		if err := c.ShouldBindQuery(req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		response, err := service.ListTransactions(c.Request.Context(), &uid, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

func getTransaction(service *transactions.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		transactionID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
			return
		}

		transaction, err := service.GetTransactionByID(c.Request.Context(), userID.(uuid.UUID), transactionID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"transaction": transaction})
	}
}

func createTransaction(service *transactions.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")

		var req transactions.CreateTransactionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// This would need product info and user info from other services
		productInfo := transactions.ProductInfo{} // Would be fetched
		sellerInfo := transactions.SellerInfo{}   // Would be fetched
		buyerInfo := transactions.BuyerInfo{}     // Would be fetched

		transaction, err := service.CreateTransaction(c.Request.Context(), userID.(uuid.UUID), &req, productInfo, sellerInfo, buyerInfo)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"transaction": transaction})
	}
}

func updateTransaction(service *transactions.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		transactionID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
			return
		}

		var req transactions.UpdateTransactionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		transaction, err := service.UpdateTransaction(c.Request.Context(), userID.(uuid.UUID), transactionID, &req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"transaction": transaction})
	}
}

func addTransactionReview(service *transactions.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		transactionID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
			return
		}

		var req transactions.AddReviewRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = service.AddReview(c.Request.Context(), userID.(uuid.UUID), transactionID, &req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Review added successfully"})
	}
}

// Inquiry handlers
func createInquiry(service *transactions.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")

		var req transactions.CreateInquiryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Would need to get seller ID from product
		sellerID := uuid.New() // This would be fetched from product service

		inquiry, err := service.CreateInquiry(c.Request.Context(), userID.(uuid.UUID), &req, sellerID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"inquiry": inquiry})
	}
}

func respondToInquiry(service *transactions.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		inquiryID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid inquiry ID"})
			return
		}

		var req transactions.RespondToInquiryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = service.RespondToInquiry(c.Request.Context(), userID.(uuid.UUID), inquiryID, &req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Response sent successfully"})
	}
}

// WhatsApp handlers
func createWhatsAppLink(service *whatsapp.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")

		var req whatsapp.CreateLinkRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		link, err := service.CreateWhatsAppLink(c.Request.Context(), userID.(uuid.UUID), req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"link": link})
	}
}

func getUserWhatsAppLinks(service *whatsapp.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		linkType := c.Query("type")

		links, err := service.GetUserWhatsAppLinks(c.Request.Context(), userID.(uuid.UUID), linkType, 50)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"links": links})
	}
}

func trackWhatsAppClick(service *whatsapp.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		linkID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid link ID"})
			return
		}

		err = service.TrackLinkClick(c.Request.Context(), linkID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Click tracked"})
	}
}

// Admin handlers (simplified)
func getUsers(service *users.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implementation would go here
		c.JSON(http.StatusOK, gin.H{"message": "Users endpoint"})
	}
}

func updateUserVerification(service *users.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implementation would go here
		c.JSON(http.StatusOK, gin.H{"message": "User verification update endpoint"})
	}
}

func getSystemStats(userService *users.Service, transactionService *transactions.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implementation would go here
		c.JSON(http.StatusOK, gin.H{"message": "System stats endpoint"})
	}
}
