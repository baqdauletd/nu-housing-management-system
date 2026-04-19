package routes

import (
	"database/sql"

	customAuth "nu-housing-management-system/backend/internal/auth"
	"nu-housing-management-system/backend/internal/database"
	"nu-housing-management-system/backend/internal/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	r *gin.Engine,
	db *sql.DB,
	minioStore *database.MinIOStore,
) {
	r.GET("/documents/:doc_id/download", handlers.DownloadDocument(db, minioStore))

	auth := r.Group("/auth")
	{
		auth.POST("/google", handlers.GoogleSignIn(db))
		auth.GET("/oauth/google", handlers.GoogleOAuthStartUnsupported())
	}

	application := r.Group("/applications")
	application.Use(customAuth.AuthMiddleware(), customAuth.RoleMiddleware("student"))
	{
		application.POST("/submit", handlers.SubmitApplication(db))
		application.GET("/my", handlers.GetMyApplications(db))
		application.PATCH("/:id", handlers.UpdateMyApplication(db))
		application.GET("/:id/status", handlers.GetApplicationStatus(db))
	}

	documents := r.Group("/documents")
	documents.Use(customAuth.AuthMiddleware(), customAuth.RoleMiddleware("student", "housing"))
	{
		documents.POST("/upload", handlers.UploadDocument(db, minioStore))
		documents.GET("/:doc_id", handlers.GetDocument(db))
		documents.GET("/application/:app_id", handlers.GetDocumentsByApplication(db, minioStore))
	}

	settings := r.Group("/settings")
	settings.Use(customAuth.AuthMiddleware(), customAuth.RoleMiddleware("student", "housing", "admin"))
	{
		settings.GET("", handlers.GetSystemSettings(db))
	}

	housing := r.Group("/housing")
	housing.Use(customAuth.AuthMiddleware(), customAuth.RoleMiddleware("housing"))
	{
		housing.GET("/applications", handlers.HousingListApplications(db))
		housing.GET("/applications/:id", handlers.HousingGetApplication(db))
		housing.PATCH("/applications/:id/approve", handlers.HousingApprove(db))
		housing.PATCH("/applications/:id/reject", handlers.HousingReject(db))
		housing.GET("/settings", handlers.GetSystemSettings(db))
		housing.PATCH("/settings", handlers.UpdateSystemSettings(db))
	}

	admin := r.Group("/admin")
	admin.Use(customAuth.AuthMiddleware(), customAuth.RoleMiddleware("admin"))
	{
		admin.GET("/users", handlers.AdminListUsers(db))
		admin.POST("/create-user", handlers.AdminCreateUser(db))
		admin.DELETE("/users/:id", handlers.AdminDeleteUser(db))
		admin.GET("/logs", handlers.AdminSystemLogs(db))
		admin.GET("/stats", handlers.AdminStats(db))
		admin.GET("/settings", handlers.GetSystemSettings(db))
		admin.PATCH("/settings", handlers.UpdateSystemSettings(db))
	}
}
