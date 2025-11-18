package routes

import (
    "github.com/gin-gonic/gin"
    // "github.com/redis/go-redis/v9"
    // "github.com/minio/minio-go/v7"
    // "gorm.io/gorm"
    "database/sql"

    "nu-housing-management-system/backend/internal/handlers"
)

func RegisterRoutes(
    r *gin.Engine,
    db *sql.DB,
    // redis *redis.Client,
    // minioClient *minio.Client,
) {
    // --- AUTH ROUTES ---
    auth := r.Group("/auth")
    {
        auth.POST("/register", handlers.Register(db))
        auth.POST("/login", handlers.Login(db))
    }

    // --- STUDENT ROUTES ---
    student := r.Group("/student")
    {
        student.GET("/me", handlers.GetProfile(db))
        student.PUT("/update", handlers.UpdateProfile(db))
    }

    // --- APPLICATION ROUTES ---
    application := r.Group("/applications")
    {
        application.POST("/submit", handlers.SubmitApplication(db)) //(db, minioClient)
        application.GET("/my", handlers.GetMyApplications(db))
        application.GET("/:id/status", handlers.GetApplicationStatus(db))
    }

    // --- DOCUMENT ROUTES ---
    documents := r.Group("/documents")
    {
        documents.POST("/upload", handlers.UploadDocument(db)) //(db, minioClient)
        documents.GET("/:doc_id", handlers.GetDocument(db))
    }

    // --- REVIEW ENGINE ROUTES ---
    review := r.Group("/review")
    {
        review.POST("/auto/:application_id", handlers.TriggerAutoReview(db)) //(db, redis)
        review.GET("/result/:application_id", handlers.GetAutoReviewResult(db)) //(db, redis)
    }

    // --- HOUSING STAFF ROUTES ---
    housing := r.Group("/housing")
    {
        housing.GET("/applications", handlers.HousingListApplications(db))
        housing.GET("/applications/:id", handlers.HousingGetApplication(db))
        housing.POST("/applications/:id/approve", handlers.HousingApprove(db))
        housing.POST("/applications/:id/reject", handlers.HousingReject(db))
    }

    // --- ADMIN ROUTES ---
    admin := r.Group("/admin")
    {
        admin.GET("/users", handlers.AdminListUsers(db))
        admin.POST("/create-user", handlers.AdminCreateUser(db))
        admin.DELETE("/users/:id", handlers.AdminDeleteUser(db))

        admin.GET("/logs", handlers.AdminSystemLogs(db))
        admin.GET("/stats", handlers.AdminStats(db))
    }
}
