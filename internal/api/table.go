package api

import (
	"net/http"
	"sort"

	"github.com/alessandro/niedduty/internal/models"
	"github.com/gin-gonic/gin"
)

func (a *API) GetTable(c *gin.Context) {
	var entries []models.LeagueEntry
	a.db.Find(&entries)
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].Points != entries[j].Points {
			return entries[i].Points > entries[j].Points
		}
		di := entries[i].GoalsFor - entries[i].GoalsAgainst
		dj := entries[j].GoalsFor - entries[j].GoalsAgainst
		if di != dj {
			return di > dj
		}
		return entries[i].GoalsFor > entries[j].GoalsFor
	})
	c.JSON(http.StatusOK, entries)
}

type tableReq struct {
	Entries []struct {
		TeamName     string `json:"teamName" binding:"required,max=100"`
		IsOwn        bool   `json:"isOwn"`
		Played       int    `json:"played"`
		Won          int    `json:"won"`
		Drawn        int    `json:"drawn"`
		Lost         int    `json:"lost"`
		GoalsFor     int    `json:"goalsFor"`
		GoalsAgainst int    `json:"goalsAgainst"`
		Points       int    `json:"points"`
	} `json:"entries" binding:"required"`
}

// ReplaceTable ersetzt die komplette Tabelle (einfachste konsistente Pflege).
func (a *API) ReplaceTable(c *gin.Context) {
	var req tableReq
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Entries) > 30 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Tabellendaten (max. 30 Teams)"})
		return
	}
	a.db.Exec("DELETE FROM league_entries")
	for _, e := range req.Entries {
		a.db.Create(&models.LeagueEntry{
			TeamName: e.TeamName, IsOwn: e.IsOwn, Played: e.Played,
			Won: e.Won, Drawn: e.Drawn, Lost: e.Lost,
			GoalsFor: e.GoalsFor, GoalsAgainst: e.GoalsAgainst, Points: e.Points,
		})
	}
	a.GetTable(c)
}
