package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Database configuration
	Database DatabaseConfig

	// JWT configuration
	JWT JWTConfig

	// Server configuration
	Server ServerConfig

	// Google Cloud configuration
	GoogleCloud GoogleCloudConfig

	// WhatsApp configuration
	WhatsApp WhatsAppConfig

	// Environment
	Environment string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type JWTConfig struct {
	Secret           string
	ExpirationHours  time.Duration
	RefreshTokenTTL  time.Duration
	AccessTokenTTL   time.Duration
}

type ServerConfig struct {
	Port    string
	GinMode string
}

type GoogleCloudConfig struct {
	ProjectID         string
	CredentialsFile   string
	StorageBucket     string
	CloudSQLInstance  string
	UseEmulator       bool
	EmulatorHost      string
}

type WhatsAppConfig struct {
	APIUrl         string
	BusinessNumber string
	WebhookSecret  string
}

func Load() (*Config, error) {
	// Load environment variables from .env file
	_ = godotenv.Load()

	config := &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "agro_mas_dev"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "your-secret-key"),
			ExpirationHours: time.Duration(getEnvAsInt("JWT_EXPIRATION_HOURS", 24)) * time.Hour,
			RefreshTokenTTL: time.Duration(getEnvAsInt("JWT_REFRESH_TOKEN_TTL_DAYS", 7)) * 24 * time.Hour,
			AccessTokenTTL:  time.Duration(getEnvAsInt("JWT_ACCESS_TOKEN_TTL_MINUTES", 15)) * time.Minute,
		},
		Server: ServerConfig{
			Port:    getEnv("PORT", "8080"),
			GinMode: getEnv("GIN_MODE", "debug"),
		},
		GoogleCloud: GoogleCloudConfig{
			ProjectID:         getEnv("GOOGLE_CLOUD_PROJECT", ""),
			CredentialsFile:   getEnv("GOOGLE_APPLICATION_CREDENTIALS", ""),
			StorageBucket:     getEnv("GOOGLE_CLOUD_STORAGE_BUCKET", ""),
			CloudSQLInstance:  getEnv("CLOUD_SQL_INSTANCE", ""),
			UseEmulator:       getEnv("GOOGLE_CLOUD_USE_EMULATOR", "false") == "true",
			EmulatorHost:      getEnv("GOOGLE_CLOUD_STORAGE_EMULATOR_HOST", ""),
		},
		WhatsApp: WhatsAppConfig{
			APIUrl:         getEnv("WHATSAPP_API_URL", "https://api.whatsapp.com/send"),
			BusinessNumber: getEnv("WHATSAPP_BUSINESS_NUMBER", ""),
			WebhookSecret:  getEnv("WHATSAPP_WEBHOOK_SECRET", ""),
		},
		Environment: getEnv("ENVIRONMENT", "development"),
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(name string, defaultValue int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

func (c *Config) GetDatabaseURL() string {
	return "postgres://" + c.Database.User + ":" + c.Database.Password +
		"@" + c.Database.Host + ":" + c.Database.Port +
		"/" + c.Database.Name + "?sslmode=" + c.Database.SSLMode
}