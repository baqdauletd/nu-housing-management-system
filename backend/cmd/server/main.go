package main

import (
    "log"

    "github.com/gin-gonic/gin"
    "nu-housing-management-system/backend/internal/config"
    "nu-housing-management-system/backend/internal/database"
    "nu-housing-management-system/backend/internal/routes"
)

func main() {
    // load environment variables / configuration
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatal("Failed to load configuration:", err)
    }

    // PostgreSQL
    db, err := database.ConnectPostgres(cfg)
    if err != nil {
        log.Fatal("Failed to connect to PostgreSQL:", err)
    }
    log.Println("Connected to PostgreSQL")

    // Redis
    redisClient, err := database.ConnectRedis(cfg)
    if err != nil {
        log.Println("Redis not running, continuing without it")
    } else {
        log.Println("Connected to Redis")
    }

    //MinIO
    minioClient, err := database.ConnectMinIO(cfg)
    if err != nil {
        log.Fatal("Failed to connect to MinIO:", err)
    }
    log.Println("Connected to MinIO")

    // Gin
    r := gin.Default()

    // Public endpoints
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "status": "OK",
            "database": db != nil,
            "redis": redisClient != nil,
            "storage": minioClient != nil,
        })
    })

    // register all routes (student, housing, admin)
    routes.RegisterRoutes(r, db, redisClient, minioClient)

    // start server
    log.Println("Server running on port", cfg.ServerPort)
    if err := r.Run(":" + cfg.ServerPort); err != nil {
        log.Fatal("Failed to start server:", err)
    }
}
