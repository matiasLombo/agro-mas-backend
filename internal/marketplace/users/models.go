package users

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                    uuid.UUID            `json:"id" db:"id"`
	Email                 string               `json:"email" db:"email"`
	PasswordHash          string               `json:"-" db:"password_hash"`
	FirstName             string               `json:"first_name" db:"first_name"`
	LastName              string               `json:"last_name" db:"last_name"`
	Phone                 *string              `json:"phone,omitempty" db:"phone"`
	CUIT                  *string              `json:"cuit,omitempty" db:"cuit"`
	CBU                   *string              `json:"cbu,omitempty" db:"cbu"`
	CBUAlias              *string              `json:"cbu_alias,omitempty" db:"cbu_alias"`
	BankName              *string              `json:"bank_name,omitempty" db:"bank_name"`
	RENSPA                *string              `json:"renspa,omitempty" db:"renspa"`
	EstablishmentName     *string              `json:"establishment_name,omitempty" db:"establishment_name"`
	EstablishmentLocation *string              `json:"establishment_location,omitempty" db:"establishment_location"`
	BusinessName          *string              `json:"business_name,omitempty" db:"business_name"`
	BusinessType          *string              `json:"business_type,omitempty" db:"business_type"`
	TaxCategory           *string              `json:"tax_category,omitempty" db:"tax_category"`
	Province              *string              `json:"province,omitempty" db:"province"`
	City                  *string              `json:"city,omitempty" db:"city"`
	Address               *string              `json:"address,omitempty" db:"address"`
	Coordinates           *Point               `json:"coordinates,omitempty" db:"coordinates"`
	Role                  string               `json:"role" db:"role"`
	VerificationLevel     int                  `json:"verification_level" db:"verification_level"`
	IsActive              bool                 `json:"is_active" db:"is_active"`
	IsVerified            bool                 `json:"is_verified" db:"is_verified"`
	Rating                float64              `json:"rating" db:"rating"`
	TotalSales            int                  `json:"total_sales" db:"total_sales"`
	TotalPurchases        int                  `json:"total_purchases" db:"total_purchases"`
	TotalReviews          int                  `json:"total_reviews" db:"total_reviews"`
	CreatedAt             time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time            `json:"updated_at" db:"updated_at"`
	LastLogin             *time.Time           `json:"last_login,omitempty" db:"last_login"`
	VerificationDocuments *VerificationDocuments `json:"verification_documents,omitempty" db:"verification_documents"`
	Preferences           *UserPreferences     `json:"preferences,omitempty" db:"preferences"`
}

type Point struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type VerificationDocuments struct {
	IdentityDocument *DocumentInfo `json:"identity_document,omitempty"`
	CUITCertificate  *DocumentInfo `json:"cuit_certificate,omitempty"`
	BusinessLicense  *DocumentInfo `json:"business_license,omitempty"`
	PropertyDeeds    *DocumentInfo `json:"property_deeds,omitempty"`
	Other            []DocumentInfo `json:"other,omitempty"`
}

type DocumentInfo struct {
	URL         string    `json:"url"`
	FileName    string    `json:"file_name"`
	FileSize    int64     `json:"file_size"`
	MimeType    string    `json:"mime_type"`
	UploadedAt  time.Time `json:"uploaded_at"`
	Status      string    `json:"status"` // pending, approved, rejected
	ReviewNotes *string   `json:"review_notes,omitempty"`
}

type UserPreferences struct {
	NotificationEmail    bool     `json:"notification_email"`
	NotificationWhatsApp bool     `json:"notification_whatsapp"`
	SearchRadius         int      `json:"search_radius_km"`
	PreferredCategories  []string `json:"preferred_categories"`
	Language             string   `json:"language"`
	Currency             string   `json:"currency"`
	PrivacyLevel         string   `json:"privacy_level"` // public, limited, private
}

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	Email                 string  `json:"email" binding:"required,email"`
	Password              string  `json:"password" binding:"required,min=8"`
	FirstName             string  `json:"first_name" binding:"required"`
	LastName              string  `json:"last_name" binding:"required"`
	Phone                 *string `json:"phone,omitempty"`
	CUIT                  *string `json:"cuit,omitempty"`
	CBU                   *string `json:"cbu,omitempty"`
	CBUAlias              *string `json:"cbu_alias,omitempty"`
	BankName              *string `json:"bank_name,omitempty"`
	RENSPA                *string `json:"renspa,omitempty"`
	EstablishmentName     *string `json:"establishment_name,omitempty"`
	EstablishmentLocation *string `json:"establishment_location,omitempty"`
	BusinessName          *string `json:"business_name,omitempty"`
	BusinessType          *string `json:"business_type,omitempty"`
	Province              *string `json:"province,omitempty"`
	City                  *string `json:"city,omitempty"`
	Address               *string `json:"address,omitempty"`
	Coordinates           *Point  `json:"coordinates,omitempty"`
	Role                  string  `json:"role" binding:"required,oneof=buyer seller"`
}

// UpdateUserRequest represents the request to update user information
type UpdateUserRequest struct {
	FirstName             *string `json:"first_name,omitempty"`
	LastName              *string `json:"last_name,omitempty"`
	Phone                 *string `json:"phone,omitempty"`
	CBU                   *string `json:"cbu,omitempty"`
	CBUAlias              *string `json:"cbu_alias,omitempty"`
	BankName              *string `json:"bank_name,omitempty"`
	RENSPA                *string `json:"renspa,omitempty"`
	EstablishmentName     *string `json:"establishment_name,omitempty"`
	EstablishmentLocation *string `json:"establishment_location,omitempty"`
	BusinessName          *string `json:"business_name,omitempty"`
	BusinessType          *string `json:"business_type,omitempty"`
	Province              *string `json:"province,omitempty"`
	City                  *string `json:"city,omitempty"`
	Address               *string `json:"address,omitempty"`
	Coordinates           *Point  `json:"coordinates,omitempty"`
}

// LoginRequest represents the login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UserResponse represents the user data returned in API responses (without sensitive data)
type UserResponse struct {
	ID                uuid.UUID         `json:"id"`
	Email             string            `json:"email"`
	FirstName         string            `json:"first_name"`
	LastName          string            `json:"last_name"`
	Phone             *string           `json:"phone,omitempty"`
	CUIT              *string           `json:"cuit,omitempty"`
	CBU                   *string           `json:"cbu,omitempty"`
	CBUAlias              *string           `json:"cbu_alias,omitempty"`
	BankName              *string           `json:"bank_name,omitempty"`
	RENSPA                *string           `json:"renspa,omitempty"`
	EstablishmentName     *string           `json:"establishment_name,omitempty"`
	EstablishmentLocation *string           `json:"establishment_location,omitempty"`
	BusinessName      *string           `json:"business_name,omitempty"`
	BusinessType      *string           `json:"business_type,omitempty"`
	Province          *string           `json:"province,omitempty"`
	City              *string           `json:"city,omitempty"`
	Address           *string           `json:"address,omitempty"`
	Coordinates       *Point            `json:"coordinates,omitempty"`
	Role              string            `json:"role"`
	VerificationLevel int               `json:"verification_level"`
	IsVerified        bool              `json:"is_verified"`
	Rating            float64           `json:"rating"`
	TotalSales        int               `json:"total_sales"`
	TotalPurchases    int               `json:"total_purchases"`
	TotalReviews      int               `json:"total_reviews"`
	CreatedAt         time.Time         `json:"created_at"`
	Preferences       *UserPreferences  `json:"preferences,omitempty"`
}

// PublicUserResponse represents limited user data for public viewing
type PublicUserResponse struct {
	ID               uuid.UUID `json:"id"`
	FirstName        string    `json:"first_name"`
	LastName         string    `json:"last_name"`
	BusinessName     *string   `json:"business_name,omitempty"`
	Province         *string   `json:"province,omitempty"`
	City             *string   `json:"city,omitempty"`
	Role             string    `json:"role"`
	VerificationLevel int      `json:"verification_level"`
	IsVerified       bool      `json:"is_verified"`
	Rating           float64   `json:"rating"`
	TotalSales       int       `json:"total_sales"`
	TotalReviews     int       `json:"total_reviews"`
	CreatedAt        time.Time `json:"created_at"`
}

// SellerProfileRequest represents the request to upgrade a buyer to seller
type SellerProfileRequest struct {
	CUIT                  string  `json:"cuit" binding:"required"`
	CBU                   string  `json:"cbu" binding:"required"`
	CBUAlias              *string `json:"cbu_alias,omitempty"`
	BankName              *string `json:"bank_name,omitempty"`
	RENSPA                string  `json:"renspa" binding:"required"`
	EstablishmentName     *string `json:"establishment_name,omitempty"`
	EstablishmentLocation *string `json:"establishment_location,omitempty"`
	BusinessName          string  `json:"business_name" binding:"required"`
	BusinessType          *string `json:"business_type,omitempty"`
	Phone                 string  `json:"phone" binding:"required"`
	Address               string  `json:"address" binding:"required"`
	City                  string  `json:"city" binding:"required"`
	Province              string  `json:"province" binding:"required"`
}

// SellerProfileResponse represents seller profile information
type SellerProfileResponse struct {
	CUIT                  *string `json:"cuit,omitempty"`
	CBU                   *string `json:"cbu,omitempty"`
	CBUAlias              *string `json:"cbu_alias,omitempty"`
	BankName              *string `json:"bank_name,omitempty"`
	RENSPA                *string `json:"renspa,omitempty"`
	EstablishmentName     *string `json:"establishment_name,omitempty"`
	EstablishmentLocation *string `json:"establishment_location,omitempty"`
	BusinessName          *string `json:"business_name,omitempty"`
	BusinessType          *string `json:"business_type,omitempty"`
	Phone                 *string `json:"phone,omitempty"`
	Address               *string `json:"address,omitempty"`
	City                  *string `json:"city,omitempty"`
	Province              *string `json:"province,omitempty"`
	IsComplete            bool    `json:"is_complete"`
}

// Implement database/sql driver interfaces for custom types
func (p *Point) Scan(value interface{}) error {
	if value == nil {
		p = nil
		return nil
	}
	// Implementation would parse PostGIS point format
	// For now, returning nil
	return nil
}

func (p Point) Value() (driver.Value, error) {
	if p.Lat == 0 && p.Lng == 0 {
		return nil, nil
	}
	return fmt.Sprintf("POINT(%f %f)", p.Lng, p.Lat), nil
}

// Convert User to UserResponse
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:                u.ID,
		Email:             u.Email,
		FirstName:         u.FirstName,
		LastName:          u.LastName,
		Phone:             u.Phone,
		CUIT:              u.CUIT,
		CBU:                   u.CBU,
		CBUAlias:              u.CBUAlias,
		BankName:              u.BankName,
		RENSPA:                u.RENSPA,
		EstablishmentName:     u.EstablishmentName,
		EstablishmentLocation: u.EstablishmentLocation,
		BusinessName:      u.BusinessName,
		BusinessType:      u.BusinessType,
		Province:          u.Province,
		City:              u.City,
		Address:           u.Address,
		Coordinates:       u.Coordinates,
		Role:              u.Role,
		VerificationLevel: u.VerificationLevel,
		IsVerified:        u.IsVerified,
		Rating:            u.Rating,
		TotalSales:        u.TotalSales,
		TotalPurchases:    u.TotalPurchases,
		TotalReviews:      u.TotalReviews,
		CreatedAt:         u.CreatedAt,
		Preferences:       u.Preferences,
	}
}

// Convert User to PublicUserResponse
func (u *User) ToPublicResponse() *PublicUserResponse {
	return &PublicUserResponse{
		ID:                u.ID,
		FirstName:         u.FirstName,
		LastName:          u.LastName,
		BusinessName:      u.BusinessName,
		Province:          u.Province,
		City:              u.City,
		Role:              u.Role,
		VerificationLevel: u.VerificationLevel,
		IsVerified:        u.IsVerified,
		Rating:            u.Rating,
		TotalSales:        u.TotalSales,
		TotalReviews:      u.TotalReviews,
		CreatedAt:         u.CreatedAt,
	}
}