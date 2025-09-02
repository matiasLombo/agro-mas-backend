package whatsapp

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

type Client struct {
	apiURL         string
	businessNumber string
}

type MessageTemplate struct {
	ProductTitle    string
	ProductCategory string
	ProductPrice    string
	SellerName      string
	BuyerName       string
	InquiryType     string
	CustomMessage   string
}

type ClickToCallConfig struct {
	PhoneNumber string
	Message     string
}

func NewClient(apiURL, businessNumber string) *Client {
	return &Client{
		apiURL:         apiURL,
		businessNumber: businessNumber,
	}
}

// GenerateWhatsAppURL generates a WhatsApp click-to-chat URL
func (c *Client) GenerateWhatsAppURL(phoneNumber, message string) (string, error) {
	// Clean and validate phone number
	cleanedPhone, err := c.cleanPhoneNumber(phoneNumber)
	if err != nil {
		return "", fmt.Errorf("invalid phone number: %w", err)
	}

	// Encode message for URL
	encodedMessage := url.QueryEscape(message)

	// Generate WhatsApp URL
	whatsappURL := fmt.Sprintf("https://wa.me/%s?text=%s", cleanedPhone, encodedMessage)
	
	return whatsappURL, nil
}

// GenerateProductInquiryURL generates a WhatsApp URL for product inquiries
func (c *Client) GenerateProductInquiryURL(phoneNumber string, template MessageTemplate) (string, error) {
	message := c.buildProductInquiryMessage(template)
	return c.GenerateWhatsAppURL(phoneNumber, message)
}

// GenerateTransactionURL generates a WhatsApp URL for transaction-related communications
func (c *Client) GenerateTransactionURL(phoneNumber string, template MessageTemplate) (string, error) {
	message := c.buildTransactionMessage(template)
	return c.GenerateWhatsAppURL(phoneNumber, message)
}

// GenerateBusinessContactURL generates a WhatsApp URL for general business contact
func (c *Client) GenerateBusinessContactURL(phoneNumber string, businessName, inquiryType string) (string, error) {
	message := c.buildBusinessContactMessage(businessName, inquiryType)
	return c.GenerateWhatsAppURL(phoneNumber, message)
}

// BuildProductInquiryMessage creates a formatted message for product inquiries
func (c *Client) buildProductInquiryMessage(template MessageTemplate) string {
	var message strings.Builder

	message.WriteString(fmt.Sprintf("🌾 *Consulta sobre producto - Agro Mas*\n\n"))
	
	if template.ProductTitle != "" {
		message.WriteString(fmt.Sprintf("📋 *Producto:* %s\n", template.ProductTitle))
	}
	
	if template.ProductCategory != "" {
		message.WriteString(fmt.Sprintf("🏷️ *Categoría:* %s\n", c.translateCategory(template.ProductCategory)))
	}
	
	if template.ProductPrice != "" {
		message.WriteString(fmt.Sprintf("💰 *Precio:* %s\n", template.ProductPrice))
	}
	
	message.WriteString("\n")
	
	if template.InquiryType != "" {
		message.WriteString(fmt.Sprintf("❓ *Tipo de consulta:* %s\n\n", c.translateInquiryType(template.InquiryType)))
	}
	
	if template.CustomMessage != "" {
		message.WriteString(fmt.Sprintf("💬 *Mensaje:*\n%s\n\n", template.CustomMessage))
	} else {
		message.WriteString("Hola! Me interesa este producto. ¿Podrías darme más información?\n\n")
	}
	
	if template.BuyerName != "" {
		message.WriteString(fmt.Sprintf("Saludos,\n%s", template.BuyerName))
	}
	
	return message.String()
}

// BuildTransactionMessage creates a formatted message for transaction communications
func (c *Client) buildTransactionMessage(template MessageTemplate) string {
	var message strings.Builder

	message.WriteString(fmt.Sprintf("🌾 *Transacción - Agro Mas*\n\n"))
	
	if template.ProductTitle != "" {
		message.WriteString(fmt.Sprintf("📋 *Producto:* %s\n", template.ProductTitle))
	}
	
	if template.SellerName != "" && template.BuyerName != "" {
		message.WriteString(fmt.Sprintf("👥 *Vendedor:* %s\n", template.SellerName))
		message.WriteString(fmt.Sprintf("👤 *Comprador:* %s\n", template.BuyerName))
	}
	
	message.WriteString("\n")
	
	if template.CustomMessage != "" {
		message.WriteString(fmt.Sprintf("💬 *Mensaje:*\n%s\n\n", template.CustomMessage))
	} else {
		message.WriteString("Hola! Me comunico contigo sobre la transacción de este producto.\n\n")
	}
	
	message.WriteString("📱 *Esta conversación está relacionada con una transacción en Agro Mas*")
	
	return message.String()
}

// BuildBusinessContactMessage creates a formatted message for general business contact
func (c *Client) buildBusinessContactMessage(businessName, inquiryType string) string {
	var message strings.Builder

	message.WriteString(fmt.Sprintf("🌾 *Contacto comercial - Agro Mas*\n\n"))
	
	if businessName != "" {
		message.WriteString(fmt.Sprintf("🏢 *Empresa:* %s\n", businessName))
	}
	
	if inquiryType != "" {
		message.WriteString(fmt.Sprintf("❓ *Tipo de consulta:* %s\n\n", c.translateInquiryType(inquiryType)))
	}
	
	message.WriteString("Hola! Me comunico contigo desde Agro Mas para hacer una consulta comercial.\n\n")
	message.WriteString("¿Podrías contactarme cuando tengas un momento disponible?\n\n")
	message.WriteString("Saludos!")
	
	return message.String()
}

// Helper functions
func (c *Client) cleanPhoneNumber(phoneNumber string) (string, error) {
	// Remove all non-digit characters
	reg := regexp.MustCompile(`\D`)
	cleaned := reg.ReplaceAllString(phoneNumber, "")
	
	// Handle Argentine phone numbers
	if strings.HasPrefix(cleaned, "54") {
		// Already has country code
		if len(cleaned) < 12 || len(cleaned) > 13 {
			return "", fmt.Errorf("invalid Argentine phone number length: %s", cleaned)
		}
	} else if strings.HasPrefix(cleaned, "0") {
		// Remove leading 0 and add country code
		cleaned = "54" + cleaned[1:]
	} else if len(cleaned) == 10 {
		// Add country code for 10-digit numbers
		cleaned = "54" + cleaned
	} else if len(cleaned) == 8 || len(cleaned) == 9 {
		// Add area code and country code for shorter numbers
		// This is a simplified approach - in production, you'd need proper area code mapping
		cleaned = "5411" + cleaned
	}
	
	// Final validation
	if len(cleaned) < 12 || len(cleaned) > 15 {
		return "", fmt.Errorf("invalid phone number format: %s", cleaned)
	}
	
	return cleaned, nil
}

func (c *Client) translateCategory(category string) string {
	translations := map[string]string{
		"transport":  "Transporte",
		"livestock":  "Ganadería",
		"supplies":   "Insumos",
	}
	
	if spanish, exists := translations[category]; exists {
		return spanish
	}
	return category
}

func (c *Client) translateInquiryType(inquiryType string) string {
	translations := map[string]string{
		"general":      "Consulta general",
		"price":        "Consulta de precio",
		"availability": "Consulta de disponibilidad",
		"technical":    "Consulta técnica",
		"logistics":    "Consulta logística",
	}
	
	if spanish, exists := translations[inquiryType]; exists {
		return spanish
	}
	return inquiryType
}

// FormatPhoneNumber formats a phone number for display
func (c *Client) FormatPhoneNumber(phoneNumber string) string {
	cleaned, err := c.cleanPhoneNumber(phoneNumber)
	if err != nil {
		return phoneNumber // Return original if cleaning fails
	}
	
	// Format Argentine numbers: +54 11 1234-5678
	if strings.HasPrefix(cleaned, "54") && len(cleaned) >= 12 {
		countryCode := cleaned[:2]
		areaCode := cleaned[2:4]
		number := cleaned[4:]
		
		// Add hyphen for better readability
		if len(number) >= 4 {
			formattedNumber := number[:len(number)-4] + "-" + number[len(number)-4:]
			return fmt.Sprintf("+%s %s %s", countryCode, areaCode, formattedNumber)
		}
		
		return fmt.Sprintf("+%s %s %s", countryCode, areaCode, number)
	}
	
	return "+" + cleaned
}

// ValidatePhoneNumber validates if a phone number is valid for WhatsApp
func (c *Client) ValidatePhoneNumber(phoneNumber string) error {
	_, err := c.cleanPhoneNumber(phoneNumber)
	return err
}

// GenerateDeepLink generates a deep link for WhatsApp
func (c *Client) GenerateDeepLink(phoneNumber, message string) (string, error) {
	cleanedPhone, err := c.cleanPhoneNumber(phoneNumber)
	if err != nil {
		return "", fmt.Errorf("invalid phone number: %w", err)
	}

	encodedMessage := url.QueryEscape(message)
	
	// Generate deep link for mobile apps
	deepLink := fmt.Sprintf("whatsapp://send?phone=%s&text=%s", cleanedPhone, encodedMessage)
	
	return deepLink, nil
}

// GenerateWebLink generates a web link for WhatsApp Web
func (c *Client) GenerateWebLink(phoneNumber, message string) (string, error) {
	cleanedPhone, err := c.cleanPhoneNumber(phoneNumber)
	if err != nil {
		return "", fmt.Errorf("invalid phone number: %w", err)
	}

	encodedMessage := url.QueryEscape(message)
	
	// Generate web link for WhatsApp Web
	webLink := fmt.Sprintf("https://web.whatsapp.com/send?phone=%s&text=%s", cleanedPhone, encodedMessage)
	
	return webLink, nil
}

// GenerateQRCodeData generates data for a WhatsApp QR code
func (c *Client) GenerateQRCodeData(phoneNumber, message string) (string, error) {
	return c.GenerateWhatsAppURL(phoneNumber, message)
}

// MessageTemplates provides pre-built message templates
type MessageTemplates struct{}

func (mt *MessageTemplates) PriceInquiry(productName, buyerName string) string {
	return fmt.Sprintf(
		"Hola! Soy %s y me interesa conocer el precio actual de %s.\n\n¿Podrías proporcionarme esta información?\n\nGracias!",
		buyerName, productName)
}

func (mt *MessageTemplates) AvailabilityInquiry(productName, quantity string, buyerName string) string {
	return fmt.Sprintf(
		"Hola! Soy %s y me gustaría saber si tienes disponible %s.\n\nCantidad necesaria: %s\n\n¿Podrías confirmar disponibilidad?\n\nSaludos!",
		buyerName, productName, quantity)
}

func (mt *MessageTemplates) LogisticsInquiry(productName, location string, buyerName string) string {
	return fmt.Sprintf(
		"Hola! Soy %s y estoy interesado en %s.\n\n¿Realizas envíos a %s? ¿Cuál sería el costo y tiempo de entrega?\n\nGracias!",
		buyerName, productName, location)
}

func (mt *MessageTemplates) TechnicalInquiry(productName, question string, buyerName string) string {
	return fmt.Sprintf(
		"Hola! Soy %s y tengo una consulta técnica sobre %s.\n\n%s\n\n¿Podrías ayudarme con esta información?\n\nSaludos!",
		buyerName, productName, question)
}

func (mt *MessageTemplates) TransactionConfirmation(productName, price string, buyerName string) string {
	return fmt.Sprintf(
		"Hola! Soy %s y quiero confirmar la compra de %s por %s.\n\n¿Podemos coordinar los detalles de la transacción?\n\nGracias!",
		buyerName, productName, price)
}

func (mt *MessageTemplates) DeliveryCoordination(productName, address string, buyerName string) string {
	return fmt.Sprintf(
		"Hola! Soy %s y necesito coordinar la entrega de %s.\n\nDirección: %s\n\n¿Cuándo podrías realizar la entrega?\n\nSaludos!",
		buyerName, productName, address)
}

// GetMessageTemplates returns a new instance of MessageTemplates
func GetMessageTemplates() *MessageTemplates {
	return &MessageTemplates{}
}