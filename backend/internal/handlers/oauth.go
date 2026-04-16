package handlers

import (
	"database/sql"
	"net/http"
	"net/url"
	"strings"

	appAuth "nu-housing-management-system/backend/internal/auth"
	"nu-housing-management-system/backend/internal/database"

	"github.com/gin-gonic/gin"
)

func GoogleSignIn(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body struct {
			IDToken string `json:"id_token" binding:"required"`
		}

		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id_token is required"})
			return
		}

		tokenInfo, err := appAuth.VerifyGoogleIDToken(body.IDToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid Google sign-in token", "details": err.Error()})
			return
		}

		email := strings.ToLower(strings.TrimSpace(tokenInfo.Email))
		user, _, err := database.EnsureOAuthUser(db, email, generateFallbackNuID())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to provision user", "details": err.Error()})
			return
		}

		roleName, err := database.GetRoleNameByID(db, user.RoleID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve user role"})
			return
		}

		jwtToken, err := appAuth.GenerateToken(user.ID, roleName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
			return
		}

		db.Exec(`INSERT INTO audit_logs (actor_id, action, entity, entity_id) VALUES ($1, 'login_google', 'user', $1)`, user.ID)

		c.JSON(http.StatusOK, gin.H{
			"token": jwtToken,
			"user": gin.H{
				"id":    user.ID,
				"nu_id": user.NuID,
				"email": user.Email,
				"role":  roleName,
				"phone": user.Phone,
			},
		})
	}
}

func GoogleOAuthStartUnsupported() gin.HandlerFunc {
	return func(c *gin.Context) {
		const message = "Google OAuth redirect flow is not implemented by this backend. Use POST /auth/google with a frontend-issued Google id_token."

		redirectURI := strings.TrimSpace(c.Query("redirect_uri"))
		if redirectURI != "" {
			target, err := url.Parse(redirectURI)
			if err == nil {
				query := target.Query()
				query.Set("error", "unsupported_oauth_flow")
				query.Set("error_description", message)
				target.RawQuery = query.Encode()
				c.Redirect(http.StatusTemporaryRedirect, target.String())
				return
			}
		}

		c.JSON(http.StatusNotImplemented, gin.H{
			"error":   "unsupported_oauth_flow",
			"message": message,
		})
	}
}
