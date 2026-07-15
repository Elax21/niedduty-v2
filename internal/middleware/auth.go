package middleware

import (
	"net/http"
	"time"

	"github.com/alessandro/niedduty/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const SessionCookie = "ndt_session"

// Auth lädt Session + User aus dem Cookie und legt den User in den Context.
func Auth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(SessionCookie)
		if err != nil || token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Nicht eingeloggt"})
			return
		}
		var sess models.Session
		if err := db.First(&sess, "token = ?", token).Error; err != nil || sess.ExpiresAt.Before(time.Now()) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Sitzung abgelaufen"})
			return
		}
		var user models.User
		if err := db.First(&user, "id = ?", sess.UserID).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unbekannter Benutzer"})
			return
		}
		c.Set("user", &user)
		c.Next()
	}
}

// RequireAdmin — nur der Admin (Einstellungen, Benutzer, Kader, Tabelle).
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if CurrentUser(c).Role != models.RoleAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Nur für den Admin"})
			return
		}
		c.Next()
	}
}

// RequirePerm — Admin oder Mitglied mit entsprechendem Recht.
func RequirePerm(perm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !CurrentUser(c).Can(perm) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Keine Berechtigung"})
			return
		}
		c.Next()
	}
}

// CurrentUser holt den eingeloggten User aus dem Context (nach Auth-Middleware).
func CurrentUser(c *gin.Context) *models.User {
	u, _ := c.Get("user")
	return u.(*models.User)
}
