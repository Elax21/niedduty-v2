package api

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/alessandro/niedduty/internal/fussball"
	"github.com/alessandro/niedduty/internal/models"
	"github.com/gin-gonic/gin"
)

// ── Gegner-Scouting ─────────────────────────────────────────────────
//
// Ziel: vor dem nächsten Pflichtspiel auf einen Blick wissen, wie der Gegner
// dasteht — ohne fussball.de aufzumachen. Datenquellen sind die bereits
// gespeicherte Ligatabelle und der öffentliche Spielplan des Gegners.

// FormEntry — ein gewertetes Spiel für die Formkurve.
type FormEntry struct {
	Result   string `json:"result"` // S | U | N
	Opponent string `json:"opponent"`
	Score    string `json:"score"` // "3:1" aus Sicht der betrachteten Mannschaft
	Date     string `json:"date"`  // YYYY-MM-DD
	Home     bool   `json:"home"`
}

// Meeting — ein früheres Aufeinandertreffen mit dem Gegner.
type Meeting struct {
	Date   string `json:"date"`
	Score  string `json:"score"` // aus unserer Sicht
	Result string `json:"result"`
	Home   bool   `json:"home"`
	Note   string `json:"note"`
}

// OpponentInfo — alles, was wir über den nächsten Gegner sagen können.
type OpponentInfo struct {
	Name         string      `json:"name"`
	LogoURL      string      `json:"logoUrl"`
	TeamID       string      `json:"teamId"`
	Position     int         `json:"position"`
	Played       int         `json:"played"`
	Won          int         `json:"won"`
	Drawn        int         `json:"drawn"`
	Lost         int         `json:"lost"`
	GoalsFor     int         `json:"goalsFor"`
	GoalsAgainst int         `json:"goalsAgainst"`
	Points       int         `json:"points"`
	InTable      bool        `json:"inTable"`
	Form         []FormEntry `json:"form"`
	Meetings     []Meeting   `json:"meetings"`
	Summary      string      `json:"summary"`
}

// ScoutingResponse — nächstes Pflichtspiel plus Gegner-Steckbrief.
type ScoutingResponse struct {
	Match    *fussball.Match `json:"match"`
	Opponent *OpponentInfo   `json:"opponent"`
	OwnForm  []FormEntry     `json:"ownForm"`
	AtHome   bool            `json:"atHome"`
}

type scoutCache struct {
	mu   sync.Mutex
	data *ScoutingResponse
	key  string
	at   time.Time
}

var scouting = &scoutCache{}

const scoutingTTL = 30 * time.Minute

// GetScouting liefert den Steckbrief zum nächsten Gegner.
func (a *API) GetScouting(c *gin.Context) {
	var club models.Club
	if err := a.db.First(&club, "id = 1").Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Verein nicht gefunden"})
		return
	}
	matches, err := a.loadMatches(club)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "fussball.de nicht erreichbar"})
		return
	}
	next := nextMatch(matches)
	if next == nil {
		c.JSON(http.StatusOK, &ScoutingResponse{})
		return
	}

	opp, atHome := opponentOf(*next)
	key := next.ID + "|" + opp.TeamID

	scouting.mu.Lock()
	if scouting.data != nil && scouting.key == key && time.Since(scouting.at) < scoutingTTL {
		cached := scouting.data
		scouting.mu.Unlock()
		c.JSON(http.StatusOK, cached)
		return
	}
	scouting.mu.Unlock()

	resp := &ScoutingResponse{Match: next, AtHome: atHome}
	resp.Opponent = a.buildOpponent(opp, club.Name)

	// Eigene Formkurve — dieselbe Logik, andere Mannschaft.
	a.db.First(&club, "id = 1") // FussballTeamId ggf. inzwischen gesetzt
	if club.FussballTeamId != "" {
		if games, err := fussball.FetchPrevGames(club.FussballTeamId); err == nil {
			resp.OwnForm = formFrom(games, club.Name, 5)
			resp.Opponent.Meetings = meetingsWith(games, club.Name, opp.Name)
		}
	}
	resp.Opponent.Summary = summarize(resp.Opponent, atHome)

	scouting.mu.Lock()
	scouting.data, scouting.key, scouting.at = resp, key, time.Now()
	scouting.mu.Unlock()
	c.JSON(http.StatusOK, resp)
}

// nextMatch nimmt das nächste angesetzte Spiel (Liste ist chronologisch).
func nextMatch(m *fussball.Matches) *fussball.Match {
	for i := range m.Next {
		if !m.Next[i].Played {
			return &m.Next[i]
		}
	}
	if len(m.Next) > 0 {
		return &m.Next[0]
	}
	return nil
}

// opponentOf liefert die gegnerische Mannschaft und ob wir zuhause spielen.
func opponentOf(m fussball.Match) (fussball.MatchTeam, bool) {
	if m.Home.IsOwn {
		return m.Guest, true
	}
	return m.Home, false
}

// buildOpponent reichert den Gegner mit Tabellenplatz und Formkurve an.
func (a *API) buildOpponent(t fussball.MatchTeam, ownName string) *OpponentInfo {
	info := &OpponentInfo{
		Name: t.Name, LogoURL: t.LogoURL, TeamID: t.TeamID,
		Form: []FormEntry{}, Meetings: []Meeting{},
	}

	var table []models.LeagueEntry
	a.db.Order("points desc, (goals_for - goals_against) desc, goals_for desc").Find(&table)
	for i, e := range table {
		if !sameTeam(e.TeamName, t.Name) {
			continue
		}
		// Vor dem ersten Spieltag stehen alle bei 0 — dann ist der „Platz"
		// bedeutungslos und wird nicht ausgewiesen.
		info.Position, info.InTable = i+1, e.Played > 0
		info.Played, info.Won, info.Drawn, info.Lost = e.Played, e.Won, e.Drawn, e.Lost
		info.GoalsFor, info.GoalsAgainst, info.Points = e.GoalsFor, e.GoalsAgainst, e.Points
		break
	}

	if t.TeamID != "" {
		if games, err := fussball.FetchPrevGames(t.TeamID); err == nil {
			info.Form = formFrom(games, t.Name, 5)
		}
	}
	return info
}

// formFrom wertet die letzten Spiele aus Sicht von teamName aus (neueste zuerst).
func formFrom(games []fussball.ClassicMatch, teamName string, limit int) []FormEntry {
	out := []FormEntry{}
	for _, g := range games {
		if g.HomeGoals == nil || g.GuestGoals == nil {
			continue // abgesetzt / gewertet ohne Ergebnis
		}
		home := sameTeam(g.Home, teamName)
		if !home && !sameTeam(g.Guest, teamName) {
			continue
		}
		own, other := *g.HomeGoals, *g.GuestGoals
		opp := g.Guest
		if !home {
			own, other, opp = *g.GuestGoals, *g.HomeGoals, g.Home
		}
		out = append(out, FormEntry{
			Result:   resultLetter(own, other),
			Opponent: opp,
			Score:    fmt.Sprintf("%d:%d", own, other),
			Date:     g.ISODate,
			Home:     home,
		})
		if len(out) >= limit {
			break
		}
	}
	return out
}

// meetingsWith sucht frühere Partien gegen genau diesen Gegner (aus unserer Sicht).
func meetingsWith(games []fussball.ClassicMatch, ownName, oppName string) []Meeting {
	out := []Meeting{}
	for _, g := range games {
		home := sameTeam(g.Home, ownName)
		if !home && !sameTeam(g.Guest, ownName) {
			continue
		}
		other := g.Guest
		if !home {
			other = g.Home
		}
		if !sameTeam(other, oppName) {
			continue
		}
		m := Meeting{Date: g.ISODate, Home: home, Note: g.Note}
		if g.HomeGoals != nil && g.GuestGoals != nil {
			own, opp := *g.HomeGoals, *g.GuestGoals
			if !home {
				own, opp = *g.GuestGoals, *g.HomeGoals
			}
			m.Score, m.Result = fmt.Sprintf("%d:%d", own, opp), resultLetter(own, opp)
		}
		out = append(out, m)
		if len(out) >= 3 {
			break
		}
	}
	return out
}

func resultLetter(own, other int) string {
	switch {
	case own > other:
		return "S"
	case own < other:
		return "N"
	default:
		return "U"
	}
}

// sameTeam vergleicht Mannschaftsnamen tolerant — fussball.de schreibt sie je
// nach Quelle mit oder ohne Zusätze („e.V.", „I", „II").
func sameTeam(a, b string) bool {
	na, nb := teamKey(a), teamKey(b)
	if na == "" || nb == "" {
		return false
	}
	return na == nb || strings.HasPrefix(na, nb) || strings.HasPrefix(nb, na)
}

var teamNoise = strings.NewReplacer(
	"e.v.", "", "e. v.", "", ".", "", ",", "", "-", " ", "  ", " ",
)

func teamKey(s string) string {
	return strings.TrimSpace(teamNoise.Replace(strings.ToLower(strings.TrimSpace(s))))
}

// summarize formuliert einen Satz Klartext — mehr liest am Handy niemand.
func summarize(o *OpponentInfo, weAreHome bool) string {
	parts := []string{}
	if o.InTable && o.Played > 0 {
		parts = append(parts, fmt.Sprintf("%d. Platz mit %d Punkten", o.Position, o.Points))
		parts = append(parts, fmt.Sprintf("%d:%d Tore", o.GoalsFor, o.GoalsAgainst))
	}
	if s := streakText(o.Form); s != "" {
		parts = append(parts, s)
	}
	if len(parts) == 0 {
		if weAreHome {
			return "Noch keine Daten zum Gegner — wir spielen zuhause."
		}
		return "Noch keine Daten zum Gegner — wir spielen auswärts."
	}
	s := strings.Join(parts, ", ")
	return strings.ToUpper(s[:1]) + s[1:] + "."
}

// streakText beschreibt die aktuelle Serie ("3 Siege in Folge", "zuletzt 2S 1U 2N").
func streakText(form []FormEntry) string {
	if len(form) == 0 {
		return ""
	}
	first := form[0].Result
	run := 0
	for _, f := range form {
		if f.Result != first {
			break
		}
		run++
	}
	if run >= 2 {
		switch first {
		case "S":
			return fmt.Sprintf("%d Siege in Folge", run)
		case "N":
			return fmt.Sprintf("%d Niederlagen in Folge", run)
		default:
			return fmt.Sprintf("%d Unentschieden in Folge", run)
		}
	}
	var s, u, n int
	for _, f := range form {
		switch f.Result {
		case "S":
			s++
		case "U":
			u++
		case "N":
			n++
		}
	}
	return fmt.Sprintf("zuletzt %dS %dU %dN", s, u, n)
}

// ── Kaderstatistik (Einsätze, Einsatzminuten, Tore) ─────────────────

type squadCache struct {
	mu   sync.Mutex
	data []fussball.PlayerStat
	key  string
	at   time.Time
}

var squadStats = &squadCache{}

const squadTTL = 6 * time.Hour

// GetSquadStats liefert die fussball.de-Kaderstatistik der Mannschaft.
// ?saison=2526 überschreibt die laufende Saison.
func (a *API) GetSquadStats(c *gin.Context) {
	var club models.Club
	if err := a.db.First(&club, "id = 1").Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Verein nicht gefunden"})
		return
	}
	if club.FussballTeamId == "" {
		// Beim ersten Aufruf ist die Mannschafts-ID evtl. noch nicht erkannt.
		if _, err := a.loadMatches(club); err == nil {
			a.db.First(&club, "id = 1")
		}
	}
	if club.FussballTeamId == "" {
		c.JSON(http.StatusOK, gin.H{"season": "", "players": []fussball.PlayerStat{}})
		return
	}

	season := c.Query("saison")
	if !seasonRe.MatchString(season) {
		season = fussball.CurrentSeason(time.Now())
	}
	key := club.FussballTeamId + "|" + season

	squadStats.mu.Lock()
	fresh := squadStats.data != nil && squadStats.key == key && time.Since(squadStats.at) < squadTTL
	cached := squadStats.data
	squadStats.mu.Unlock()
	if fresh {
		c.JSON(http.StatusOK, gin.H{"season": season, "players": cached})
		return
	}

	stats, err := fussball.FetchSquadStats(club.FussballTeamId, season)
	if err != nil {
		if cached != nil {
			c.JSON(http.StatusOK, gin.H{"season": season, "players": cached})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": "fussball.de nicht erreichbar"})
		return
	}
	squadStats.mu.Lock()
	squadStats.data, squadStats.key, squadStats.at = stats, key, time.Now()
	squadStats.mu.Unlock()
	c.JSON(http.StatusOK, gin.H{"season": season, "players": stats})
}
