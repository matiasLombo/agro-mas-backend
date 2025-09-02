package transactions

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTransactionNotFound      = errors.New("transaction not found")
	ErrTransactionNotAuthorized = errors.New("user not authorized for this transaction")
	ErrInvalidTransactionStatus = errors.New("invalid transaction status")
	ErrTransactionAlreadyExists = errors.New("transaction already exists for this product and buyer")
	ErrProductNotAvailable      = errors.New("product is not available for transaction")
	ErrInsufficientQuantity     = errors.New("insufficient product quantity")
	ErrInvalidReviewRating      = errors.New("review rating must be between 1 and 5")
	ErrReviewAlreadyExists      = errors.New("review already exists for this transaction")
	ErrInquiryNotFound          = errors.New("inquiry not found")
	ErrInquiryNotAuthorized     = errors.New("user not authorized for this inquiry")
)

type Service struct {
	repo *Repository
}

type ProductInfo struct {
	ID                uuid.UUID `json:"id"`
	Title             string    `json:"title"`
	Category          string    `json:"category"`
	Price             *float64  `json:"price"`
	PriceType         string    `json:"price_type"`
	Currency          string    `json:"currency"`
	Unit              *string   `json:"unit"`
	Quantity          *int      `json:"quantity"`
	IsActive          bool      `json:"is_active"`
	IsAvailable       bool      `json:"is_available"`
	SellerID          uuid.UUID `json:"seller_id"`
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// CreateTransaction creates a new transaction
func (s *Service) CreateTransaction(ctx context.Context, buyerID uuid.UUID, req *CreateTransactionRequest, productInfo ProductInfo, sellerInfo SellerInfo, buyerInfo BuyerInfo) (*Transaction, error) {
	// Validate product availability
	if !productInfo.IsActive || !productInfo.IsAvailable {
		return nil, ErrProductNotAvailable
	}

	// Check quantity availability
	if productInfo.Quantity != nil && req.Quantity > *productInfo.Quantity {
		return nil, ErrInsufficientQuantity
	}

	// Prevent self-transactions
	if buyerID == productInfo.SellerID {
		return nil, errors.New("cannot create transaction for your own product")
	}

	// Calculate final price
	var finalPrice float64
	if req.NegotiatedPrice != nil {
		finalPrice = *req.NegotiatedPrice * float64(req.Quantity)
	} else if productInfo.Price != nil {
		finalPrice = *productInfo.Price * float64(req.Quantity)
	} else {
		// For quote-based products, final price will be set later
		finalPrice = 0
	}

	// Create transaction object
	transaction := &Transaction{
		ID:                   uuid.New(),
		ProductID:            req.ProductID,
		BuyerID:              buyerID,
		SellerID:             productInfo.SellerID,
		Status:               StatusPending,
		TransactionType:      TransactionTypeSale,
		OriginalPrice:        productInfo.Price,
		NegotiatedPrice:      req.NegotiatedPrice,
		FinalPrice:           finalPrice,
		Currency:             productInfo.Currency,
		Quantity:             req.Quantity,
		Unit:                 productInfo.Unit,
		PaymentMethod:        req.PaymentMethod,
		PaymentStatus:        PaymentStatusPending,
		PickupAddress:        req.PickupAddress,
		PickupCoordinates:    req.PickupCoordinates,
		PickupDate:           req.PickupDate,
		PickupContactName:    req.PickupContactName,
		PickupContactPhone:   req.PickupContactPhone,
		DeliveryAddress:      req.DeliveryAddress,
		DeliveryCoordinates:  req.DeliveryCoordinates,
		DeliveryDate:         req.DeliveryDate,
		DeliveryContactName:  req.DeliveryContactName,
		DeliveryContactPhone: req.DeliveryContactPhone,
		Notes:                req.Notes,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
		CommunicationLog:     &CommunicationLog{Messages: []CommunicationMessage{}},
		Metadata: &TransactionMetadata{
			ProductTitle:    productInfo.Title,
			ProductCategory: productInfo.Category,
			SellerInfo: SellerInfo{
				Name:              sellerInfo.Name,
				Email:             sellerInfo.Email,
				Phone:             sellerInfo.Phone,
				VerificationLevel: sellerInfo.VerificationLevel,
			},
			BuyerInfo: BuyerInfo{
				Name:              buyerInfo.Name,
				Email:             buyerInfo.Email,
				Phone:             buyerInfo.Phone,
				VerificationLevel: buyerInfo.VerificationLevel,
			},
		},
	}

	// Create transaction in database
	if err := s.repo.CreateTransaction(ctx, transaction); err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	return transaction, nil
}

// GetTransactionByID retrieves a transaction by ID
func (s *Service) GetTransactionByID(ctx context.Context, userID, transactionID uuid.UUID) (*Transaction, error) {
	transaction, err := s.repo.GetTransactionByID(ctx, transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}
	if transaction == nil {
		return nil, ErrTransactionNotFound
	}

	// Check authorization (user must be buyer, seller, or admin)
	if transaction.BuyerID != userID && transaction.SellerID != userID {
		return nil, ErrTransactionNotAuthorized
	}

	return transaction, nil
}

// UpdateTransactionStatus updates the status of a transaction
func (s *Service) UpdateTransactionStatus(ctx context.Context, userID, transactionID uuid.UUID, newStatus string) error {
	// Validate status
	if !IsValidTransactionStatus(newStatus) {
		return ErrInvalidTransactionStatus
	}

	// Get transaction
	transaction, err := s.repo.GetTransactionByID(ctx, transactionID)
	if err != nil {
		return fmt.Errorf("failed to get transaction: %w", err)
	}
	if transaction == nil {
		return ErrTransactionNotFound
	}

	// Check authorization
	if transaction.BuyerID != userID && transaction.SellerID != userID {
		return ErrTransactionNotAuthorized
	}

	// Validate status transition
	if err := s.validateStatusTransition(transaction.Status, newStatus, userID, transaction); err != nil {
		return err
	}

	// Prepare updates
	updates := map[string]interface{}{
		"status": newStatus,
	}

	// Add timestamp fields based on status
	switch newStatus {
	case StatusCompleted:
		updates["completed_at"] = time.Now()
	case StatusCancelled:
		updates["cancelled_at"] = time.Now()
	}

	// Update transaction
	return s.repo.UpdateTransaction(ctx, transactionID, updates)
}

// UpdateTransaction updates transaction details
func (s *Service) UpdateTransaction(ctx context.Context, userID, transactionID uuid.UUID, req *UpdateTransactionRequest) (*Transaction, error) {
	// Get transaction
	transaction, err := s.repo.GetTransactionByID(ctx, transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}
	if transaction == nil {
		return nil, ErrTransactionNotFound
	}

	// Check authorization
	if transaction.BuyerID != userID && transaction.SellerID != userID {
		return nil, ErrTransactionNotAuthorized
	}

	// Prepare updates
	updates := make(map[string]interface{})

	if req.Status != nil {
		if !IsValidTransactionStatus(*req.Status) {
			return nil, ErrInvalidTransactionStatus
		}
		if err := s.validateStatusTransition(transaction.Status, *req.Status, userID, transaction); err != nil {
			return nil, err
		}
		updates["status"] = *req.Status
	}

	if req.NegotiatedPrice != nil {
		updates["negotiated_price"] = *req.NegotiatedPrice
		// Recalculate final price
		updates["final_price"] = *req.NegotiatedPrice * float64(transaction.Quantity)
	}

	if req.PaymentMethod != nil {
		updates["payment_method"] = *req.PaymentMethod
	}

	if req.PaymentStatus != nil {
		if !IsValidPaymentStatus(*req.PaymentStatus) {
			return nil, errors.New("invalid payment status")
		}
		updates["payment_status"] = *req.PaymentStatus
		if *req.PaymentStatus == PaymentStatusCompleted {
			updates["payment_date"] = time.Now()
		}
	}

	if req.PickupAddress != nil {
		updates["pickup_address"] = *req.PickupAddress
	}
	if req.PickupCoordinates != nil {
		updates["pickup_coordinates"] = req.PickupCoordinates
	}
	if req.PickupDate != nil {
		updates["pickup_date"] = *req.PickupDate
	}
	if req.PickupContactName != nil {
		updates["pickup_contact_name"] = *req.PickupContactName
	}
	if req.PickupContactPhone != nil {
		updates["pickup_contact_phone"] = *req.PickupContactPhone
	}

	if req.DeliveryAddress != nil {
		updates["delivery_address"] = *req.DeliveryAddress
	}
	if req.DeliveryCoordinates != nil {
		updates["delivery_coordinates"] = req.DeliveryCoordinates
	}
	if req.DeliveryDate != nil {
		updates["delivery_date"] = *req.DeliveryDate
	}
	if req.DeliveryContactName != nil {
		updates["delivery_contact_name"] = *req.DeliveryContactName
	}
	if req.DeliveryContactPhone != nil {
		updates["delivery_contact_phone"] = *req.DeliveryContactPhone
	}

	if req.Notes != nil {
		updates["notes"] = *req.Notes
	}

	// Update transaction
	if err := s.repo.UpdateTransaction(ctx, transactionID, updates); err != nil {
		return nil, fmt.Errorf("failed to update transaction: %w", err)
	}

	// Return updated transaction
	return s.repo.GetTransactionByID(ctx, transactionID)
}

// AddReview adds a review for a completed transaction
func (s *Service) AddReview(ctx context.Context, userID, transactionID uuid.UUID, req *AddReviewRequest) error {
	// Validate rating
	if req.Rating < 1 || req.Rating > 5 {
		return ErrInvalidReviewRating
	}

	// Get transaction
	transaction, err := s.repo.GetTransactionByID(ctx, transactionID)
	if err != nil {
		return fmt.Errorf("failed to get transaction: %w", err)
	}
	if transaction == nil {
		return ErrTransactionNotFound
	}

	// Check authorization
	if transaction.BuyerID != userID && transaction.SellerID != userID {
		return ErrTransactionNotAuthorized
	}

	// Only allow reviews for completed transactions
	if transaction.Status != StatusCompleted {
		return errors.New("can only review completed transactions")
	}

	// Prepare updates based on who is reviewing
	updates := make(map[string]interface{})
	
	if userID == transaction.BuyerID {
		// Buyer reviewing seller
		if transaction.BuyerRating != nil {
			return ErrReviewAlreadyExists
		}
		updates["buyer_rating"] = req.Rating
		updates["buyer_review"] = req.Review
		updates["buyer_review_date"] = time.Now()
	} else {
		// Seller reviewing buyer
		if transaction.SellerRating != nil {
			return ErrReviewAlreadyExists
		}
		updates["seller_rating"] = req.Rating
		updates["seller_review"] = req.Review
		updates["seller_review_date"] = time.Now()
	}

	return s.repo.UpdateTransaction(ctx, transactionID, updates)
}

// ListTransactions retrieves transactions with filters and pagination
func (s *Service) ListTransactions(ctx context.Context, userID *uuid.UUID, req *TransactionListRequest) (*TransactionListResponse, error) {
	// Set default values
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Parse filters
	filters := TransactionFilters{
		Status: req.Status,
		SortBy: req.SortBy,
	}

	if userID != nil {
		filters.UserID = userID
	}

	if req.ProductID != "" {
		if productID, err := uuid.Parse(req.ProductID); err == nil {
			filters.ProductID = &productID
		}
	}

	if req.BuyerID != "" {
		if buyerID, err := uuid.Parse(req.BuyerID); err == nil {
			filters.BuyerID = &buyerID
		}
	}

	if req.SellerID != "" {
		if sellerID, err := uuid.Parse(req.SellerID); err == nil {
			filters.SellerID = &sellerID
		}
	}

	if req.DateFrom != "" {
		if dateFrom, err := time.Parse("2006-01-02", req.DateFrom); err == nil {
			filters.DateFrom = &dateFrom
		}
	}

	if req.DateTo != "" {
		if dateTo, err := time.Parse("2006-01-02", req.DateTo); err == nil {
			filters.DateTo = &dateTo
		}
	}

	offset := (page - 1) * pageSize

	transactions, totalCount, err := s.repo.ListTransactions(ctx, filters, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list transactions: %w", err)
	}

	// Convert to response format
	transactionList := make([]Transaction, len(transactions))
	for i, t := range transactions {
		transactionList[i] = *t
	}

	totalPages := (totalCount + pageSize - 1) / pageSize

	return &TransactionListResponse{
		Transactions: transactionList,
		TotalCount:   totalCount,
		Page:         page,
		PageSize:     pageSize,
		TotalPages:   totalPages,
	}, nil
}

// GetTransactionStats retrieves transaction statistics
func (s *Service) GetTransactionStats(ctx context.Context, userID *uuid.UUID, dateFrom, dateTo *time.Time) (*TransactionStatsResponse, error) {
	filters := TransactionStatsFilters{
		DateFrom: dateFrom,
		DateTo:   dateTo,
	}

	return s.repo.GetTransactionStats(ctx, userID, filters)
}

// Product Inquiries
func (s *Service) CreateInquiry(ctx context.Context, buyerID uuid.UUID, req *CreateInquiryRequest, sellerID uuid.UUID) (*ProductInquiry, error) {
	// Validate inquiry type
	if !IsValidInquiryType(req.InquiryType) {
		return nil, errors.New("invalid inquiry type")
	}

	inquiry := &ProductInquiry{
		ID:          uuid.New(),
		ProductID:   req.ProductID,
		BuyerID:     buyerID,
		SellerID:    sellerID,
		InquiryType: req.InquiryType,
		Subject:     req.Subject,
		Message:     req.Message,
		IsResponded: false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.CreateInquiry(ctx, inquiry); err != nil {
		return nil, fmt.Errorf("failed to create inquiry: %w", err)
	}

	return inquiry, nil
}

func (s *Service) RespondToInquiry(ctx context.Context, sellerID, inquiryID uuid.UUID, req *RespondToInquiryRequest) error {
	// Get inquiry
	inquiry, err := s.repo.GetInquiryByID(ctx, inquiryID)
	if err != nil {
		return fmt.Errorf("failed to get inquiry: %w", err)
	}
	if inquiry == nil {
		return ErrInquiryNotFound
	}

	// Check authorization
	if inquiry.SellerID != sellerID {
		return ErrInquiryNotAuthorized
	}

	// Update inquiry with response
	updates := map[string]interface{}{
		"response":     req.Response,
		"responded_at": time.Now(),
		"is_responded": true,
	}

	return s.repo.UpdateInquiry(ctx, inquiryID, updates)
}

// Helper functions
func (s *Service) validateStatusTransition(currentStatus, newStatus string, userID uuid.UUID, transaction *Transaction) error {
	// Define allowed transitions
	allowedTransitions := map[string]map[string]bool{
		StatusPending: {
			StatusConfirmed:  true,
			StatusCancelled:  true,
		},
		StatusConfirmed: {
			StatusInProgress: true,
			StatusCancelled:  true,
		},
		StatusInProgress: {
			StatusCompleted: true,
			StatusDisputed:  true,
		},
		StatusDisputed: {
			StatusCompleted: true,
			StatusCancelled: true,
		},
	}

	// Check if transition is allowed
	if transitions, exists := allowedTransitions[currentStatus]; !exists {
		return fmt.Errorf("no transitions allowed from status %s", currentStatus)
	} else if !transitions[newStatus] {
		return fmt.Errorf("transition from %s to %s is not allowed", currentStatus, newStatus)
	}

	// Additional business rules
	switch newStatus {
	case StatusConfirmed:
		// Only seller can confirm
		if userID != transaction.SellerID {
			return errors.New("only seller can confirm transaction")
		}
	case StatusCompleted:
		// Both parties can mark as completed, but payment should be completed
		if transaction.PaymentStatus != PaymentStatusCompleted {
			return errors.New("payment must be completed before transaction can be marked as completed")
		}
	}

	return nil
}