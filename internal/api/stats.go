package api

import (
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/alessandro/niedduty/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Auswertungen für die Diagramme auf der Beteiligungs-Seite. Alles wird aus
// den vorhandenen Tabellen gerechnet (Termin-Serien + event_attendances) —
// es gibt keinen zweiten Datenbestand, der auseinanderlaufen könnte.

// MonthStat — ein Monat im Beteiligungsverlauf.
type MonthStat struct {
	Month string `json:"month"` // "2026-08"
	Label string `json:"label"` // "Aug 26"
	// Trainings/Spiele im Monat (nur Vorkommen bis heute).
	Trainings int `json:"trainings"`
	Matches   int `json:"matches"`
	// Summierte Rückmeldungen über alle Spieler.
	Attending int `json:"attending"`
	Declined  int `json:"declined"`
	NoAnswer  int `json:"noAnswer"`
	// QuotePct — Zusagen im Verhältnis zu allen möglichen Teilnahmen.
	QuotePct int `json:"quotePct"`
	// AvgAttending — Schnitt an Zusagen je Einheit.
	AvgAttending int `json:"avgAttending"`
}

// WeekdayStat — an welchem Wochentag kommen die meisten?
type WeekdayStat struct {
	Weekday  int    `json:"weekday"` // 1 = Montag … 7 = Sonntag
	Label    string `json:"label"`
	Count    int    `json:"count"` // Einheiten an diesem Tag
	QuotePct int    `json:"quotePct"`
}

// KasseMonth — Strafen je Monat, damit man den Verlauf der Kasse sieht.
type KasseMonth struct {
	Month string `json:"month"`
	Label string `json:"label"`
	Open  int    `json:"open"` // Cent
	Paid  int    `json:"paid"`
	Count int    `json:"count"`
}

var monthShort = [...]string{"", "Jan", "Feb", "Mär", "Apr", "Mai", "Jun", "Jul", "Aug", "Sep", "Okt", "Nov", "Dez"}

func monthLabel(t time.Time) string {
	return monthShort[int(t.Month())] + " " + t.Format("06")
}

// StatsOverview — alle Diagrammdaten in einem Abruf. `months` = wie viele
// Monate zurück (1–24, Vorgabe 6).
func (a *API) StatsOverview(c *gin.Context) {
	months := clampInt(queryInt(c, "months", 6), 1, 24)

	now := time.Now()
	today, _ := time.Parse(dateLayout, now.Format(dateLayout))
	first := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local).AddDate(0, -(months - 1), 0)

	var events []models.Event
	a.db.Find(&events)
	occs := expandEvents(events, first, today)

	var squadSize int64
	a.db.Model(&models.Player{}).Count(&squadSize)

	keys := make([]string, 0, len(occs))
	for _, o := range occs {
		keys = append(keys, o.EventKey)
	}
	var atts []models.EventAttendance
	if len(keys) > 0 {
		a.db.Where("event_key IN ?", keys).Find(&atts)
	}
	byKey := map[string]map[string]int{}
	for _, at := range atts {
		if byKey[at.EventKey] == nil {
			byKey[at.EventKey] = map[string]int{}
		}
		byKey[at.EventKey][at.Status]++
	}

	// Monate vorbelegen, damit auch leere Monate im Diagramm auftauchen.
	byMonth := map[string]*MonthStat{}
	order := []string{}
	for i := 0; i < months; i++ {
		m := first.AddDate(0, i, 0)
		key := m.Format("2006-01")
		byMonth[key] = &MonthStat{Month: key, Label: monthLabel(m)}
		order = append(order, key)
	}

	byWeekday := map[int]*WeekdayStat{}
	for wd := 1; wd <= 7; wd++ {
		byWeekday[wd] = &WeekdayStat{Weekday: wd, Label: weekdayShort[wd]}
	}
	wdAttending := map[int]int{}
	wdPossible := map[int]int{}

	for _, o := range occs {
		if o.Type != "training" && o.Type != "spiel" {
			continue
		}
		d, err := time.ParseInLocation(dateLayout, o.OccDate, time.Local)
		if err != nil {
			continue
		}
		m := byMonth[o.OccDate[:7]]
		if m == nil {
			continue
		}
		counts := byKey[o.EventKey]
		att, dec := counts["attending"], counts["declined"]

		if o.Type == "training" {
			m.Trainings++
		} else {
			m.Matches++
		}
		m.Attending += att
		m.Declined += dec
		if rest := int(squadSize) - att - dec; rest > 0 {
			m.NoAnswer += rest
		}

		wd := isoWeekday(d)
		byWeekday[wd].Count++
		wdAttending[wd] += att
		wdPossible[wd] += int(squadSize)
	}

	monthList := make([]MonthStat, 0, months)
	for _, key := range order {
		m := byMonth[key]
		units := m.Trainings + m.Matches
		if units > 0 {
			m.AvgAttending = m.Attending / units
			if possible := units * int(squadSize); possible > 0 {
				m.QuotePct = m.Attending * 100 / possible
			}
		}
		monthList = append(monthList, *m)
	}

	weekdayList := make([]WeekdayStat, 0, 7)
	for wd := 1; wd <= 7; wd++ {
		w := byWeekday[wd]
		if w.Count == 0 {
			continue
		}
		if wdPossible[wd] > 0 {
			w.QuotePct = wdAttending[wd] * 100 / wdPossible[wd]
		}
		weekdayList = append(weekdayList, *w)
	}

	c.JSON(http.StatusOK, gin.H{
		"months":     monthList,
		"weekdays":   weekdayList,
		"kasse":      a.kasseByMonth(first, order),
		"squadSize":  squadSize,
		"topPlayers": a.topAttenders(occs, atts),
	})
}

var weekdayShort = map[int]string{1: "Mo", 2: "Di", 3: "Mi", 4: "Do", 5: "Fr", 6: "Sa", 7: "So"}

// kasseByMonth — offene und bezahlte Strafen je Monat der Zuweisung.
func (a *API) kasseByMonth(first time.Time, order []string) []KasseMonth {
	var rows []models.PlayerPenalty
	a.db.Where("created_at >= ?", first).Find(&rows)

	byMonth := map[string]*KasseMonth{}
	for _, key := range order {
		t, _ := time.Parse("2006-01", key)
		byMonth[key] = &KasseMonth{Month: key, Label: monthLabel(t)}
	}
	for _, r := range rows {
		m := byMonth[r.CreatedAt.Format("2006-01")]
		if m == nil {
			continue
		}
		m.Count++
		if r.Paid {
			m.Paid += r.Amount
		} else {
			m.Open += r.Amount
		}
	}
	out := make([]KasseMonth, 0, len(order))
	for _, key := range order {
		out = append(out, *byMonth[key])
	}
	return out
}

// topAttenders — die fleißigsten fünf über den ganzen Zeitraum.
type attenderStat struct {
	Name     string `json:"name"`
	Attended int    `json:"attended"`
	Total    int    `json:"total"`
	QuotePct int    `json:"quotePct"`
}

func (a *API) topAttenders(occs []models.Occurrence, atts []models.EventAttendance) []attenderStat {
	relevant := map[string]bool{}
	for _, o := range occs {
		if o.Type == "training" || o.Type == "spiel" {
			relevant[o.EventKey] = true
		}
	}
	total := len(relevant)

	count := map[uuid.UUID]int{}
	for _, at := range atts {
		if at.Status == "attending" && relevant[at.EventKey] {
			count[at.PlayerID]++
		}
	}
	var players []models.Player
	a.db.Find(&players)

	out := make([]attenderStat, 0, len(players))
	for _, p := range players {
		s := attenderStat{Name: p.Name, Attended: count[p.ID], Total: total}
		if total > 0 {
			s.QuotePct = s.Attended * 100 / total
		}
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Attended > out[j].Attended })
	if len(out) > 5 {
		out = out[:5]
	}
	return out
}

func queryInt(c *gin.Context, key string, fallback int) int {
	var v int
	if _, err := fmt.Sscanf(c.Query(key), "%d", &v); err != nil {
		return fallback
	}
	return v
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
