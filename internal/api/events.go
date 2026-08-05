package api

import (
	"net/http"
	"sort"
	"time"

	"github.com/alessandro/niedduty/internal/middleware"
	"github.com/alessandro/niedduty/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const dateLayout = "2006-01-02"

var validEventTypes = map[string]bool{"training": true, "spiel": true, "mannschaftsabend": true, "sonstiges": true}

type eventReq struct {
	Title          string `json:"title" binding:"required,max=120"`
	Type           string `json:"type" binding:"required"`
	Date           string `json:"date" binding:"required"`
	EndDate        string `json:"endDate"`
	StartTime      string `json:"startTime" binding:"max=5"`
	EndTime        string `json:"endTime" binding:"max=5"`
	Location       string `json:"location" binding:"max=120"`
	Notes          string `json:"notes" binding:"max=2000"`
	Recurring      bool   `json:"recurring"`
	RecurrenceType string `json:"recurrenceType"`
	RecurrenceEnd  string `json:"recurrenceEnd"`
}

func (r *eventReq) validate() string {
	if !validEventTypes[r.Type] {
		return "Ungültiger Termin-Typ"
	}
	if _, err := time.Parse(dateLayout, r.Date); err != nil {
		return "Datum muss YYYY-MM-DD sein"
	}
	if r.EndDate != "" {
		if _, err := time.Parse(dateLayout, r.EndDate); err != nil {
			return "Enddatum muss YYYY-MM-DD sein"
		}
	}
	if r.Recurring {
		if r.RecurrenceType != "weekly" && r.RecurrenceType != "biweekly" {
			return "Wiederholung: weekly oder biweekly"
		}
		if r.RecurrenceEnd != "" {
			if _, err := time.Parse(dateLayout, r.RecurrenceEnd); err != nil {
				return "Wiederholungs-Ende muss YYYY-MM-DD sein"
			}
		}
	}
	return ""
}

func (r *eventReq) apply(e *models.Event) {
	e.Title, e.Type, e.Date, e.EndDate = r.Title, r.Type, r.Date, r.EndDate
	e.StartTime, e.EndTime, e.Location, e.Notes = r.StartTime, r.EndTime, r.Location, r.Notes
	e.Recurring, e.RecurrenceType, e.RecurrenceEnd = r.Recurring, r.RecurrenceType, r.RecurrenceEnd
}

// ListEvents expandiert Termine (inkl. Wiederholungen) im Bereich ?from&to.
func (a *API) ListEvents(c *gin.Context) {
	from, err1 := time.Parse(dateLayout, c.DefaultQuery("from", time.Now().AddDate(0, -1, 0).Format(dateLayout)))
	to, err2 := time.Parse(dateLayout, c.DefaultQuery("to", time.Now().AddDate(0, 2, 0).Format(dateLayout)))
	if err1 != nil || err2 != nil || to.Before(from) || to.Sub(from) > 400*24*time.Hour {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültiger Zeitraum (from/to, max. 400 Tage)"})
		return
	}
	var events []models.Event
	a.db.Find(&events)

	occs := expandEvents(events, from, to)
	sort.Slice(occs, func(i, j int) bool {
		if occs[i].OccDate != occs[j].OccDate {
			return occs[i].OccDate < occs[j].OccDate
		}
		return occs[i].StartTime < occs[j].StartTime
	})
	a.fillAttendance(c, occs)
	a.fillNotes(occs)
	c.JSON(http.StatusOK, occs)
}

// fillNotes hängt die Notizen einzelner Vorkommen an (z. B. Trainingsinhalt).
func (a *API) fillNotes(occs []models.Occurrence) {
	if len(occs) == 0 {
		return
	}
	keys := make([]string, 0, len(occs))
	for _, o := range occs {
		keys = append(keys, o.EventKey)
	}
	var notes []models.EventNote
	a.db.Where("event_key IN ?", keys).Find(&notes)
	byKey := make(map[string]string, len(notes))
	for _, n := range notes {
		byKey[n.EventKey] = n.Text
	}
	for i := range occs {
		occs[i].OccNote = byKey[occs[i].EventKey]
	}
}

// fillAttendance ergänzt jedes Vorkommen um Zusagen/Absagen/Offen und die
// eigene Rückmeldung — damit die Terminliste ohne Nachladen auskommt.
func (a *API) fillAttendance(c *gin.Context, occs []models.Occurrence) {
	if len(occs) == 0 {
		return
	}
	keys := make([]string, 0, len(occs))
	for _, o := range occs {
		keys = append(keys, o.EventKey)
	}
	var atts []models.EventAttendance
	a.db.Where("event_key IN ?", keys).Find(&atts)

	var squadSize int64
	a.db.Model(&models.Player{}).Count(&squadSize)

	var me uuid.UUID
	if u := middleware.CurrentUser(c); u != nil && u.PlayerID != nil {
		me = *u.PlayerID
	}

	counts := map[string]*models.Occurrence{}
	for i := range occs {
		counts[occs[i].EventKey] = &occs[i]
	}
	for _, at := range atts {
		o := counts[at.EventKey]
		if o == nil {
			continue
		}
		switch at.Status {
		case "attending":
			o.Attending++
		case "declined":
			o.Declined++
		}
		if me != uuid.Nil && at.PlayerID == me {
			o.MyStatus, o.MyReason = at.Status, at.Reason
		}
	}
	for i := range occs {
		if open := int(squadSize) - occs[i].Attending - occs[i].Declined; open > 0 {
			occs[i].Open = open
		}
	}
}

// expandEvents erzeugt Vorkommen im Bereich [from, to].
// EventKey: ID bei Einzelterminen, "ID_YYYY-MM-DD" bei Wiederholungen.
func expandEvents(events []models.Event, from, to time.Time) []models.Occurrence {
	occs := []models.Occurrence{}
	for _, e := range events {
		start, err := time.Parse(dateLayout, e.Date)
		if err != nil {
			continue
		}
		if !e.Recurring {
			if !start.Before(from) && !start.After(to) || spansRange(e, from, to) {
				occs = append(occs, models.Occurrence{Event: e, EventKey: e.ID.String(), OccDate: e.Date})
			}
			continue
		}
		step := 7
		if e.RecurrenceType == "biweekly" {
			step = 14
		}
		end := to
		if e.RecurrenceEnd != "" {
			if re, err := time.Parse(dateLayout, e.RecurrenceEnd); err == nil && re.Before(to) {
				end = re
			}
		}
		for d := start; !d.After(end); d = d.AddDate(0, 0, step) {
			if d.Before(from) {
				continue
			}
			ds := d.Format(dateLayout)
			occs = append(occs, models.Occurrence{Event: e, EventKey: e.ID.String() + "_" + ds, OccDate: ds})
		}
	}
	return occs
}

// spansRange — mehrtägiger Einzeltermin, der in den Bereich hineinragt.
func spansRange(e models.Event, from, to time.Time) bool {
	if e.EndDate == "" {
		return false
	}
	start, err1 := time.Parse(dateLayout, e.Date)
	end, err2 := time.Parse(dateLayout, e.EndDate)
	if err1 != nil || err2 != nil {
		return false
	}
	return !end.Before(from) && !start.After(to)
}

func (a *API) CreateEvent(c *gin.Context) {
	var req eventReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Titel, Typ und Datum angeben"})
		return
	}
	if msg := req.validate(); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}
	var e models.Event
	req.apply(&e)
	a.db.Create(&e)
	c.JSON(http.StatusCreated, e)
}

func (a *API) UpdateEvent(c *gin.Context) {
	var e models.Event
	if err := a.db.First(&e, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Termin nicht gefunden"})
		return
	}
	var req eventReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Titel, Typ und Datum angeben"})
		return
	}
	if msg := req.validate(); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}
	req.apply(&e)
	a.db.Save(&e)
	c.JSON(http.StatusOK, e)
}

func (a *API) DeleteEvent(c *gin.Context) {
	a.db.Delete(&models.Event{}, "id = ?", c.Param("id"))
	a.db.Delete(&models.EventAttendance{}, "event_key = ? OR event_key LIKE ?", c.Param("id"), c.Param("id")+"\\_%")
	a.db.Delete(&models.EventNote{}, "event_key = ? OR event_key LIKE ?", c.Param("id"), c.Param("id")+"\\_%")
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
