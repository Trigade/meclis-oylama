package auth

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RequireSession — cookie'den member_id okur, context'e yazar
func RequireSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		val, err := c.Cookie("session_member_id")
		if err != nil || val == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "oturum gerekli"})
			return
		}

		memberID, err := strconv.Atoi(val)
		if err != nil || memberID <= 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "geçersiz oturum"})
			return
		}

		c.Set("member_id", memberID)
		c.Next()
	}
}

// RequireModerator — role kontrolü yapar
func RequireModerator(db interface {
	QueryRow(query string, args ...any) interface {
		Scan(dest ...any) error
	}
}) gin.HandlerFunc {
	return func(c *gin.Context) {
		memberID, exists := c.Get("member_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "oturum gerekli"})
			return
		}
		_ = memberID
		// TODO: DB'den role kontrolü eklenecek
		c.Next()
	}
}
