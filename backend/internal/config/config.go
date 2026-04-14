package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	ServerPort string

	PostgresURL string

	RedisAddr string

	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string
	MinioUseSSL    bool
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		// Allow startup without a .env file when the process environment already
		// provides the required variables.
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config file: %w", err)
		}
	}

	cfg := &Config{
		ServerPort:  viper.GetString("PORT"),
		PostgresURL: viper.GetString("POSTGRES_URL"),
		// RedisAddr:      viper.GetString("REDIS_ADDR"),
		MinioEndpoint:  viper.GetString("MINIO_ENDPOINT"),
		MinioAccessKey: viper.GetString("MINIO_ACCESS_KEY"),
		MinioSecretKey: viper.GetString("MINIO_SECRET_KEY"),
		MinioBucket:    viper.GetString("MINIO_BUCKET"),
		MinioUseSSL:    viper.GetBool("MINIO_USE_SSL"),
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
