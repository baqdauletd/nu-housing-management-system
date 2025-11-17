package routes

import (
    "database/sql"

    "github.com/gin-gonic/gin"
    "github.com/minio/minio-go/v7"
    "github.com/redis/go-redis/v9"
)

func RegisterRoutes(
    r *gin.Engine,
    db *sql.DB,
    redis *redis.Client,
    minio *minio.Client,
) {
    // Auth routes
    // AuthRoutes(r, db)

    // Student routes
    // StudentRoutes(r.Group("/student"), db, minio)

    // Housing staff routes
    // HousingRoutes(r.Group("/housing"), db)

    // Admin routes
    // AdminRoutes(r.Group("/admin"), db)
}
