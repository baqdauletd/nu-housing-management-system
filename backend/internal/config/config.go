package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	ServerPort string

	PostgresURL string
	FrontendOrigins []string

	RedisAddr string

	MinioEndpoint       string
	MinioPublicEndpoint string
	MinioAccessKey      string
	MinioSecretKey      string
	MinioBucket         string
	MinioUseSSL         bool
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		// Allow startup without a .env file when the process environment already
		// provides the required variables.
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok && !os.IsNotExist(err) {
			return nil, fmt.Errorf("read config file: %w", err)
		}
	}

	cfg := &Config{
		ServerPort:         viper.GetString("PORT"),
		PostgresURL:        viper.GetString("POSTGRES_URL"),
		FrontendOrigins:    loadFrontendOrigins(),
		// RedisAddr:      viper.GetString("REDIS_ADDR"),
		MinioEndpoint:       viper.GetString("MINIO_ENDPOINT"),
		MinioPublicEndpoint: viper.GetString("MINIO_PUBLIC_ENDPOINT"),
		MinioAccessKey:      viper.GetString("MINIO_ACCESS_KEY"),
		MinioSecretKey:      viper.GetString("MINIO_SECRET_KEY"),
		MinioBucket:         viper.GetString("MINIO_BUCKET"),
		MinioUseSSL:         viper.GetBool("MINIO_USE_SSL"),
	}

	var missing []string
	if cfg.ServerPort == "" {
		missing = append(missing, "PORT")
	}
	if cfg.PostgresURL == "" {
		missing = append(missing, "POSTGRES_URL")
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
