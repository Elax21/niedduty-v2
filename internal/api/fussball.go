package api

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/alessandro/niedduty/internal/fussball"
	"github.com/alessandro/niedduty/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ── Tabellen-Sync ───────────────────────────────────────────────────

// SyncTableFromFussball holt die Tabelle vom hinterlegten fussball.de-Widget
// und ersetzt die gespeicherte Ligatabelle. Läuft on-demand (Admin-Button)
// und periodisch aus dem Hintergrund (siehe cmd/server).
func SyncTableFromFussball(db *gorm.DB) (int, error) {
	var club models.Club
	if err := db.First(&club, "id = 1").Error; err != nil {
		return 0, err
	}
	if club.FussballTableId == "" {
		return 0, nil
	}
	entries, err := fussball.FetchTable(club.FussballTableId, club.Name)
	if err != nil {
		return 0, err
	}
	if len(entries) == 0 {
		return 0, nil
	}
	rows := make([]models.LeagueEntry, 0, len(entries))
	for _, e := range entries {
		rows = append(rows, models.LeagueEntry{
			TeamName: e.TeamName, IsOwn: e.IsOwn, Played: e.Played,
			Won: e.Won, Drawn: e.Drawn, Lost: e.Lost,
			GoalsFor: e.GoalsFor, GoalsAgainst: e.GoalsAgainst, Points: e.Points,
		})
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM league_entries").Error; err != nil {
			return err
		}
		return tx.Create(&rows).Error
	})
	if err != nil {
		return 0, err
	}
	return len(rows), nil
}

// SyncTable — Admin-Endpoint „Jetzt von fussball.de holen".
func (a *API) SyncTable(c *gin.Context) {
	n, err := SyncTableFromFussball(a.db)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "fussball.de nicht erreichbar oder Widget-ID fehlt"})
		return
	}
	if n == 0 {
		c.JSON(http.StatusOK, gin.H{"synced": 0, "note": "Keine Widget-ID hinterlegt oder Tabelle leer"})
		return
	}
	a.GetTable(c)
}

// StartTableSyncLoop synchronisiert die Tabelle beim Start und danach
// periodisch (kein externer Cron nötig).
func StartTableSyncLoop(db *gorm.DB, every time.Duration) {
	go func() {
		for {
			if n, err := SyncTableFromFussball(db); err != nil {
				log.Printf("Tabellen-Sync fehlgeschlagen: %v", err)
			} else if n > 0 {
				log.Printf("Tabellen-Sync: %d Teams von fussball.de übernommen", n)
			}
			time.Sleep(every)
		}
	}()
}

// ── Spiele (live, mit kurzem Cache) ─────────────────────────────────

type matchCache struct {
	mu   sync.Mutex
	data *fussball.Matches
	at   time.Time
	id   string
}

var matchesCache = &matchCache{}

const matchesTTL = 10 * time.Minute

// GetMatches liefert letzte + kommende Spiele (dekodiert), max. alle 10 Min neu.
func (a *API) GetMatches(c *gin.Context) {
	var club models.Club
	if err := a.db.First(&club, "id = 1").Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Verein nicht gefunden"})
		return
	}
	if club.FussballMatchesId == "" {
		c.JSON(http.StatusOK, &fussball.Matches{Previous: []fussball.Match{}, Next: []fussball.Match{}})
		return
	}

	matchesCache.mu.Lock()
	fresh := matchesCache.data != nil && matchesCache.id == club.FussballMatchesId && time.Since(matchesCache.at) < matchesTTL
	cached := matchesCache.data
	matchesCache.mu.Unlock()

	if fresh {
		c.JSON(http.StatusOK, cached)
		return
	}

	m, err := fussball.FetchMatches(club.FussballMatchesId, club.Name)
	if err != nil {
		if cached != nil { // im Fehlerfall lieber alte Daten als nichts
			c.JSON(http.StatusOK, cached)
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": "fussball.de nicht erreichbar"})
		return
	}
	matchesCache.mu.Lock()
	matchesCache.data, matchesCache.at, matchesCache.id = m, time.Now(), club.FussballMatchesId
	matchesCache.mu.Unlock()
	c.JSON(http.StatusOK, m)
}
