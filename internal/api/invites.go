package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/alessandro/niedduty/internal/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var aliasRe = regexp.MustCompile(`^[a-z0-9._-]{3,24}$`)

func newToken(n int) string {
	buf := make([]byte, n)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}

// GetInvite (Admin) — aktueller aktiver Einladungslink oder null.
func (a *API) GetInvite(c *gin.Context) {
	var inv models.Invite
	if err := a.db.Where("active = ?", true).Order("created_at DESC").First(&inv).Error; err != nil {
		c.JSON(http.StatusOK, nil)
		return
	}
	c.JSON(http.StatusOK, inv)
}

// CreateInvite (Admin) — erzeugt einen neuen Link und deaktiviert alte.
func (a *API) CreateInvite(c *gin.Context) {
	a.db.Model(&models.Invite{}).Where("active = ?", true).Update("active", false)
	inv := models.Invite{Token: newToken(16), Active: true}
	if err := a.db.Create(&inv).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Einladung fehlgeschlagen"})
		return
	}
	c.JSON(http.StatusCreated, inv)
}

// DeactivateInvite (Admin) — macht alle Links ungültig.
func (a *API) DeactivateInvite(c *gin.Context) {
	a.db.Model(&models.Invite{}).Where("active = ?", true).Update("active", false)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (a *API) findValidInvite(token string) (*models.Invite, bool) {
	var inv models.Invite
	if err := a.db.First(&inv, "token = ?", token).Error; err != nil {
		return nil, false
	}
	if !inv.Active {
		return nil, false
	}
	if inv.ExpiresAt != nil && inv.ExpiresAt.Before(time.Now()) {
		return nil, false
	}
	if inv.MaxUses > 0 && inv.UseCount >= inv.MaxUses {
		return nil, false
	}
	return &inv, true
}

// CheckInvite (öffentlich) — prüft einen Token für die Registrierungsseite.
func (a *API) CheckInvite(c *gin.Context) {
	_, ok := a.findValidInvite(c.Param("token"))
	if !ok {
		c.JSON(http.StatusOK, gin.H{"valid": false})
		return
	}
	var club models.Club
	a.db.First(&club, "id = 1")
	c.JSON(http.StatusOK, gin.H{"valid": true, "clubName": club.Name})
}

type registerReq struct {
	Token     string `json:"token" binding:"required"`
	FirstName string `json:"firstName" binding:"required,max=60"`
	LastName  string `json:"lastName" binding:"required,max=60"`
	Alias     string `json:"alias" binding:"required"`
	Password  string `json:"password" binding:"required,min=8,max=100"`
}

// Register (öffentlich) — Selbstregistrierung per Einladungslink.
// Legt Konto + verknüpften Kader-Eintrag an und loggt direkt ein.
func (a *API) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vorname, Nachname, Alias und Passwort (min. 8 Zeichen) angeben"})
		return
	}
	inv, ok := a.findValidInvite(req.Token)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Einladungslink ungültig oder abgelaufen"})
		return
	}
	alias := strings.ToLower(strings.TrimSpace(req.Alias))
	if !aliasRe.MatchString(alias) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Alias: 3–24 Zeichen, nur a–z, 0–9, . _ -"})
		return
	}
	var exists int64
	a.db.Model(&models.User{}).Where("lower(alias) = ?", alias).Count(&exists)
	if exists > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Alias ist schon vergeben"})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Passwort-Hash fehlgeschlagen"})
		return
	}
	name := strings.TrimSpace(req.FirstName) + " " + strings.TrimSpace(req.LastName)

	var user models.User
	txErr := a.db.Transaction(func(tx *gorm.DB) error {
		player := models.Player{Name: name, Position: "MF", Status: "fit"}
		if err := tx.Create(&player).Error; err != nil {
			return err
		}
		user = models.User{
			Alias: alias, Name: name, PasswordHash: string(hash),
			Role: models.RoleMember, Permissions: []string{}, PlayerID: &player.ID,
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		return tx.Model(&models.Invite{}).Where("token = ?", inv.Token).
			Updates(map[string]any{"use_count": gorm.Expr("use_count + 1")}).Error
	})
	if txErr != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Registrierung fehlgeschlagen – Alias evtl. vergeben"})
		return
	}
	if inv.MaxUses > 0 && inv.UseCount+1 >= inv.MaxUses {
		a.db.Model(&models.Invite{}).Where("token = ?", inv.Token).Update("active", false)
	}
	if a.startSession(c, user) {
		c.JSON(http.StatusCreated, user)
	}
}
