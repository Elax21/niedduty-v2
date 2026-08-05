package api

import (
	"net/http"

	"github.com/alessandro/niedduty/internal/models"
	"github.com/gin-gonic/gin"
)

var validPositions = map[string]bool{"TW": true, "AB": true, "MF": true, "ST": true}
var validStatus = map[string]bool{"fit": true, "verletzt": true, "gesperrt": true, "krank": true}

type playerReq struct {
	Name     string `json:"name" binding:"required,max=100"`
	Number   *int   `json:"number"`
	Position string `json:"position" binding:"required"`
	Status   string `json:"status"`
	Birthday string `json:"birthday" binding:"omitempty,len=10"`
}

func (r *playerReq) validate() string {
	if !validPositions[r.Position] {
		return "Ungültige Position (TW/AB/MF/ST)"
	}
	if r.Status != "" && !validStatus[r.Status] {
		return "Ungültiger Status"
	}
	if r.Number != nil && (*r.Number < 1 || *r.Number > 99) {
		return "Rückennummer muss 1–99 sein"
	}
	return ""
}

func (a *API) ListPlayers(c *gin.Context) {
	var players []models.Player
	a.db.Order("number nulls last, name").Find(&players)
	c.JSON(http.StatusOK, players)
}

func (a *API) CreatePlayer(c *gin.Context) {
	var req playerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name und Position angeben"})
		return
	}
	if msg := req.validate(); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}
	p := models.Player{Name: req.Name, Number: req.Number, Position: req.Position, Status: "fit", Birthday: req.Birthday}
	if req.Status != "" {
		p.Status = req.Status
	}
	if err := a.db.Create(&p).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Spieler konnte nicht angelegt werden"})
		return
	}
	c.JSON(http.StatusCreated, p)
}

func (a *API) UpdatePlayer(c *gin.Context) {
	var p models.Player
	if err := a.db.First(&p, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Spieler nicht gefunden"})
		return
	}
	var req playerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name und Position angeben"})
		return
	}
	if msg := req.validate(); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}
	p.Name, p.Number, p.Position, p.Birthday = req.Name, req.Number, req.Position, req.Birthday
	if req.Status != "" {
		p.Status = req.Status
	}
	a.db.Save(&p)
	c.JSON(http.StatusOK, p)
}

func (a *API) DeletePlayer(c *gin.Context) {
	a.db.Delete(&models.Player{}, "id = ?", c.Param("id"))
	a.db.Delete(&models.PlayerPenalty{}, "player_id = ?", c.Param("id"))
	a.db.Delete(&models.EventAttendance{}, "player_id = ?", c.Param("id"))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
