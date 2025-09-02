package transactions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// CreateTransaction creates a new transaction in the database
func (r *Repository) CreateTransaction(ctx context.Context, transaction *Transaction) error {
	query := `
		INSERT INTO transactions (
			id, product_id, buyer_id, seller_id, status, transaction_type,
			original_price, negotiated_price, final_price, currency, quantity, unit,
			payment_method, payment_status, pickup_address, pickup_coordinates,
			pickup_date, pickup_contact_name, pickup_contact_phone,
			delivery_address, delivery_coordinates, delivery_date,
			delivery_contact_name, delivery_contact_phone, whatsapp_thread_id,
			communication_log, notes, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			ST_GeomFromText('POINT(' || $16 || ' ' || $17 || ')', 4326),
			$18, $19, $20, $21,
			ST_GeomFromText('POINT(' || $22 || ' ' || $23 || ')', 4326),
			$24, $25, $26, $27, $28, $29, $30
		)`

	var pickupLng, pickupLat, deliveryLng, deliveryLat sql.NullFloat64

	if transaction.PickupCoordinates != nil {
		pickupLng.Float64 = transaction.PickupCoordinates.Lng
		pickupLng.Valid = true
		pickupLat.Float64 = transaction.PickupCoordinates.Lat
		pickupLat.Valid = true
	}

	if transaction.DeliveryCoordinates != nil {
		deliveryLng.Float64 = transaction.DeliveryCoordinates.Lng
		deliveryLng.Valid = true
		deliveryLat.Float64 = transaction.DeliveryCoordinates.Lat
		deliveryLat.Valid = true
	}

	var communicationLogJSON, metadataJSON []byte
	var err error

	if transaction.CommunicationLog != nil {
		communicationLogJSON, err = json.Marshal(transaction.CommunicationLog)
		if err != nil {
			return fmt.Errorf("failed to marshal communication log: %w", err)
		}
	}

	if transaction.Metadata != nil {
		metadataJSON, err = json.Marshal(transaction.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	_, err = r.db.ExecContext(ctx, query,
		transaction.ID, transaction.ProductID, transaction.BuyerID, transaction.SellerID,
		transaction.Status, transaction.TransactionType, transaction.OriginalPrice,
		transaction.NegotiatedPrice, transaction.FinalPrice, transaction.Currency,
		transaction.Quantity, transaction.Unit, transaction.PaymentMethod,
		transaction.PaymentStatus, transaction.PickupAddress, pickupLng, pickupLat,
		transaction.PickupDate, transaction.PickupContactName, transaction.PickupContactPhone,
		transaction.DeliveryAddress, deliveryLng, deliveryLat, transaction.DeliveryDate,
		transaction.DeliveryContactName, transaction.DeliveryContactPhone,
		transaction.WhatsAppThreadID, communicationLogJSON, transaction.Notes, metadataJSON)

	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

// GetTransactionByID retrieves a transaction by its ID
func (r *Repository) GetTransactionByID(ctx context.Context, id uuid.UUID) (*Transaction, error) {
	query := `
		SELECT 
			id, product_id, buyer_id, seller_id, status, transaction_type,
			original_price, negotiated_price, final_price, currency, quantity, unit,
			payment_method, payment_status, payment_date, pickup_address,
			ST_X(pickup_coordinates) as pickup_lng, ST_Y(pickup_coordinates) as pickup_lat,
			pickup_date, pickup_contact_name, pickup_contact_phone, delivery_address,
			ST_X(delivery_coordinates) as delivery_lng, ST_Y(delivery_coordinates) as delivery_lat,
			delivery_date, delivery_contact_name, delivery_contact_phone,
			whatsapp_thread_id, communication_log, buyer_rating, seller_rating,
			buyer_review, seller_review, buyer_review_date, seller_review_date,
			dispute_reason, dispute_resolution, dispute_resolved_at, dispute_resolved_by,
			created_at, updated_at, completed_at, cancelled_at, cancellation_reason,
			notes, metadata
		FROM transactions 
		WHERE id = $1`

	transaction := &Transaction{}
	var pickupLng, pickupLat, deliveryLng, deliveryLat sql.NullFloat64
	var communicationLogJSON, metadataJSON sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&transaction.ID, &transaction.ProductID, &transaction.BuyerID, &transaction.SellerID,
		&transaction.Status, &transaction.TransactionType, &transaction.OriginalPrice,
		&transaction.NegotiatedPrice, &transaction.FinalPrice, &transaction.Currency,
		&transaction.Quantity, &transaction.Unit, &transaction.PaymentMethod,
		&transaction.PaymentStatus, &transaction.PaymentDate, &transaction.PickupAddress,
		&pickupLng, &pickupLat, &transaction.PickupDate, &transaction.PickupContactName,
		&transaction.PickupContactPhone, &transaction.DeliveryAddress, &deliveryLng,
		&deliveryLat, &transaction.DeliveryDate, &transaction.DeliveryContactName,
		&transaction.DeliveryContactPhone, &transaction.WhatsAppThreadID,
		&communicationLogJSON, &transaction.BuyerRating, &transaction.SellerRating,
		&transaction.BuyerReview, &transaction.SellerReview, &transaction.BuyerReviewDate,
		&transaction.SellerReviewDate, &transaction.DisputeReason, &transaction.DisputeResolution,
		&transaction.DisputeResolvedAt, &transaction.DisputeResolvedBy, &transaction.CreatedAt,
		&transaction.UpdatedAt, &transaction.CompletedAt, &transaction.CancelledAt,
		&transaction.CancellationReason, &transaction.Notes, &metadataJSON)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	// Parse coordinates
	if pickupLng.Valid && pickupLat.Valid {
		transaction.PickupCoordinates = &Point{
			Lng: pickupLng.Float64,
			Lat: pickupLat.Float64,
		}
	}

	if deliveryLng.Valid && deliveryLat.Valid {
		transaction.DeliveryCoordinates = &Point{
			Lng: deliveryLng.Float64,
			Lat: deliveryLat.Float64,
		}
	}

	// Parse JSON fields
	if communicationLogJSON.Valid && communicationLogJSON.String != "" {
		if err := json.Unmarshal([]byte(communicationLogJSON.String), &transaction.CommunicationLog); err != nil {
			return nil, fmt.Errorf("failed to unmarshal communication log: %w", err)
		}
	}

	if metadataJSON.Valid && metadataJSON.String != "" {
		if err := json.Unmarshal([]byte(metadataJSON.String), &transaction.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return transaction, nil
}

// ListTransactions retrieves transactions with filtering and pagination
func (r *Repository) ListTransactions(ctx context.Context, filters TransactionFilters, limit, offset int) ([]*Transaction, int, error) {
	whereConditions := []string{"1=1"}
	args := []interface{}{}
	argIndex := 1

	if filters.Status != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, filters.Status)
		argIndex++
	}

	if filters.ProductID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("product_id = $%d", argIndex))
		args = append(args, *filters.ProductID)
		argIndex++
	}

	if filters.BuyerID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("buyer_id = $%d", argIndex))
		args = append(args, *filters.BuyerID)
		argIndex++
	}

	if filters.SellerID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("seller_id = $%d", argIndex))
		args = append(args, *filters.SellerID)
		argIndex++
	}

	if filters.UserID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("(buyer_id = $%d OR seller_id = $%d)", argIndex, argIndex))
		args = append(args, *filters.UserID)
		argIndex++
	}

	if filters.DateFrom != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, *filters.DateFrom)
		argIndex++
	}

	if filters.DateTo != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, *filters.DateTo)
		argIndex++
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM transactions WHERE %s", whereClause)
	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count transactions: %w", err)
	}

	// Determine sorting
	orderBy := "created_at DESC"
	switch filters.SortBy {
	case "date_asc":
		orderBy = "created_at ASC"
	case "date_desc":
		orderBy = "created_at DESC"
	case "amount_asc":
		orderBy = "final_price ASC"
	case "amount_desc":
		orderBy = "final_price DESC"
	}

	// Get paginated results
	query := fmt.Sprintf(`
		SELECT 
			id, product_id, buyer_id, seller_id, status, transaction_type,
			original_price, negotiated_price, final_price, currency, quantity, unit,
			payment_method, payment_status, payment_date, pickup_address,
			ST_X(pickup_coordinates) as pickup_lng, ST_Y(pickup_coordinates) as pickup_lat,
			pickup_date, pickup_contact_name, pickup_contact_phone, delivery_address,
			ST_X(delivery_coordinates) as delivery_lng, ST_Y(delivery_coordinates) as delivery_lat,
			delivery_date, delivery_contact_name, delivery_contact_phone,
			whatsapp_thread_id, communication_log, buyer_rating, seller_rating,
			buyer_review, seller_review, buyer_review_date, seller_review_date,
			dispute_reason, dispute_resolution, dispute_resolved_at, dispute_resolved_by,
			created_at, updated_at, completed_at, cancelled_at, cancellation_reason,
			notes, metadata
		FROM transactions 
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d`, whereClause, orderBy, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list transactions: %w", err)
	}
	defer rows.Close()

	transactions := make([]*Transaction, 0)
	for rows.Next() {
		transaction := &Transaction{}
		var pickupLng, pickupLat, deliveryLng, deliveryLat sql.NullFloat64
		var communicationLogJSON, metadataJSON sql.NullString

		err := rows.Scan(
			&transaction.ID, &transaction.ProductID, &transaction.BuyerID, &transaction.SellerID,
			&transaction.Status, &transaction.TransactionType, &transaction.OriginalPrice,
			&transaction.NegotiatedPrice, &transaction.FinalPrice, &transaction.Currency,
			&transaction.Quantity, &transaction.Unit, &transaction.PaymentMethod,
			&transaction.PaymentStatus, &transaction.PaymentDate, &transaction.PickupAddress,
			&pickupLng, &pickupLat, &transaction.PickupDate, &transaction.PickupContactName,
			&transaction.PickupContactPhone, &transaction.DeliveryAddress, &deliveryLng,
			&deliveryLat, &transaction.DeliveryDate, &transaction.DeliveryContactName,
			&transaction.DeliveryContactPhone, &transaction.WhatsAppThreadID,
			&communicationLogJSON, &transaction.BuyerRating, &transaction.SellerRating,
			&transaction.BuyerReview, &transaction.SellerReview, &transaction.BuyerReviewDate,
			&transaction.SellerReviewDate, &transaction.DisputeReason, &transaction.DisputeResolution,
			&transaction.DisputeResolvedAt, &transaction.DisputeResolvedBy, &transaction.CreatedAt,
			&transaction.UpdatedAt, &transaction.CompletedAt, &transaction.CancelledAt,
			&transaction.CancellationReason, &transaction.Notes, &metadataJSON)

		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan transaction: %w", err)
		}

		// Parse coordinates
		if pickupLng.Valid && pickupLat.Valid {
			transaction.PickupCoordinates = &Point{
				Lng: pickupLng.Float64,
				Lat: pickupLat.Float64,
			}
		}

		if deliveryLng.Valid && deliveryLat.Valid {
			transaction.DeliveryCoordinates = &Point{
				Lng: deliveryLng.Float64,
				Lat: deliveryLat.Float64,
			}
		}

		// Parse JSON fields
		if communicationLogJSON.Valid && communicationLogJSON.String != "" {
			if err := json.Unmarshal([]byte(communicationLogJSON.String), &transaction.CommunicationLog); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal communication log: %w", err)
			}
		}

		if metadataJSON.Valid && metadataJSON.String != "" {
			if err := json.Unmarshal([]byte(metadataJSON.String), &transaction.Metadata); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		transactions = append(transactions, transaction)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate transactions: %w", err)
	}

	return transactions, totalCount, nil
}

// UpdateTransaction updates an existing transaction
func (r *Repository) UpdateTransaction(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	setParts := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)
	argIndex := 1

	for field, value := range updates {
		switch field {
		case "pickup_coordinates":
			if coords, ok := value.(*Point); ok && coords != nil {
				setParts = append(setParts, fmt.Sprintf("pickup_coordinates = ST_GeomFromText('POINT(%f %f)', 4326)", coords.Lng, coords.Lat))
				continue
			}
		case "delivery_coordinates":
			if coords, ok := value.(*Point); ok && coords != nil {
				setParts = append(setParts, fmt.Sprintf("delivery_coordinates = ST_GeomFromText('POINT(%f %f)', 4326)", coords.Lng, coords.Lat))
				continue
			}
		}
		setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
		args = append(args, value)
		argIndex++
	}

	// Always update the updated_at timestamp
	setParts = append(setParts, "updated_at = NOW()")

	query := fmt.Sprintf("UPDATE transactions SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIndex)
	args = append(args, id)

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	return nil
}

// Product Inquiries
func (r *Repository) CreateInquiry(ctx context.Context, inquiry *ProductInquiry) error {
	query := `
		INSERT INTO product_inquiries (
			id, product_id, buyer_id, seller_id, inquiry_type, subject, message,
			is_responded, whatsapp_sent, whatsapp_message_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.db.ExecContext(ctx, query,
		inquiry.ID, inquiry.ProductID, inquiry.BuyerID, inquiry.SellerID,
		inquiry.InquiryType, inquiry.Subject, inquiry.Message,
		inquiry.IsResponded, inquiry.WhatsAppSent, inquiry.WhatsAppMessageID)

	if err != nil {
		return fmt.Errorf("failed to create inquiry: %w", err)
	}

	return nil
}

func (r *Repository) GetInquiryByID(ctx context.Context, id uuid.UUID) (*ProductInquiry, error) {
	query := `
		SELECT id, product_id, buyer_id, seller_id, inquiry_type, subject, message,
			   response, responded_at, is_responded, whatsapp_sent, whatsapp_message_id,
			   created_at, updated_at
		FROM product_inquiries 
		WHERE id = $1`

	inquiry := &ProductInquiry{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&inquiry.ID, &inquiry.ProductID, &inquiry.BuyerID, &inquiry.SellerID,
		&inquiry.InquiryType, &inquiry.Subject, &inquiry.Message, &inquiry.Response,
		&inquiry.RespondedAt, &inquiry.IsResponded, &inquiry.WhatsAppSent,
		&inquiry.WhatsAppMessageID, &inquiry.CreatedAt, &inquiry.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get inquiry: %w", err)
	}

	return inquiry, nil
}

func (r *Repository) UpdateInquiry(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
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

	setParts = append(setParts, "updated_at = NOW()")

	query := fmt.Sprintf("UPDATE product_inquiries SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIndex)
	args = append(args, id)

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update inquiry: %w", err)
	}

	return nil
}

// GetTransactionStats retrieves transaction statistics
func (r *Repository) GetTransactionStats(ctx context.Context, userID *uuid.UUID, filters TransactionStatsFilters) (*TransactionStatsResponse, error) {
	whereConditions := []string{"1=1"}
	args := []interface{}{}
	argIndex := 1

	if userID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("(buyer_id = $%d OR seller_id = $%d)", argIndex, argIndex))
		args = append(args, *userID)
		argIndex++
	}

	if filters.DateFrom != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, *filters.DateFrom)
		argIndex++
	}

	if filters.DateTo != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, *filters.DateTo)
		argIndex++
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// Get basic stats
	statsQuery := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total_transactions,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_transactions,
			COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_transactions,
			COUNT(CASE WHEN status = 'cancelled' THEN 1 END) as cancelled_transactions,
			COALESCE(SUM(CASE WHEN status = 'completed' THEN final_price ELSE 0 END), 0) as total_revenue,
			COALESCE(AVG(CASE WHEN status = 'completed' THEN final_price END), 0) as avg_transaction_value
		FROM transactions 
		WHERE %s`, whereClause)

	stats := &TransactionStatsResponse{
		TransactionsByStatus: make(map[string]int),
		TransactionsByMonth:  make(map[string]float64),
	}

	err := r.db.QueryRowContext(ctx, statsQuery, args...).Scan(
		&stats.TotalTransactions,
		&stats.CompletedTransactions,
		&stats.PendingTransactions,
		&stats.CancelledTransactions,
		&stats.TotalRevenue,
		&stats.AverageTransactionValue)

	if err != nil {
		return nil, fmt.Errorf("failed to get transaction stats: %w", err)
	}

	// Get transactions by status
	statusQuery := fmt.Sprintf(`
		SELECT status, COUNT(*) 
		FROM transactions 
		WHERE %s 
		GROUP BY status`, whereClause)

	statusRows, err := r.db.QueryContext(ctx, statusQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions by status: %w", err)
	}
	defer statusRows.Close()

	for statusRows.Next() {
		var status string
		var count int
		if err := statusRows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("failed to scan status row: %w", err)
		}
		stats.TransactionsByStatus[status] = count
	}

	// Get transactions by month
	monthQuery := fmt.Sprintf(`
		SELECT 
			TO_CHAR(created_at, 'YYYY-MM') as month,
			COALESCE(SUM(final_price), 0) as revenue
		FROM transactions 
		WHERE %s 
		GROUP BY TO_CHAR(created_at, 'YYYY-MM')
		ORDER BY month`, whereClause)

	monthRows, err := r.db.QueryContext(ctx, monthQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions by month: %w", err)
	}
	defer monthRows.Close()

	for monthRows.Next() {
		var month string
		var revenue float64
		if err := monthRows.Scan(&month, &revenue); err != nil {
			return nil, fmt.Errorf("failed to scan month row: %w", err)
		}
		stats.TransactionsByMonth[month] = revenue
	}

	return stats, nil
}

// Filter types
type TransactionFilters struct {
	Status    string     `json:"status"`
	ProductID *uuid.UUID `json:"product_id"`
	BuyerID   *uuid.UUID `json:"buyer_id"`
	SellerID  *uuid.UUID `json:"seller_id"`
	UserID    *uuid.UUID `json:"user_id"` // For getting transactions where user is either buyer or seller
	DateFrom  *time.Time `json:"date_from"`
	DateTo    *time.Time `json:"date_to"`
	SortBy    string     `json:"sort_by"`
}

type TransactionStatsFilters struct {
	DateFrom *time.Time `json:"date_from"`
	DateTo   *time.Time `json:"date_to"`
}