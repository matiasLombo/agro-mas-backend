package whatsapp

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	client *Client
	db     *sql.DB
}

type WhatsAppLink struct {
	ID              uuid.UUID `json:"id"`
	ProductID       *uuid.UUID `json:"product_id,omitempty"`
	TransactionID   *uuid.UUID `json:"transaction_id,omitempty"`
	InquiryID       *uuid.UUID `json:"inquiry_id,omitempty"`
	FromUserID      uuid.UUID `json:"from_user_id"`
	ToUserID        uuid.UUID `json:"to_user_id"`
	PhoneNumber     string    `json:"phone_number"`
	Message         string    `json:"message"`
	WhatsAppURL     string    `json:"whatsapp_url"`
	DeepLink        string    `json:"deep_link"`
	WebLink         string    `json:"web_link"`
	LinkType        string    `json:"link_type"` // inquiry, transaction, business
	Status          string    `json:"status"`    // created, clicked, expired
	ClickCount      int       `json:"click_count"`
	LastClickedAt   *time.Time `json:"last_clicked_at,omitempty"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type CreateLinkRequest struct {
	ProductID       *uuid.UUID `json:"product_id,omitempty"`
	TransactionID   *uuid.UUID `json:"transaction_id,omitempty"`
	InquiryID       *uuid.UUID `json:"inquiry_id,omitempty"`
	ToUserID        uuid.UUID  `json:"to_user_id"`
	PhoneNumber     string     `json:"phone_number"`
	LinkType        string     `json:"link_type"`
	MessageTemplate MessageTemplate `json:"message_template"`
	ExpirationHours int        `json:"expiration_hours,omitempty"` // 0 means no expiration
}

func NewService(client *Client, db *sql.DB) *Service {
	return &Service{
		client: client,
		db:     db,
	}
}

// CreateWhatsAppLink creates a new WhatsApp communication link
func (s *Service) CreateWhatsAppLink(ctx context.Context, fromUserID uuid.UUID, req CreateLinkRequest) (*WhatsAppLink, error) {
	// Validate phone number
	if err := s.client.ValidatePhoneNumber(req.PhoneNumber); err != nil {
		return nil, fmt.Errorf("invalid phone number: %w", err)
	}

	// Build message based on link type
	var message string
	var err error

	switch req.LinkType {
	case "inquiry":
		message = s.client.buildProductInquiryMessage(req.MessageTemplate)
	case "transaction":
		message = s.client.buildTransactionMessage(req.MessageTemplate)
	case "business":
		message = s.client.buildBusinessContactMessage(req.MessageTemplate.SellerName, req.MessageTemplate.InquiryType)
	default:
		return nil, fmt.Errorf("invalid link type: %s", req.LinkType)
	}

	// Generate WhatsApp URLs
	whatsappURL, err := s.client.GenerateWhatsAppURL(req.PhoneNumber, message)
	if err != nil {
		return nil, fmt.Errorf("failed to generate WhatsApp URL: %w", err)
	}

	deepLink, err := s.client.GenerateDeepLink(req.PhoneNumber, message)
	if err != nil {
		return nil, fmt.Errorf("failed to generate deep link: %w", err)
	}

	webLink, err := s.client.GenerateWebLink(req.PhoneNumber, message)
	if err != nil {
		return nil, fmt.Errorf("failed to generate web link: %w", err)
	}

	// Calculate expiration
	var expiresAt *time.Time
	if req.ExpirationHours > 0 {
		expiry := time.Now().Add(time.Duration(req.ExpirationHours) * time.Hour)
		expiresAt = &expiry
	}

	// Create WhatsApp link object
	link := &WhatsAppLink{
		ID:            uuid.New(),
		ProductID:     req.ProductID,
		TransactionID: req.TransactionID,
		InquiryID:     req.InquiryID,
		FromUserID:    fromUserID,
		ToUserID:      req.ToUserID,
		PhoneNumber:   req.PhoneNumber,
		Message:       message,
		WhatsAppURL:   whatsappURL,
		DeepLink:      deepLink,
		WebLink:       webLink,
		LinkType:      req.LinkType,
		Status:        "created",
		ClickCount:    0,
		ExpiresAt:     expiresAt,
		CreatedAt:     time.Now(),
	}

	// Save to database
	if err := s.saveWhatsAppLink(ctx, link); err != nil {
		return nil, fmt.Errorf("failed to save WhatsApp link: %w", err)
	}

	return link, nil
}

// GetWhatsAppLink retrieves a WhatsApp link by ID
func (s *Service) GetWhatsAppLink(ctx context.Context, linkID uuid.UUID) (*WhatsAppLink, error) {
	query := `
		SELECT id, product_id, transaction_id, inquiry_id, from_user_id, to_user_id,
			   phone_number, message, whatsapp_url, deep_link, web_link, link_type,
			   status, click_count, last_clicked_at, expires_at, created_at
		FROM whatsapp_links 
		WHERE id = $1`

	link := &WhatsAppLink{}
	err := s.db.QueryRowContext(ctx, query, linkID).Scan(
		&link.ID, &link.ProductID, &link.TransactionID, &link.InquiryID,
		&link.FromUserID, &link.ToUserID, &link.PhoneNumber, &link.Message,
		&link.WhatsAppURL, &link.DeepLink, &link.WebLink, &link.LinkType,
		&link.Status, &link.ClickCount, &link.LastClickedAt, &link.ExpiresAt,
		&link.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("WhatsApp link not found")
		}
		return nil, fmt.Errorf("failed to get WhatsApp link: %w", err)
	}

	return link, nil
}

// TrackLinkClick tracks when a WhatsApp link is clicked
func (s *Service) TrackLinkClick(ctx context.Context, linkID uuid.UUID) error {
	// Check if link exists and is not expired
	link, err := s.GetWhatsAppLink(ctx, linkID)
	if err != nil {
		return err
	}

	// Check expiration
	if link.ExpiresAt != nil && time.Now().After(*link.ExpiresAt) {
		// Mark as expired
		if err := s.updateLinkStatus(ctx, linkID, "expired"); err != nil {
			return fmt.Errorf("failed to mark link as expired: %w", err)
		}
		return fmt.Errorf("WhatsApp link has expired")
	}

	// Update click count and timestamp
	query := `
		UPDATE whatsapp_links 
		SET click_count = click_count + 1, 
			last_clicked_at = NOW(),
			status = CASE WHEN status = 'created' THEN 'clicked' ELSE status END
		WHERE id = $1`

	_, err = s.db.ExecContext(ctx, query, linkID)
	if err != nil {
		return fmt.Errorf("failed to track link click: %w", err)
	}

	return nil
}

// GetUserWhatsAppLinks retrieves WhatsApp links for a user
func (s *Service) GetUserWhatsAppLinks(ctx context.Context, userID uuid.UUID, linkType string, limit int) ([]*WhatsAppLink, error) {
	query := `
		SELECT id, product_id, transaction_id, inquiry_id, from_user_id, to_user_id,
			   phone_number, message, whatsapp_url, deep_link, web_link, link_type,
			   status, click_count, last_clicked_at, expires_at, created_at
		FROM whatsapp_links 
		WHERE (from_user_id = $1 OR to_user_id = $1)`

	args := []interface{}{userID}
	argIndex := 2

	if linkType != "" {
		query += fmt.Sprintf(" AND link_type = $%d", argIndex)
		args = append(args, linkType)
		argIndex++
	}

	query += " ORDER BY created_at DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, limit)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get user WhatsApp links: %w", err)
	}
	defer rows.Close()

	links := make([]*WhatsAppLink, 0)
	for rows.Next() {
		link := &WhatsAppLink{}
		err := rows.Scan(
			&link.ID, &link.ProductID, &link.TransactionID, &link.InquiryID,
			&link.FromUserID, &link.ToUserID, &link.PhoneNumber, &link.Message,
			&link.WhatsAppURL, &link.DeepLink, &link.WebLink, &link.LinkType,
			&link.Status, &link.ClickCount, &link.LastClickedAt, &link.ExpiresAt,
			&link.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan WhatsApp link: %w", err)
		}
		links = append(links, link)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate WhatsApp links: %w", err)
	}

	return links, nil
}

// ExpireOldLinks marks old links as expired
func (s *Service) ExpireOldLinks(ctx context.Context) error {
	query := `
		UPDATE whatsapp_links 
		SET status = 'expired' 
		WHERE expires_at IS NOT NULL 
		AND expires_at < NOW() 
		AND status != 'expired'`

	result, err := s.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to expire old links: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		fmt.Printf("Expired %d WhatsApp links\n", rowsAffected)
	}

	return nil
}

// GetLinkStats retrieves statistics about WhatsApp links
func (s *Service) GetLinkStats(ctx context.Context, userID *uuid.UUID, fromDate, toDate time.Time) (*LinkStats, error) {
	whereClause := "WHERE created_at >= $1 AND created_at <= $2"
	args := []interface{}{fromDate, toDate}
	argIndex := 3

	if userID != nil {
		whereClause += fmt.Sprintf(" AND (from_user_id = $%d OR to_user_id = $%d)", argIndex, argIndex)
		args = append(args, *userID)
		argIndex++
	}

	query := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total_links,
			COUNT(CASE WHEN status = 'clicked' THEN 1 END) as clicked_links,
			COUNT(CASE WHEN status = 'expired' THEN 1 END) as expired_links,
			COUNT(CASE WHEN link_type = 'inquiry' THEN 1 END) as inquiry_links,
			COUNT(CASE WHEN link_type = 'transaction' THEN 1 END) as transaction_links,
			COUNT(CASE WHEN link_type = 'business' THEN 1 END) as business_links,
			COALESCE(SUM(click_count), 0) as total_clicks,
			COALESCE(AVG(click_count), 0) as avg_clicks_per_link
		FROM whatsapp_links %s`, whereClause)

	stats := &LinkStats{}
	err := s.db.QueryRowContext(ctx, query, args...).Scan(
		&stats.TotalLinks, &stats.ClickedLinks, &stats.ExpiredLinks,
		&stats.InquiryLinks, &stats.TransactionLinks, &stats.BusinessLinks,
		&stats.TotalClicks, &stats.AvgClicksPerLink)

	if err != nil {
		return nil, fmt.Errorf("failed to get link stats: %w", err)
	}

	// Calculate click rate
	if stats.TotalLinks > 0 {
		stats.ClickRate = float64(stats.ClickedLinks) / float64(stats.TotalLinks) * 100
	}

	return stats, nil
}

// Helper functions
func (s *Service) saveWhatsAppLink(ctx context.Context, link *WhatsAppLink) error {
	query := `
		INSERT INTO whatsapp_links (
			id, product_id, transaction_id, inquiry_id, from_user_id, to_user_id,
			phone_number, message, whatsapp_url, deep_link, web_link, link_type,
			status, click_count, expires_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`

	_, err := s.db.ExecContext(ctx, query,
		link.ID, link.ProductID, link.TransactionID, link.InquiryID,
		link.FromUserID, link.ToUserID, link.PhoneNumber, link.Message,
		link.WhatsAppURL, link.DeepLink, link.WebLink, link.LinkType,
		link.Status, link.ClickCount, link.ExpiresAt, link.CreatedAt)

	return err
}

func (s *Service) updateLinkStatus(ctx context.Context, linkID uuid.UUID, status string) error {
	query := `UPDATE whatsapp_links SET status = $1 WHERE id = $2`
	_, err := s.db.ExecContext(ctx, query, status, linkID)
	return err
}

// Migration for whatsapp_links table
func CreateWhatsAppLinksTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS whatsapp_links (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			product_id UUID REFERENCES products(id) ON DELETE SET NULL,
			transaction_id UUID REFERENCES transactions(id) ON DELETE SET NULL,
			inquiry_id UUID REFERENCES product_inquiries(id) ON DELETE SET NULL,
			from_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			to_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			phone_number VARCHAR(20) NOT NULL,
			message TEXT NOT NULL,
			whatsapp_url TEXT NOT NULL,
			deep_link TEXT NOT NULL,
			web_link TEXT NOT NULL,
			link_type VARCHAR(20) NOT NULL CHECK (link_type IN ('inquiry', 'transaction', 'business')),
			status VARCHAR(20) DEFAULT 'created' CHECK (status IN ('created', 'clicked', 'expired')),
			click_count INTEGER DEFAULT 0,
			last_clicked_at TIMESTAMP WITH TIME ZONE,
			expires_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_whatsapp_links_from_user ON whatsapp_links(from_user_id);
		CREATE INDEX IF NOT EXISTS idx_whatsapp_links_to_user ON whatsapp_links(to_user_id);
		CREATE INDEX IF NOT EXISTS idx_whatsapp_links_product ON whatsapp_links(product_id);
		CREATE INDEX IF NOT EXISTS idx_whatsapp_links_transaction ON whatsapp_links(transaction_id);
		CREATE INDEX IF NOT EXISTS idx_whatsapp_links_created_at ON whatsapp_links(created_at);
		CREATE INDEX IF NOT EXISTS idx_whatsapp_links_expires_at ON whatsapp_links(expires_at) WHERE expires_at IS NOT NULL;`

	_, err := db.Exec(query)
	return err
}

type LinkStats struct {
	TotalLinks         int     `json:"total_links"`
	ClickedLinks       int     `json:"clicked_links"`
	ExpiredLinks       int     `json:"expired_links"`
	InquiryLinks       int     `json:"inquiry_links"`
	TransactionLinks   int     `json:"transaction_links"`
	BusinessLinks      int     `json:"business_links"`
	TotalClicks        int     `json:"total_clicks"`
	AvgClicksPerLink   float64 `json:"avg_clicks_per_link"`
	ClickRate          float64 `json:"click_rate"` // percentage
}