package users

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// CreateUser creates a new user in the database
func (r *Repository) CreateUser(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (
			id, email, password_hash, first_name, last_name, phone, cuit,
			business_name, business_type, province, city, address, coordinates,
			role, verification_documents, preferences
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, 
			ST_GeomFromText('POINT(' || $13 || ' ' || $14 || ')', 4326),
			$15, $16, $17
		)`

	var lng, lat sql.NullFloat64
	if user.Coordinates != nil {
		lng.Float64 = user.Coordinates.Lng
		lng.Valid = true
		lat.Float64 = user.Coordinates.Lat
		lat.Valid = true
	}

	var verificationDocsJSON, preferencesJSON []byte
	var err error

	if user.VerificationDocuments != nil {
		verificationDocsJSON, err = json.Marshal(user.VerificationDocuments)
		if err != nil {
			return fmt.Errorf("failed to marshal verification documents: %w", err)
		}
	}

	if user.Preferences != nil {
		preferencesJSON, err = json.Marshal(user.Preferences)
		if err != nil {
			return fmt.Errorf("failed to marshal preferences: %w", err)
		}
	}

	_, err = r.db.ExecContext(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.FirstName, user.LastName,
		user.Phone, user.CUIT, user.BusinessName, user.BusinessType,
		user.Province, user.City, user.Address, lng, lat, user.Role,
		verificationDocsJSON, preferencesJSON)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetUserByID retrieves a user by their ID
func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	query := `
		SELECT 
			id, email, password_hash, first_name, last_name, phone, cuit,
			business_name, business_type, tax_category, province, city, address,
			ST_X(coordinates) as lng, ST_Y(coordinates) as lat,
			role, verification_level, is_active, is_verified, rating,
			total_sales, total_purchases, total_reviews, created_at, updated_at,
			last_login, verification_documents, preferences
		FROM users 
		WHERE id = $1 AND is_active = true`

	user := &User{}
	var lng, lat sql.NullFloat64
	var verificationDocsJSON, preferencesJSON sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName,
		&user.Phone, &user.CUIT, &user.BusinessName, &user.BusinessType,
		&user.TaxCategory, &user.Province, &user.City, &user.Address,
		&lng, &lat, &user.Role, &user.VerificationLevel, &user.IsActive,
		&user.IsVerified, &user.Rating, &user.TotalSales, &user.TotalPurchases,
		&user.TotalReviews, &user.CreatedAt, &user.UpdatedAt, &user.LastLogin,
		&verificationDocsJSON, &preferencesJSON)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	// Parse coordinates
	if lng.Valid && lat.Valid {
		user.Coordinates = &Point{
			Lng: lng.Float64,
			Lat: lat.Float64,
		}
	}

	// Parse JSON fields
	if verificationDocsJSON.Valid && verificationDocsJSON.String != "" {
		if err := json.Unmarshal([]byte(verificationDocsJSON.String), &user.VerificationDocuments); err != nil {
			return nil, fmt.Errorf("failed to unmarshal verification documents: %w", err)
		}
	}

	if preferencesJSON.Valid && preferencesJSON.String != "" {
		if err := json.Unmarshal([]byte(preferencesJSON.String), &user.Preferences); err != nil {
			return nil, fmt.Errorf("failed to unmarshal preferences: %w", err)
		}
	}

	return user, nil
}

// GetUserByEmail retrieves a user by their email
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT 
			id, email, password_hash, first_name, last_name, phone, cuit,
			business_name, business_type, tax_category, province, city, address,
			ST_X(coordinates) as lng, ST_Y(coordinates) as lat,
			role, verification_level, is_active, is_verified, rating,
			total_sales, total_purchases, total_reviews, created_at, updated_at,
			last_login, verification_documents, preferences
		FROM users 
		WHERE email = $1 AND is_active = true`

	user := &User{}
	var lng, lat sql.NullFloat64
	var verificationDocsJSON, preferencesJSON sql.NullString

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName,
		&user.Phone, &user.CUIT, &user.BusinessName, &user.BusinessType,
		&user.TaxCategory, &user.Province, &user.City, &user.Address,
		&lng, &lat, &user.Role, &user.VerificationLevel, &user.IsActive,
		&user.IsVerified, &user.Rating, &user.TotalSales, &user.TotalPurchases,
		&user.TotalReviews, &user.CreatedAt, &user.UpdatedAt, &user.LastLogin,
		&verificationDocsJSON, &preferencesJSON)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	// Parse coordinates
	if lng.Valid && lat.Valid {
		user.Coordinates = &Point{
			Lng: lng.Float64,
			Lat: lat.Float64,
		}
	}

	// Parse JSON fields
	if verificationDocsJSON.Valid && verificationDocsJSON.String != "" {
		if err := json.Unmarshal([]byte(verificationDocsJSON.String), &user.VerificationDocuments); err != nil {
			return nil, fmt.Errorf("failed to unmarshal verification documents: %w", err)
		}
	}

	if preferencesJSON.Valid && preferencesJSON.String != "" {
		if err := json.Unmarshal([]byte(preferencesJSON.String), &user.Preferences); err != nil {
			return nil, fmt.Errorf("failed to unmarshal preferences: %w", err)
		}
	}

	return user, nil
}

// GetUserByCUIT retrieves a user by their CUIT
func (r *Repository) GetUserByCUIT(ctx context.Context, cuit string) (*User, error) {
	query := `
		SELECT 
			id, email, password_hash, first_name, last_name, phone, cuit,
			business_name, business_type, tax_category, province, city, address,
			ST_X(coordinates) as lng, ST_Y(coordinates) as lat,
			role, verification_level, is_active, is_verified, rating,
			total_sales, total_purchases, total_reviews, created_at, updated_at,
			last_login, verification_documents, preferences
		FROM users 
		WHERE cuit = $1 AND is_active = true`

	user := &User{}
	var lng, lat sql.NullFloat64
	var verificationDocsJSON, preferencesJSON sql.NullString

	err := r.db.QueryRowContext(ctx, query, cuit).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName,
		&user.Phone, &user.CUIT, &user.BusinessName, &user.BusinessType,
		&user.TaxCategory, &user.Province, &user.City, &user.Address,
		&lng, &lat, &user.Role, &user.VerificationLevel, &user.IsActive,
		&user.IsVerified, &user.Rating, &user.TotalSales, &user.TotalPurchases,
		&user.TotalReviews, &user.CreatedAt, &user.UpdatedAt, &user.LastLogin,
		&verificationDocsJSON, &preferencesJSON)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by CUIT: %w", err)
	}

	// Parse coordinates and JSON fields (same as above methods)
	if lng.Valid && lat.Valid {
		user.Coordinates = &Point{
			Lng: lng.Float64,
			Lat: lat.Float64,
		}
	}

	if verificationDocsJSON.Valid && verificationDocsJSON.String != "" {
		if err := json.Unmarshal([]byte(verificationDocsJSON.String), &user.VerificationDocuments); err != nil {
			return nil, fmt.Errorf("failed to unmarshal verification documents: %w", err)
		}
	}

	if preferencesJSON.Valid && preferencesJSON.String != "" {
		if err := json.Unmarshal([]byte(preferencesJSON.String), &user.Preferences); err != nil {
			return nil, fmt.Errorf("failed to unmarshal preferences: %w", err)
		}
	}

	return user, nil
}

// UpdateUser updates an existing user
func (r *Repository) UpdateUser(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	setParts := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)
	argIndex := 1

	for field, value := range updates {
		if field == "coordinates" && value != nil {
			if coords, ok := value.(*Point); ok {
				setParts = append(setParts, fmt.Sprintf("coordinates = ST_GeomFromText('POINT(%f %f)', 4326)", coords.Lng, coords.Lat))
				continue
			}
		}
		setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
		args = append(args, value)
		argIndex++
	}

	// Always update the updated_at timestamp
	setParts = append(setParts, "updated_at = NOW()")

	query := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIndex)
	args = append(args, id)

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// UpdateLastLogin updates the last login timestamp for a user
func (r *Repository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET last_login = NOW(), updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}
	return nil
}

// ListUsers retrieves users with filtering and pagination
func (r *Repository) ListUsers(ctx context.Context, filters UserFilters, limit, offset int) ([]*User, int, error) {
	whereConditions := []string{"is_active = true"}
	args := []interface{}{}
	argIndex := 1

	if filters.Role != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("role = $%d", argIndex))
		args = append(args, filters.Role)
		argIndex++
	}

	if filters.Province != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("province = $%d", argIndex))
		args = append(args, filters.Province)
		argIndex++
	}

	if filters.VerificationLevel > 0 {
		whereConditions = append(whereConditions, fmt.Sprintf("verification_level >= $%d", argIndex))
		args = append(args, filters.VerificationLevel)
		argIndex++
	}

	if filters.IsVerified != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("is_verified = $%d", argIndex))
		args = append(args, *filters.IsVerified)
		argIndex++
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users WHERE %s", whereClause)
	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Get paginated results
	query := fmt.Sprintf(`
		SELECT 
			id, email, first_name, last_name, phone, cuit,
			business_name, business_type, tax_category, province, city, address,
			ST_X(coordinates) as lng, ST_Y(coordinates) as lat,
			role, verification_level, is_active, is_verified, rating,
			total_sales, total_purchases, total_reviews, created_at, updated_at,
			last_login
		FROM users 
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	users := make([]*User, 0)
	for rows.Next() {
		user := &User{}
		var lng, lat sql.NullFloat64

		err := rows.Scan(
			&user.ID, &user.Email, &user.FirstName, &user.LastName,
			&user.Phone, &user.CUIT, &user.BusinessName, &user.BusinessType,
			&user.TaxCategory, &user.Province, &user.City, &user.Address,
			&lng, &lat, &user.Role, &user.VerificationLevel, &user.IsActive,
			&user.IsVerified, &user.Rating, &user.TotalSales, &user.TotalPurchases,
			&user.TotalReviews, &user.CreatedAt, &user.UpdatedAt, &user.LastLogin)

		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}

		if lng.Valid && lat.Valid {
			user.Coordinates = &Point{
				Lng: lng.Float64,
				Lat: lat.Float64,
			}
		}

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate users: %w", err)
	}

	return users, totalCount, nil
}

// DeleteUser soft deletes a user by setting is_active to false
func (r *Repository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET is_active = false, updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

type UserFilters struct {
	Role              string `json:"role"`
	Province          string `json:"province"`
	VerificationLevel int    `json:"verification_level"`
	IsVerified        *bool  `json:"is_verified"`
}