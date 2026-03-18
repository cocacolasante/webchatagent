package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// App
	Port    string
	BaseURL string
	Env     string
	LogLevel string

	// Admin Auth
	AdminKey string

	// Database
	DatabaseURL string

	// Redis
	RedisURL            string
	RedisSessionTTL     time.Duration
	RedisConfigCacheTTL time.Duration

	// Anthropic
	AnthropicAPIKey   string
	AnthropicModel    string
	AnthropicMaxTokens int

	// Encryption
	EncryptionKey string

	// Scheduler
	DefaultScheduler string
	CalComAPIBase    string
	CalendlyAPIBase  string

	// Email
	SMTPHost          string
	SMTPPort          int
	SMTPUser          string
	SMTPPass          string
	NotificationFromEmail string
	NotificationFromName  string

	// Discord
	DiscordWebhookURL string

	// n8n
	N8NBaseURL string
	N8NAPIKey  string

	// Rate Limits
	RateLimitPerMin int
	RateLimitPerDay int

	// Widget
	WidgetBundlePath string
	WidgetMaxSizeKB  int

	// Blueprint Command billing
	PortalsURL     string
	BPAInternalKey string
}

func Load() (*Config, error) {
	// Load .env if present (ignore error — .env is optional in production)
	_ = godotenv.Load()

	cfg := &Config{
		Port:                  getEnv("PORT", "8080"),
		BaseURL:               getEnv("BASE_URL", "http://localhost:8080"),
		Env:                   getEnv("ENV", "development"),
		LogLevel:              getEnv("LOG_LEVEL", "info"),
		DatabaseURL:           getEnv("DATABASE_URL", ""),
		RedisURL:              getEnv("REDIS_URL", "redis://localhost:6379"),
		AnthropicModel:        getEnv("ANTHROPIC_MODEL", "claude-sonnet-4-20250514"),
		DefaultScheduler:      getEnv("DEFAULT_SCHEDULER", "calcom"),
		CalComAPIBase:         getEnv("CALCOM_API_BASE", "https://api.cal.com/v1"),
		CalendlyAPIBase:       getEnv("CALENDLY_API_BASE", "https://api.calendly.com"),
		SMTPHost:              getEnv("SMTP_HOST", ""),
		SMTPPort:              getEnvInt("SMTP_PORT", 587),
		SMTPUser:              getEnv("SMTP_USER", ""),
		SMTPPass:              getEnv("SMTP_PASS", ""),
		NotificationFromEmail: getEnv("NOTIFICATION_FROM_EMAIL", "chat@blueprintautomation.tech"),
		NotificationFromName:  getEnv("NOTIFICATION_FROM_NAME", "Blueprint Chat"),
		DiscordWebhookURL:     getEnv("DISCORD_WEBHOOK_URL", ""),
		N8NBaseURL:            getEnv("N8N_BASE_URL", "http://localhost:5678"),
		N8NAPIKey:             getEnv("N8N_API_KEY", ""),
		RateLimitPerMin:       getEnvInt("RATE_LIMIT_MESSAGES_PER_MIN", 30),
		RateLimitPerDay:       getEnvInt("RATE_LIMIT_MESSAGES_PER_DAY", 1000),
		WidgetBundlePath:      getEnv("WIDGET_BUNDLE_PATH", "./widget/dist/widget.js"),
		WidgetMaxSizeKB:       getEnvInt("WIDGET_MAX_SIZE_KB", 50),
		AnthropicMaxTokens:    getEnvInt("ANTHROPIC_MAX_TOKENS", 1024),
		PortalsURL:            getEnv("PORTALS_URL", ""),
		BPAInternalKey:        getEnv("BPA_INTERNAL_KEY", ""),
	}

	// Required fields — fail fast
	required := map[string]string{
		"ANTHROPIC_API_KEY":    getEnv("ANTHROPIC_API_KEY", ""),
		"BLUEPRINT_ADMIN_KEY": getEnv("BLUEPRINT_ADMIN_KEY", ""),
		"DATABASE_URL":        cfg.DatabaseURL,
		"ENCRYPTION_KEY":      getEnv("ENCRYPTION_KEY", ""),
	}

	for key, val := range required {
		if val == "" {
			return nil, fmt.Errorf("required environment variable %s is not set", key)
		}
	}

	cfg.AnthropicAPIKey = required["ANTHROPIC_API_KEY"]
	cfg.AdminKey = required["BLUEPRINT_ADMIN_KEY"]
	cfg.EncryptionKey = required["ENCRYPTION_KEY"]

	sessionTTL := getEnvInt("REDIS_SESSION_TTL_MINUTES", 30)
	cfg.RedisSessionTTL = time.Duration(sessionTTL) * time.Minute

	configCacheTTL := getEnvInt("REDIS_CONFIG_CACHE_TTL_MINUTES", 5)
	cfg.RedisConfigCacheTTL = time.Duration(configCacheTTL) * time.Minute

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultVal
}
