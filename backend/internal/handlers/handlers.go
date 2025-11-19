package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	// "github.com/minio/minio-go/v7"
	// "github.com/redis/go-redis/v9"
	"nu-housing-management-system/backend/internal/auth"
	"nu-housing-management-system/backend/internal/database"
	"nu-housing-management-system/backend/internal/models"
	"strings"
)

//////////////////////////////////////////////////////////
// AUTH HANDLERS
//////////////////////////////////////////////////////////

// Register allows students and housing staff to create accounts.
// Admin should NOT be allowed to self-register.
func Register(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user struct {
			NuID     string `json:"nu_id" binding:"required"`
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required,min=6"`
			Phone    string `json:"phone"`
			Role     string `json:"role"` // optional: "student" or "housing"
		}
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input", "details": err.Error()})
			return
		}

		role := strings.ToLower(user.Role)
		if role == "" {
			role = "student"
		}
		if role != "student" && role != "housing" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role: only 'student' or 'housing' allowed"})
			return
		}

		roleID, err := database.GetRoleIDByName(db, role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "role lookup failed"})
			return
		}

		// check uniqueness of email or nu_id
		if _, err := database.GetUserByEmail(db, user.Email); err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "email already used"})
			return
		}

		hashed, err := auth.HashPassword(user.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}

		u := models.User{
			NuID:         user.NuID,
			Email:        user.Email,
			PasswordHash: hashed,
			RoleID:       roleID,
			Phone:        user.Phone,
		}
		id, err := database.CreateUser(db, u)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user", "details": err.Error()})
			return
		}

		// return created user id
		c.JSON(http.StatusCreated, gin.H{"id": id})
	}
}

func Login(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
			return
		}

		user, err := database.ValidateUserCredentials(db, body.Email)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		if err := auth.CheckPassword(user.PasswordHash, body.Password); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		// fetch role name
		roleName, _ := database.GetRoleNameByID(db, user.RoleID)

		token, err := auth.GenerateToken(user.ID, roleName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token": token,
			"user": gin.H{
				"id":     user.ID,
				"nu_id":  user.NuID,
				"email":  user.Email,
				"role":   roleName,
				"phone":  user.Phone,
			},
		})
	}
}


//////////////////////////////////////////////////////////
// STUDENT PROFILE HANDLERS
//////////////////////////////////////////////////////////

// func GetProfile(db *sql.DB) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		// TODO: Fetch student data using student_id from JWT
// 		c.JSON(http.StatusOK, gin.H{"message": "get profile"})
// 	}
// }

// func UpdateProfile(db *sql.DB) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		// TODO: Update student fields in DB
// 		c.JSON(http.StatusOK, gin.H{"message": "update profile"})
// 	}
// }

//////////////////////////////////////////////////////////
// APPLICATION HANDLERS
//////////////////////////////////////////////////////////

// SubmitApplication expects authenticated student (user_id in context)
func SubmitApplication(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body struct {
			Year           int    `json:"year" binding:"required"`
			Major          string `json:"major" binding:"required"`
			Gender         string `json:"gender" binding:"required"`
			RoomPreference string `json:"room_preference"`
			AdditionalInfo string `json:"additional_info"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input", "details": err.Error()})
			return
		}

		uid, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id missing in token"})
			return
		}
		studentID := uid.(int)

		app := models.Application{
			StudentID:      studentID,
			Year:           body.Year,
			Major:          body.Major,
			Gender:         body.Gender,
			RoomPreference: body.RoomPreference,
			AdditionalInfo: body.AdditionalInfo,
		}

		id, err := database.SubmitApplication(db, app)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to submit application", "details": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"application_id": id})
	}
}

func GetMyApplications(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id missing in token"})
			return
		}
		studentID := uid.(int)

		apps, err := database.GetApplicationsByStudent(db, studentID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch applications", "details": err.Error()})
			return
		}
		c.JSON(http.StatusOK, apps)
	}
}

func GetApplicationStatus(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, _ := strconv.Atoi(idStr)
		app, err := database.GetApplicationByID(db, id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": app.Status, "rejected_reason": app.RejectedReason})
	}
}

//////////////////////////////////////////////////////////
// DOCUMENT HANDLERS
//////////////////////////////////////////////////////////

// UploadDocument: for now, expects JSON with application_id, type, file_url
// Later to replace with multipart upload + MinIO.
func UploadDocument(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body struct {
			ApplicationID int    `json:"application_id" binding:"required"`
			Type          string `json:"type" binding:"required"`
			FileURL       string `json:"file_url" binding:"required"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
			return
		}

		doc := models.Document{
			ApplicationID: body.ApplicationID,
			Type:          body.Type,
			FileURL:       body.FileURL,
		}
		id, err := database.InsertDocument(db, doc)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save document", "details": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"document_id": id})
	}
}

func GetDocument(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("doc_id")
		id, _ := strconv.Atoi(idStr)
		doc, err := database.GetDocument(db, id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
			return
		}
		c.JSON(http.StatusOK, doc)
	}
}

//////////////////////////////////////////////////////////
// REVIEW ENGINE HANDLERS
//////////////////////////////////////////////////////////

// func TriggerAutoReview(db *sql.DB) gin.HandlerFunc { //(db *sql.DB, redis *redis.Client)
// 	return func(c *gin.Context) {
// 		// TODO:
// 		// 1. Push job to Redis queue
// 		// 2. Worker processes automatic review
// 		c.JSON(http.StatusOK, gin.H{"message": "auto review triggered"})
// 	}
// }

// func GetAutoReviewResult(db *sql.DB) gin.HandlerFunc { //(db *sql.DB, redis *redis.Client)
// 	return func(c *gin.Context) {
// 		// TODO:
// 		// 1. Check Redis for result
// 		// 2. Return status (accepted, rejected, needs manual review)
// 		c.JSON(http.StatusOK, gin.H{"message": "auto review result"})
// 	}
// }

//////////////////////////////////////////////////////////
// HOUSING STAFF HANDLERS
//////////////////////////////////////////////////////////

// List all applications (housing staff)
func HousingListApplications(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		apps, err := database.HousingListApplications(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list applications", "details": err.Error()})
			return
		}
		c.JSON(http.StatusOK, apps)
	}
}

func HousingGetApplication(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, _ := strconv.Atoi(idStr)
		app, err := database.HousingGetApplication(db, id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
			return
		}
		c.JSON(http.StatusOK, app)
	}
}

func HousingApprove(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, _ := strconv.Atoi(idStr)
		uid, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user id"})
			return
		}
		reviewerID := uid.(int)

		if err := database.HousingApprove(db, id, reviewerID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "approve failed", "details": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "approved"})
	}
}

func HousingReject(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, _ := strconv.Atoi(idStr)
		var body struct {
			Reason string `json:"reason" binding:"required"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "reason is required"})
			return
		}
		uid, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user id"})
			return
		}
		reviewerID := uid.(int)

		if err := database.HousingReject(db, id, body.Reason, reviewerID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "reject failed", "details": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "rejected"})
	}
}

//////////////////////////////////////////////////////////
// ADMIN HANDLERS
//////////////////////////////////////////////////////////


func AdminListUsers(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		users, err := database.ListUsers(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
			return
		}
		c.JSON(http.StatusOK, users)
	}
}

func AdminCreateUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body struct {
			NuID     string `json:"nu_id" binding:"required"`
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required,min=6"`
			Phone    string `json:"phone"`
			Role     string `json:"role" binding:"required"` // allow admin if you want; here admin will be allowed since this endpoint is admin-only
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
			return
		}
		roleID, err := database.GetRoleIDByName(db, body.Role)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
			return
		}
		hash, err := auth.HashPassword(body.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "password hashing failed"})
			return
		}
		u := models.User{
			NuID:         body.NuID,
			Email:        body.Email,
			PasswordHash: hash,
			RoleID:       roleID,
			Phone:        body.Phone,
		}
		id, err := database.CreateUser(db, u)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "create user failed", "details": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"id": id})
	}
}

func AdminDeleteUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, _ := strconv.Atoi(idStr)
		if err := database.DeleteUser(db, id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed", "details": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	}
}

func AdminSystemLogs(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		logs, err := database.AdminSystemLogs(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch logs"})
			return
		}
		c.JSON(http.StatusOK, logs)
	}
}

func AdminStats(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats, err := database.AdminStats(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch stats"})
			return
		}
		c.JSON(http.StatusOK, stats)
	}
}
