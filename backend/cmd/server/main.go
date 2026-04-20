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
	if err := database.EnsureRuntimeSchema(db); err != nil {
		log.Fatal("Failed to ensure database schema:", err)
	}
	log.Println("Connected to PostgreSQL")

	// MinIO
	minioClient, err := database.ConnectMinIO(cfg)
	if err != nil {
		log.Fatal("Failed to connect to MinIO:", err)
	}
	log.Println("Connected to MinIO")

	// Gin
	router := gin.Default()
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
			"storage": minioClient != nil,
		})
	})

	routes.RegisterRoutes(router, db, minioClient, cfg)

	// start server
	log.Println("Server running on port", cfg.ServerPort)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
