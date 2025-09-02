package products

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrProductNotFound     = errors.New("product not found")
	ErrProductNotOwnedByUser = errors.New("product not owned by user")
	ErrInvalidCategory     = errors.New("invalid product category")
	ErrInvalidPriceType    = errors.New("invalid price type")
	ErrProductNotActive    = errors.New("product is not active")
)

type Service struct {
	repo         *Repository
	imageService *ImageService
}

func NewService(repo *Repository, imageService *ImageService) *Service {
	return &Service{
		repo:         repo,
		imageService: imageService,
	}
}

// CreateProduct creates a new product with validation
func (s *Service) CreateProduct(ctx context.Context, userID uuid.UUID, req *CreateProductRequest, sellerInfo SellerInfo) (*Product, error) {
	// Validate category
	if !isValidCategory(req.Category) {
		return nil, ErrInvalidCategory
	}

	// Validate price type
	if !isValidPriceType(req.PriceType) {
		return nil, ErrInvalidPriceType
	}

	// Validate category-specific details
	if err := s.validateCategoryDetails(req); err != nil {
		return nil, fmt.Errorf("category validation failed: %w", err)
	}

	// Generate search keywords
	searchKeywords := s.generateSearchKeywords(req)

	// Create product object
	product := &Product{
		ID:                      uuid.New(),
		UserID:                  userID,
		Title:                   req.Title,
		Description:             req.Description,
		Category:                req.Category,
		Subcategory:             req.Subcategory,
		Price:                   req.Price,
		PriceType:               req.PriceType,
		Currency:                "ARS", // Default to Argentine Peso
		Unit:                    req.Unit,
		Quantity:                req.Quantity,
		AvailableFrom:           req.AvailableFrom,
		AvailableUntil:          req.AvailableUntil,
		IsActive:                true,
		IsFeatured:              false,
		Province:                req.Province,
		City:                    req.City,
		LocationCoordinates:     req.LocationCoordinates,
		PickupAvailable:         req.PickupAvailable,
		DeliveryAvailable:       req.DeliveryAvailable,
		DeliveryRadius:          req.DeliveryRadius,
		SellerName:              &sellerInfo.Name,
		SellerPhone:             &sellerInfo.Phone,
		SellerRating:            &sellerInfo.Rating,
		SellerVerificationLevel: &sellerInfo.VerificationLevel,
		ViewsCount:              0,
		FavoritesCount:          0,
		InquiriesCount:          0,
		SearchKeywords:          &searchKeywords,
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
		PublishedAt:             func() *time.Time { t := time.Now(); return &t }(), // Set to current time to auto-publish
		Tags:                    req.Tags,
	}

	// Set category-specific details
	switch req.Category {
	case "transport":
		if req.TransportDetails != nil {
			req.TransportDetails.ProductID = product.ID
			req.TransportDetails.CreatedAt = time.Now()
			req.TransportDetails.UpdatedAt = time.Now()
			product.TransportDetails = req.TransportDetails
		}
	case "livestock":
		if req.LivestockDetails != nil {
			req.LivestockDetails.ProductID = product.ID
			req.LivestockDetails.CreatedAt = time.Now()
			req.LivestockDetails.UpdatedAt = time.Now()
			product.LivestockDetails = req.LivestockDetails
		}
	case "supplies":
		if req.SuppliesDetails != nil {
			req.SuppliesDetails.ProductID = product.ID
			req.SuppliesDetails.CreatedAt = time.Now()
			req.SuppliesDetails.UpdatedAt = time.Now()
			product.SuppliesDetails = req.SuppliesDetails
		}
	}

	// Create product in database
	if err := s.repo.CreateProduct(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to create product in database: %w", err)
	}

	return product, nil
}

// CreateProductWithImages creates a new product with images
func (s *Service) CreateProductWithImages(ctx context.Context, userID uuid.UUID, req *CreateProductRequest, sellerInfo SellerInfo, imageFiles []*multipart.FileHeader) (*Product, error) {
	// Create the product first
	product, err := s.CreateProduct(ctx, userID, req, sellerInfo)
	if err != nil {
		return nil, err
	}

	// Process images if provided
	if len(imageFiles) > 0 {
		for i, fileHeader := range imageFiles {
			file, err := fileHeader.Open()
			if err != nil {
				fmt.Printf("Failed to open image file %d: %v\n", i, err)
				continue
			}

			// Create upload request
			uploadReq := UploadImageRequest{
				ProductID:    product.ID,
				AltText:      fmt.Sprintf("Product image %d", i+1),
				IsPrimary:    i == 0, // First image is primary
				DisplayOrder: i + 1,
			}

			// Upload image
			productImage, err := s.imageService.UploadProductImage(ctx, userID, file, fileHeader, uploadReq)
			if err != nil {
				fmt.Printf("Failed to upload image %d: %v\n", i, err)
				file.Close()
				continue
			}

			// Add image to product
			product.Images = append(product.Images, *productImage)
			file.Close()
		}
	}

	return product, nil
}

// GetProductByID retrieves a product by its ID and increments view count
func (s *Service) GetProductByID(ctx context.Context, id uuid.UUID, incrementView bool) (*Product, error) {
	product, err := s.repo.GetProductByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	if product == nil {
		return nil, ErrProductNotFound
	}

	// Increment view count if requested and product is active
	if incrementView && product.IsActive {
		if err := s.repo.IncrementViewsCount(ctx, id); err != nil {
			// Log error but don't fail the request
			fmt.Printf("Failed to increment view count for product %s: %v\n", id, err)
		}
	}

	return product, nil
}

// SearchProducts searches for products with filters and pagination
func (s *Service) SearchProducts(ctx context.Context, req *ProductSearchRequest) (*ProductListResponse, error) {
	// Set default values
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	// Perform search
	products, totalCount, err := s.repo.SearchProducts(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to search products: %w", err)
	}

	// Calculate total pages
	totalPages := (totalCount + req.PageSize - 1) / req.PageSize

	// Convert to slice of Product structs instead of pointers for response
	productList := make([]Product, len(products))
	for i, p := range products {
		productList[i] = *p
	}

	return &ProductListResponse{
		Products:   productList,
		TotalCount: totalCount,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateProduct updates an existing product
func (s *Service) UpdateProduct(ctx context.Context, userID, productID uuid.UUID, req *UpdateProductRequest) (*Product, error) {
	fmt.Printf("[DEBUG] UpdateProduct called with userID: %v, productID: %v\n", userID, productID)
	fmt.Printf("[DEBUG] Update request: %+v\n", req)
	
	// Get existing product
	existingProduct, err := s.repo.GetProductByID(ctx, productID)
	if err != nil {
		fmt.Printf("[ERROR] Failed to get existing product: %v\n", err)
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	if existingProduct == nil {
		return nil, ErrProductNotFound
	}

	// Check ownership
	if existingProduct.UserID != userID {
		return nil, ErrProductNotOwnedByUser
	}

	// Prepare updates map
	updates := make(map[string]interface{})

	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Subcategory != nil {
		updates["subcategory"] = *req.Subcategory
	}
	if req.Price != nil {
		updates["price"] = *req.Price
	}
	if req.PriceType != nil {
		if !isValidPriceType(*req.PriceType) {
			return nil, ErrInvalidPriceType
		}
		updates["price_type"] = *req.PriceType
	}
	if req.Unit != nil {
		updates["unit"] = *req.Unit
	}
	if req.Quantity != nil {
		updates["quantity"] = *req.Quantity
	}
	if req.AvailableFrom != nil {
		updates["available_from"] = *req.AvailableFrom
	}
	if req.AvailableUntil != nil {
		updates["available_until"] = *req.AvailableUntil
	}
	if req.Province != nil {
		updates["province"] = *req.Province
	}
	if req.Department != nil {
		updates["department"] = *req.Department
	}
	if req.Settlement != nil {
		updates["settlement"] = *req.Settlement
	}
	if req.City != nil {
		updates["city"] = *req.City
	}
	if req.LocationCoordinates != nil {
		updates["location_coordinates"] = req.LocationCoordinates
	}
	if req.PickupAvailable != nil {
		updates["pickup_available"] = *req.PickupAvailable
	}
	if req.DeliveryAvailable != nil {
		updates["delivery_available"] = *req.DeliveryAvailable
	}
	if req.DeliveryRadius != nil {
		updates["delivery_radius"] = *req.DeliveryRadius
	}
	if req.Tags != nil {
		updates["tags"] = req.Tags
	}

	// Update search keywords if title or description changed
	if req.Title != nil || req.Description != nil {
		// Create a temporary request to generate new keywords
		tempReq := &CreateProductRequest{
			Title:       getStringValue(req.Title, existingProduct.Title),
			Description: getStringPtr(req.Description, existingProduct.Description),
			Category:    existingProduct.Category,
			Subcategory: getStringPtr(req.Subcategory, existingProduct.Subcategory),
			Tags:        getSliceValue(req.Tags, existingProduct.Tags),
		}
		searchKeywords := s.generateSearchKeywords(tempReq)
		updates["search_keywords"] = searchKeywords
	}

	// Update product in database
	if err := s.repo.UpdateProduct(ctx, productID, updates); err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	// Return updated product
	return s.repo.GetProductByID(ctx, productID)
}

// UpdateProductWithImages updates an existing product and processes new images
func (s *Service) UpdateProductWithImages(ctx context.Context, userID, productID uuid.UUID, req *UpdateProductRequest, imageFiles []*multipart.FileHeader) (*Product, error) {
	// Update the product first
	product, err := s.UpdateProduct(ctx, userID, productID, req)
	if err != nil {
		return nil, err
	}

	// Process images if provided
	if len(imageFiles) > 0 {
		for i, fileHeader := range imageFiles {
			file, err := fileHeader.Open()
			if err != nil {
				fmt.Printf("Failed to open image file %d: %v\n", i, err)
				continue
			}

			// Create upload request
			uploadReq := UploadImageRequest{
				ProductID:    product.ID,
				AltText:      fmt.Sprintf("Product image %d", i+1),
				IsPrimary:    false, // For updates, don't set as primary by default
				DisplayOrder: i + 1,
			}

			// Upload the image
			_, err = s.imageService.UploadProductImage(ctx, userID, file, fileHeader, uploadReq)
			if err != nil {
				fmt.Printf("Failed to upload image %d: %v\n", i, err)
				continue
			}

			file.Close()
		}
	}

	// Return the updated product with fresh images
	return s.repo.GetProductByID(ctx, productID)
}

// PublishProduct publishes a product to make it visible in searches
func (s *Service) PublishProduct(ctx context.Context, userID, productID uuid.UUID) error {
	// Get existing product
	existingProduct, err := s.repo.GetProductByID(ctx, productID)
	if err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}
	if existingProduct == nil {
		return ErrProductNotFound
	}

	// Check ownership
	if existingProduct.UserID != userID {
		return ErrProductNotOwnedByUser
	}

	// Update published_at timestamp
	updates := map[string]interface{}{
		"published_at": time.Now(),
	}

	return s.repo.UpdateProduct(ctx, productID, updates)
}

// UnpublishProduct unpublishes a product to hide it from searches
func (s *Service) UnpublishProduct(ctx context.Context, userID, productID uuid.UUID) error {
	// Get existing product
	existingProduct, err := s.repo.GetProductByID(ctx, productID)
	if err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}
	if existingProduct == nil {
		return ErrProductNotFound
	}

	// Check ownership
	if existingProduct.UserID != userID {
		return ErrProductNotOwnedByUser
	}

	// Set published_at to null
	updates := map[string]interface{}{
		"published_at": nil,
	}

	return s.repo.UpdateProduct(ctx, productID, updates)
}

// DeleteProduct soft deletes a product
func (s *Service) DeleteProduct(ctx context.Context, userID, productID uuid.UUID) error {
	// Get existing product
	existingProduct, err := s.repo.GetProductByID(ctx, productID)
	if err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}
	if existingProduct == nil {
		return ErrProductNotFound
	}

	// Check ownership
	if existingProduct.UserID != userID {
		return ErrProductNotOwnedByUser
	}

	return s.repo.DeleteProduct(ctx, productID)
}

// GetUserProducts retrieves products belonging to a specific user
func (s *Service) GetUserProducts(ctx context.Context, userID uuid.UUID, page, pageSize int) (*ProductListResponse, error) {
	// Use dedicated method to get user products without published filter
	userProducts, totalCount, err := s.repo.GetProductsByUserID(ctx, userID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get user products: %w", err)
	}

	// Convert to response format
	productList := make([]Product, len(userProducts))
	for i, p := range userProducts {
		productList[i] = *p
	}

	return &ProductListResponse{
		Products:   productList,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: (totalCount + pageSize - 1) / pageSize,
	}, nil
}

// Helper functions
func isValidCategory(category string) bool {
	validCategories := []string{"transport", "livestock", "supplies"}
	for _, valid := range validCategories {
		if category == valid {
			return true
		}
	}
	return false
}

func isValidPriceType(priceType string) bool {
	validTypes := []string{"fixed", "negotiable", "per_unit", "quote"}
	for _, valid := range validTypes {
		if priceType == valid {
			return true
		}
	}
	return false
}

func (s *Service) validateCategoryDetails(req *CreateProductRequest) error {
	switch req.Category {
	case "transport":
		if req.TransportDetails == nil {
			return errors.New("transport details are required for transport category")
		}
		// Add more specific transport validation here
	case "livestock":
		if req.LivestockDetails == nil {
			return errors.New("livestock details are required for livestock category")
		}
		// Add more specific livestock validation here
	case "supplies":
		if req.SuppliesDetails == nil {
			return errors.New("supplies details are required for supplies category")
		}
		// Add more specific supplies validation here
	}
	return nil
}

func (s *Service) generateSearchKeywords(req *CreateProductRequest) string {
	keywords := []string{
		req.Title,
	}

	if req.Description != nil {
		keywords = append(keywords, *req.Description)
	}

	if req.Subcategory != nil {
		keywords = append(keywords, *req.Subcategory)
	}

	keywords = append(keywords, req.Category)
	keywords = append(keywords, req.Tags...)

	// Add category-specific keywords
	switch req.Category {
	case "transport":
		if req.TransportDetails != nil && req.TransportDetails.VehicleType != nil {
			keywords = append(keywords, *req.TransportDetails.VehicleType)
		}
	case "livestock":
		if req.LivestockDetails != nil {
			if req.LivestockDetails.AnimalType != nil {
				keywords = append(keywords, *req.LivestockDetails.AnimalType)
			}
			if req.LivestockDetails.Breed != nil {
				keywords = append(keywords, *req.LivestockDetails.Breed)
			}
		}
	case "supplies":
		if req.SuppliesDetails != nil {
			if req.SuppliesDetails.SupplyType != nil {
				keywords = append(keywords, *req.SuppliesDetails.SupplyType)
			}
			if req.SuppliesDetails.Brand != nil {
				keywords = append(keywords, *req.SuppliesDetails.Brand)
			}
		}
	}

	return strings.Join(keywords, " ")
}

// Helper types and functions
type SellerInfo struct {
	Name              string
	Phone             string
	Rating            float64
	VerificationLevel int
}

func getStringValue(ptr *string, defaultValue string) string {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

func getStringPtr(ptr *string, defaultPtr *string) *string {
	if ptr != nil {
		return ptr
	}
	return defaultPtr
}

func getSliceValue(slice []string, defaultSlice []string) []string {
	if slice != nil {
		return slice
	}
	return defaultSlice
}