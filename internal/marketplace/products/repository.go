package products

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// CreateProduct creates a new product in the database
func (r *Repository) CreateProduct(ctx context.Context, product *Product) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert main product
	query := `
		INSERT INTO products (
			id, user_id, title, description, category, subcategory, price, price_type,
			currency, unit, quantity, available_from, available_until, is_active,
			is_featured, province, department, settlement, city, province_name, department_name, settlement_name, location_coordinates, pickup_available,
			delivery_available, delivery_radius, seller_name, seller_phone,
			seller_rating, seller_verification_level, search_keywords, metadata, tags
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22,
			CASE WHEN $23::float IS NOT NULL AND $24::float IS NOT NULL THEN POINT($23, $24) ELSE NULL END,
			$25, $26, $27, $28, $29, $30, $31, $32, $33, $34
		)`

	var lng, lat sql.NullFloat64
	if product.LocationCoordinates != nil {
		lng.Float64 = product.LocationCoordinates.Lng
		lng.Valid = true
		lat.Float64 = product.LocationCoordinates.Lat
		lat.Valid = true
	}

	var metadataJSON []byte
	if product.Metadata != nil {
		metadataJSON, err = json.Marshal(product.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	} else {
		metadataJSON = []byte("{}")
	}

	_, err = tx.ExecContext(ctx, query,
		product.ID, product.UserID, product.Title, product.Description,
		product.Category, product.Subcategory, product.Price, product.PriceType,
		product.Currency, product.Unit, product.Quantity, product.AvailableFrom,
		product.AvailableUntil, product.IsActive, product.IsFeatured,
		product.Province, product.Department, product.Settlement, product.City, product.ProvinceName, product.DepartmentName, product.SettlementName, lng, lat, product.PickupAvailable,
		product.DeliveryAvailable, product.DeliveryRadius, product.SellerName,
		product.SellerPhone, product.SellerRating, product.SellerVerificationLevel,
		product.SearchKeywords, metadataJSON, pq.Array(product.Tags))

	if err != nil {
		return fmt.Errorf("failed to insert product: %w", err)
	}

	// Insert category-specific details
	switch product.Category {
	case "transport":
		if product.TransportDetails != nil {
			if err := r.insertTransportDetails(ctx, tx, product.TransportDetails); err != nil {
				return fmt.Errorf("failed to insert transport details: %w", err)
			}
		}
	case "livestock":
		if product.LivestockDetails != nil {
			if err := r.insertLivestockDetails(ctx, tx, product.LivestockDetails); err != nil {
				return fmt.Errorf("failed to insert livestock details: %w", err)
			}
		}
	case "supplies":
		if product.SuppliesDetails != nil {
			if err := r.insertSuppliesDetails(ctx, tx, product.SuppliesDetails); err != nil {
				return fmt.Errorf("failed to insert supplies details: %w", err)
			}
		}
	}

	return tx.Commit()
}

// GetProductByID retrieves a product by its ID
func (r *Repository) GetProductByID(ctx context.Context, id uuid.UUID) (*Product, error) {
	query := `
		SELECT 
			id, user_id, title, description, category, subcategory, price, price_type,
			currency, unit, quantity, available_from, available_until, is_active,
 			is_featured, province, department, settlement, city, province_name, department_name, settlement_name,
			CASE WHEN location_coordinates IS NOT NULL THEN location_coordinates[0] ELSE NULL END as lng,
			CASE WHEN location_coordinates IS NOT NULL THEN location_coordinates[1] ELSE NULL END as lat, 
			pickup_available, delivery_available,
			delivery_radius, seller_name, seller_phone, seller_rating,
			seller_verification_level, views_count, favorites_count, inquiries_count,
			search_keywords, created_at, updated_at, published_at, expires_at,
			metadata, tags
		FROM products 
		WHERE id = $1`

	product := &Product{}
	var lng, lat sql.NullFloat64
	var metadataJSON sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID, &product.UserID, &product.Title, &product.Description,
		&product.Category, &product.Subcategory, &product.Price, &product.PriceType,
		&product.Currency, &product.Unit, &product.Quantity, &product.AvailableFrom,
		&product.AvailableUntil, &product.IsActive, &product.IsFeatured,
		&product.Province, &product.Department, &product.Settlement, &product.City, &product.ProvinceName, &product.DepartmentName, &product.SettlementName, &lng, &lat, &product.PickupAvailable,
		&product.DeliveryAvailable, &product.DeliveryRadius, &product.SellerName,
		&product.SellerPhone, &product.SellerRating, &product.SellerVerificationLevel,
		&product.ViewsCount, &product.FavoritesCount, &product.InquiriesCount,
		&product.SearchKeywords, &product.CreatedAt, &product.UpdatedAt,
		&product.PublishedAt, &product.ExpiresAt, &metadataJSON, pq.Array(&product.Tags))

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Parse coordinates
	if lng.Valid && lat.Valid {
		product.LocationCoordinates = &Point{
			Lng: lng.Float64,
			Lat: lat.Float64,
		}
	}

	// Parse metadata
	if metadataJSON.Valid && metadataJSON.String != "" {
		if err := json.Unmarshal([]byte(metadataJSON.String), &product.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	// Load category-specific details and images
	if err := r.loadProductDetails(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to load product details: %w", err)
	}

	return product, nil
}

// SearchProducts searches for products with filters
func (r *Repository) SearchProducts(ctx context.Context, req *ProductSearchRequest) ([]*Product, int, error) {
	whereConditions := []string{"p.is_active = true"}
	args := []interface{}{}
	argIndex := 1

	// Add filters
	if req.Query != "" {
		whereConditions = append(whereConditions,
			fmt.Sprintf("to_tsvector('spanish', p.title || ' ' || COALESCE(p.description, '') || ' ' || COALESCE(p.search_keywords, '')) @@ plainto_tsquery('spanish', $%d)", argIndex))
		args = append(args, req.Query)
		argIndex++
	}

	if req.Category != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("p.category = $%d", argIndex))
		args = append(args, req.Category)
		argIndex++
	}

	if req.Subcategory != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("p.subcategory = $%d", argIndex))
		args = append(args, req.Subcategory)
		argIndex++
	}

	if req.Province != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("p.province = $%d", argIndex))
		args = append(args, req.Province)
		argIndex++
	}

	if req.City != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("p.city = $%d", argIndex))
		args = append(args, req.City)
		argIndex++
	}

	if req.MinPrice != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("p.price >= $%d", argIndex))
		args = append(args, *req.MinPrice)
		argIndex++
	}

	if req.MaxPrice != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("p.price <= $%d", argIndex))
		args = append(args, *req.MaxPrice)
		argIndex++
	}

	if req.PriceType != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("p.price_type = $%d", argIndex))
		args = append(args, req.PriceType)
		argIndex++
	}

	if req.PickupAvailable != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("p.pickup_available = $%d", argIndex))
		args = append(args, *req.PickupAvailable)
		argIndex++
	}

	if req.DeliveryAvailable != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("p.delivery_available = $%d", argIndex))
		args = append(args, *req.DeliveryAvailable)
		argIndex++
	}

	if req.IsVerifiedSeller != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("u.is_verified = $%d", argIndex))
		args = append(args, *req.IsVerifiedSeller)
		argIndex++
	}

	if len(req.Tags) > 0 {
		whereConditions = append(whereConditions, fmt.Sprintf("p.tags && $%d", argIndex))
		args = append(args, pq.Array(req.Tags))
		argIndex++
	}

	if req.ExcludeUserID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("p.user_id != $%d", argIndex))
		args = append(args, *req.ExcludeUserID)
		argIndex++
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// Count total results
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM products p
		LEFT JOIN users u ON p.user_id = u.id
		WHERE %s`, whereClause)

	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	// Determine sorting
	orderBy := "p.created_at DESC"
	switch req.SortBy {
	case "price_asc":
		orderBy = "p.price ASC NULLS LAST"
	case "price_desc":
		orderBy = "p.price DESC NULLS LAST"
	case "date_asc":
		orderBy = "p.created_at ASC"
	case "date_desc":
		orderBy = "p.created_at DESC"
	case "rating":
		orderBy = "p.seller_rating DESC NULLS LAST"
	case "relevance":
		if req.Query != "" {
			orderBy = "ts_rank(to_tsvector('spanish', p.title || ' ' || COALESCE(p.description, '') || ' ' || COALESCE(p.search_keywords, '')), plainto_tsquery('spanish', $1)) DESC"
		}
	}

	// Set pagination defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	offset := (req.Page - 1) * req.PageSize

	// Get products
	query := fmt.Sprintf(`
		SELECT 
			p.id, p.user_id, p.title, p.description, p.category, p.subcategory,
			p.price, p.price_type, p.currency, p.unit, p.quantity, p.available_from,
			p.available_until, p.is_active, p.is_featured, p.province, p.city, p.province_name, p.department_name, p.settlement_name,
			CASE WHEN p.location_coordinates IS NOT NULL THEN p.location_coordinates[0] ELSE NULL END as lng,
			CASE WHEN p.location_coordinates IS NOT NULL THEN p.location_coordinates[1] ELSE NULL END as lat,
			p.pickup_available, p.delivery_available, p.delivery_radius,
			p.seller_name, p.seller_phone, p.seller_rating, p.seller_verification_level,
			p.views_count, p.favorites_count, p.inquiries_count, p.search_keywords,
			p.created_at, p.updated_at, p.published_at, p.expires_at, p.metadata, p.tags
		FROM products p
		LEFT JOIN users u ON p.user_id = u.id
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d`, whereClause, orderBy, argIndex, argIndex+1)

	args = append(args, req.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search products: %w", err)
	}
	defer rows.Close()

	products := make([]*Product, 0)
	for rows.Next() {
		product := &Product{}
		var lng, lat sql.NullFloat64
		var metadataJSON sql.NullString

		err := rows.Scan(
			&product.ID, &product.UserID, &product.Title, &product.Description,
			&product.Category, &product.Subcategory, &product.Price, &product.PriceType,
			&product.Currency, &product.Unit, &product.Quantity, &product.AvailableFrom,
			&product.AvailableUntil, &product.IsActive, &product.IsFeatured,
			&product.Province, &product.City, &product.ProvinceName, &product.DepartmentName, &product.SettlementName, &lng, &lat, &product.PickupAvailable,
			&product.DeliveryAvailable, &product.DeliveryRadius, &product.SellerName,
			&product.SellerPhone, &product.SellerRating, &product.SellerVerificationLevel,
			&product.ViewsCount, &product.FavoritesCount, &product.InquiriesCount,
			&product.SearchKeywords, &product.CreatedAt, &product.UpdatedAt,
			&product.PublishedAt, &product.ExpiresAt, &metadataJSON, pq.Array(&product.Tags))

		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan product: %w", err)
		}

		// Parse coordinates
		if lng.Valid && lat.Valid {
			product.LocationCoordinates = &Point{
				Lng: lng.Float64,
				Lat: lat.Float64,
			}
		}

		// Parse metadata
		if metadataJSON.Valid && metadataJSON.String != "" {
			if err := json.Unmarshal([]byte(metadataJSON.String), &product.Metadata); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		products = append(products, product)
	}

	// Load details for all products (in batches for performance)
	for _, product := range products {
		if err := r.loadProductDetails(ctx, product); err != nil {
			return nil, 0, fmt.Errorf("failed to load product details: %w", err)
		}
	}

	return products, totalCount, nil
}

// Helper methods for category-specific details
func (r *Repository) insertTransportDetails(ctx context.Context, tx *sql.Tx, details *TransportDetails) error {
	query := `
		INSERT INTO transport_details (
			product_id, vehicle_type, capacity_tons, capacity_cubic_meters,
			price_per_km, has_refrigeration, has_livestock_equipment,
			service_provinces, min_distance_km, max_distance_km,
			license_plate, license_expiry, insurance_expiry, vehicle_year
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	_, err := tx.ExecContext(ctx, query,
		details.ProductID, details.VehicleType, details.CapacityTons,
		details.CapacityCubicMeters, details.PricePerKm, details.HasRefrigeration,
		details.HasLivestockEquipment, pq.Array(details.ServiceProvinces),
		details.MinDistanceKm, details.MaxDistanceKm, details.LicensePlate,
		details.LicenseExpiry, details.InsuranceExpiry, details.VehicleYear)

	return err
}

func (r *Repository) insertLivestockDetails(ctx context.Context, tx *sql.Tx, details *LivestockDetails) error {
	var vaccinationsJSON, breedingHistoryJSON sql.NullString

	if details.Vaccinations != nil {
		jsonData, err := json.Marshal(details.Vaccinations)
		if err != nil {
			return fmt.Errorf("failed to marshal vaccinations: %w", err)
		}
		vaccinationsJSON.String = string(jsonData)
		vaccinationsJSON.Valid = true
	}

	if details.BreedingHistory != nil {
		jsonData, err := json.Marshal(details.BreedingHistory)
		if err != nil {
			return fmt.Errorf("failed to marshal breeding history: %w", err)
		}
		breedingHistoryJSON.String = string(jsonData)
		breedingHistoryJSON.Valid = true
	}

	query := `
		INSERT INTO livestock_details (
			product_id, animal_type, breed, age_months, weight_kg, gender,
			health_certificates, vaccinations, last_veterinary_check,
			is_organic, is_pregnant, breeding_history, genetic_information
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err := tx.ExecContext(ctx, query,
		details.ProductID, details.AnimalType, details.Breed, details.AgeMonths,
		details.WeightKg, details.Gender, pq.Array(details.HealthCertificates),
		vaccinationsJSON, details.LastVeterinaryCheck, details.IsOrganic,
		details.IsPregnant, breedingHistoryJSON, details.GeneticInformation)

	return err
}

func (r *Repository) insertSuppliesDetails(ctx context.Context, tx *sql.Tx, details *SuppliesDetails) error {
	query := `
		INSERT INTO supplies_details (
			product_id, supply_type, brand, model, active_ingredients,
			concentration, expiry_date, batch_number, registration_number,
			required_licenses, safety_data_sheet_url, storage_requirements,
			handling_instructions, disposal_instructions
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	_, err := tx.ExecContext(ctx, query,
		details.ProductID, details.SupplyType, details.Brand, details.Model,
		pq.Array(details.ActiveIngredients), details.Concentration,
		details.ExpiryDate, details.BatchNumber, details.RegistrationNumber,
		pq.Array(details.RequiredLicenses), details.SafetyDataSheetURL,
		details.StorageRequirements, details.HandlingInstructions,
		details.DisposalInstructions)

	return err
}

func (r *Repository) loadProductDetails(ctx context.Context, product *Product) error {
	// Load images
	images, err := r.getProductImages(ctx, product.ID)
	if err != nil {
		return fmt.Errorf("failed to load product images: %w", err)
	}
	product.Images = images

	// Load category-specific details
	switch product.Category {
	case "transport":
		details, err := r.getTransportDetails(ctx, product.ID)
		if err != nil {
			return fmt.Errorf("failed to load transport details: %w", err)
		}
		product.TransportDetails = details
	case "livestock":
		details, err := r.getLivestockDetails(ctx, product.ID)
		if err != nil {
			return fmt.Errorf("failed to load livestock details: %w", err)
		}
		product.LivestockDetails = details
	case "supplies":
		details, err := r.getSuppliesDetails(ctx, product.ID)
		if err != nil {
			return fmt.Errorf("failed to load supplies details: %w", err)
		}
		product.SuppliesDetails = details
	}

	return nil
}

func (r *Repository) getProductImages(ctx context.Context, productID uuid.UUID) ([]ProductImage, error) {
	query := `
		SELECT id, product_id, image_url, cloud_storage_path, alt_text,
			   is_primary, display_order, file_size, mime_type, uploaded_at
		FROM product_images 
		WHERE product_id = $1 
		ORDER BY is_primary DESC, display_order ASC`

	rows, err := r.db.QueryContext(ctx, query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	images := make([]ProductImage, 0)
	for rows.Next() {
		img := ProductImage{}
		err := rows.Scan(&img.ID, &img.ProductID, &img.ImageURL, &img.CloudStoragePath,
			&img.AltText, &img.IsPrimary, &img.DisplayOrder, &img.FileSize,
			&img.MimeType, &img.UploadedAt)
		if err != nil {
			return nil, err
		}
		images = append(images, img)
	}

	return images, nil
}

func (r *Repository) getTransportDetails(ctx context.Context, productID uuid.UUID) (*TransportDetails, error) {
	query := `
		SELECT product_id, vehicle_type, capacity_tons, capacity_cubic_meters,
			   price_per_km, has_refrigeration, has_livestock_equipment,
			   service_provinces, min_distance_km, max_distance_km,
			   license_plate, license_expiry, insurance_expiry, vehicle_year,
			   created_at, updated_at
		FROM transport_details WHERE product_id = $1`

	details := &TransportDetails{}
	err := r.db.QueryRowContext(ctx, query, productID).Scan(
		&details.ProductID, &details.VehicleType, &details.CapacityTons,
		&details.CapacityCubicMeters, &details.PricePerKm, &details.HasRefrigeration,
		&details.HasLivestockEquipment, pq.Array(&details.ServiceProvinces),
		&details.MinDistanceKm, &details.MaxDistanceKm, &details.LicensePlate,
		&details.LicenseExpiry, &details.InsuranceExpiry, &details.VehicleYear,
		&details.CreatedAt, &details.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return details, nil
}

func (r *Repository) getLivestockDetails(ctx context.Context, productID uuid.UUID) (*LivestockDetails, error) {
	query := `
		SELECT product_id, animal_type, breed, age_months, weight_kg, gender,
			   health_certificates, vaccinations, last_veterinary_check,
			   is_organic, is_pregnant, breeding_history, genetic_information,
			   created_at, updated_at
		FROM livestock_details WHERE product_id = $1`

	details := &LivestockDetails{}
	var vaccinationsJSON, breedingHistoryJSON sql.NullString

	err := r.db.QueryRowContext(ctx, query, productID).Scan(
		&details.ProductID, &details.AnimalType, &details.Breed, &details.AgeMonths,
		&details.WeightKg, &details.Gender, pq.Array(&details.HealthCertificates),
		&vaccinationsJSON, &details.LastVeterinaryCheck, &details.IsOrganic,
		&details.IsPregnant, &breedingHistoryJSON, &details.GeneticInformation,
		&details.CreatedAt, &details.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Parse JSON fields
	if vaccinationsJSON.Valid && vaccinationsJSON.String != "" {
		if err := json.Unmarshal([]byte(vaccinationsJSON.String), &details.Vaccinations); err != nil {
			return nil, fmt.Errorf("failed to unmarshal vaccinations: %w", err)
		}
	}

	if breedingHistoryJSON.Valid && breedingHistoryJSON.String != "" {
		if err := json.Unmarshal([]byte(breedingHistoryJSON.String), &details.BreedingHistory); err != nil {
			return nil, fmt.Errorf("failed to unmarshal breeding history: %w", err)
		}
	}

	return details, nil
}

func (r *Repository) getSuppliesDetails(ctx context.Context, productID uuid.UUID) (*SuppliesDetails, error) {
	query := `
		SELECT product_id, supply_type, brand, model, active_ingredients,
			   concentration, expiry_date, batch_number, registration_number,
			   required_licenses, safety_data_sheet_url, storage_requirements,
			   handling_instructions, disposal_instructions, created_at, updated_at
		FROM supplies_details WHERE product_id = $1`

	details := &SuppliesDetails{}
	err := r.db.QueryRowContext(ctx, query, productID).Scan(
		&details.ProductID, &details.SupplyType, &details.Brand, &details.Model,
		pq.Array(&details.ActiveIngredients), &details.Concentration,
		&details.ExpiryDate, &details.BatchNumber, &details.RegistrationNumber,
		pq.Array(&details.RequiredLicenses), &details.SafetyDataSheetURL,
		&details.StorageRequirements, &details.HandlingInstructions,
		&details.DisposalInstructions, &details.CreatedAt, &details.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return details, nil
}

// UpdateProduct updates an existing product
func (r *Repository) UpdateProduct(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	fmt.Printf("[DEBUG] Repository UpdateProduct called with id: %v, updates: %+v\n", id, updates)
	if len(updates) == 0 {
		return nil
	}

	setParts := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)
	argIndex := 1

	for field, value := range updates {
		if field == "location_coordinates" && value != nil {
			if coords, ok := value.(*Point); ok {
				setParts = append(setParts, fmt.Sprintf("location_coordinates = POINT(%f, %f)", coords.Lng, coords.Lat))
				continue
			} else if coordsMap, ok := value.(map[string]interface{}); ok {
				if lat, latOk := coordsMap["lat"].(float64); latOk {
					if lng, lngOk := coordsMap["lng"].(float64); lngOk {
						setParts = append(setParts, fmt.Sprintf("location_coordinates = POINT(%f, %f)", lng, lat))
						continue
					}
				}
			}
		}
		// Handle tags field for PostgreSQL array
		if field == "tags" && value != nil {
			if tags, ok := value.([]string); ok {
				setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
				args = append(args, pq.Array(tags))
				argIndex++
				continue
			}
		}
		// Skip nil values for optional fields
		if value == nil {
			continue
		}
		setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
		args = append(args, value)
		argIndex++
	}

	setParts = append(setParts, "updated_at = NOW()")

	if len(setParts) == 1 { // Only "updated_at = NOW()" was added
		// No actual fields to update, just update the timestamp
		query := "UPDATE products SET updated_at = NOW() WHERE id = $1"
		_, err := r.db.ExecContext(ctx, query, id)
		if err != nil {
			fmt.Printf("[ERROR] SQL execution failed: %v\n", err)
			return fmt.Errorf("failed to update product: %w", err)
		}
		return nil
	}

	query := fmt.Sprintf("UPDATE products SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIndex)
	args = append(args, id)

	fmt.Printf("[DEBUG] Executing query: %s\n", query)
	fmt.Printf("[DEBUG] Args: %v\n", args)

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		fmt.Printf("[ERROR] SQL execution failed: %v\n", err)
		return fmt.Errorf("failed to update product: %w", err)
	}

	return nil
}

// DeleteProduct soft deletes a product
func (r *Repository) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE products SET is_active = false, updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}
	return nil
}

// IncrementViewsCount increments the views count for a product
func (r *Repository) IncrementViewsCount(ctx context.Context, productID uuid.UUID) error {
	query := `UPDATE products SET views_count = views_count + 1, updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, productID)
	return err
}

// GetProductsByUserID retrieves products belonging to a specific user (including unpublished ones)
func (r *Repository) GetProductsByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*Product, int, error) {
	offset := (page - 1) * pageSize

	// Count query
	countQuery := `SELECT COUNT(*) FROM products WHERE user_id = $1 AND is_active = true`
	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count user products: %w", err)
	}

	// Main query - get products without published filter
	query := `
		SELECT 
			p.id, p.user_id, p.title, p.description, p.category, p.subcategory,
			p.price, p.price_type, p.currency, p.unit, p.quantity,
			p.available_from, p.available_until, p.is_active, p.is_featured,
			p.province, p.city, p.province_name, p.department_name, p.settlement_name,
			CASE WHEN p.location_coordinates IS NOT NULL THEN p.location_coordinates[0] ELSE NULL END as lng,
			CASE WHEN p.location_coordinates IS NOT NULL THEN p.location_coordinates[1] ELSE NULL END as lat, 
			p.pickup_available, p.delivery_available, p.delivery_radius,
			p.seller_name, p.seller_phone, p.seller_rating, p.seller_verification_level,
			p.views_count, p.favorites_count, p.inquiries_count,
			p.search_keywords, p.created_at, p.updated_at, p.published_at, p.expires_at,
			p.metadata, p.tags
		FROM products p
		WHERE p.user_id = $1 AND p.is_active = true
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, userID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query user products: %w", err)
	}
	defer rows.Close()

	var products []*Product
	for rows.Next() {
		product := &Product{}
		var lng, lat sql.NullFloat64
		var metadataJSON []byte

		err := rows.Scan(
			&product.ID, &product.UserID, &product.Title, &product.Description,
			&product.Category, &product.Subcategory, &product.Price, &product.PriceType,
			&product.Currency, &product.Unit, &product.Quantity, &product.AvailableFrom,
			&product.AvailableUntil, &product.IsActive, &product.IsFeatured,
			&product.Province, &product.City, &product.ProvinceName, &product.DepartmentName, &product.SettlementName, &lng, &lat,
			&product.PickupAvailable, &product.DeliveryAvailable, &product.DeliveryRadius,
			&product.SellerName, &product.SellerPhone, &product.SellerRating,
			&product.SellerVerificationLevel, &product.ViewsCount, &product.FavoritesCount,
			&product.InquiriesCount, &product.SearchKeywords, &product.CreatedAt,
			&product.UpdatedAt, &product.PublishedAt, &product.ExpiresAt,
			&metadataJSON, pq.Array(&product.Tags))

		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user product: %w", err)
		}

		// Handle location coordinates
		if lng.Valid && lat.Valid {
			product.LocationCoordinates = &Point{
				Lng: lng.Float64,
				Lat: lat.Float64,
			}
		}

		// Handle metadata
		if len(metadataJSON) > 0 {
			var metadata ProductMetadata
			if err := json.Unmarshal(metadataJSON, &metadata); err == nil {
				product.Metadata = &metadata
			}
		}

		// Load category-specific details and images for each product
		if err := r.loadProductDetails(ctx, product); err != nil {
			return nil, 0, fmt.Errorf("failed to load product details: %w", err)
		}

		products = append(products, product)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate user products: %w", err)
	}

	return products, totalCount, nil
}
