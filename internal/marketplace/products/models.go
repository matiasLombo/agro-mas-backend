package products

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID                      uuid.UUID           `json:"id" db:"id"`
	UserID                  uuid.UUID           `json:"user_id" db:"user_id"`
	Title                   string              `json:"title" db:"title"`
	Description             *string             `json:"description,omitempty" db:"description"`
	Category                string              `json:"category" db:"category"`
	Subcategory             *string             `json:"subcategory,omitempty" db:"subcategory"`
	Price                   *float64            `json:"price,omitempty" db:"price"`
	PriceType               string              `json:"price_type" db:"price_type"`
	Currency                string              `json:"currency" db:"currency"`
	Unit                    *string             `json:"unit,omitempty" db:"unit"`
	Quantity                *int                `json:"quantity,omitempty" db:"quantity"`
	AvailableFrom           *time.Time          `json:"available_from,omitempty" db:"available_from"`
	AvailableUntil          *time.Time          `json:"available_until,omitempty" db:"available_until"`
	IsActive                bool                `json:"is_active" db:"is_active"`
	IsFeatured              bool                `json:"is_featured" db:"is_featured"`
	Province                *string             `json:"province,omitempty" db:"province"`
	Department              *string             `json:"department,omitempty" db:"department"`
	Settlement              *string             `json:"settlement,omitempty" db:"settlement"`
	City                    *string             `json:"city,omitempty" db:"city"`
	ProvinceName            *string             `json:"province_name,omitempty" db:"province_name"`
	DepartmentName          *string             `json:"department_name,omitempty" db:"department_name"`
	SettlementName          *string             `json:"settlement_name,omitempty" db:"settlement_name"`
	LocationCoordinates     *Point              `json:"location_coordinates,omitempty" db:"location_coordinates"`
	PickupAvailable         bool                `json:"pickup_available" db:"pickup_available"`
	DeliveryAvailable       bool                `json:"delivery_available" db:"delivery_available"`
	DeliveryRadius          *int                `json:"delivery_radius,omitempty" db:"delivery_radius"`
	SellerName              *string             `json:"seller_name,omitempty" db:"seller_name"`
	SellerPhone             *string             `json:"seller_phone,omitempty" db:"seller_phone"`
	SellerRating            *float64            `json:"seller_rating,omitempty" db:"seller_rating"`
	SellerVerificationLevel *int                `json:"seller_verification_level,omitempty" db:"seller_verification_level"`
	ViewsCount              int                 `json:"views_count" db:"views_count"`
	FavoritesCount          int                 `json:"favorites_count" db:"favorites_count"`
	InquiriesCount          int                 `json:"inquiries_count" db:"inquiries_count"`
	SearchKeywords          *string             `json:"search_keywords,omitempty" db:"search_keywords"`
	CreatedAt               time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt               time.Time           `json:"updated_at" db:"updated_at"`
	PublishedAt             *time.Time          `json:"published_at,omitempty" db:"published_at"`
	ExpiresAt               *time.Time          `json:"expires_at,omitempty" db:"expires_at"`
	Metadata                *ProductMetadata    `json:"metadata,omitempty" db:"metadata"`
	Tags                    []string            `json:"tags,omitempty" db:"tags"`
	Images                  []ProductImage      `json:"images,omitempty"`
	TransportDetails        *TransportDetails   `json:"transport_details,omitempty"`
	LivestockDetails        *LivestockDetails   `json:"livestock_details,omitempty"`
	SuppliesDetails         *SuppliesDetails    `json:"supplies_details,omitempty"`
}

type Point struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type ProductMetadata struct {
	AdditionalInfo map[string]interface{} `json:"additional_info,omitempty"`
	SEOKeywords    []string               `json:"seo_keywords,omitempty"`
	InternalNotes  string                 `json:"internal_notes,omitempty"`
}

type ProductImage struct {
	ID               uuid.UUID `json:"id" db:"id"`
	ProductID        uuid.UUID `json:"product_id" db:"product_id"`
	ImageURL         string    `json:"image_url" db:"image_url"`
	CloudStoragePath string    `json:"cloud_storage_path" db:"cloud_storage_path"`
	AltText          *string   `json:"alt_text,omitempty" db:"alt_text"`
	IsPrimary        bool      `json:"is_primary" db:"is_primary"`
	DisplayOrder     int       `json:"display_order" db:"display_order"`
	FileSize         *int      `json:"file_size,omitempty" db:"file_size"`
	MimeType         *string   `json:"mime_type,omitempty" db:"mime_type"`
	UploadedAt       time.Time `json:"uploaded_at" db:"uploaded_at"`
}

type TransportDetails struct {
	ProductID             uuid.UUID   `json:"product_id" db:"product_id"`
	VehicleType           *string     `json:"vehicle_type,omitempty" db:"vehicle_type"`
	CapacityTons          *float64    `json:"capacity_tons,omitempty" db:"capacity_tons"`
	CapacityCubicMeters   *float64    `json:"capacity_cubic_meters,omitempty" db:"capacity_cubic_meters"`
	PricePerKm            *float64    `json:"price_per_km,omitempty" db:"price_per_km"`
	HasRefrigeration      bool        `json:"has_refrigeration" db:"has_refrigeration"`
	HasLivestockEquipment bool        `json:"has_livestock_equipment" db:"has_livestock_equipment"`
	ServiceProvinces      []string    `json:"service_provinces,omitempty" db:"service_provinces"`
	MinDistanceKm         *int        `json:"min_distance_km,omitempty" db:"min_distance_km"`
	MaxDistanceKm         *int        `json:"max_distance_km,omitempty" db:"max_distance_km"`
	LicensePlate          *string     `json:"license_plate,omitempty" db:"license_plate"`
	LicenseExpiry         *time.Time  `json:"license_expiry,omitempty" db:"license_expiry"`
	InsuranceExpiry       *time.Time  `json:"insurance_expiry,omitempty" db:"insurance_expiry"`
	VehicleYear           *int        `json:"vehicle_year,omitempty" db:"vehicle_year"`
	CreatedAt             time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time   `json:"updated_at" db:"updated_at"`
}

type LivestockDetails struct {
	ProductID              uuid.UUID                `json:"product_id" db:"product_id"`
	AnimalType             *string                  `json:"animal_type,omitempty" db:"animal_type"`
	Breed                  *string                  `json:"breed,omitempty" db:"breed"`
	AgeMonths              *int                     `json:"age_months,omitempty" db:"age_months"`
	WeightKg               *float64                 `json:"weight_kg,omitempty" db:"weight_kg"`
	Gender                 *string                  `json:"gender,omitempty" db:"gender"`
	HealthCertificates     []string                 `json:"health_certificates,omitempty" db:"health_certificates"`
	Vaccinations           *VaccinationRecords      `json:"vaccinations,omitempty" db:"vaccinations"`
	LastVeterinaryCheck    *time.Time               `json:"last_veterinary_check,omitempty" db:"last_veterinary_check"`
	IsOrganic              bool                     `json:"is_organic" db:"is_organic"`
	IsPregnant             *bool                    `json:"is_pregnant,omitempty" db:"is_pregnant"`
	BreedingHistory        *BreedingHistory         `json:"breeding_history,omitempty" db:"breeding_history"`
	GeneticInformation     *string                  `json:"genetic_information,omitempty" db:"genetic_information"`
	CreatedAt              time.Time                `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time                `json:"updated_at" db:"updated_at"`
}

type VaccinationRecords struct {
	Vaccines []VaccinationRecord `json:"vaccines"`
}

type VaccinationRecord struct {
	Name         string    `json:"name"`
	Date         time.Time `json:"date"`
	VetLicense   string    `json:"vet_license"`
	BatchNumber  string    `json:"batch_number,omitempty"`
	ExpiryDate   *time.Time `json:"expiry_date,omitempty"`
}

type BreedingHistory struct {
	TotalCalves     int                `json:"total_calves"`
	LastBreeding    *time.Time         `json:"last_breeding,omitempty"`
	BreedingRecords []BreedingRecord   `json:"breeding_records,omitempty"`
}

type BreedingRecord struct {
	Date        time.Time `json:"date"`
	SireID      *string   `json:"sire_id,omitempty"`
	Result      string    `json:"result"` // successful, failed, pending
	OffspringID *string   `json:"offspring_id,omitempty"`
}

type SuppliesDetails struct {
	ProductID             uuid.UUID   `json:"product_id" db:"product_id"`
	SupplyType            *string     `json:"supply_type,omitempty" db:"supply_type"`
	Brand                 *string     `json:"brand,omitempty" db:"brand"`
	Model                 *string     `json:"model,omitempty" db:"model"`
	ActiveIngredients     []string    `json:"active_ingredients,omitempty" db:"active_ingredients"`
	Concentration         *string     `json:"concentration,omitempty" db:"concentration"`
	ExpiryDate            *time.Time  `json:"expiry_date,omitempty" db:"expiry_date"`
	BatchNumber           *string     `json:"batch_number,omitempty" db:"batch_number"`
	RegistrationNumber    *string     `json:"registration_number,omitempty" db:"registration_number"`
	RequiredLicenses      []string    `json:"required_licenses,omitempty" db:"required_licenses"`
	SafetyDataSheetURL    *string     `json:"safety_data_sheet_url,omitempty" db:"safety_data_sheet_url"`
	StorageRequirements   *string     `json:"storage_requirements,omitempty" db:"storage_requirements"`
	HandlingInstructions  *string     `json:"handling_instructions,omitempty" db:"handling_instructions"`
	DisposalInstructions  *string     `json:"disposal_instructions,omitempty" db:"disposal_instructions"`
	CreatedAt             time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time   `json:"updated_at" db:"updated_at"`
}

// Request/Response types
type CreateProductRequest struct {
	Title               string              `json:"title" binding:"required"`
	Description         *string             `json:"description,omitempty"`
	Category            string              `json:"category" binding:"required,oneof=transport livestock supplies"`
	Subcategory         *string             `json:"subcategory,omitempty"`
	Price               *float64            `json:"price,omitempty"`
	PriceType           string              `json:"price_type" binding:"required,oneof=fixed negotiable per_unit quote"`
	Unit                *string             `json:"unit,omitempty"`
	Quantity            *int                `json:"quantity,omitempty"`
	AvailableFrom       *time.Time          `json:"available_from,omitempty"`
	AvailableUntil      *time.Time          `json:"available_until,omitempty"`
	Province            *string             `json:"province,omitempty"`
	Department          *string             `json:"department,omitempty"`
	Settlement          *string             `json:"settlement,omitempty"`
	City                *string             `json:"city,omitempty"`
	LocationCoordinates *Point              `json:"location_coordinates,omitempty"`
	PickupAvailable     bool                `json:"pickup_available"`
	DeliveryAvailable   bool                `json:"delivery_available"`
	DeliveryRadius      *int                `json:"delivery_radius,omitempty"`
	Tags                []string            `json:"tags,omitempty"`
	TransportDetails    *TransportDetails   `json:"transport_details,omitempty"`
	LivestockDetails    *LivestockDetails   `json:"livestock_details,omitempty"`
	SuppliesDetails     *SuppliesDetails    `json:"supplies_details,omitempty"`
}

type UpdateProductRequest struct {
	Title               *string             `json:"title,omitempty"`
	Description         *string             `json:"description,omitempty"`
	Subcategory         *string             `json:"subcategory,omitempty"`
	Price               *float64            `json:"price,omitempty"`
	PriceType           *string             `json:"price_type,omitempty"`
	Unit                *string             `json:"unit,omitempty"`
	Quantity            *int                `json:"quantity,omitempty"`
	AvailableFrom       *time.Time          `json:"available_from,omitempty"`
	AvailableUntil      *time.Time          `json:"available_until,omitempty"`
	Province            *string             `json:"province,omitempty"`
	Department          *string             `json:"department,omitempty"`
	Settlement          *string             `json:"settlement,omitempty"`
	City                *string             `json:"city,omitempty"`
	LocationCoordinates *Point              `json:"location_coordinates,omitempty"`
	PickupAvailable     *bool               `json:"pickup_available,omitempty"`
	DeliveryAvailable   *bool               `json:"delivery_available,omitempty"`
	DeliveryRadius      *int                `json:"delivery_radius,omitempty"`
	Tags                []string            `json:"tags,omitempty"`
	TransportDetails    *TransportDetails   `json:"transport_details,omitempty"`
	LivestockDetails    *LivestockDetails   `json:"livestock_details,omitempty"`
	SuppliesDetails     *SuppliesDetails    `json:"supplies_details,omitempty"`
}

type ProductSearchRequest struct {
	Query            string    `json:"query,omitempty"`
	Category         string    `json:"category,omitempty"`
	Subcategory      string    `json:"subcategory,omitempty"`
	Province         string    `json:"province,omitempty"`
	City             string    `json:"city,omitempty"`
	MinPrice         *float64  `json:"min_price,omitempty"`
	MaxPrice         *float64  `json:"max_price,omitempty"`
	PriceType        string    `json:"price_type,omitempty"`
	PickupAvailable  *bool     `json:"pickup_available,omitempty"`
	DeliveryAvailable *bool    `json:"delivery_available,omitempty"`
	IsVerifiedSeller *bool     `json:"is_verified_seller,omitempty"`
	Tags             []string  `json:"tags,omitempty"`
	SortBy           string    `json:"sort_by,omitempty"` // price_asc, price_desc, date_asc, date_desc, relevance, rating
	Page             int       `json:"page,omitempty"`
	PageSize         int       `json:"page_size,omitempty"`
}

type ProductListResponse struct {
	Products    []Product `json:"products"`
	TotalCount  int       `json:"total_count"`
	Page        int       `json:"page"`
	PageSize    int       `json:"page_size"`
	TotalPages  int       `json:"total_pages"`
}

// Database driver interfaces
func (p *Point) Scan(value interface{}) error {
	if value == nil {
		p = nil
		return nil
	}
	// Implementation would parse PostGIS point format
	return nil
}

func (p Point) Value() (driver.Value, error) {
	if p.Lat == 0 && p.Lng == 0 {
		return nil, nil
	}
	// Return as string that PostgreSQL can convert to POINT
	return fmt.Sprintf("(%f,%f)", p.Lat, p.Lng), nil
}

func (pm *ProductMetadata) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into ProductMetadata", value)
	}
	
	return json.Unmarshal(bytes, pm)
}

func (pm ProductMetadata) Value() (driver.Value, error) {
	return json.Marshal(pm)
}

func (vr *VaccinationRecords) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into VaccinationRecords", value)
	}
	
	return json.Unmarshal(bytes, vr)
}

func (vr VaccinationRecords) Value() (driver.Value, error) {
	return json.Marshal(vr)
}

func (bh *BreedingHistory) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into BreedingHistory", value)
	}
	
	return json.Unmarshal(bytes, bh)
}

func (bh BreedingHistory) Value() (driver.Value, error) {
	return json.Marshal(bh)
}