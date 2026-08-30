package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Redis    RedisConfig
	Storage  StorageConfig
	SMTP     SMTPConfig
}

type AppConfig struct {
	Env     string
	Version string
	Port    string
	URL     string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
	DSN      string
}

type AuthConfig struct {
	JWTSecret        string
	AccessExpMinutes int
	RefreshExpDays   int
}

type RedisConfig struct {
	URL string
}

type StorageConfig struct {
	Endpoint        string
	AccessKey       string
	SecretKey       string
	BucketDocuments string
	UseSSL          bool
	// LocalPath is the local development storage root for document files
	// (P1-005). In production this should be replaced by MinIO/S3-compatible
	// object storage (Endpoint/AccessKey/SecretKey/BucketDocuments).
	LocalPath string
	// MaxSizeBytes caps an uploaded document file size (default 20 MB).
	MaxSizeBytes int64
}

type SMTPConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	From     string
}

// Load reads environment variables and returns a validated Config.
// It will load .env file if present, but env vars already set take precedence.
func Load() (*Config, error) {
	// Load .env file if present — errors are intentionally ignored
	// (file may not exist in production where env vars are injected directly)
	_ = godotenv.Load()

	cfg := &Config{}

	// App
	cfg.App.Env = getEnv("APP_ENV", "development")
	cfg.App.Version = getEnv("APP_VERSION", "0.1.0")
	cfg.App.Port = getEnv("PORT", "8080")
	cfg.App.URL = getEnv("APP_URL", "http://localhost:8080")

	// Database
	cfg.Database.Host = requireEnv("DB_HOST")
	cfg.Database.Port = getEnv("DB_PORT", "5432")
	cfg.Database.User = requireEnv("DB_USER")
	cfg.Database.Password = requireEnv("DB_PASSWORD")
	cfg.Database.Name = requireEnv("DB_NAME")
	cfg.Database.SSLMode = getEnv("DB_SSLMODE", "disable")
	cfg.Database.DSN = fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Jakarta",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	// Auth
	cfg.Auth.JWTSecret = requireEnv("JWT_SECRET")
	if len(cfg.Auth.JWTSecret) < 32 {
		return nil, fmt.Errorf("config: JWT_SECRET must be at least 32 characters")
	}
	cfg.Auth.AccessExpMinutes = getEnvInt("JWT_ACCESS_EXP_MINUTES", 15)
	cfg.Auth.RefreshExpDays = getEnvInt("JWT_REFRESH_EXP_DAYS", 7)

	// Redis
	cfg.Redis.URL = getEnv("REDIS_URL", "redis://localhost:6379/0")

	// Storage
	cfg.Storage.Endpoint = getEnv("STORAGE_ENDPOINT", "localhost:9000")
	cfg.Storage.AccessKey = getEnv("STORAGE_ACCESS_KEY", "")
	cfg.Storage.SecretKey = getEnv("STORAGE_SECRET_KEY", "")
	cfg.Storage.BucketDocuments = getEnv("STORAGE_BUCKET_DOCUMENTS", "cankora-documents")
	cfg.Storage.UseSSL = getEnvBool("STORAGE_USE_SSL", false)
	cfg.Storage.LocalPath = getEnv("STORAGE_LOCAL_PATH", "storage/documents")
	cfg.Storage.MaxSizeBytes = getEnvInt64("STORAGE_MAX_SIZE_BYTES", 20*1024*1024)

	// SMTP
	cfg.SMTP.Host = getEnv("SMTP_HOST", "")
	cfg.SMTP.Port = getEnv("SMTP_PORT", "587")
	cfg.SMTP.User = getEnv("SMTP_USER", "")
	cfg.SMTP.Password = getEnv("SMTP_PASSWORD", "")
	cfg.SMTP.From = getEnv("SMTP_FROM", "PMO <noreply@cankora.local>")

	return cfg, nil
}

// IsDevelopment returns true if APP_ENV is "development".
func (c *Config) IsDevelopment() bool {
	return c.App.Env == "development"
}

// IsProduction returns true if APP_ENV is "production".
func (c *Config) IsProduction() bool {
	return c.App.Env == "production"
}

// requireEnv returns the value of an env var or panics with a clear message.
func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf(
			"FATAL: Required environment variable %q is not set.\n"+
				"Please check your .env file or environment configuration.",
			key,
		))
	}
	return v
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvInt64(key string, fallback int64) int64 {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}
