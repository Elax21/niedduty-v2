package api

import (
	"log"
	"strings"
	"time"

	"github.com/alessandro/niedduty/internal/fussball"
	"github.com/alessandro/niedduty/internal/models"
)

// squadImportKey ist der Merker in `settings`, der den einmaligen Kader-Import
// von fussball.de kennzeichnet — genau wie bei den Seed-Migrationen.
const squadImportKey = "migration.squadImport"

// ImportSquadOnce zieht den kompletten Kader einmalig von fussball.de und legt
// jeden noch fehlenden Spieler an. Hintergrund: viele Spieler haben sich nie
// selbst registriert, sollen aber Strafen bekommen können — dafür muss der
// ganze Kader als Spieler in der DB stehen.
//
// Läuft dank Merker genau einmal. Ist die Mannschafts-ID beim Start noch nicht
// bekannt (der Sync-Loop erkennt sie erst kurz danach) oder fussball.de gerade
// nicht erreichbar, wird der Merker NICHT gesetzt — der Import versucht es beim
// nächsten Start erneut. Namen sind auf fussball.de font-verschleiert und werden
// von FetchSquadStats bereits dekodiert.
func (a *API) ImportSquadOnce() {
	var flag models.Setting
	if err := a.db.First(&flag, "key = ?", squadImportKey).Error; err == nil {
		return // schon erledigt
	}

	var club models.Club
	if err := a.db.First(&club, "id = 1").Error; err != nil {
		log.Printf("Kader-Import: Verein nicht gefunden, überspringe")
		return
	}

	// Mannschafts-ID nötig; falls noch leer, über den Spielplan erkennen lassen.
	if club.FussballTeamId == "" {
		if _, err := a.loadMatches(club); err == nil {
			a.db.First(&club, "id = 1")
		}
	}
	if club.FussballTeamId == "" {
		log.Printf("Kader-Import: Mannschafts-ID noch unbekannt, versuche es beim nächsten Start erneut")
		return
	}

	season := fussball.CurrentSeason(time.Now())
	stats, err := fussball.FetchSquadStats(club.FussballTeamId, season)
	if err != nil {
		log.Printf("Kader-Import: fussball.de nicht erreichbar (%v), versuche es beim nächsten Start erneut", err)
		return
	}

	// Vorhandene Spieler nach kleingeschriebenem Namen indizieren, damit der
	// Import wiederholbar bleibt und keine Dubletten entstehen.
	var existing []models.Player
	a.db.Find(&existing)
	known := make(map[string]bool, len(existing))
	for _, p := range existing {
		known[strings.ToLower(strings.TrimSpace(p.Name))] = true
	}

	added := 0
	for _, s := range stats {
		name := strings.TrimSpace(s.Name)
		if name == "" || known[strings.ToLower(name)] {
			continue
		}
		if err := a.db.Create(&models.Player{Name: name, Position: "MF", Status: "fit"}).Error; err != nil {
			log.Printf("Kader-Import: Spieler %q konnte nicht angelegt werden: %v", name, err)
			continue
		}
		known[strings.ToLower(name)] = true
		added++
	}

	a.db.Create(&models.Setting{Key: squadImportKey, Value: "done"})
	log.Printf("Kader-Import: %d von %d fussball.de-Spielern neu angelegt (Saison %s)", added, len(stats), season)
}
