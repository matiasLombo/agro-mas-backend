package products

import (
	"context"
	"database/sql"
	"fmt"
	"mime/multipart"
	"strings"

	"agro-mas-backend/pkg/gcloud"
	"github.com/google/uuid"
)

type ImageService struct {
	db            *sql.DB
	storageClient *gcloud.StorageClient
}

type UploadImageRequest struct {
	ProductID    uuid.UUID `form:"product_id" binding:"required"`
	AltText      string    `form:"alt_text"`
	IsPrimary    bool      `form:"is_primary"`
	DisplayOrder int       `form:"display_order"`
}

type UploadImageResponse struct {
	Image ProductImage `json:"image"`
}

func NewImageService(db *sql.DB, storageClient *gcloud.StorageClient) *ImageService {
	return &ImageService{
		db:            db,
		storageClient: storageClient,
	}
}

// UploadProductImage uploads an image for a product
func (s *ImageService) UploadProductImage(ctx context.Context, userID uuid.UUID, file multipart.File, header *multipart.FileHeader, req UploadImageRequest) (*ProductImage, error) {
	// Validate image file
	if err := gcloud.ValidateImageFile(header); err != nil {
		return nil, fmt.Errorf("image validation failed: %w", err)
	}

	// Check if product exists and user owns it
	if err := s.validateProductOwnership(ctx, userID, req.ProductID); err != nil {
		return nil, err
	}

	// If this is set as primary, remove primary flag from other images
	if req.IsPrimary {
		if err := s.removePrimaryFlag(ctx, req.ProductID); err != nil {
			return nil, fmt.Errorf("failed to remove primary flag from other images: %w", err)
		}
	}

	// Set display order if not provided
	if req.DisplayOrder == 0 {
		req.DisplayOrder = s.getNextDisplayOrder(ctx, req.ProductID)
	}

	// Upload to Cloud Storage
	uploadOptions := gcloud.UploadOptions{
		Directory:    "products",
		SubDirectory: req.ProductID.String(),
		PublicRead:   true,
		CacheControl: "public, max-age=31536000", // 1 year for product images
		Metadata: map[string]string{
			"product_id": req.ProductID.String(),
			"user_id":    userID.String(),
			"alt_text":   req.AltText,
		},
	}

	uploadResult, err := s.storageClient.UploadFile(ctx, file, header, uploadOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to upload image to storage: %w", err)
	}

	// Create product image record
	productImage := &ProductImage{
		ID:               uuid.New(),
		ProductID:        req.ProductID,
		ImageURL:         uploadResult.URL,
		CloudStoragePath: uploadResult.StoragePath,
		AltText:          &req.AltText,
		IsPrimary:        req.IsPrimary,
		DisplayOrder:     req.DisplayOrder,
		FileSize:         func() *int { size := int(uploadResult.FileSize); return &size }(),
		MimeType:         &uploadResult.MimeType,
		UploadedAt:       uploadResult.UploadedAt,
	}

	// Save to database
	if err := s.createProductImage(ctx, productImage); err != nil {
		// If database save fails, clean up uploaded file
		if deleteErr := s.storageClient.DeleteFile(ctx, uploadResult.StoragePath); deleteErr != nil {
			fmt.Printf("Failed to clean up uploaded file after database error: %v\n", deleteErr)
		}
		return nil, fmt.Errorf("failed to save image to database: %w", err)
	}

	return productImage, nil
}

// UpdateProductImage updates an existing product image
func (s *ImageService) UpdateProductImage(ctx context.Context, userID, imageID uuid.UUID, updates map[string]interface{}) error {
	// Get existing image
	image, err := s.getProductImageByID(ctx, imageID)
	if err != nil {
		return fmt.Errorf("failed to get image: %w", err)
	}
	if image == nil {
		return fmt.Errorf("image not found")
	}

	// Check product ownership
	if err := s.validateProductOwnership(ctx, userID, image.ProductID); err != nil {
		return err
	}

	// If setting as primary, remove primary flag from other images
	if isPrimary, exists := updates["is_primary"]; exists && isPrimary.(bool) {
		if err := s.removePrimaryFlag(ctx, image.ProductID); err != nil {
			return fmt.Errorf("failed to remove primary flag from other images: %w", err)
		}
	}

	// Update image in database
	return s.updateProductImage(ctx, imageID, updates)
}

// DeleteProductImage deletes a product image
func (s *ImageService) DeleteProductImage(ctx context.Context, userID, imageID uuid.UUID) error {
	// Get existing image
	image, err := s.getProductImageByID(ctx, imageID)
	if err != nil {
		return fmt.Errorf("failed to get image: %w", err)
	}
	if image == nil {
		return fmt.Errorf("image not found")
	}

	// Check product ownership
	if err := s.validateProductOwnership(ctx, userID, image.ProductID); err != nil {
		return err
	}

	// Delete from Cloud Storage
	if err := s.storageClient.DeleteFile(ctx, image.CloudStoragePath); err != nil {
		fmt.Printf("Failed to delete file from storage: %v\n", err)
		// Continue with database deletion even if storage deletion fails
	}

	// Delete from database
	if err := s.deleteProductImage(ctx, imageID); err != nil {
		return fmt.Errorf("failed to delete image from database: %w", err)
	}

	// If this was the primary image, set another image as primary
	if image.IsPrimary {
		if err := s.setPrimaryImageIfNeeded(ctx, image.ProductID); err != nil {
			fmt.Printf("Failed to set new primary image: %v\n", err)
		}
	}

	return nil
}

// GetProductImages retrieves all images for a product
func (s *ImageService) GetProductImages(ctx context.Context, productID uuid.UUID) ([]ProductImage, error) {
	query := `
		SELECT id, product_id, image_url, cloud_storage_path, alt_text,
			   is_primary, display_order, file_size, mime_type, uploaded_at
		FROM product_images 
		WHERE product_id = $1 
		ORDER BY is_primary DESC, display_order ASC`

	rows, err := s.db.QueryContext(ctx, query, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to query product images: %w", err)
	}
	defer rows.Close()

	images := make([]ProductImage, 0)
	for rows.Next() {
		img := ProductImage{}
		err := rows.Scan(&img.ID, &img.ProductID, &img.ImageURL, &img.CloudStoragePath,
			&img.AltText, &img.IsPrimary, &img.DisplayOrder, &img.FileSize,
			&img.MimeType, &img.UploadedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan image row: %w", err)
		}
		images = append(images, img)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate image rows: %w", err)
	}

	return images, nil
}

// GetResizedImageURL generates a URL for a resized image
func (s *ImageService) GetResizedImageURL(storagePath string, width, height, quality int) string {
	return s.storageClient.GenerateResizedImageURL(storagePath, width, height, quality)
}

// Helper methods
func (s *ImageService) validateProductOwnership(ctx context.Context, userID, productID uuid.UUID) error {
	query := `SELECT user_id FROM products WHERE id = $1 AND is_active = true`
	var ownerID uuid.UUID
	err := s.db.QueryRowContext(ctx, query, productID).Scan(&ownerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("product not found")
		}
		return fmt.Errorf("failed to check product ownership: %w", err)
	}

	if ownerID != userID {
		return fmt.Errorf("user does not own this product")
	}

	return nil
}

func (s *ImageService) removePrimaryFlag(ctx context.Context, productID uuid.UUID) error {
	query := `UPDATE product_images SET is_primary = false WHERE product_id = $1 AND is_primary = true`
	_, err := s.db.ExecContext(ctx, query, productID)
	return err
}

func (s *ImageService) getNextDisplayOrder(ctx context.Context, productID uuid.UUID) int {
	query := `SELECT COALESCE(MAX(display_order), 0) + 1 FROM product_images WHERE product_id = $1`
	var nextOrder int
	err := s.db.QueryRowContext(ctx, query, productID).Scan(&nextOrder)
	if err != nil {
		return 1 // Default to 1 if query fails
	}
	return nextOrder
}

func (s *ImageService) createProductImage(ctx context.Context, image *ProductImage) error {
	query := `
		INSERT INTO product_images (
			id, product_id, image_url, cloud_storage_path, alt_text,
			is_primary, display_order, file_size, mime_type, uploaded_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := s.db.ExecContext(ctx, query,
		image.ID, image.ProductID, image.ImageURL, image.CloudStoragePath,
		image.AltText, image.IsPrimary, image.DisplayOrder, image.FileSize,
		image.MimeType, image.UploadedAt)

	return err
}

func (s *ImageService) getProductImageByID(ctx context.Context, imageID uuid.UUID) (*ProductImage, error) {
	query := `
		SELECT id, product_id, image_url, cloud_storage_path, alt_text,
			   is_primary, display_order, file_size, mime_type, uploaded_at
		FROM product_images 
		WHERE id = $1`

	image := &ProductImage{}
	err := s.db.QueryRowContext(ctx, query, imageID).Scan(
		&image.ID, &image.ProductID, &image.ImageURL, &image.CloudStoragePath,
		&image.AltText, &image.IsPrimary, &image.DisplayOrder, &image.FileSize,
		&image.MimeType, &image.UploadedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return image, nil
}

func (s *ImageService) updateProductImage(ctx context.Context, imageID uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	setParts := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)
	argIndex := 1

	for field, value := range updates {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
		args = append(args, value)
		argIndex++
	}

	query := fmt.Sprintf("UPDATE product_images SET %s WHERE id = $%d", 
		strings.Join(setParts, ", "), argIndex)
	args = append(args, imageID)

	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

func (s *ImageService) deleteProductImage(ctx context.Context, imageID uuid.UUID) error {
	query := `DELETE FROM product_images WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, imageID)
	return err
}

func (s *ImageService) setPrimaryImageIfNeeded(ctx context.Context, productID uuid.UUID) error {
	// Check if there are any images left
	query := `SELECT id FROM product_images WHERE product_id = $1 ORDER BY display_order ASC LIMIT 1`
	var imageID uuid.UUID
	err := s.db.QueryRowContext(ctx, query, productID).Scan(&imageID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil // No images left, nothing to do
		}
		return err
	}

	// Set the first image as primary
	updateQuery := `UPDATE product_images SET is_primary = true WHERE id = $1`
	_, err = s.db.ExecContext(ctx, updateQuery, imageID)
	return err
}