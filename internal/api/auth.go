package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/alessandro/niedduty/internal/middleware"
	"github.com/alessandro/niedduty/internal/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

const sessionTTL = 30 * 24 * time.Hour

// startSession legt eine Session an und setzt das Cookie.
func (a *API) startSession(c *gin.Context, user models.User) bool {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token-Erzeugung fehlgeschlagen"})
		return false
	}
	token := hex.EncodeToString(buf)
	sess := models.Session{Token: token, UserID: user.ID, ExpiresAt: time.Now().Add(sessionTTL)}
	if err := a.db.Create(&sess).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Sitzung konnte nicht angelegt werden"})
		return false
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(middleware.SessionCookie, token, int(sessionTTL.Seconds()), "/", "", a.secureCookies, true)
	return true
}

type loginReq struct {
	Login    string `json:"login" binding:"required"` // Alias oder E-Mail
	Password string `json:"password" binding:"required"`
}

func (a *API) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Alias und Passwort angeben"})
		return
	}
	id := strings.TrimSpace(req.Login)
	var user models.User
	if err := a.db.First(&user, "lower(alias) = lower(?) OR lower(email) = lower(?)", id, id).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Alias oder Passwort falsch"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Alias oder Passwort falsch"})
		return
	}
	if a.startSession(c, user) {
		c.JSON(http.StatusOK, user)
	}
}

func (a *API) Logout(c *gin.Context) {
	if token, err := c.Cookie(middleware.SessionCookie); err == nil {
		a.db.Delete(&models.Session{}, "token = ?", token)
	}
	c.SetCookie(middleware.SessionCookie, "", -1, "/", "", a.secureCookies, true)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (a *API) Me(c *gin.Context) {
	c.JSON(http.StatusOK, middleware.CurrentUser(c))
}

// SetTutorialDone merkt am eigenen Konto, dass der Rundgang durch ist.
// Gilt immer nur für die eigene Session — es gibt bewusst keine User-ID im
// Request, damit niemand an fremden Konten schrauben kann. `done` erlaubt das
// Zurücksetzen, um den Rundgang noch einmal zu starten.
func (a *API) SetTutorialDone(c *gin.Context) {
	var req struct {
		Done *bool `json:"done"`
	}
	_ = c.ShouldBindJSON(&req) // leerer Body = fertig
	done := true
	if req.Done != nil {
		done = *req.Done
	}
	user := middleware.CurrentUser(c)
	if err := a.db.Model(&models.User{}).Where("id = ?", user.ID).
		Update("tutorial_done", done).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Konnte nicht gespeichert werden"})
		return
	}
	user.TutorialDone = done
	c.JSON(http.StatusOK, gin.H{"tutorialDone": done})
}

// SetChangelogSeen merkt am eigenen Konto, welche Neuerungen gelesen wurden.
// Wie beim Rundgang ohne User-ID im Request — immer nur die eigene Session.
func (a *API) SetChangelogSeen(c *gin.Context) {
	var req struct {
		Version string `json:"version" binding:"required,max=20"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Version fehlt"})
		return
	}
	user := middleware.CurrentUser(c)
	if err := a.db.Model(&models.User{}).Where("id = ?", user.ID).
		Update("seen_changelog", req.Version).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Konnte nicht gespeichert werden"})
		return
	}
	user.SeenChangelog = req.Version
	c.JSON(http.StatusOK, gin.H{"seenChangelog": req.Version})
}
