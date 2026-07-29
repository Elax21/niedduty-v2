package api

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/alessandro/niedduty/internal/models"
	"github.com/gin-gonic/gin"
)

// widgetIdRe akzeptiert fussball.de-Widget-IDs (UUID-Format oder lange Hex/Alnum-Kennung).
var widgetIdRe = regexp.MustCompile(`^[a-zA-Z0-9-]{8,64}$`)

func (a *API) GetClub(c *gin.Context) {
	var club models.Club
	if err := a.db.First(&club, "id = 1").Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Verein nicht angelegt"})
		return
	}
	c.JSON(http.StatusOK, club)
}

type clubReq struct {
	Name               string `json:"name" binding:"required,max=100"`
	Short              string `json:"short" binding:"max=5"`
	PrimaryColor       string `json:"primaryColor" binding:"max=9"`
	SecondaryColor     string `json:"secondaryColor" binding:"max=9"`
	KasseIban          string `json:"kasseIban" binding:"max=40"`
	KasseInhaber       string `json:"kasseInhaber" binding:"max=100"`
	Liga                string `json:"liga" binding:"max=60"`
	FussballTableId     string `json:"fussballTableId" binding:"max=64"`
	FussballMatchesId   string `json:"fussballMatchesId" binding:"max=64"`
	FussballNextMatchId string `json:"fussballNextMatchId" binding:"max=64"`
	GoogleCalendarUrl   string `json:"googleCalendarUrl" binding:"max=400"`
}

func (a *API) UpdateClub(c *gin.Context) {
	var req clubReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vereinsname angeben"})
		return
	}
	var club models.Club
	if err := a.db.First(&club, "id = 1").Error; err != nil {
		club = models.Club{ID: 1}
	}
	for _, id := range []string{req.FussballTableId, req.FussballMatchesId, req.FussballNextMatchId} {
		if id != "" && !widgetIdRe.MatchString(id) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Widget-ID muss eine fussball.de-ID sein (z.B. aab8a3a1-12c9-4a0a-bd06-da2911b780ea)"})
			return
		}
	}
	if req.GoogleCalendarUrl != "" && !strings.HasPrefix(req.GoogleCalendarUrl, "https://calendar.google.com/") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Kalender-Link muss mit https://calendar.google.com/ beginnen"})
		return
	}
	club.Name, club.Short = req.Name, req.Short
	club.PrimaryColor, club.SecondaryColor = req.PrimaryColor, req.SecondaryColor
	club.KasseIban, club.KasseInhaber, club.Liga = req.KasseIban, req.KasseInhaber, req.Liga
	club.FussballTableId = req.FussballTableId
	club.FussballMatchesId = req.FussballMatchesId
	club.FussballNextMatchId = req.FussballNextMatchId
	club.GoogleCalendarUrl = req.GoogleCalendarUrl
	a.db.Save(&club)
	c.JSON(http.StatusOK, club)
}
