package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
	ErrInvalidClaims = errors.New("invalid token claims")
)

type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
}

type UserClaims struct {
	UserID              uuid.UUID `json:"user_id"`
	Email               string    `json:"email"`
	Role                string    `json:"role"`
	CUIT                *string   `json:"cuit,omitempty"`
	Province            *string   `json:"province,omitempty"`
	VerificationLevel   int       `json:"verification_level"`
	IsVerified          bool      `json:"is_verified"`
	jwt.RegisteredClaims
}

type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

func NewJWTManager(secretKey string, tokenDuration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:     secretKey,
		tokenDuration: tokenDuration,
	}
}

func (manager *JWTManager) GenerateToken(
	userID uuid.UUID,
	email string,
	role string,
	cuit *string,
	province *string,
	verificationLevel int,
	isVerified bool,
) (*TokenResponse, error) {
	now := time.Now()
	expiresAt := now.Add(manager.tokenDuration)

	claims := UserClaims{
		UserID:            userID,
		Email:             email,
		Role:              role,
		CUIT:              cuit,
		Province:          province,
		VerificationLevel: verificationLevel,
		IsVerified:        isVerified,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
			Subject:   userID.String(),
			Issuer:    "agro-mas-backend",
			Audience:  jwt.ClaimStrings{"agro-mas-frontend"},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString([]byte(manager.secretKey))
	if err != nil {
		return nil, err
	}

	// Generate refresh token with longer expiration
	refreshExpiresAt := now.Add(time.Hour * 24 * 7) // 7 days
	refreshClaims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(refreshExpiresAt),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ID:        uuid.New().String(),
		Subject:   userID.String(),
		Issuer:    "agro-mas-backend",
		Audience:  jwt.ClaimStrings{"agro-mas-refresh"},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(manager.secretKey))
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
		ExpiresAt:    expiresAt,
		TokenType:    "Bearer",
	}, nil
}

func (manager *JWTManager) VerifyToken(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&UserClaims{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, ErrInvalidToken
			}
			return []byte(manager.secretKey), nil
		},
	)

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, ErrInvalidClaims
	}

	if time.Now().After(claims.ExpiresAt.Time) {
		return nil, ErrExpiredToken
	}

	return claims, nil
}

func (manager *JWTManager) RefreshToken(refreshTokenString string) (*TokenResponse, error) {
	token, err := jwt.ParseWithClaims(
		refreshTokenString,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, ErrInvalidToken
			}
			return []byte(manager.secretKey), nil
		},
	)

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return nil, ErrInvalidClaims
	}

	if time.Now().After(claims.ExpiresAt.Time) {
		return nil, ErrExpiredToken
	}

	// Validate audience for refresh token
	validAudience := false
	for _, aud := range claims.Audience {
		if aud == "agro-mas-refresh" {
			validAudience = true
			break
		}
	}
	if !validAudience {
		return nil, ErrInvalidToken
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, ErrInvalidClaims
	}

	// Here you would typically fetch the user from database to get updated information
	// For now, we'll return a placeholder response
	// In a real implementation, you'd call a user service to get current user data
	return manager.GenerateToken(
		userID,
		"", // These would be fetched from database
		"",
		nil,
		nil,
		0,
		false,
	)
}

// ExtractTokenFromHeader extracts JWT token from Authorization header
func ExtractTokenFromHeader(header string) (string, error) {
	if len(header) < 7 || header[:7] != "Bearer " {
		return "", ErrInvalidToken
	}
	return header[7:], nil
}

// ValidateRole checks if the user has the required role
func ValidateRole(userRole string, requiredRoles ...string) bool {
	for _, role := range requiredRoles {
		if userRole == role {
			return true
		}
	}
	return false
}

// ValidateVerificationLevel checks if the user meets minimum verification level
func ValidateVerificationLevel(userLevel, requiredLevel int) bool {
	return userLevel >= requiredLevel
}