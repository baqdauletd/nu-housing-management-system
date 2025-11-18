package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	// "github.com/minio/minio-go/v7"
	// "github.com/redis/go-redis/v9"
)

//////////////////////////////////////////////////////////
// AUTH HANDLERS
//////////////////////////////////////////////////////////

func Register(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Parse JSON, hash password, insert into DB
		c.JSON(http.StatusOK, gin.H{"message": "register endpoint"})
	}
}

func Login(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Authenticate user, generate JWT
		c.JSON(http.StatusOK, gin.H{"message": "login endpoint"})
	}
}

//////////////////////////////////////////////////////////
// STUDENT PROFILE HANDLERS
//////////////////////////////////////////////////////////

func GetProfile(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Fetch student data using student_id from JWT
		c.JSON(http.StatusOK, gin.H{"message": "get profile"})
	}
}

func UpdateProfile(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Update student fields in DB
		c.JSON(http.StatusOK, gin.H{"message": "update profile"})
	}
}

//////////////////////////////////////////////////////////
// APPLICATION HANDLERS
//////////////////////////////////////////////////////////

func SubmitApplication(db *sql.DB) gin.HandlerFunc { //(db *sql.DB, minio *minio.Client)
	return func(c *gin.Context) {
		// TODO:
		// 1. Parse form data
		// 2. Upload docs to MinIO
		// 3. Insert application record
		c.JSON(http.StatusOK, gin.H{"message": "application submitted"})
	}
}

func GetMyApplications(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Query applications WHERE student_id = ?
		c.JSON(http.StatusOK, gin.H{"message": "my applications list"})
	}
}

func GetApplicationStatus(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: SELECT status FROM applications WHERE id = ?
		c.JSON(http.StatusOK, gin.H{"message": "application status"})
	}
}

//////////////////////////////////////////////////////////
// DOCUMENT HANDLERS
//////////////////////////////////////////////////////////

func UploadDocument(db *sql.DB) gin.HandlerFunc { //(db *sql.DB, minio *minio.Client)
	return func(c *gin.Context) {
		// TODO:
		// 1. Accept file
		// 2. Upload to MinIO
		// 3. Insert record into documents table
		c.JSON(http.StatusOK, gin.H{"message": "document uploaded"})
	}
}

func GetDocument(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Get document info → generate presigned MinIO URL
		c.JSON(http.StatusOK, gin.H{"message": "document download"})
	}
}

//////////////////////////////////////////////////////////
// REVIEW ENGINE HANDLERS
//////////////////////////////////////////////////////////

func TriggerAutoReview(db *sql.DB) gin.HandlerFunc { //(db *sql.DB, redis *redis.Client)
	return func(c *gin.Context) {
		// TODO:
		// 1. Push job to Redis queue
		// 2. Worker processes automatic review
		c.JSON(http.StatusOK, gin.H{"message": "auto review triggered"})
	}
}

func GetAutoReviewResult(db *sql.DB) gin.HandlerFunc { //(db *sql.DB, redis *redis.Client)
	return func(c *gin.Context) {
		// TODO:
		// 1. Check Redis for result
		// 2. Return status (accepted, rejected, needs manual review)
		c.JSON(http.StatusOK, gin.H{"message": "auto review result"})
	}
}

//////////////////////////////////////////////////////////
// HOUSING STAFF HANDLERS
//////////////////////////////////////////////////////////

func HousingListApplications(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: SELECT * FROM applications (with filters)
		c.JSON(http.StatusOK, gin.H{"message": "housing list applications"})
	}
}

func HousingGetApplication(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: SELECT * FROM applications WHERE id = ?
		c.JSON(http.StatusOK, gin.H{"message": "housing application details"})
	}
}

func HousingApprove(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: UPDATE applications SET status='approved'
		c.JSON(http.StatusOK, gin.H{"message": "application approved"})
	}
}

func HousingReject(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO:
		// 1. Add rejection reason
		// 2. Update status = rejected
		c.JSON(http.StatusOK, gin.H{"message": "application rejected"})
	}
}

//////////////////////////////////////////////////////////
// ADMIN HANDLERS
//////////////////////////////////////////////////////////

func AdminListUsers(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: SELECT * FROM users
		c.JSON(http.StatusOK, gin.H{"message": "admin user list"})
	}
}

func AdminCreateUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: INSERT INTO users (...)
		c.JSON(http.StatusOK, gin.H{"message": "admin user created"})
	}
}

func AdminDeleteUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: DELETE FROM users WHERE id = ?
		c.JSON(http.StatusOK, gin.H{"message": "admin user deleted"})
	}
}

func AdminSystemLogs(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: SELECT * FROM logs ORDER BY timestamp DESC
		c.JSON(http.StatusOK, gin.H{"message": "system logs"})
	}
}

func AdminStats(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Query counts/statistics for dashboard
		c.JSON(http.StatusOK, gin.H{"message": "system stats"})
	}
}
