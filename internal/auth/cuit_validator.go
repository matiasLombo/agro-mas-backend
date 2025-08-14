package auth

import (
	"errors"
	"regexp"
	"strconv"
)

var (
	ErrInvalidCUITFormat = errors.New("CUIT must be 11 digits")
	ErrInvalidCUITChecksum = errors.New("invalid CUIT checksum")
)

// CUITValidator validates Argentine CUIT (tax identification numbers)
type CUITValidator struct{}

func NewCUITValidator() *CUITValidator {
	return &CUITValidator{}
}

// ValidateCUIT validates a CUIT number according to Argentine tax authority rules
func (cv *CUITValidator) ValidateCUIT(cuit string) error {
	// Remove any non-digit characters
	reg := regexp.MustCompile(`\D`)
	cleanCUIT := reg.ReplaceAllString(cuit, "")

	// Check if it's exactly 11 digits
	if len(cleanCUIT) != 11 {
		return ErrInvalidCUITFormat
	}

	// Convert to slice of integers
	digits := make([]int, 11)
	for i, char := range cleanCUIT {
		digit, err := strconv.Atoi(string(char))
		if err != nil {
			return ErrInvalidCUITFormat
		}
		digits[i] = digit
	}

	// Validate checksum using Argentine algorithm
	if !cv.validateChecksum(digits) {
		return ErrInvalidCUITChecksum
	}

	return nil
}

// validateChecksum validates the CUIT checksum according to Argentine algorithm
func (cv *CUITValidator) validateChecksum(digits []int) bool {
	// CUIT validation multipliers
	multipliers := []int{5, 4, 3, 2, 7, 6, 5, 4, 3, 2}

	sum := 0
	for i := 0; i < 10; i++ {
		sum += digits[i] * multipliers[i]
	}

	remainder := sum % 11
	checkDigit := 11 - remainder

	// Special cases for check digit
	if checkDigit == 11 {
		checkDigit = 0
	} else if checkDigit == 10 {
		// For CUIT, check digit 10 is invalid
		return false
	}

	return digits[10] == checkDigit
}

// FormatCUIT formats a CUIT with standard Argentine formatting (XX-XXXXXXXX-X)
func (cv *CUITValidator) FormatCUIT(cuit string) string {
	// Remove any non-digit characters
	reg := regexp.MustCompile(`\D`)
	cleanCUIT := reg.ReplaceAllString(cuit, "")

	if len(cleanCUIT) != 11 {
		return cuit // Return original if invalid
	}

	return cleanCUIT[:2] + "-" + cleanCUIT[2:10] + "-" + cleanCUIT[10:]
}

// GetCUITType returns the type of CUIT based on the first two digits
func (cv *CUITValidator) GetCUITType(cuit string) string {
	// Remove any non-digit characters
	reg := regexp.MustCompile(`\D`)
	cleanCUIT := reg.ReplaceAllString(cuit, "")

	if len(cleanCUIT) != 11 {
		return "invalid"
	}

	prefix := cleanCUIT[:2]

	switch prefix {
	case "20":
		return "individual_male" // Masculine individual
	case "27":
		return "individual_female" // Feminine individual
	case "23":
		return "individual_male_single" // Masculine individual (single)
	case "24":
		return "individual_female_single" // Feminine individual (single)
	case "30":
		return "company" // Company
	case "33":
		return "company_public" // Public company
	case "34":
		return "company_foreign" // Foreign company
	default:
		if prefix >= "50" && prefix <= "59" {
			return "association" // Associations and foundations
		}
		return "other"
	}
}

// IsCompanyCUIT checks if the CUIT belongs to a company (not individual)
func (cv *CUITValidator) IsCompanyCUIT(cuit string) bool {
	cuitType := cv.GetCUITType(cuit)
	return cuitType == "company" || cuitType == "company_public" || 
		   cuitType == "company_foreign" || cuitType == "association"
}

// IsIndividualCUIT checks if the CUIT belongs to an individual
func (cv *CUITValidator) IsIndividualCUIT(cuit string) bool {
	cuitType := cv.GetCUITType(cuit)
	return cuitType == "individual_male" || cuitType == "individual_female" ||
		   cuitType == "individual_male_single" || cuitType == "individual_female_single"
}