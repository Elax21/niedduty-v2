package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/alessandro/niedduty/internal/middleware"
	"github.com/alessandro/niedduty/internal/models"
	"github.com/alessandro/niedduty/internal/push"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ── Abo-Verwaltung ──────────────────────────────────────────────────

// GetPushKey liefert den öffentlichen VAPID-Schlüssel für den Browser.
func (a *API) GetPushKey(c *gin.Context) {
	keys, err := push.LoadKeys(a.db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Push nicht verfügbar"})
		return
	}
	var count int64
	a.db.Model(&models.PushSubscription{}).Where("user_id = ?", middleware.CurrentUser(c).ID).Count(&count)
	c.JSON(http.StatusOK, gin.H{"publicKey": keys.Public, "devices": count})
}

type subscribeReq struct {
	Endpoint string `json:"endpoint" binding:"required,max=600"`
	Keys     struct {
		P256dh string `json:"p256dh" binding:"required,max=200"`
		Auth   string `json:"auth" binding:"required,max=100"`
	} `json:"keys"`
}

// Subscribe speichert ein Gerät. Ein bereits bekanntes Endpoint wird dem
// aktuellen Konto zugeordnet (z.B. nach Kontowechsel auf demselben Handy).
func (a *API) Subscribe(c *gin.Context) {
	var req subscribeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültiges Push-Abo"})
		return
	}
	if !strings.HasPrefix(req.Endpoint, "https://") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültiges Push-Abo"})
		return
	}
	sub := models.PushSubscription{
		UserID:    middleware.CurrentUser(c).ID,
		Endpoint:  req.Endpoint,
		P256dh:    req.Keys.P256dh,
		Auth:      req.Keys.Auth,
		UserAgent: truncate(c.Request.UserAgent(), 200),
	}
	a.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "endpoint"}},
		DoUpdates: clause.AssignmentColumns([]string{"user_id", "p256dh", "auth", "user_agent", "failures"}),
	}).Create(&sub)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type unsubscribeReq struct {
	Endpoint string `json:"endpoint" binding:"required,max=600"`
}

// Unsubscribe entfernt ein Gerät wieder.
func (a *API) Unsubscribe(c *gin.Context) {
	var req unsubscribeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Endpoint fehlt"})
		return
	}
	a.db.Delete(&models.PushSubscription{}, "endpoint = ? AND user_id = ?", req.Endpoint, middleware.CurrentUser(c).ID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// TestPush schickt dem eigenen Konto eine Probenachricht.
func (a *API) TestPush(c *gin.Context) {
	user := middleware.CurrentUser(c)
	n := push.SendToUsers(a.db, []uuid.UUID{user.ID}, push.Payload{
		Title: "Niedduty",
		Body:  "Push funktioniert. 🟡⚫",
		URL:   "/",
		Tag:   "test",
	})
	c.JSON(http.StatusOK, gin.H{"sent": n})
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

// ── Erinnerungen ────────────────────────────────────────────────────

// Treffpunkt vor einem Spiel — fest, weil es in der Mannschaft so gilt.
const meetBeforeMatch = 90 * time.Minute

// defaultPushSetting gilt für alle, die nichts eingestellt haben.
func defaultPushSetting(userID uuid.UUID) models.PushSetting {
	return models.PushSetting{
		UserID:           userID,
		TrainingLeadMin:  60,
		MatchLeadMin:     180,
		MeetLeadMin:      30,
		VorschauSpiel:    1440,
		VorschauTraining: 300,
		Birthdays:        true,
	}
}

// StartReminderLoop prüft regelmäßig, welche Erinnerung für wen fällig ist.
// Kein externer Cron nötig; push_deliveries verhindert Doppler je Konto.
func StartReminderLoop(db *gorm.DB, every time.Duration) {
	go func() {
		for {
			runReminders(db, time.Now())
			runBirthdays(db, time.Now())
			runPollReminders(db, time.Now())
			time.Sleep(every)
		}
	}()
}

// settingsByUser lädt alle Konten mit ihren Vorlaufzeiten (Vorgaben, wo nichts
// gespeichert ist).
func settingsByUser(db *gorm.DB) map[uuid.UUID]models.PushSetting {
	var users []models.User
	db.Find(&users)
	var stored []models.PushSetting
	db.Find(&stored)
	byUser := map[uuid.UUID]models.PushSetting{}
	for _, s := range stored {
		byUser[s.UserID] = s
	}
	out := map[uuid.UUID]models.PushSetting{}
	for _, u := range users {
		if s, ok := byUser[u.ID]; ok {
			out[u.ID] = s
		} else {
			out[u.ID] = defaultPushSetting(u.ID)
		}
	}
	return out
}

func runReminders(db *gorm.DB, now time.Time) {
	var events []models.Event
	db.Find(&events)
	occs := expandEvents(events, now.Truncate(24*time.Hour), now.Add(72*time.Hour))
	if len(occs) == 0 {
		return
	}
	settings := settingsByUser(db)

	for _, o := range occs {
		start, ok := occurrenceStart(o)
		if !ok || !start.After(now) {
			continue
		}
		open := openAnswers(db, o.EventKey)

		for userID, s := range settings {
			// Kurz vorher — an alle, die ihr Gerät angemeldet haben.
			if lead := kurzvorherLead(o.Type, s); lead > 0 && start.Sub(now) <= lead {
				if claimDelivery(db, o.EventKey, "kurzvorher", userID) {
					sendKurzvorher(db, userID, o, start)
				}
				continue // Treffpunkt und Vorschau wären jetzt überholt
			}

			// Treffpunkt — nur bei Spielen, 1:30 vor Anpfiff.
			if o.Type == "spiel" && s.MeetLeadMin > 0 {
				meet := start.Add(-meetBeforeMatch)
				if meet.After(now) && meet.Sub(now) <= time.Duration(s.MeetLeadMin)*time.Minute {
					if claimDelivery(db, o.EventKey, "treffpunkt", userID) {
						sendTreffpunkt(db, userID, o, meet, start)
					}
					continue
				}
			}

			// Vorschau — nur an die, von denen noch keine Rückmeldung da ist.
			if !open[userID] {
				continue
			}
			if lead := vorschauLead(o.Type, s); lead > 0 && start.Sub(now) <= lead {
				if claimDelivery(db, o.EventKey, "vorschau", userID) {
					sendVorschau(db, userID, o, start)
				}
			}
		}
	}
}

func kurzvorherLead(typ string, s models.PushSetting) time.Duration {
	if typ == "spiel" {
		return time.Duration(s.MatchLeadMin) * time.Minute
	}
	return time.Duration(s.TrainingLeadMin) * time.Minute
}

func vorschauLead(typ string, s models.PushSetting) time.Duration {
	if typ == "spiel" {
		return time.Duration(s.VorschauSpiel) * time.Minute
	}
	return time.Duration(s.VorschauTraining) * time.Minute
}

// occurrenceStart baut den Startzeitpunkt aus Datum + Uhrzeit des Vorkommens.
// Ohne Uhrzeit gilt 18:00 als Ansatz, damit auch solche Termine erinnert werden.
func occurrenceStart(o models.Occurrence) (time.Time, bool) {
	d, err := time.ParseInLocation(dateLayout, o.OccDate, time.Local)
	if err != nil {
		return time.Time{}, false
	}
	hh, mm := 18, 0
	if _, err := fmt.Sscanf(o.StartTime, "%d:%d", &hh, &mm); err != nil && o.StartTime != "" {
		return time.Time{}, false
	}
	return time.Date(d.Year(), d.Month(), d.Day(), hh, mm, 0, 0, time.Local), true
}

// claimDelivery trägt die Erinnerung ein und meldet, ob *wir* sie verschicken
// dürfen (false = war für dieses Konto schon raus).
func claimDelivery(db *gorm.DB, eventKey, kind string, userID uuid.UUID) bool {
	res := db.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&models.PushDelivery{EventKey: eventKey, Kind: kind, UserID: userID, SentAt: time.Now()})
	return res.Error == nil && res.RowsAffected == 1
}

func sendVorschau(db *gorm.DB, userID uuid.UUID, o models.Occurrence, start time.Time) {
	push.SendToUsers(db, []uuid.UUID{userID}, push.Payload{
		Title: eventHeadline(o),
		Body:  fmt.Sprintf("%s%s — sag zu oder ab.", weekdayTime(start), locationSuffix(o)),
		URL:   "/termine",
		Tag:   "vorschau-" + o.EventKey,
	})
}

func sendTreffpunkt(db *gorm.DB, userID uuid.UUID, o models.Occurrence, meet, start time.Time) {
	push.SendToUsers(db, []uuid.UUID{userID}, push.Payload{
		Title: eventHeadline(o),
		Body: fmt.Sprintf("Treffpunkt %s Uhr (Anpfiff %s)%s.",
			meet.Format("15:04"), start.Format("15:04"), locationSuffix(o)),
		URL: "/termine",
		Tag: "treffpunkt-" + o.EventKey,
	})
}

func sendKurzvorher(db *gorm.DB, userID uuid.UUID, o models.Occurrence, start time.Time) {
	body := fmt.Sprintf("Beginn %s Uhr%s.", start.Format("15:04"), locationSuffix(o))
	var att int64
	db.Model(&models.EventAttendance{}).Where("event_key = ? AND status = ?", o.EventKey, "attending").Count(&att)
	if att > 0 {
		body += fmt.Sprintf(" %d Zusagen.", att)
	}
	push.SendToUsers(db, []uuid.UUID{userID}, push.Payload{
		Title: eventHeadline(o),
		Body:  body,
		URL:   "/termine",
		Tag:   "kurzvorher-" + o.EventKey,
	})
}

// runBirthdays gratuliert einmal am Tag — der Schlüssel enthält das Datum,
// damit es im nächsten Jahr wieder losgeht.
func runBirthdays(db *gorm.DB, now time.Time) {
	if now.Hour() < 8 { // nicht mitten in der Nacht
		return
	}
	today := now.Format("01-02")
	var players []models.Player
	db.Where("birthday <> ''").Find(&players)

	settings := settingsByUser(db)
	for _, p := range players {
		if len(p.Birthday) < 10 || p.Birthday[5:10] != today {
			continue
		}
		key := "bday_" + p.ID.String() + "_" + now.Format("2006")
		for userID, s := range settings {
			if !s.Birthdays {
				continue
			}
			if !claimDelivery(db, key, "geburtstag", userID) {
				continue
			}
			push.SendToUsers(db, []uuid.UUID{userID}, push.Payload{
				Title: "🎂 " + p.Name,
				Body:  birthdayBody(p, now),
				URL:   "/kader",
				Tag:   key,
			})
		}
		log.Printf("Push: Geburtstag %s", p.Name)
	}
}

func birthdayBody(p models.Player, now time.Time) string {
	if year, err := strconv.Atoi(p.Birthday[:4]); err == nil && year > 1900 {
		age := now.Year() - year
		return fmt.Sprintf("hat heute Geburtstag — wird %d. Gratulieren nicht vergessen.", age)
	}
	return "hat heute Geburtstag. Gratulieren nicht vergessen."
}

// runPollReminders stupst alle an, die vor Ablauf einer Abstimmung noch nicht
// abgestimmt haben — einmal je Abstimmung und Konto.
func runPollReminders(db *gorm.DB, now time.Time) {
	var polls []models.Poll
	db.Where("closed_at IS NULL AND ends_at IS NOT NULL AND ends_at > ?", now).Find(&polls)

	for _, p := range polls {
		if p.EndsAt.Sub(now) > pollReminderLead {
			continue
		}
		var votes []models.PollVote
		db.Where("poll_id = ?", p.ID).Find(&votes)
		voted := map[uuid.UUID]bool{}
		for _, v := range votes {
			voted[v.UserID] = true
		}

		var users []models.User
		db.Find(&users)
		key := "poll_" + p.ID.String()
		for _, u := range users {
			if voted[u.ID] || !claimDelivery(db, key, "abstimmung", u.ID) {
				continue
			}
			push.SendToUsers(db, []uuid.UUID{u.ID}, push.Payload{
				Title: "🗳 Abstimmung läuft ab",
				Body:  p.Question + " — endet " + p.EndsAt.Format("02.01. um 15:04") + " Uhr.",
				URL:   "/abstimmungen",
				Tag:   key,
			})
		}
	}
}

// openAnswers — Konten, von denen für dieses Vorkommen noch nichts kam.
func openAnswers(db *gorm.DB, eventKey string) map[uuid.UUID]bool {
	var users []models.User
	db.Where("player_id IS NOT NULL").Find(&users)
	var answered []models.EventAttendance
	db.Where("event_key = ?", eventKey).Find(&answered)

	done := map[uuid.UUID]bool{}
	for _, a := range answered {
		done[a.PlayerID] = true
	}
	out := map[uuid.UUID]bool{}
	for _, u := range users {
		out[u.ID] = !done[*u.PlayerID]
	}
	return out
}

func eventHeadline(o models.Occurrence) string {
	switch o.Type {
	case "spiel":
		return "⚽ " + o.Title
	case "training":
		return "🏃 " + o.Title
	case "mannschaftsabend":
		return "🍻 " + o.Title
	default:
		return o.Title
	}
}

var weekdays = [...]string{"Sonntag", "Montag", "Dienstag", "Mittwoch", "Donnerstag", "Freitag", "Samstag"}

func weekdayTime(t time.Time) string {
	return fmt.Sprintf("%s, %s Uhr", weekdays[int(t.Weekday())], t.Format("15:04"))
}

func locationSuffix(o models.Occurrence) string {
	if o.Location == "" {
		return ""
	}
	return " · " + o.Location
}

// ── Persönliche Einstellungen ───────────────────────────────────────

// GetPushSettings liefert die eigenen Vorlaufzeiten (oder die Vorgaben).
func (a *API) GetPushSettings(c *gin.Context) {
	user := middleware.CurrentUser(c)
	var s models.PushSetting
	if err := a.db.First(&s, "user_id = ?", user.ID).Error; err != nil {
		s = defaultPushSetting(user.ID)
	}
	c.JSON(http.StatusOK, gin.H{"settings": s, "meetBeforeMatchMin": int(meetBeforeMatch / time.Minute)})
}

type pushSettingsReq struct {
	TrainingLeadMin  int  `json:"trainingLeadMin" binding:"min=0,max=1440"`
	MatchLeadMin     int  `json:"matchLeadMin" binding:"min=0,max=2880"`
	MeetLeadMin      int  `json:"meetLeadMin" binding:"min=0,max=1440"`
	VorschauSpiel    int  `json:"vorschauSpiel" binding:"min=0,max=10080"`
	VorschauTraining int  `json:"vorschauTraining" binding:"min=0,max=10080"`
	Birthdays        bool `json:"birthdays"`
}

// PutPushSettings speichert die eigenen Vorlaufzeiten.
func (a *API) PutPushSettings(c *gin.Context) {
	var req pushSettingsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Vorlaufzeit"})
		return
	}
	user := middleware.CurrentUser(c)
	s := models.PushSetting{
		UserID:           user.ID,
		TrainingLeadMin:  req.TrainingLeadMin,
		MatchLeadMin:     req.MatchLeadMin,
		MeetLeadMin:      req.MeetLeadMin,
		VorschauSpiel:    req.VorschauSpiel,
		VorschauTraining: req.VorschauTraining,
		Birthdays:        req.Birthdays,
	}
	if err := a.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		UpdateAll: true,
	}).Create(&s).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Konnte nicht gespeichert werden"})
		return
	}
	c.JSON(http.StatusOK, s)
}
