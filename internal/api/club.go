package api

import (
	"net/http"
	"strings"

	"github.com/alessandro/niedduty/internal/models"
	"github.com/gin-gonic/gin"
)

func (a *API) GetClub(c *gin.Context) {
	var club models.Club
	if err := a.db.First(&club, "id = 1").Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Verein nicht angelegt"})
		return
	}
	c.JSON(http.StatusOK, club)
}

type clubReq struct {
	Name             string `json:"name" binding:"required,max=100"`
	Short            string `json:"short" binding:"max=5"`
	PrimaryColor     string `json:"primaryColor" binding:"max=9"`
	SecondaryColor   string `json:"secondaryColor" binding:"max=9"`
	KasseIban        string `json:"kasseIban" binding:"max=40"`
	KasseInhaber     string `json:"kasseInhaber" binding:"max=100"`
	Liga             string `json:"liga" binding:"max=60"`
	FussballDeWidget string `json:"fussballDeWidget" binding:"max=300"`
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
	if req.FussballDeWidget != "" && !strings.HasPrefix(req.FussballDeWidget, "https://www.fussball.de/") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Widget-URL muss mit https://www.fussball.de/ beginnen"})
		return
	}
	club.Name, club.Short = req.Name, req.Short
	club.PrimaryColor, club.SecondaryColor = req.PrimaryColor, req.SecondaryColor
	club.KasseIban, club.KasseInhaber, club.Liga = req.KasseIban, req.KasseInhaber, req.Liga
	club.FussballDeWidget = req.FussballDeWidget
	a.db.Save(&club)
	c.JSON(http.StatusOK, club)
}
