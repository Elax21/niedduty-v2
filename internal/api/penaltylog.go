package api

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/alessandro/niedduty/internal/middleware"
	"github.com/alessandro/niedduty/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Kassen-Protokoll. Wer darf Strafen aufschreiben, kann auch löschen — damit
// das niemand still tut, wird jede Bewegung protokolliert. Die Kette aus
// Hashes macht nachträgliche Eingriffe sichtbar, auch direkt in der Datenbank.

// logMu serialisiert das Anhängen, sonst könnten zwei Requests denselben
// Vorgänger-Hash erwischen und die Kette zerreißen.
var logMu sync.Mutex

type logEntry struct {
	Action     string
	PlayerID   *uuid.UUID
	PlayerName string
	Label      string
	Amount     int
}

// hashLog bildet den Fingerabdruck eines Eintrags inklusive Vorgänger.
func hashLog(e models.PenaltyLog) string {
	player := ""
	if e.PlayerID != nil {
		player = e.PlayerID.String()
	}
	// Mikrosekunden: PostgreSQL speichert timestamptz nicht feiner. Mit
	// Nanosekunden würde der Hash nach dem Zurücklesen nicht mehr passen.
	raw := fmt.Sprintf("%s|%d|%s|%s|%s|%s|%d|%s",
		e.PrevHash, e.CreatedAt.UTC().UnixMicro(), e.ActorID,
		e.Action, player, e.Label, e.Amount, e.ActorAlias)
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// writePenaltyLog hängt Einträge an die Kette. Fehler hier dürfen die
// eigentliche Aktion nicht kippen — protokolliert wird best effort, aber der
// Bruch fällt bei der Prüfung auf.
func (a *API) writePenaltyLog(c *gin.Context, entries ...logEntry) {
	if len(entries) == 0 {
		return
	}
	user := middleware.CurrentUser(c)

	logMu.Lock()
	defer logMu.Unlock()

	var last models.PenaltyLog
	prev := ""
	if err := a.db.Order("seq desc").First(&last).Error; err == nil {
		prev = last.Hash
	}

	for _, e := range entries {
		row := models.PenaltyLog{
			CreatedAt:  time.Now().Truncate(time.Microsecond),
			ActorID:    user.ID,
			ActorName:  user.Name,
			ActorAlias: user.Alias,
			Action:     e.Action,
			PlayerID:   e.PlayerID,
			PlayerName: e.PlayerName,
			Label:      e.Label,
			Amount:     e.Amount,
			PrevHash:   prev,
		}
		row.Hash = hashLog(row)
		if err := a.db.Create(&row).Error; err != nil {
			return
		}
		prev = row.Hash
	}
}

// ListPenaltyLog — jüngste Einträge zuerst (nur Aufschreiber/Admin).
func (a *API) ListPenaltyLog(c *gin.Context) {
	var list []models.PenaltyLog
	a.db.Order("seq desc").Limit(300).Find(&list)
	c.JSON(http.StatusOK, list)
}

// VerifyPenaltyLog rechnet die Kette nach und meldet den ersten Bruch.
func (a *API) VerifyPenaltyLog(c *gin.Context) {
	var list []models.PenaltyLog
	a.db.Order("seq asc").Find(&list)

	prev := ""
	for _, e := range list {
		if e.PrevHash != prev || hashLog(e) != e.Hash {
			c.JSON(http.StatusOK, gin.H{
				"ok":       false,
				"count":    len(list),
				"brokenAt": e.Seq,
				"message":  fmt.Sprintf("Protokoll wurde ab Eintrag %d verändert", e.Seq),
			})
			return
		}
		prev = e.Hash
	}
	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"count":   len(list),
		"message": "Protokoll unverändert",
	})
}

// playerNames — Namen für das Protokoll, damit es ohne Joins lesbar bleibt.
func (a *API) playerNames(ids []uuid.UUID) map[uuid.UUID]string {
	out := map[uuid.UUID]string{}
	if len(ids) == 0 {
		return out
	}
	var players []models.Player
	a.db.Where("id IN ?", ids).Find(&players)
	for _, p := range players {
		out[p.ID] = p.Name
	}
	return out
}
