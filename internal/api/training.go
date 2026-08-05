package api

import (
	"net/http"
	"sort"
	"time"

	"github.com/alessandro/niedduty/internal/models"
	"github.com/gin-gonic/gin"
)

// Trainingsplan — die festen Wochentage, an denen trainiert wird.
//
// Dahinter stecken ganz normale wöchentliche Termine (Series = "training"),
// einer pro Wochentag. So bleiben Zu-/Absagen, Kalender-Links und die
// Beteiligungs-Statistik unverändert; hier gibt es nur eine Oberfläche, die
// alle Wochentage auf einmal setzt.
const trainingSeries = "training"

type trainingScheduleReq struct {
	// Weekdays: 1 = Montag … 7 = Sonntag.
	Weekdays      []int  `json:"weekdays"`
	Title         string `json:"title" binding:"max=120"`
	StartTime     string `json:"startTime" binding:"max=5"`
	EndTime       string `json:"endTime" binding:"max=5"`
	Location      string `json:"location" binding:"max=120"`
	Notes         string `json:"notes" binding:"max=2000"`
	RecurrenceEnd string `json:"recurrenceEnd"`
}

type trainingScheduleRes struct {
	Weekdays      []int  `json:"weekdays"`
	Title         string `json:"title"`
	StartTime     string `json:"startTime"`
	EndTime       string `json:"endTime"`
	Location      string `json:"location"`
	Notes         string `json:"notes"`
	RecurrenceEnd string `json:"recurrenceEnd"`
}

// GetTrainingSchedule liefert die aktuell gepflegten Trainings-Wochentage.
func (a *API) GetTrainingSchedule(c *gin.Context) {
	var events []models.Event
	a.db.Where("series = ?", trainingSeries).Find(&events)

	res := trainingScheduleRes{Weekdays: []int{}, Title: "Training"}
	for _, e := range events {
		d, err := time.Parse(dateLayout, e.Date)
		if err != nil {
			continue
		}
		res.Weekdays = append(res.Weekdays, isoWeekday(d))
		// Alle Serien teilen sich Zeiten und Ort — der erste Treffer genügt.
		if res.StartTime == "" {
			res.Title, res.StartTime, res.EndTime = e.Title, e.StartTime, e.EndTime
			res.Location, res.Notes, res.RecurrenceEnd = e.Location, e.Notes, e.RecurrenceEnd
		}
	}
	sort.Ints(res.Weekdays)
	c.JSON(http.StatusOK, res)
}

// PutTrainingSchedule ersetzt den Plan: fehlende Wochentage werden angelegt,
// bestehende aktualisiert, abgewählte gelöscht.
func (a *API) PutTrainingSchedule(c *gin.Context) {
	var req trainingScheduleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Angaben"})
		return
	}
	wanted := map[int]bool{}
	for _, w := range req.Weekdays {
		if w < 1 || w > 7 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Wochentag muss 1 (Mo) bis 7 (So) sein"})
			return
		}
		wanted[w] = true
	}
	if req.RecurrenceEnd != "" {
		if _, err := time.Parse(dateLayout, req.RecurrenceEnd); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Serien-Ende muss YYYY-MM-DD sein"})
			return
		}
	}
	title := req.Title
	if title == "" {
		title = "Training"
	}

	var existing []models.Event
	a.db.Where("series = ?", trainingSeries).Find(&existing)

	byWeekday := map[int]models.Event{}
	for _, e := range existing {
		d, err := time.Parse(dateLayout, e.Date)
		if err != nil {
			continue
		}
		byWeekday[isoWeekday(d)] = e
	}

	// Abgewählte Wochentage entfernen — samt Rückmeldungen und Notizen.
	for wd, e := range byWeekday {
		if wanted[wd] {
			continue
		}
		id := e.ID.String()
		a.db.Delete(&models.Event{}, "id = ?", id)
		a.db.Delete(&models.EventAttendance{}, "event_key = ? OR event_key LIKE ?", id, id+"\\_%")
		a.db.Delete(&models.EventNote{}, "event_key = ? OR event_key LIKE ?", id, id+"\\_%")
	}

	// Gewählte Wochentage anlegen oder aktualisieren. Das Startdatum bleibt
	// bestehen, solange es auf demselben Wochentag liegt — sonst verlöre die
	// Serie ihre bisherigen Rückmeldungen.
	for wd := 1; wd <= 7; wd++ {
		if !wanted[wd] {
			continue
		}
		e, ok := byWeekday[wd]
		if !ok {
			e = models.Event{Date: nextWeekday(time.Now(), wd).Format(dateLayout)}
		}
		e.Title, e.Type, e.Series = title, "training", trainingSeries
		e.StartTime, e.EndTime = req.StartTime, req.EndTime
		e.Location, e.Notes = req.Location, req.Notes
		e.Recurring, e.RecurrenceType, e.RecurrenceEnd = true, "weekly", req.RecurrenceEnd
		if ok {
			a.db.Save(&e)
		} else {
			a.db.Create(&e)
		}
	}

	a.GetTrainingSchedule(c)
}

type eventNoteReq struct {
	EventKey string `json:"eventKey" binding:"required,max=80"`
	Text     string `json:"text" binding:"max=2000"`
}

// PutEventNote hängt eine Notiz an ein einzelnes Vorkommen. Leerer Text löscht.
func (a *API) PutEventNote(c *gin.Context) {
	var req eventNoteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "eventKey angeben"})
		return
	}
	if req.Text == "" {
		a.db.Delete(&models.EventNote{}, "event_key = ?", req.EventKey)
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}
	note := models.EventNote{EventKey: req.EventKey, Text: req.Text, UpdatedAt: time.Now()}
	if err := a.db.Save(&note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Notiz konnte nicht gespeichert werden"})
		return
	}
	c.JSON(http.StatusOK, note)
}

// isoWeekday liefert 1 = Montag … 7 = Sonntag.
func isoWeekday(d time.Time) int {
	wd := int(d.Weekday())
	if wd == 0 {
		return 7
	}
	return wd
}

// nextWeekday — nächstes Datum ab (einschließlich) from mit dem Wochentag wd.
func nextWeekday(from time.Time, wd int) time.Time {
	d := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.Local)
	for i := 0; i < 7; i++ {
		if isoWeekday(d) == wd {
			return d
		}
		d = d.AddDate(0, 0, 1)
	}
	return d
}
