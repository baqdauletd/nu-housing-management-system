package database

import (
	"context"
    "github.com/redis/go-redis/v9"
    "nu-housing-management-system/backend/internal/config"
)

func ConnectRedis(cfg *config.Config) (*redis.Client, error) {
    client := redis.NewClient(&redis.Options{
        Addr: cfg.RedisAddr,
    })

    ctx := context.Background()
    _, err := client.Ping(ctx).Result()
    if err != nil {
        return nil, err
    }
    return client, err
}
