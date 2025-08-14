package transactions

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	ID                      uuid.UUID              `json:"id" db:"id"`
	ProductID               uuid.UUID              `json:"product_id" db:"product_id"`
	BuyerID                 uuid.UUID              `json:"buyer_id" db:"buyer_id"`
	SellerID                uuid.UUID              `json:"seller_id" db:"seller_id"`
	Status                  string                 `json:"status" db:"status"`
	TransactionType         string                 `json:"transaction_type" db:"transaction_type"`
	OriginalPrice           *float64               `json:"original_price,omitempty" db:"original_price"`
	NegotiatedPrice         *float64               `json:"negotiated_price,omitempty" db:"negotiated_price"`
	FinalPrice              float64                `json:"final_price" db:"final_price"`
	Currency                string                 `json:"currency" db:"currency"`
	Quantity                int                    `json:"quantity" db:"quantity"`
	Unit                    *string                `json:"unit,omitempty" db:"unit"`
	PaymentMethod           *string                `json:"payment_method,omitempty" db:"payment_method"`
	PaymentStatus           string                 `json:"payment_status" db:"payment_status"`
	PaymentDate             *time.Time             `json:"payment_date,omitempty" db:"payment_date"`
	PickupAddress           *string                `json:"pickup_address,omitempty" db:"pickup_address"`
	PickupCoordinates       *Point                 `json:"pickup_coordinates,omitempty" db:"pickup_coordinates"`
	PickupDate              *time.Time             `json:"pickup_date,omitempty" db:"pickup_date"`
	PickupContactName       *string                `json:"pickup_contact_name,omitempty" db:"pickup_contact_name"`
	PickupContactPhone      *string                `json:"pickup_contact_phone,omitempty" db:"pickup_contact_phone"`
	DeliveryAddress         *string                `json:"delivery_address,omitempty" db:"delivery_address"`
	DeliveryCoordinates     *Point                 `json:"delivery_coordinates,omitempty" db:"delivery_coordinates"`
	DeliveryDate            *time.Time             `json:"delivery_date,omitempty" db:"delivery_date"`
	DeliveryContactName     *string                `json:"delivery_contact_name,omitempty" db:"delivery_contact_name"`
	DeliveryContactPhone    *string                `json:"delivery_contact_phone,omitempty" db:"delivery_contact_phone"`
	WhatsAppThreadID        *string                `json:"whatsapp_thread_id,omitempty" db:"whatsapp_thread_id"`
	CommunicationLog        *CommunicationLog      `json:"communication_log,omitempty" db:"communication_log"`
	BuyerRating             *int                   `json:"buyer_rating,omitempty" db:"buyer_rating"`
	SellerRating            *int                   `json:"seller_rating,omitempty" db:"seller_rating"`
	BuyerReview             *string                `json:"buyer_review,omitempty" db:"buyer_review"`
	SellerReview            *string                `json:"seller_review,omitempty" db:"seller_review"`
	BuyerReviewDate         *time.Time             `json:"buyer_review_date,omitempty" db:"buyer_review_date"`
	SellerReviewDate        *time.Time             `json:"seller_review_date,omitempty" db:"seller_review_date"`
	DisputeReason           *string                `json:"dispute_reason,omitempty" db:"dispute_reason"`
	DisputeResolution       *string                `json:"dispute_resolution,omitempty" db:"dispute_resolution"`
	DisputeResolvedAt       *time.Time             `json:"dispute_resolved_at,omitempty" db:"dispute_resolved_at"`
	DisputeResolvedBy       *uuid.UUID             `json:"dispute_resolved_by,omitempty" db:"dispute_resolved_by"`
	CreatedAt               time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt               time.Time              `json:"updated_at" db:"updated_at"`
	CompletedAt             *time.Time             `json:"completed_at,omitempty" db:"completed_at"`
	CancelledAt             *time.Time             `json:"cancelled_at,omitempty" db:"cancelled_at"`
	CancellationReason      *string                `json:"cancellation_reason,omitempty" db:"cancellation_reason"`
	Notes                   *string                `json:"notes,omitempty" db:"notes"`
	Metadata                *TransactionMetadata   `json:"metadata,omitempty" db:"metadata"`
}

type Point struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type CommunicationLog struct {
	Messages []CommunicationMessage `json:"messages"`
}

type CommunicationMessage struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	SenderID    uuid.UUID              `json:"sender_id"`
	ReceiverID  uuid.UUID              `json:"receiver_id"`
	Channel     string                 `json:"channel"` // whatsapp, email, internal
	MessageType string                 `json:"message_type"` // text, image, location, document
	Content     string                 `json:"content"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type TransactionMetadata struct {
	ProductTitle       string                 `json:"product_title"`
	ProductCategory    string                 `json:"product_category"`
	SellerInfo         SellerInfo            `json:"seller_info"`
	BuyerInfo          BuyerInfo             `json:"buyer_info"`
	AdditionalData     map[string]interface{} `json:"additional_data,omitempty"`
	InternalNotes      string                 `json:"internal_notes,omitempty"`
}

type SellerInfo struct {
	Name              string `json:"name"`
	Email             string `json:"email"`
	Phone             string `json:"phone"`
	VerificationLevel int    `json:"verification_level"`
}

type BuyerInfo struct {
	Name              string `json:"name"`
	Email             string `json:"email"`
	Phone             string `json:"phone"`
	VerificationLevel int    `json:"verification_level"`
}

// Request/Response types
type CreateTransactionRequest struct {
	ProductID               uuid.UUID  `json:"product_id" binding:"required"`
	Quantity                int        `json:"quantity" binding:"required,min=1"`
	NegotiatedPrice         *float64   `json:"negotiated_price,omitempty"`
	PaymentMethod           *string    `json:"payment_method,omitempty"`
	PickupAddress           *string    `json:"pickup_address,omitempty"`
	PickupCoordinates       *Point     `json:"pickup_coordinates,omitempty"`
	PickupDate              *time.Time `json:"pickup_date,omitempty"`
	PickupContactName       *string    `json:"pickup_contact_name,omitempty"`
	PickupContactPhone      *string    `json:"pickup_contact_phone,omitempty"`
	DeliveryAddress         *string    `json:"delivery_address,omitempty"`
	DeliveryCoordinates     *Point     `json:"delivery_coordinates,omitempty"`
	DeliveryDate            *time.Time `json:"delivery_date,omitempty"`
	DeliveryContactName     *string    `json:"delivery_contact_name,omitempty"`
	DeliveryContactPhone    *string    `json:"delivery_contact_phone,omitempty"`
	Notes                   *string    `json:"notes,omitempty"`
}

type UpdateTransactionRequest struct {
	Status                  *string    `json:"status,omitempty"`
	NegotiatedPrice         *float64   `json:"negotiated_price,omitempty"`
	PaymentMethod           *string    `json:"payment_method,omitempty"`
	PaymentStatus           *string    `json:"payment_status,omitempty"`
	PickupAddress           *string    `json:"pickup_address,omitempty"`
	PickupCoordinates       *Point     `json:"pickup_coordinates,omitempty"`
	PickupDate              *time.Time `json:"pickup_date,omitempty"`
	PickupContactName       *string    `json:"pickup_contact_name,omitempty"`
	PickupContactPhone      *string    `json:"pickup_contact_phone,omitempty"`
	DeliveryAddress         *string    `json:"delivery_address,omitempty"`
	DeliveryCoordinates     *Point     `json:"delivery_coordinates,omitempty"`
	DeliveryDate            *time.Time `json:"delivery_date,omitempty"`
	DeliveryContactName     *string    `json:"delivery_contact_name,omitempty"`
	DeliveryContactPhone    *string    `json:"delivery_contact_phone,omitempty"`
	Notes                   *string    `json:"notes,omitempty"`
}

type AddReviewRequest struct {
	Rating  int     `json:"rating" binding:"required,min=1,max=5"`
	Review  *string `json:"review,omitempty"`
}

type TransactionListRequest struct {
	Status       string `json:"status,omitempty"`
	ProductID    string `json:"product_id,omitempty"`
	BuyerID      string `json:"buyer_id,omitempty"`
	SellerID     string `json:"seller_id,omitempty"`
	DateFrom     string `json:"date_from,omitempty"`
	DateTo       string `json:"date_to,omitempty"`
	SortBy       string `json:"sort_by,omitempty"` // date_asc, date_desc, amount_asc, amount_desc
	Page         int    `json:"page,omitempty"`
	PageSize     int    `json:"page_size,omitempty"`
}

type TransactionListResponse struct {
	Transactions []Transaction `json:"transactions"`
	TotalCount   int           `json:"total_count"`
	Page         int           `json:"page"`
	PageSize     int           `json:"page_size"`
	TotalPages   int           `json:"total_pages"`
}

type TransactionStatsResponse struct {
	TotalTransactions      int                `json:"total_transactions"`
	CompletedTransactions  int                `json:"completed_transactions"`
	PendingTransactions    int                `json:"pending_transactions"`
	CancelledTransactions  int                `json:"cancelled_transactions"`
	TotalRevenue           float64            `json:"total_revenue"`
	AverageTransactionValue float64           `json:"average_transaction_value"`
	TransactionsByStatus   map[string]int     `json:"transactions_by_status"`
	TransactionsByMonth    map[string]float64 `json:"transactions_by_month"`
}

// Inquiry related types
type ProductInquiry struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	ProductID        uuid.UUID  `json:"product_id" db:"product_id"`
	BuyerID          uuid.UUID  `json:"buyer_id" db:"buyer_id"`
	SellerID         uuid.UUID  `json:"seller_id" db:"seller_id"`
	InquiryType      string     `json:"inquiry_type" db:"inquiry_type"`
	Subject          *string    `json:"subject,omitempty" db:"subject"`
	Message          string     `json:"message" db:"message"`
	Response         *string    `json:"response,omitempty" db:"response"`
	RespondedAt      *time.Time `json:"responded_at,omitempty" db:"responded_at"`
	IsResponded      bool       `json:"is_responded" db:"is_responded"`
	WhatsAppSent     bool       `json:"whatsapp_sent" db:"whatsapp_sent"`
	WhatsAppMessageID *string   `json:"whatsapp_message_id,omitempty" db:"whatsapp_message_id"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

type CreateInquiryRequest struct {
	ProductID   uuid.UUID `json:"product_id" binding:"required"`
	InquiryType string    `json:"inquiry_type" binding:"required,oneof=general price availability technical logistics"`
	Subject     *string   `json:"subject,omitempty"`
	Message     string    `json:"message" binding:"required"`
}

type RespondToInquiryRequest struct {
	Response string `json:"response" binding:"required"`
}

// Database driver interfaces
func (p *Point) Scan(value interface{}) error {
	if value == nil {
		p = nil
		return nil
	}
	return nil
}

func (p Point) Value() (driver.Value, error) {
	if p.Lat == 0 && p.Lng == 0 {
		return nil, nil
	}
	return fmt.Sprintf("POINT(%f %f)", p.Lng, p.Lat), nil
}

func (cl *CommunicationLog) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into CommunicationLog", value)
	}
	
	return json.Unmarshal(bytes, cl)
}

func (cl CommunicationLog) Value() (driver.Value, error) {
	return json.Marshal(cl)
}

func (tm *TransactionMetadata) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into TransactionMetadata", value)
	}
	
	return json.Unmarshal(bytes, tm)
}

func (tm TransactionMetadata) Value() (driver.Value, error) {
	return json.Marshal(tm)
}

// Constants for transaction statuses
const (
	StatusPending     = "pending"
	StatusConfirmed   = "confirmed"
	StatusInProgress  = "in_progress"
	StatusCompleted   = "completed"
	StatusCancelled   = "cancelled"
	StatusDisputed    = "disputed"
)

// Constants for payment statuses
const (
	PaymentStatusPending   = "pending"
	PaymentStatusPartial   = "partial"
	PaymentStatusCompleted = "completed"
	PaymentStatusRefunded  = "refunded"
)

// Constants for transaction types
const (
	TransactionTypeSale    = "sale"
	TransactionTypeRental  = "rental"
	TransactionTypeService = "service"
)

// Constants for inquiry types
const (
	InquiryTypeGeneral   = "general"
	InquiryTypePrice     = "price"
	InquiryTypeAvailability = "availability"
	InquiryTypeTechnical = "technical"
	InquiryTypeLogistics = "logistics"
)

// Validation functions
func IsValidTransactionStatus(status string) bool {
	validStatuses := []string{StatusPending, StatusConfirmed, StatusInProgress, StatusCompleted, StatusCancelled, StatusDisputed}
	for _, valid := range validStatuses {
		if status == valid {
			return true
		}
	}
	return false
}

func IsValidPaymentStatus(status string) bool {
	validStatuses := []string{PaymentStatusPending, PaymentStatusPartial, PaymentStatusCompleted, PaymentStatusRefunded}
	for _, valid := range validStatuses {
		if status == valid {
			return true
		}
	}
	return false
}

func IsValidInquiryType(inquiryType string) bool {
	validTypes := []string{InquiryTypeGeneral, InquiryTypePrice, InquiryTypeAvailability, InquiryTypeTechnical, InquiryTypeLogistics}
	for _, valid := range validTypes {
		if inquiryType == valid {
			return true
		}
	}
	return false
}