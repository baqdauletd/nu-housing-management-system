package main

import (
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"nu-housing-management-system/backend/internal/auth"
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

	auth.LoadJWTSecret()

	// PostgreSQL
	db, err := database.ConnectPostgres(cfg)
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}
	log.Println("Connected to PostgreSQL")

	// Redis
	// redisClient, err := database.ConnectRedis(cfg)
	// if err != nil {
	//     log.Println("Redis not running, continuing without it")
	// } else {
	//     log.Println("Connected to Redis")
	// }

	// MinIO
	minioClient, err := database.ConnectMinIO(cfg)
	if err != nil {
		log.Fatal("Failed to connect to MinIO:", err)
	}
	log.Println("Connected to MinIO")

	// Gin
	router := gin.Default()

	// This API uses bearer tokens, not browser cookies. Restrict CORS to known
	// frontend origins and avoid credentialed requests entirely.
	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.FrontendOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	// Public endpoints
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":   "OK",
			"database": db != nil,
			// "redis": redisClient != nil,
			// "storage": minioClient != nil,
		})
	})

	// register all routes (student, housing, admin)
	// routes.RegisterRoutes(r, db, redisClient, minioClient)
	routes.RegisterRoutes(router, db, minioClient, cfg)

	// start server
	log.Println("Server running on port", cfg.ServerPort)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
