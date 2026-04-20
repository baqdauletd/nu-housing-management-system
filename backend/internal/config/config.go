package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	ServerPort string

	PostgresURL     string
	FrontendOrigins []string

	RedisAddr string

	MinioEndpoint       string
	MinioPublicEndpoint string
	MinioAccessKey      string
	MinioSecretKey      string
	MinioBucket         string
	MinioUseSSL         bool

	FrontendBaseURL          string
	StripeSecretKey          string
	StripeWebhookSecret      string
	StripeMerchantName       string
	StripeProductName        string
	StripePaymentCurrency    string
	StripePaymentAmountKZT   int
	StripePaymentDescription string

	SMTPHost              string
	SMTPPort              int
	SMTPUsername          string
	SMTPPassword          string
	SMTPFrom              string
	SMTPAllowedRecipients map[string]struct{}
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		// Allow startup without a .env file when the process environment already
		// provides the required variables.
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok && !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("read config file: %w", err)
		}
	}

	cfg := &Config{
		ServerPort:      viper.GetString("PORT"),
		PostgresURL:     viper.GetString("POSTGRES_URL"),
		FrontendOrigins: loadFrontendOrigins(),
		FrontendBaseURL: viper.GetString("FRONTEND_BASE_URL"),
		// RedisAddr:      viper.GetString("REDIS_ADDR"),
		MinioEndpoint:            firstNonEmpty(viper.GetString("MINIO_ENDPOINT"), viper.GetString("MINIO_PRIVATE_ENDPOINT")),
		MinioPublicEndpoint:      viper.GetString("MINIO_PUBLIC_ENDPOINT"),
		MinioAccessKey:           firstNonEmpty(viper.GetString("MINIO_ACCESS_KEY"), viper.GetString("MINIO_ROOT_USER")),
		MinioSecretKey:           firstNonEmpty(viper.GetString("MINIO_SECRET_KEY"), viper.GetString("MINIO_ROOT_PASSWORD")),
		MinioBucket:              viper.GetString("MINIO_BUCKET"),
		MinioUseSSL:              viper.GetBool("MINIO_USE_SSL"),
		StripeSecretKey:          viper.GetString("STRIPE_SECRET_KEY"),
		StripeWebhookSecret:      viper.GetString("STRIPE_WEBHOOK_SECRET"),
		StripeMerchantName:       viper.GetString("STRIPE_MERCHANT_NAME"),
		StripeProductName:        viper.GetString("STRIPE_PRODUCT_NAME"),
		StripePaymentCurrency:    normalizeStripeCurrency(viper.GetString("STRIPE_PAYMENT_CURRENCY")),
		StripePaymentAmountKZT:   viper.GetInt("STRIPE_PAYMENT_AMOUNT_KZT"),
		StripePaymentDescription: viper.GetString("STRIPE_PAYMENT_DESCRIPTION"),
		SMTPHost:                 viper.GetString("SMTP_HOST"),
		SMTPPort:                 viper.GetInt("SMTP_PORT"),
		SMTPUsername:             viper.GetString("SMTP_USERNAME"),
		SMTPPassword:             viper.GetString("SMTP_PASSWORD"),
		SMTPFrom:                 viper.GetString("SMTP_FROM"),
		SMTPAllowedRecipients:    loadAllowedEmails("SMTP_ALLOWED_RECIPIENTS"),
	}

	var missing []string
	if cfg.ServerPort == "" {
		missing = append(missing, "PORT")
	}
	if cfg.PostgresURL == "" {
		missing = append(missing, "POSTGRES_URL")
	}
	if cfg.FrontendBaseURL == "" {
		missing = append(missing, "FRONTEND_BASE_URL")
	}
	if cfg.MinioEndpoint == "" {
		missing = append(missing, "MINIO_ENDPOINT")
	}
	if cfg.MinioAccessKey == "" {
		missing = append(missing, "MINIO_ACCESS_KEY")
	}
	if cfg.MinioSecretKey == "" {
		missing = append(missing, "MINIO_SECRET_KEY")
	}
	if cfg.MinioBucket == "" {
		missing = append(missing, "MINIO_BUCKET")
	}
	if viper.GetString("JWT_SECRET") == "" {
		missing = append(missing, "JWT_SECRET")
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required configuration: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}

func loadFrontendOrigins() []string {
	raw := strings.TrimSpace(viper.GetString("FRONTEND_ORIGINS"))
	if raw == "" {
		return []string{
			"http://localhost:3000",
			"http://127.0.0.1:3000",
		}
	}

	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		origin := strings.TrimSpace(part)
		if origin == "" {
			continue
		}
		origins = append(origins, origin)
	}

	if len(origins) == 0 {
		return []string{
			"http://localhost:3000",
			"http://127.0.0.1:3000",
		}
	}

	return origins
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func normalizeStripeCurrency(value string) string {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	if trimmed == "" {
		return "usd"
	}
	return trimmed
}

func loadAllowedEmails(key string) map[string]struct{} {
	raw := strings.TrimSpace(viper.GetString(key))
	allowed := make(map[string]struct{})
	if raw == "" {
		return allowed
	}

	for _, part := range strings.Split(raw, ",") {
		email := strings.ToLower(strings.TrimSpace(part))
		if email == "" {
			continue
		}
		allowed[email] = struct{}{}
	}

	return allowed
}

func (c *Config) IsEmailRecipientAllowed(email string) bool {
	if c == nil {
		return false
	}

	normalized := strings.ToLower(strings.TrimSpace(email))
	if normalized == "" {
		return false
	}
	if len(c.SMTPAllowedRecipients) == 0 {
		return true
	}

	_, ok := c.SMTPAllowedRecipients[normalized]
	return ok
}
