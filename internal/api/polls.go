package api

import (
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/alessandro/niedduty/internal/middleware"
	"github.com/alessandro/niedduty/internal/models"
	"github.com/alessandro/niedduty/internal/push"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Abstimmungen. Laufende erscheinen auf der Startseite; wer abgestimmt hat,
// sieht das Ergebnis. Vorher bleiben die Zahlen verborgen, damit die ersten
// Stimmen den Rest nicht in eine Richtung ziehen.

const (
	maxPollOptions = 10
	// Stimmen sind nicht geheim: in einer Mannschaft ist ohnehin klar, wer
	// fehlt. Der Name hängt an der Stimme, damit man nachfassen kann.
	pollReminderLead = 24 * time.Hour
)

// pollResult — Abstimmung samt Zählung für die Oberfläche.
type pollResult struct {
	models.Poll
	// Counts je Option, gleiche Reihenfolge wie Options.
	Counts []int `json:"counts"`
	// Voters je Option — Namen, damit man weiß, wer noch fehlt.
	Voters [][]string `json:"voters"`
	// MyVotes — die eigenen angekreuzten Optionen.
	MyVotes []int `json:"myVotes"`
	// Total — wie viele Konten abgestimmt haben (nicht Stimmen).
	Total int `json:"total"`
	// Running — läuft noch (nicht geschlossen, Ende nicht erreicht).
	Running bool `json:"running"`
}

func pollRunning(p models.Poll, now time.Time) bool {
	if p.ClosedAt != nil {
		return false
	}
	return p.EndsAt == nil || p.EndsAt.After(now)
}

// buildResults hängt an jede Abstimmung ihre Zählung.
func (a *API) buildResults(polls []models.Poll, me uuid.UUID) []pollResult {
	out := make([]pollResult, 0, len(polls))
	if len(polls) == 0 {
		return out
	}
	ids := make([]uuid.UUID, 0, len(polls))
	for _, p := range polls {
		ids = append(ids, p.ID)
	}
	var votes []models.PollVote
	a.db.Where("poll_id IN ?", ids).Order("created_at").Find(&votes)

	byPoll := map[uuid.UUID][]models.PollVote{}
	for _, v := range votes {
		byPoll[v.PollID] = append(byPoll[v.PollID], v)
	}

	now := time.Now()
	for _, p := range polls {
		r := pollResult{
			Poll:    p,
			Counts:  make([]int, len(p.Options)),
			Voters:  make([][]string, len(p.Options)),
			MyVotes: []int{},
			Running: pollRunning(p, now),
		}
		for i := range r.Voters {
			r.Voters[i] = []string{}
		}
		seen := map[uuid.UUID]bool{}
		for _, v := range byPoll[p.ID] {
			if v.OptionIdx < 0 || v.OptionIdx >= len(p.Options) {
				continue
			}
			r.Counts[v.OptionIdx]++
			r.Voters[v.OptionIdx] = append(r.Voters[v.OptionIdx], v.VoterName)
			if !seen[v.UserID] {
				seen[v.UserID] = true
				r.Total++
			}
			if v.UserID == me {
				r.MyVotes = append(r.MyVotes, v.OptionIdx)
			}
		}
		sort.Ints(r.MyVotes)
		out = append(out, r)
	}
	return out
}

// ListPolls — laufende zuerst, danach die beendeten.
func (a *API) ListPolls(c *gin.Context) {
	var polls []models.Poll
	a.db.Order("created_at desc").Find(&polls)
	c.JSON(http.StatusOK, a.buildResults(polls, middleware.CurrentUser(c).ID))
}

// ListRunningPolls — nur laufende, für die Startseite.
func (a *API) ListRunningPolls(c *gin.Context) {
	var polls []models.Poll
	a.db.Where("closed_at IS NULL AND (ends_at IS NULL OR ends_at > ?)", time.Now()).
		Order("created_at desc").Find(&polls)
	c.JSON(http.StatusOK, a.buildResults(polls, middleware.CurrentUser(c).ID))
}

type createPollReq struct {
	Question    string   `json:"question" binding:"required,max=200"`
	Options     []string `json:"options" binding:"required,min=2,max=10"`
	MultiChoice bool     `json:"multiChoice"`
	// EndsAt als "YYYY-MM-DDTHH:MM" (lokale Zeit) oder leer.
	EndsAt string `json:"endsAt"`
}

// CreatePoll startet eine Abstimmung und schickt sie als Push an alle.
func (a *API) CreatePoll(c *gin.Context) {
	var req createPollReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Frage und mindestens zwei Antworten angeben"})
		return
	}
	options := make([]string, 0, len(req.Options))
	for _, o := range req.Options {
		if o = strings.TrimSpace(o); o != "" {
			options = append(options, truncate(o, 120))
		}
	}
	if len(options) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Mindestens zwei Antworten angeben"})
		return
	}
	if len(options) > maxPollOptions {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Höchstens zehn Antworten"})
		return
	}

	var endsAt *time.Time
	if req.EndsAt != "" {
		t, err := time.ParseInLocation("2006-01-02T15:04", req.EndsAt, time.Local)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ende muss ein gültiger Zeitpunkt sein"})
			return
		}
		if !t.After(time.Now()) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Das Ende muss in der Zukunft liegen"})
			return
		}
		endsAt = &t
	}

	user := middleware.CurrentUser(c)
	p := models.Poll{
		Question:    strings.TrimSpace(req.Question),
		Options:     options,
		MultiChoice: req.MultiChoice,
		EndsAt:      endsAt,
		CreatedBy:   user.ID,
		CreatorName: user.Name,
	}
	if err := a.db.Create(&p).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Abstimmung konnte nicht angelegt werden"})
		return
	}

	push.SendToAll(a.db, push.Payload{
		Title: "🗳 Neue Abstimmung",
		Body:  p.Question + pollDeadlineSuffix(p),
		URL:   "/abstimmungen",
		Tag:   "poll-" + p.ID.String(),
	})

	c.JSON(http.StatusCreated, a.buildResults([]models.Poll{p}, user.ID)[0])
}

func pollDeadlineSuffix(p models.Poll) string {
	if p.EndsAt == nil {
		return ""
	}
	return " — bis " + p.EndsAt.Format("02.01., 15:04") + " Uhr"
}

type voteReq struct {
	// Options — angekreuzte Indizes. Leer = Stimme zurückziehen.
	Options []int `json:"options"`
}

// Vote setzt die eigene Stimme. Erneutes Abstimmen ersetzt die alte Wahl.
func (a *API) Vote(c *gin.Context) {
	var p models.Poll
	if err := a.db.First(&p, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Abstimmung nicht gefunden"})
		return
	}
	if !pollRunning(p, time.Now()) {
		c.JSON(http.StatusConflict, gin.H{"error": "Die Abstimmung ist beendet"})
		return
	}
	var req voteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Auswahl"})
		return
	}
	if !p.MultiChoice && len(req.Options) > 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Hier ist nur eine Antwort erlaubt"})
		return
	}
	seen := map[int]bool{}
	for _, idx := range req.Options {
		if idx < 0 || idx >= len(p.Options) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unbekannte Antwort"})
			return
		}
		seen[idx] = true
	}

	user := middleware.CurrentUser(c)
	a.db.Delete(&models.PollVote{}, "poll_id = ? AND user_id = ?", p.ID, user.ID)

	rows := make([]models.PollVote, 0, len(seen))
	for idx := range seen {
		rows = append(rows, models.PollVote{
			PollID: p.ID, UserID: user.ID, OptionIdx: idx,
			VoterName: user.Name, CreatedAt: time.Now(),
		})
	}
	if len(rows) > 0 {
		if err := a.db.Create(&rows).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Stimme konnte nicht gespeichert werden"})
			return
		}
	}
	c.JSON(http.StatusOK, a.buildResults([]models.Poll{p}, user.ID)[0])
}

// ClosePoll beendet vorzeitig; das Ergebnis bleibt sichtbar.
func (a *API) ClosePoll(c *gin.Context) {
	var p models.Poll
	if err := a.db.First(&p, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Abstimmung nicht gefunden"})
		return
	}
	now := time.Now()
	p.ClosedAt = &now
	a.db.Save(&p)
	c.JSON(http.StatusOK, a.buildResults([]models.Poll{p}, middleware.CurrentUser(c).ID)[0])
}

// DeletePoll entfernt Abstimmung samt Stimmen.
func (a *API) DeletePoll(c *gin.Context) {
	a.db.Delete(&models.Poll{}, "id = ?", c.Param("id"))
	a.db.Delete(&models.PollVote{}, "poll_id = ?", c.Param("id"))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
