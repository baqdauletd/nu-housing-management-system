package routes

import (
    "github.com/gin-gonic/gin"
    "github.com/minio/minio-go/v7"
    "database/sql"
    "nu-housing-management-system/backend/internal/handlers"
    customAuth "nu-housing-management-system/backend/internal/auth"
)

func RegisterRoutes(
    r *gin.Engine,
    db *sql.DB,
    minioClient *minio.Client,
) {
    // --- AUTH ROUTES ---
    auth := r.Group("/auth")
    {
        auth.POST("/register", handlers.Register(db))
        auth.POST("/login", handlers.Login(db))
    }

    // --- APPLICATION ROUTES (student only) ---
    application := r.Group("/applications")
    application.Use(customAuth.AuthMiddleware(), customAuth.RoleMiddleware("student"))
    {
        application.POST("/submit", handlers.SubmitApplication(db))
        application.GET("/my", handlers.GetMyApplications(db))
        application.GET("/:id/status", handlers.GetApplicationStatus(db))
    }

    // --- DOCUMENT ROUTES (student + housing - shared middleware) ---
    documents := r.Group("/documents")
    documents.Use(customAuth.AuthMiddleware(), customAuth.RoleMiddleware("student", "housing"))
    {
        documents.POST("/upload", handlers.UploadDocument(db, minioClient))
        documents.GET("/:doc_id", handlers.GetDocument(db))
        documents.GET("/application/:app_id", handlers.GetDocumentsByApplication(db, minioClient))
    }

    // --- HOUSING STAFF ROUTES ---
    housing := r.Group("/housing")
    housing.Use(customAuth.AuthMiddleware(), customAuth.RoleMiddleware("housing"))
    {
        housing.GET("/applications", handlers.HousingListApplications(db))
        housing.GET("/applications/:id", handlers.HousingGetApplication(db))
        housing.PATCH("/applications/:id/approve", handlers.HousingApprove(db))
        housing.PATCH("/applications/:id/reject", handlers.HousingReject(db))
    }

    // --- ADMIN ROUTES ---
    admin := r.Group("/admin")
    admin.Use(customAuth.AuthMiddleware(), customAuth.RoleMiddleware("admin"))
    {
        admin.GET("/users", handlers.AdminListUsers(db))
        admin.POST("/create-user", handlers.AdminCreateUser(db))
        admin.DELETE("/users/:id", handlers.AdminDeleteUser(db))
        admin.GET("/logs", handlers.AdminSystemLogs(db))
        admin.GET("/stats", handlers.AdminStats(db))
    }
}