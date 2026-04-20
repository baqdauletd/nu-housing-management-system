package routes

import (
	"database/sql"

	customAuth "nu-housing-management-system/backend/internal/auth"
	"nu-housing-management-system/backend/internal/config"
	"nu-housing-management-system/backend/internal/database"
	"nu-housing-management-system/backend/internal/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	r *gin.Engine,
	db *sql.DB,
	minioStore *database.MinIOStore,
	cfg *config.Config,
) {
	r.GET("/documents/:doc_id/download", handlers.DownloadDocument(db, minioStore))
	r.POST("/payments/stripe/webhook", handlers.HandleStripeWebhook(db, cfg))

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
		application.PATCH("/:id", handlers.UpdateMyApplication(db, minioStore))
		application.GET("/:id/status", handlers.GetApplicationStatus(db))
	}

	payments := r.Group("/payments")
	payments.Use(customAuth.AuthMiddleware(), customAuth.RoleMiddleware("student"))
	{
		payments.GET("/application/:app_id", handlers.GetApplicationPaymentSummary(db, cfg))
		payments.POST("/application/:app_id/initiate", handlers.InitiateApplicationPayment(db, cfg))
		payments.POST("/application/:app_id/sync", handlers.SyncStripePayment(db, cfg))
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
		housing.GET("/dorm-inventory", handlers.HousingDormInventory(db))
		housing.PATCH("/applications/:id/approve", handlers.HousingApprove(db))
		housing.PATCH("/applications/:id/reject", handlers.HousingReject(db))
		housing.POST("/notify-rejected", handlers.HousingNotifyRejected(db, cfg))
		housing.GET("/settings", handlers.GetSystemSettings(db))
		housing.PATCH("/settings", handlers.UpdateSystemSettings(db, cfg))
	}

	admin := r.Group("/admin")
	admin.Use(customAuth.AuthMiddleware(), customAuth.RoleMiddleware("admin"))
	{
		admin.GET("/users", handlers.AdminListUsers(db))
		admin.POST("/create-user", handlers.AdminCreateUser(db))
		admin.PATCH("/users/:id/role", handlers.AdminUpdateUserRole(db))
		admin.DELETE("/users/:id", handlers.AdminDeleteUser(db))
		admin.GET("/logs", handlers.AdminSystemLogs(db))
		admin.GET("/stats", handlers.AdminStats(db))
		admin.GET("/settings", handlers.GetSystemSettings(db))
		admin.PATCH("/settings", handlers.UpdateSystemSettings(db, cfg))
	}
}
