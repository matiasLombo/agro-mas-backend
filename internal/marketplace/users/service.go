package users

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"agro-mas-backend/internal/auth"
	"github.com/google/uuid"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrUserExists       = errors.New("user already exists")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrUserNotActive    = errors.New("user account is not active")
	ErrInvalidRole      = errors.New("invalid user role")
	ErrCUITExists       = errors.New("CUIT already registered")
)

type Service struct {
	repo            *Repository
	passwordManager *auth.PasswordManager
	jwtManager      *auth.JWTManager
	cuitValidator   *auth.CUITValidator
}

func NewService(repo *Repository, passwordManager *auth.PasswordManager, jwtManager *auth.JWTManager) *Service {
	return &Service{
		repo:            repo,
		passwordManager: passwordManager,
		jwtManager:      jwtManager,
		cuitValidator:   auth.NewCUITValidator(),
	}
}

// CreateUser creates a new user account with validation
func (s *Service) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
	// Validate password strength
	if err := auth.ValidatePasswordStrength(req.Password); err != nil {
		return nil, fmt.Errorf("password validation failed: %w", err)
	}

	// Validate role
	if req.Role != "buyer" && req.Role != "seller" {
		return nil, ErrInvalidRole
	}

	// Check if email already exists
	existingUser, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, ErrUserExists
	}

	// Validate and check CUIT if provided
	if req.CUIT != nil && *req.CUIT != "" {
		if err := s.cuitValidator.ValidateCUIT(*req.CUIT); err != nil {
			return nil, fmt.Errorf("CUIT validation failed: %w", err)
		}

		// Format CUIT
		formattedCUIT := s.cuitValidator.FormatCUIT(*req.CUIT)
		req.CUIT = &formattedCUIT

		// Check if CUIT already exists
		existingCUITUser, err := s.repo.GetUserByCUIT(ctx, *req.CUIT)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing CUIT: %w", err)
		}
		if existingCUITUser != nil {
			return nil, ErrCUITExists
		}

		// For sellers with CUIT, set higher verification level and determine business type
		if req.Role == "seller" {
			if s.cuitValidator.IsCompanyCUIT(*req.CUIT) && req.BusinessType == nil {
				businessType := "company"
				req.BusinessType = &businessType
			}
		}
	}

	// Hash password
	passwordHash, err := s.passwordManager.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user object
	user := &User{
		ID:           uuid.New(),
		Email:        strings.ToLower(req.Email),
		PasswordHash: passwordHash,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Phone:        req.Phone,
		CUIT:         req.CUIT,
		CBU:          req.CBU,
		CBUAlias:     req.CBUAlias,
		BankName:     req.BankName,
		BusinessName: req.BusinessName,
		BusinessType: req.BusinessType,
		Province:     req.Province,
		City:         req.City,
		Address:      req.Address,
		Coordinates:  req.Coordinates,
		Role:         req.Role,
		VerificationLevel: 0,
		IsActive:     true,
		IsVerified:   false,
		Rating:       0.0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Preferences: &UserPreferences{
			NotificationEmail:    true,
			NotificationWhatsApp: true,
			SearchRadius:         50, // 50km default
			Language:             "es",
			Currency:             "ARS",
			PrivacyLevel:         "limited",
		},
	}

	// Set initial verification level based on available information
	if user.CUIT != nil {
		user.VerificationLevel = 1 // Email + CUIT provided
	}

	// Create user in database
	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user in database: %w", err)
	}

	return user, nil
}

// Authenticate validates user credentials and returns a JWT token
func (s *Service) Authenticate(ctx context.Context, req *LoginRequest) (*auth.TokenResponse, *User, error) {
	// Get user by email
	user, err := s.repo.GetUserByEmail(ctx, strings.ToLower(req.Email))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, nil, ErrUserNotFound
	}

	// Check if user is active
	if !user.IsActive {
		return nil, nil, ErrUserNotActive
	}

	// Verify password
	isValid, err := s.passwordManager.ComparePasswordAndHash(req.Password, user.PasswordHash)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to verify password: %w", err)
	}
	if !isValid {
		return nil, nil, ErrInvalidPassword
	}

	// Generate JWT token
	tokenResponse, err := s.jwtManager.GenerateToken(
		user.ID,
		user.Email,
		user.Role,
		user.CUIT,
		user.Province,
		user.VerificationLevel,
		user.IsVerified,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Update last login
	if err := s.repo.UpdateLastLogin(ctx, user.ID); err != nil {
		// Log error but don't fail authentication
		// In production, you'd use a proper logger
		fmt.Printf("Failed to update last login for user %s: %v\n", user.ID, err)
	}

	return tokenResponse, user, nil
}

// GetUserByID retrieves a user by their ID
func (s *Service) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// GetPublicUserByID retrieves public user information by ID
func (s *Service) GetPublicUserByID(ctx context.Context, id uuid.UUID) (*PublicUserResponse, error) {
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user.ToPublicResponse(), nil
}

// UpdateUser updates user information
func (s *Service) UpdateUser(ctx context.Context, id uuid.UUID, req *UpdateUserRequest) (*User, error) {
	// Get existing user
	existingUser, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if existingUser == nil {
		return nil, ErrUserNotFound
	}

	// Prepare updates map
	updates := make(map[string]interface{})

	if req.FirstName != nil {
		updates["first_name"] = *req.FirstName
	}
	if req.LastName != nil {
		updates["last_name"] = *req.LastName
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.CBU != nil {
		updates["cbu"] = *req.CBU
	}
	if req.CBUAlias != nil {
		updates["cbu_alias"] = *req.CBUAlias
	}
	if req.BankName != nil {
		updates["bank_name"] = *req.BankName
	}
	if req.BusinessName != nil {
		updates["business_name"] = *req.BusinessName
	}
	if req.BusinessType != nil {
		updates["business_type"] = *req.BusinessType
	}
	if req.Province != nil {
		updates["province"] = *req.Province
	}
	if req.City != nil {
		updates["city"] = *req.City
	}
	if req.Address != nil {
		updates["address"] = *req.Address
	}
	if req.Coordinates != nil {
		updates["coordinates"] = req.Coordinates
	}

	// Update user in database
	if err := s.repo.UpdateUser(ctx, id, updates); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Return updated user
	return s.repo.GetUserByID(ctx, id)
}

// ChangePassword allows a user to change their password
func (s *Service) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	// Get user
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Verify current password
	isValid, err := s.passwordManager.ComparePasswordAndHash(currentPassword, user.PasswordHash)
	if err != nil {
		return fmt.Errorf("failed to verify current password: %w", err)
	}
	if !isValid {
		return ErrInvalidPassword
	}

	// Validate new password strength
	if err := auth.ValidatePasswordStrength(newPassword); err != nil {
		return fmt.Errorf("new password validation failed: %w", err)
	}

	// Hash new password
	newPasswordHash, err := s.passwordManager.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update password in database
	updates := map[string]interface{}{
		"password_hash": newPasswordHash,
	}

	if err := s.repo.UpdateUser(ctx, userID, updates); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// ListUsers retrieves users with filtering and pagination
func (s *Service) ListUsers(ctx context.Context, filters UserFilters, page, pageSize int) ([]*User, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	users, totalCount, err := s.repo.ListUsers(ctx, filters, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, totalCount, nil
}

// UpdateVerificationLevel updates the verification level of a user (admin only)
func (s *Service) UpdateVerificationLevel(ctx context.Context, userID uuid.UUID, level int, isVerified bool) error {
	if level < 0 || level > 4 {
		return errors.New("verification level must be between 0 and 4")
	}

	updates := map[string]interface{}{
		"verification_level": level,
		"is_verified":        isVerified,
	}

	if err := s.repo.UpdateUser(ctx, userID, updates); err != nil {
		return fmt.Errorf("failed to update verification level: %w", err)
	}

	return nil
}

// IsSellerProfileComplete checks if a user has all required information to be a seller
func (s *Service) IsSellerProfileComplete(ctx context.Context, userID uuid.UUID) (bool, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return false, ErrUserNotFound
	}

	// Check required fields for seller profile
	requiredFields := []bool{
		user.CUIT != nil && *user.CUIT != "",
		user.CBU != nil && *user.CBU != "",
		user.BusinessName != nil && *user.BusinessName != "",
		user.Phone != nil && *user.Phone != "",
		user.Address != nil && *user.Address != "",
		user.City != nil && *user.City != "",
		user.Province != nil && *user.Province != "",
	}

	for _, required := range requiredFields {
		if !required {
			return false, nil
		}
	}

	return true, nil
}

// UpgradeToSeller upgrades a buyer to seller with required profile information
func (s *Service) UpgradeToSeller(ctx context.Context, userID uuid.UUID, req *SellerProfileRequest) error {
	// Get existing user
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Validate CUIT if provided
	if req.CUIT != "" {
		if err := s.cuitValidator.ValidateCUIT(req.CUIT); err != nil {
			return fmt.Errorf("CUIT validation failed: %w", err)
		}

		// Clean CUIT (remove dashes for database storage)
		cleanCUIT := strings.ReplaceAll(req.CUIT, "-", "")
		req.CUIT = cleanCUIT

		// Check if CUIT already exists (but not for this user)
		existingCUITUser, err := s.repo.GetUserByCUIT(ctx, req.CUIT)
		if err != nil {
			return fmt.Errorf("failed to check existing CUIT: %w", err)
		}
		if existingCUITUser != nil && existingCUITUser.ID != userID {
			return ErrCUITExists
		}
	}

	// Clean and validate CBU format (remove spaces for validation)
	if req.CBU != "" {
		cleanCBU := strings.ReplaceAll(req.CBU, " ", "")
		if len(cleanCBU) != 22 {
			return errors.New("CBU must be 22 digits long")
		}
		// Store the clean CBU (without spaces)
		req.CBU = cleanCBU
	}

	// Prepare updates for seller profile
	updates := map[string]interface{}{
		"role":                   "seller",
		"cuit":                   req.CUIT,
		"cbu":                    req.CBU,
		"cbu_alias":              req.CBUAlias,
		"bank_name":              req.BankName,
		"renspa":                 req.RENSPA,
		"establishment_name":     req.EstablishmentName,
		"establishment_location": req.EstablishmentLocation,
		"business_name":          req.BusinessName,
		"business_type":          req.BusinessType,
		"phone":                  req.Phone,
		"address":                req.Address,
		"city":                   req.City,
		"province":               req.Province,
	}

	// Update user in database
	if err := s.repo.UpdateUser(ctx, userID, updates); err != nil {
		// Log the error details for debugging
		fmt.Printf("Failed to upgrade user to seller. UserID: %s, Updates: %+v, Error: %v\n", userID.String(), updates, err)
		return fmt.Errorf("failed to upgrade user to seller: %w", err)
	}

	return nil
}

// GetSellerProfile gets the seller profile information for a user
func (s *Service) GetSellerProfile(ctx context.Context, userID uuid.UUID) (*SellerProfileResponse, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	isComplete, err := s.IsSellerProfileComplete(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &SellerProfileResponse{
		CUIT:                  user.CUIT,
		CBU:                   user.CBU,
		CBUAlias:              user.CBUAlias,
		BankName:              user.BankName,
		RENSPA:                user.RENSPA,
		EstablishmentName:     user.EstablishmentName,
		EstablishmentLocation: user.EstablishmentLocation,
		BusinessName:          user.BusinessName,
		BusinessType:          user.BusinessType,
		Phone:                 user.Phone,
		Address:               user.Address,
		City:                  user.City,
		Province:              user.Province,
		IsComplete:            isComplete,
	}, nil
}

// DeactivateUser deactivates a user account
func (s *Service) DeactivateUser(ctx context.Context, userID uuid.UUID) error {
	if err := s.repo.DeleteUser(ctx, userID); err != nil {
		return fmt.Errorf("failed to deactivate user: %w", err)
	}
	return nil
}

// GenerateTokenForUser generates a JWT token for a given user (used in registration)
func (s *Service) GenerateTokenForUser(ctx context.Context, user *User) (*auth.TokenResponse, error) {
	// Generate JWT token
	tokenResponse, err := s.jwtManager.GenerateToken(
		user.ID,
		user.Email,
		user.Role,
		user.CUIT,
		user.Province,
		user.VerificationLevel,
		user.IsVerified,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return tokenResponse, nil
}