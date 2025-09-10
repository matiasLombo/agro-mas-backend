package handlers

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"agro-mas-backend/internal/marketplace/products"
	"agro-mas-backend/internal/marketplace/users"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProductsHandler struct {
	productService *products.Service
	imageService   *products.ImageService
	userService    *users.Service
}

func NewProductsHandler(productService *products.Service, imageService *products.ImageService, userService *users.Service) *ProductsHandler {
	return &ProductsHandler{
		productService: productService,
		imageService:   imageService,
		userService:    userService,
	}
}

// CreateProduct handles product creation
func (h *ProductsHandler) CreateProduct(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
			"code":  "AUTH_REQUIRED",
		})
		return
	}

	// Check if user has complete seller profile
	isComplete, err := h.userService.IsSellerProfileComplete(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to check seller profile status",
			"code":  "SELLER_PROFILE_CHECK_FAILED",
		})
		return
	}

	if !isComplete {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Seller profile incomplete",
			"code":    "SELLER_PROFILE_INCOMPLETE",
			"message": "You must complete your seller profile before publishing products",
		})
		return
	}

	var req products.CreateProductRequest
	var imageFiles []*multipart.FileHeader

	// Check if this is multipart form data (with images) or JSON
	contentType := c.GetHeader("Content-Type")
	if contentType != "" && strings.HasPrefix(contentType, "multipart/form-data") {
		// Handle FormData with images
		productJSON := c.PostForm("product")
		if productJSON == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Missing product data",
				"code":  "MISSING_PRODUCT_DATA",
			})
			return
		}

		fmt.Printf("[DEBUG] POST: About to parse JSON: '%s'\n", productJSON)
		if err := json.Unmarshal([]byte(productJSON), &req); err != nil {
			fmt.Printf("[ERROR] POST: Failed to unmarshal product JSON: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid product JSON in form data",
				"code":    "INVALID_PRODUCT_JSON",
				"details": err.Error(),
			})
			return
		}

		// Handle image files
		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Failed to parse multipart form",
				"code":    "MULTIPART_PARSE_ERROR",
				"details": err.Error(),
			})
			return
		}

		// Process image files
		if files := form.File["image"]; len(files) > 0 {
			fmt.Printf("[DEBUG] Received %d image files\n", len(files))
			imageFiles = files
			for i, file := range files {
				fmt.Printf("[DEBUG] Image %d: %s (%d bytes)\n", i+1, file.Filename, file.Size)
			}
		}

	} else {
		// Handle regular JSON request
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request format",
				"code":    "INVALID_REQUEST",
				"details": err.Error(),
			})
			return
		}
	}

	// Get seller information from context
	sellerInfo := products.SellerInfo{
		Name:              "Seller Name", // This would come from user service
		Phone:             "Seller Phone",
		Rating:            4.5,
		VerificationLevel: 2,
	}

	var product *products.Product

	// Create product - use CreateProductWithImages if we have image files, otherwise use CreateProduct
	if len(imageFiles) > 0 {
		// Create product with images
		product, err = h.productService.CreateProductWithImages(c.Request.Context(), userID.(uuid.UUID), &req, sellerInfo, imageFiles)
	} else {
		// Create product without images
		product, err = h.productService.CreateProduct(c.Request.Context(), userID.(uuid.UUID), &req, sellerInfo)
	}
	if err != nil {
		// Enhanced logging for debugging
		fmt.Printf("[ERROR] Product creation failed: %v\n", err)
		fmt.Printf("[DEBUG] Request data: %+v\n", req)
		fmt.Printf("[DEBUG] User ID: %v\n", userID)
		fmt.Printf("[DEBUG] Seller info: %+v\n", sellerInfo)

		status := http.StatusInternalServerError
		code := "PRODUCT_CREATION_FAILED"

		switch err {
		case products.ErrInvalidCategory:
			status = http.StatusBadRequest
			code = "INVALID_CATEGORY"
		case products.ErrInvalidPriceType:
			status = http.StatusBadRequest
			code = "INVALID_PRICE_TYPE"
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
			"code":  code,
			"debug": fmt.Sprintf("Full error: %v", err),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Product created successfully",
		"product": product,
	})
}

// GetProduct retrieves a product by ID
func (h *ProductsHandler) GetProduct(c *gin.Context) {
	productIDStr := c.Param("id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid product ID format",
			"code":  "INVALID_PRODUCT_ID",
		})
		return
	}

	// Check if user is authenticated to decide whether to increment view count
	_, authenticated := c.Get("user_id")
	incrementView := authenticated

	product, err := h.productService.GetProductByID(c.Request.Context(), productID, incrementView)
	if err != nil {
		status := http.StatusNotFound
		if err != products.ErrProductNotFound {
			status = http.StatusInternalServerError
		}

		c.JSON(status, gin.H{
			"error": "Product not found",
			"code":  "PRODUCT_NOT_FOUND",
		})
		return
	}

	// Get real seller information
	if seller, err := h.userService.GetUserByID(c.Request.Context(), product.UserID); err == nil && seller != nil {
		// Update seller information with real data
		if seller.BusinessName != nil && *seller.BusinessName != "" {
			product.SellerName = seller.BusinessName
		} else {
			fullName := fmt.Sprintf("%s %s", seller.FirstName, seller.LastName)
			product.SellerName = &fullName
		}

		// Update other seller fields if available
		if seller.Rating > 0 {
			product.SellerRating = &seller.Rating
		}
		product.SellerVerificationLevel = &seller.VerificationLevel
	}

	c.JSON(http.StatusOK, product)
}

// SearchProducts handles product search
func (h *ProductsHandler) SearchProducts(c *gin.Context) {
	req := &products.ProductSearchRequest{
		Query:       c.Query("query"),
		Category:    c.Query("category"),
		Subcategory: c.Query("subcategory"),
		Province:    c.Query("province"),
		City:        c.Query("city"),
		PriceType:   c.Query("price_type"),
		SortBy:      c.Query("sort_by"),
	}

	// Parse optional numeric parameters
	if minPriceStr := c.Query("min_price"); minPriceStr != "" {
		if minPrice, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
			req.MinPrice = &minPrice
		}
	}

	if maxPriceStr := c.Query("max_price"); maxPriceStr != "" {
		if maxPrice, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
			req.MaxPrice = &maxPrice
		}
	}

	// Parse boolean parameters
	if pickupStr := c.Query("pickup_available"); pickupStr != "" {
		if pickup, err := strconv.ParseBool(pickupStr); err == nil {
			req.PickupAvailable = &pickup
		}
	}

	if deliveryStr := c.Query("delivery_available"); deliveryStr != "" {
		if delivery, err := strconv.ParseBool(deliveryStr); err == nil {
			req.DeliveryAvailable = &delivery
		}
	}

	if verifiedStr := c.Query("is_verified_seller"); verifiedStr != "" {
		if verified, err := strconv.ParseBool(verifiedStr); err == nil {
			req.IsVerifiedSeller = &verified
		}
	}

	// Parse pagination
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			req.Page = page
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil {
			req.PageSize = pageSize
		}
	}

	// Parse tags
	if tags := c.QueryArray("tags"); len(tags) > 0 {
		req.Tags = tags
	}

	// If user is authenticated, exclude their own products
	if userID, exists := c.Get("user_id"); exists {
		userUUID := userID.(uuid.UUID)
		req.ExcludeUserID = &userUUID
	}

	response, err := h.productService.SearchProducts(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to search products",
			"code":    "SEARCH_FAILED",
			"details": err.Error(),
		})
		return
	}

	// Enrich products with real seller information
	for i := range response.Products {
		if seller, err := h.userService.GetUserByID(c.Request.Context(), response.Products[i].UserID); err == nil && seller != nil {
			// Update seller information with real data
			if seller.BusinessName != nil && *seller.BusinessName != "" {
				response.Products[i].SellerName = seller.BusinessName
			} else {
				fullName := fmt.Sprintf("%s %s", seller.FirstName, seller.LastName)
				response.Products[i].SellerName = &fullName
			}

			// Update other seller fields if available
			if seller.Rating > 0 {
				response.Products[i].SellerRating = &seller.Rating
			}
			response.Products[i].SellerVerificationLevel = &seller.VerificationLevel
		}
	}

	c.JSON(http.StatusOK, response)
}

// UpdateProduct handles product updates
func (h *ProductsHandler) UpdateProduct(c *gin.Context) {
	fmt.Printf("[DEBUG] UpdateProduct handler called - Method: %s, URL: %s\n", c.Request.Method, c.Request.URL.String())
	fmt.Printf("[DEBUG] Request Headers: %v\n", c.Request.Header)

	userID, exists := c.Get("user_id")
	if !exists {
		fmt.Printf("[ERROR] User not authenticated\n")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
			"code":  "AUTH_REQUIRED",
		})
		return
	}

	productIDStr := c.Param("id")
	fmt.Printf("[DEBUG] Product ID from URL: %s\n", productIDStr)
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		fmt.Printf("[ERROR] Invalid product ID format: %s\n", productIDStr)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid product ID format",
			"code":  "INVALID_PRODUCT_ID",
		})
		return
	}

	var req products.UpdateProductRequest
	var imageFiles []*multipart.FileHeader

	// Check if this is multipart form data (with images) or JSON
	contentType := c.GetHeader("Content-Type")
	fmt.Printf("[DEBUG] Update PUT request Content-Type: %s\n", contentType)
	if contentType != "" && len(contentType) >= 19 && contentType[:19] == "multipart/form-data" {
		fmt.Printf("[DEBUG] Processing multipart form data for update\n")
		// Handle FormData with images
		productJSON := c.PostForm("product")
		fmt.Printf("[DEBUG] Product JSON from form: %s\n", productJSON)
		if productJSON == "" {
			fmt.Printf("[ERROR] Missing product data in form\n")
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Missing product data",
				"code":  "MISSING_PRODUCT_DATA",
			})
			return
		}

		fmt.Printf("[DEBUG] PUT: About to parse JSON: '%s'\n", productJSON)
		fmt.Printf("[DEBUG] PUT: JSON length: %d\n", len(productJSON))

		if err := json.Unmarshal([]byte(productJSON), &req); err != nil {
			fmt.Printf("[ERROR] PUT: Failed to unmarshal product JSON: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request format",
				"code":    "INVALID_REQUEST",
				"details": err.Error(),
			})
			return
		}

		// Handle image files
		form, err := c.MultipartForm()
		if err != nil {
			fmt.Printf("[ERROR] Failed to parse multipart form: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Failed to parse multipart form",
				"code":    "MULTIPART_PARSE_ERROR",
				"details": err.Error(),
			})
			return
		}

		// Process image files
		if files := form.File["image"]; len(files) > 0 {
			fmt.Printf("[DEBUG] Received %d image files for update\n", len(files))
			imageFiles = files
			for i, file := range files {
				fmt.Printf("[DEBUG] Image %d: %s (%d bytes)\n", i+1, file.Filename, file.Size)
			}
		} else {
			fmt.Printf("[DEBUG] No image files found in form\n")
		}

	} else {
		// Handle regular JSON request
		fmt.Printf("[DEBUG] Processing JSON request for update\n")
		fmt.Printf("[DEBUG] Content-Type: %s\n", contentType)
		if err := c.ShouldBindJSON(&req); err != nil {
			fmt.Printf("[ERROR] Failed to bind JSON: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request format",
				"code":    "INVALID_REQUEST",
				"details": err.Error(),
			})
			return
		}
		fmt.Printf("[DEBUG] Successfully parsed JSON request: %+v\n", req)
	}

	var product *products.Product

	// Update product - use UpdateProductWithImages if we have image files OR existing_images changes, otherwise use UpdateProduct
	hasImageChanges := len(imageFiles) > 0 || req.ExistingImages != nil
	if hasImageChanges {
		// Update product with images (new images or existing image changes)
		product, err = h.productService.UpdateProductWithImages(c.Request.Context(), userID.(uuid.UUID), productID, &req, imageFiles)
	} else {
		// Update product without images
		product, err = h.productService.UpdateProduct(c.Request.Context(), userID.(uuid.UUID), productID, &req)
	}

	if err != nil {
		// Enhanced logging for debugging
		fmt.Printf("[ERROR] Product update failed: %v\n", err)
		fmt.Printf("[DEBUG] Update request data: %+v\n", req)
		fmt.Printf("[DEBUG] User ID: %v\n", userID)
		fmt.Printf("[DEBUG] Product ID: %v\n", productID)

		status := http.StatusInternalServerError
		code := "PRODUCT_UPDATE_FAILED"

		switch err {
		case products.ErrProductNotFound:
			status = http.StatusNotFound
			code = "PRODUCT_NOT_FOUND"
		case products.ErrProductNotOwnedByUser:
			status = http.StatusForbidden
			code = "NOT_PRODUCT_OWNER"
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
			"code":  code,
			"debug": fmt.Sprintf("Full error: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product updated successfully",
		"product": product,
	})
}

// PublishProduct publishes a product
func (h *ProductsHandler) PublishProduct(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
			"code":  "AUTH_REQUIRED",
		})
		return
	}

	productIDStr := c.Param("id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid product ID format",
			"code":  "INVALID_PRODUCT_ID",
		})
		return
	}

	err = h.productService.PublishProduct(c.Request.Context(), userID.(uuid.UUID), productID)
	if err != nil {
		status := http.StatusInternalServerError
		code := "PUBLISH_FAILED"

		switch err {
		case products.ErrProductNotFound:
			status = http.StatusNotFound
			code = "PRODUCT_NOT_FOUND"
		case products.ErrProductNotOwnedByUser:
			status = http.StatusForbidden
			code = "NOT_PRODUCT_OWNER"
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
			"code":  code,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product published successfully",
	})
}

// UnpublishProduct unpublishes a product
func (h *ProductsHandler) UnpublishProduct(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
			"code":  "AUTH_REQUIRED",
		})
		return
	}

	productIDStr := c.Param("id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid product ID format",
			"code":  "INVALID_PRODUCT_ID",
		})
		return
	}

	err = h.productService.UnpublishProduct(c.Request.Context(), userID.(uuid.UUID), productID)
	if err != nil {
		status := http.StatusInternalServerError
		code := "UNPUBLISH_FAILED"

		switch err {
		case products.ErrProductNotFound:
			status = http.StatusNotFound
			code = "PRODUCT_NOT_FOUND"
		case products.ErrProductNotOwnedByUser:
			status = http.StatusForbidden
			code = "NOT_PRODUCT_OWNER"
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
			"code":  code,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product unpublished successfully",
	})
}

// DeleteProduct handles product deletion
func (h *ProductsHandler) DeleteProduct(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
			"code":  "AUTH_REQUIRED",
		})
		return
	}

	productIDStr := c.Param("id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid product ID format",
			"code":  "INVALID_PRODUCT_ID",
		})
		return
	}

	err = h.productService.DeleteProduct(c.Request.Context(), userID.(uuid.UUID), productID)
	if err != nil {
		status := http.StatusInternalServerError
		code := "DELETE_FAILED"

		switch err {
		case products.ErrProductNotFound:
			status = http.StatusNotFound
			code = "PRODUCT_NOT_FOUND"
		case products.ErrProductNotOwnedByUser:
			status = http.StatusForbidden
			code = "NOT_PRODUCT_OWNER"
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
			"code":  code,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product deleted successfully",
	})
}

// GetUserProducts retrieves products belonging to a user
func (h *ProductsHandler) GetUserProducts(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
			"code":  "AUTH_REQUIRED",
		})
		return
	}

	page := 1
	pageSize := 20

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil {
			page = p
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil {
			pageSize = ps
		}
	}

	response, err := h.productService.GetUserProducts(c.Request.Context(), userID.(uuid.UUID), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get user products",
			"code":    "USER_PRODUCTS_FETCH_FAILED",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetProductImages retrieves all images for a product
func (h *ProductsHandler) GetProductImages(c *gin.Context) {
	productIDStr := c.Param("id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid product ID format",
			"code":  "INVALID_PRODUCT_ID",
		})
		return
	}

	images, err := h.imageService.GetProductImages(c.Request.Context(), productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get product images",
			"code":    "IMAGES_FETCH_FAILED",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, images)
}

// UploadProductImage handles product image uploads
func (h *ProductsHandler) UploadProductImage(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
			"code":  "AUTH_REQUIRED",
		})
		return
	}

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil { // 10MB max
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to parse multipart form",
			"code":  "INVALID_FORM",
		})
		return
	}

	// Get file from form
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No image file provided",
			"code":  "NO_IMAGE_FILE",
		})
		return
	}
	defer file.Close()

	// Parse form data
	var req products.UploadImageRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid form data",
			"code":    "INVALID_FORM_DATA",
			"details": err.Error(),
		})
		return
	}

	image, err := h.imageService.UploadProductImage(c.Request.Context(), userID.(uuid.UUID), file, header, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to upload image",
			"code":    "IMAGE_UPLOAD_FAILED",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Image uploaded successfully",
		"image":   image,
	})
}

// RegisterRoutes registers product routes
func (h *ProductsHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware, optionalAuthMiddleware, sellerMiddleware gin.HandlerFunc) {
	products := router.Group("/products")
	{
		// Public routes with optional authentication
		products.GET("/search", optionalAuthMiddleware, h.SearchProducts)
		products.GET("/:id", h.GetProduct)
		products.GET("/:id/images", h.GetProductImages)

		// Protected routes
		protected := products.Group("/")
		protected.Use(authMiddleware)
		{
			protected.GET("/my", h.GetUserProducts)

			// Seller-only routes
			seller := protected.Group("/")
			seller.Use(sellerMiddleware)
			{
				seller.POST("/", h.CreateProduct)
				seller.PUT("/:id", h.UpdateProduct)
				seller.DELETE("/:id", h.DeleteProduct)
				seller.POST("/:id/publish", h.PublishProduct)
				seller.POST("/:id/unpublish", h.UnpublishProduct)
				seller.POST("/images", h.UploadProductImage)
			}
		}
	}
}
