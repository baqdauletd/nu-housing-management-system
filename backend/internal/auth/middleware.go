package auth

import (
   "net/http"
   "strings"

   "github.com/gin-gonic/gin"
)

// AuthMiddleware ensures a request has a valid JWT.
func AuthMiddleware() gin.HandlerFunc {
   return func(c *gin.Context) {
      authHeader := c.GetHeader("Authorization")
      if authHeader == "" {
         c.JSON(http.StatusUnauthorized, gin.H{"error": "missing Authorization header"})
         c.Abort()
         return
      }

      parts := strings.Split(authHeader, " ")
      if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
         c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid auth header"})
         c.Abort()
         return
      }

      tokenStr := parts[1]

      claims, err := ParseToken(tokenStr)
      if err != nil {
         c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token", "details": err.Error()})
         c.Abort()
         return
      }

      c.Set("user_id", claims.UserID)
      c.Set("role", claims.Role)

      c.Next()
   }
}

// RoleMiddleware restricts routes by one or more allowed roles
func RoleMiddleware(allowed ...string) gin.HandlerFunc {
   return func(c *gin.Context) {
      roleVal, exists := c.Get("role")
      if !exists {
         c.JSON(http.StatusForbidden, gin.H{"error": "role missing"})
         c.Abort()
         return
      }

      role := roleVal.(string)

      for _, r := range allowed {
         if role == r {
            c.Next()
            return
         }
      }

      c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
      c.Abort()
   }
}