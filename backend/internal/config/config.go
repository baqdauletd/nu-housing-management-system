package config

import (
	"fmt"

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
        fmt.Println("debug: Error in env reading")
        return nil, err
    }

    cfg := &Config{
        ServerPort:     viper.GetString("PORT"),
        PostgresURL:    viper.GetString("POSTGRES_URL"),
        // RedisAddr:      viper.GetString("REDIS_ADDR"),
        MinioEndpoint:  viper.GetString("MINIO_ENDPOINT"),
        MinioAccessKey: viper.GetString("MINIO_ACCESS_KEY"),
        MinioSecretKey: viper.GetString("MINIO_SECRET_KEY"),
        MinioBucket:    viper.GetString("MINIO_BUCKET"),
        MinioUseSSL:    viper.GetBool("MINIO_USE_SSL"),
    }

    return cfg, nil
}
