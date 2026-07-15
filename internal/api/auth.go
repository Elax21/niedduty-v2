package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/alessandro/niedduty/internal/middleware"
	"github.com/alessandro/niedduty/internal/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

const sessionTTL = 30 * 24 * time.Hour

type loginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (a *API) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "E-Mail und Passwort angeben"})
		return
	}
	var user models.User
	if err := a.db.First(&user, "lower(email) = lower(?)", req.Email).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "E-Mail oder Passwort falsch"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "E-Mail oder Passwort falsch"})
		return
	}

	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token-Erzeugung fehlgeschlagen"})
		return
	}
	token := hex.EncodeToString(buf)
	sess := models.Session{Token: token, UserID: user.ID, ExpiresAt: time.Now().Add(sessionTTL)}
	if err := a.db.Create(&sess).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Sitzung konnte nicht angelegt werden"})
		return
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(middleware.SessionCookie, token, int(sessionTTL.Seconds()), "/", "", false, true)
	c.JSON(http.StatusOK, user)
}

func (a *API) Logout(c *gin.Context) {
	if token, err := c.Cookie(middleware.SessionCookie); err == nil {
		a.db.Delete(&models.Session{}, "token = ?", token)
	}
	c.SetCookie(middleware.SessionCookie, "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (a *API) Me(c *gin.Context) {
	c.JSON(http.StatusOK, middleware.CurrentUser(c))
}
