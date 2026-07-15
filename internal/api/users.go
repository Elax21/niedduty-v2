package api

import (
	"net/http"

	"github.com/alessandro/niedduty/internal/middleware"
	"github.com/alessandro/niedduty/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var validPerms = map[string]bool{
	models.PermStrafen:     true,
	models.PermTermine:     true,
	models.PermBeteiligung: true,
}

func checkPerms(perms []string) bool {
	for _, p := range perms {
		if !validPerms[p] {
			return false
		}
	}
	return true
}

func (a *API) ListUsers(c *gin.Context) {
	var users []models.User
	a.db.Order("created_at").Find(&users)
	c.JSON(http.StatusOK, users)
}

type createUserReq struct {
	Email       string     `json:"email" binding:"required,email,max=120"`
	Name        string     `json:"name" binding:"required,max=100"`
	Password    string     `json:"password" binding:"required,min=8,max=100"`
	Permissions []string   `json:"permissions"`
	PlayerID    *uuid.UUID `json:"playerId"`
}

func (a *API) CreateUser(c *gin.Context) {
	var req createUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "E-Mail, Name und Passwort (min. 8 Zeichen) angeben"})
		return
	}
	if !checkPerms(req.Permissions) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unbekanntes Recht"})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Passwort-Hash fehlgeschlagen"})
		return
	}
	if req.Permissions == nil {
		req.Permissions = []string{}
	}
	u := models.User{
		Email: req.Email, Name: req.Name, PasswordHash: string(hash),
		Role: models.RoleMember, Permissions: req.Permissions, PlayerID: req.PlayerID,
	}
	if err := a.db.Create(&u).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "E-Mail bereits vergeben"})
		return
	}
	c.JSON(http.StatusCreated, u)
}

type updateUserReq struct {
	Name        string     `json:"name" binding:"required,max=100"`
	Permissions []string   `json:"permissions"`
	PlayerID    *uuid.UUID `json:"playerId"`
	Password    string     `json:"password" binding:"omitempty,min=8,max=100"`
}

func (a *API) UpdateUser(c *gin.Context) {
	var u models.User
	if err := a.db.First(&u, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Benutzer nicht gefunden"})
		return
	}
	var req updateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name angeben; Passwort min. 8 Zeichen"})
		return
	}
	if !checkPerms(req.Permissions) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unbekanntes Recht"})
		return
	}
	if req.Permissions == nil {
		req.Permissions = []string{}
	}
	u.Name, u.PlayerID = req.Name, req.PlayerID
	if u.Role != models.RoleAdmin { // Admin-Rechte sind implizit, nicht editierbar
		u.Permissions = req.Permissions
	}
	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Passwort-Hash fehlgeschlagen"})
			return
		}
		u.PasswordHash = string(hash)
	}
	a.db.Save(&u)
	c.JSON(http.StatusOK, u)
}

func (a *API) DeleteUser(c *gin.Context) {
	if middleware.CurrentUser(c).ID.String() == c.Param("id") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Eigenes Konto kann nicht gelöscht werden"})
		return
	}
	a.db.Delete(&models.User{}, "id = ? AND role <> ?", c.Param("id"), models.RoleAdmin)
	a.db.Delete(&models.Session{}, "user_id = ?", c.Param("id"))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
