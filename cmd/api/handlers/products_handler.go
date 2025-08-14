package handlers

import (
	"net/http"
	"strconv"

	"agro-mas-backend/internal/marketplace/products"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProductsHandler struct {
	productService *products.Service
	imageService   *products.ImageService
}

func NewProductsHandler(productService *products.Service, imageService *products.ImageService) *ProductsHandler {
	return &ProductsHandler{
		productService: productService,
		imageService:   imageService,
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

	var req products.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
			"details": err.Error(),
		})
		return
	}

	// Get seller information from context
	sellerInfo := products.SellerInfo{
		Name:              "Seller Name", // This would come from user service
		Phone:             "Seller Phone",
		Rating:            4.5,
		VerificationLevel: 2,
	}

	product, err := h.productService.CreateProduct(c.Request.Context(), userID.(uuid.UUID), &req, sellerInfo)
	if err != nil {
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

	c.JSON(http.StatusOK, gin.H{
		"product": product,
	})
}

// SearchProducts handles product search
func (h *ProductsHandler) SearchProducts(c *gin.Context) {
	req := &products.ProductSearchRequest{
		Query:             c.Query("query"),
		Category:          c.Query("category"),
		Subcategory:       c.Query("subcategory"),
		Province:          c.Query("province"),
		City:              c.Query("city"),
		PriceType:         c.Query("price_type"),
		SortBy:            c.Query("sort_by"),
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

	response, err := h.productService.SearchProducts(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to search products",
			"code":  "SEARCH_FAILED",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// UpdateProduct handles product updates
func (h *ProductsHandler) UpdateProduct(c *gin.Context) {
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

	var req products.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
			"details": err.Error(),
		})
		return
	}

	product, err := h.productService.UpdateProduct(c.Request.Context(), userID.(uuid.UUID), productID, &req)
	if err != nil {
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
			"error": "Failed to get user products",
			"code":  "USER_PRODUCTS_FETCH_FAILED",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
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
			"error": "Invalid form data",
			"code":  "INVALID_FORM_DATA",
			"details": err.Error(),
		})
		return
	}

	image, err := h.imageService.UploadProductImage(c.Request.Context(), userID.(uuid.UUID), file, header, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to upload image",
			"code":  "IMAGE_UPLOAD_FAILED",
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
func (h *ProductsHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware, sellerMiddleware gin.HandlerFunc) {
	products := router.Group("/products")
	{
		// Public routes
		products.GET("/search", h.SearchProducts)
		products.GET("/:id", h.GetProduct)

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