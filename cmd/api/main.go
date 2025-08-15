package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
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

	// Run migrations
	if err := db.RunMigrations("./migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize Google Cloud Storage (only in production)
	ctx := context.Background()
	var storageClient *gcloud.StorageClient
	if cfg.IsProduction() && cfg.GoogleCloud.ProjectID != "" && cfg.GoogleCloud.StorageBucket != "" {
		var err error
		storageClient, err = gcloud.NewStorageClient(ctx, cfg.GoogleCloud.ProjectID, cfg.GoogleCloud.CredentialsFile, cfg.GoogleCloud.StorageBucket)
		if err != nil {
			log.Fatalf("Failed to initialize Google Cloud Storage: %v", err)
		}
		defer storageClient.Close()
	} else {
		log.Println("⚠️  Running in development mode - Google Cloud Storage disabled")
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
	productService := products.NewService(productRepo)
	imageService := products.NewImageService(db.GetDB(), storageClient)
	transactionService := transactions.NewService(transactionRepo)
	whatsappService := whatsapp.NewService(whatsappClient, db.GetDB())

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userService)
	productsHandler := handlers.NewProductsHandler(productService, imageService)

	// Initialize Gin router
	router := gin.New()

	// Add middleware
	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.SecurityHeadersMiddleware())
	router.Use(middleware.APIVersionMiddleware("v1"))
	router.Use(middleware.ContentTypeMiddleware())

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
			"environment":   environment,
			"message":       message,
			"color":         color,
			"timestamp":     time.Now(),
			"project_id":    cfg.GoogleCloud.ProjectID,
			"gin_mode":      cfg.Server.GinMode,
			"db_host":       cfg.Database.Host,
		})
	})

	// API routes
	api := router.Group("/api/v1")

	// Initialize middleware for protected routes
	authMiddleware := middleware.AuthMiddleware(jwtManager)
	sellerMiddleware := middleware.SellerOnly()
	adminMiddleware := middleware.AdminOnly()

	// Register routes
	authHandler.RegisterRoutes(api, authMiddleware)
	productsHandler.RegisterRoutes(api, authMiddleware, sellerMiddleware)

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