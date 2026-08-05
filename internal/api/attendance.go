package api

import (
	"net/http"
	"time"

	"github.com/alessandro/niedduty/internal/middleware"
	"github.com/alessandro/niedduty/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm/clause"
)

// ListAttendance — Rückmeldungen eines Termin-Vorkommens.
//
// Wer zu- oder abgesagt hat, sieht nur der Mannschaftsrat (Recht
// "beteiligung"). Alle anderen bekommen ausschließlich ihre eigene
// Rückmeldung; die reinen Zähler stecken ohnehin schon am Vorkommen.
func (a *API) ListAttendance(c *gin.Context) {
	key := c.Query("eventKey")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "eventKey fehlt"})
		return
	}
	user := middleware.CurrentUser(c)
	q := a.db.Where("event_key = ?", key)
	if !user.Can(models.PermBeteiligung) {
		if user.PlayerID == nil {
			c.JSON(http.StatusOK, []models.EventAttendance{})
			return
		}
		q = q.Where("player_id = ?", *user.PlayerID)
	}
	var list []models.EventAttendance
	q.Find(&list)
	c.JSON(http.StatusOK, list)
}

type attendanceReq struct {
	EventKey string    `json:"eventKey" binding:"required,max=60"`
	PlayerID uuid.UUID `json:"playerId" binding:"required"`
	Status   string    `json:"status" binding:"required"`
	Reason   string    `json:"reason" binding:"max=200"`
}

// SetAttendance — Coach darf für alle setzen, Spieler nur für sich selbst.
func (a *API) SetAttendance(c *gin.Context) {
	var req attendanceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "eventKey, playerId und status angeben"})
		return
	}
	if req.Status != "attending" && req.Status != "declined" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status: attending oder declined"})
		return
	}
	user := middleware.CurrentUser(c)
	ownEntry := user.PlayerID != nil && *user.PlayerID == req.PlayerID
	if !ownEntry && !user.Can(models.PermTermine) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Nur die eigene Rückmeldung erlaubt"})
		return
	}
	att := models.EventAttendance{
		EventKey: req.EventKey, PlayerID: req.PlayerID,
		Status: req.Status, Reason: req.Reason, UpdatedAt: time.Now(),
	}
	a.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "event_key"}, {Name: "player_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"status", "reason", "updated_at"}),
	}).Create(&att)
	c.JSON(http.StatusOK, att)
}

type playerStats struct {
	PlayerID uuid.UUID `json:"playerId"`
	Name     string    `json:"name"`
	Number   *int      `json:"number"`
	Attended int       `json:"attended"`
	Declined int       `json:"declined"`
	NoAnswer int       `json:"noAnswer"`
	Total    int       `json:"total"`
	QuotePct int       `json:"quotePct"`
}

// AttendanceStats — Trainingsbeteiligung pro Spieler über einen Zeitraum.
// Zählt nur Trainings-Vorkommen bis heute (Zukunft zählt nicht als verpasst).
func (a *API) AttendanceStats(c *gin.Context) {
	from, err1 := time.Parse(dateLayout, c.DefaultQuery("from", time.Now().AddDate(0, -3, 0).Format(dateLayout)))
	to, err2 := time.Parse(dateLayout, c.DefaultQuery("to", time.Now().Format(dateLayout)))
	if err1 != nil || err2 != nil || to.Before(from) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültiger Zeitraum (from/to)"})
		return
	}
	if today, _ := time.Parse(dateLayout, time.Now().Format(dateLayout)); to.After(today) {
		to = today
	}

	var events []models.Event
	a.db.Where("type = ?", "training").Find(&events)
	occs := expandEvents(events, from, to)
	keys := make([]string, 0, len(occs))
	for _, o := range occs {
		keys = append(keys, o.EventKey)
	}

	var atts []models.EventAttendance
	if len(keys) > 0 {
		a.db.Where("event_key IN ?", keys).Find(&atts)
	}
	byPlayer := map[uuid.UUID]map[string]int{}
	for _, at := range atts {
		if byPlayer[at.PlayerID] == nil {
			byPlayer[at.PlayerID] = map[string]int{}
		}
		byPlayer[at.PlayerID][at.Status]++
	}

	var players []models.Player
	a.db.Order("number nulls last, name").Find(&players)

	total := len(occs)
	stats := make([]playerStats, 0, len(players))
	for _, p := range players {
		att := byPlayer[p.ID]["attending"]
		dec := byPlayer[p.ID]["declined"]
		s := playerStats{
			PlayerID: p.ID, Name: p.Name, Number: p.Number,
			Attended: att, Declined: dec, NoAnswer: total - att - dec, Total: total,
		}
		if total > 0 {
			s.QuotePct = att * 100 / total
		}
		stats = append(stats, s)
	}
	c.JSON(http.StatusOK, gin.H{
		"from": from.Format(dateLayout), "to": to.Format(dateLayout),
		"trainings": total, "players": stats,
	})
}
