package api

import (
	"net/http"

	"github.com/alessandro/niedduty/internal/middleware"
	"github.com/alessandro/niedduty/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type penaltyReq struct {
	Label     string `json:"label" binding:"required,max=120"`
	Amount    int    `json:"amount" binding:"required,min=1,max=100000"` // Cent
	Unit      string `json:"unit" binding:"max=40"`
	SortOrder int    `json:"sortOrder"`
}

func (a *API) ListPenalties(c *gin.Context) {
	var list []models.Penalty
	a.db.Order("sort_order, label").Find(&list)
	c.JSON(http.StatusOK, list)
}

func (a *API) CreatePenalty(c *gin.Context) {
	var req penaltyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bezeichnung und Betrag (Cent) angeben"})
		return
	}
	p := models.Penalty{Label: req.Label, Amount: req.Amount, Unit: req.Unit, SortOrder: req.SortOrder}
	a.db.Create(&p)
	a.writePenaltyLog(c, logEntry{Action: models.PenaltyActionCatalog, Label: p.Label, Amount: p.Amount})
	c.JSON(http.StatusCreated, p)
}

func (a *API) UpdatePenalty(c *gin.Context) {
	var p models.Penalty
	if err := a.db.First(&p, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Eintrag nicht gefunden"})
		return
	}
	var req penaltyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bezeichnung und Betrag (Cent) angeben"})
		return
	}
	p.Label, p.Amount, p.Unit, p.SortOrder = req.Label, req.Amount, req.Unit, req.SortOrder
	a.db.Save(&p)
	a.writePenaltyLog(c, logEntry{Action: models.PenaltyActionCatalog, Label: p.Label, Amount: p.Amount})
	c.JSON(http.StatusOK, p)
}

func (a *API) DeletePenalty(c *gin.Context) {
	var p models.Penalty
	a.db.First(&p, "id = ?", c.Param("id"))
	a.db.Delete(&models.Penalty{}, "id = ?", c.Param("id"))
	a.writePenaltyLog(c, logEntry{Action: models.PenaltyActionCatalogX, Label: p.Label, Amount: p.Amount})
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ── Zugewiesene Strafen (Mannschaftskasse) ──────────────────────

// ListPlayerPenalties — Aufschreiber/Admin sehen alle; normale Spieler nur
// ihre eigenen (Privatsphäre: fremde Beträge sind serverseitig ausgeblendet).
func (a *API) ListPlayerPenalties(c *gin.Context) {
	user := middleware.CurrentUser(c)
	q := a.db.Order("created_at desc")
	if user.Can(models.PermStrafen) {
		if pid := c.Query("playerId"); pid != "" {
			q = q.Where("player_id = ?", pid)
		}
	} else {
		if user.PlayerID == nil {
			c.JSON(http.StatusOK, []models.PlayerPenalty{})
			return
		}
		q = q.Where("player_id = ?", *user.PlayerID)
	}
	var list []models.PlayerPenalty
	q.Find(&list)
	c.JSON(http.StatusOK, list)
}

// PenaltiesSummary — sichere Team-Aggregatsummen (offen/bezahlt), ohne
// Aufschlüsselung pro Spieler. Für alle eingeloggten Nutzer sichtbar.
func (a *API) PenaltiesSummary(c *gin.Context) {
	var open, paid int64
	a.db.Model(&models.PlayerPenalty{}).Where("paid = ?", false).Select("COALESCE(SUM(amount),0)").Scan(&open)
	a.db.Model(&models.PlayerPenalty{}).Where("paid = ?", true).Select("COALESCE(SUM(amount),0)").Scan(&paid)
	c.JSON(http.StatusOK, gin.H{"totalOpen": open, "totalPaid": paid})
}

// assignReq — Mehrfach-Zuweisung: jeder gewählte Spieler bekommt jedes
// gewählte Vergehen (Label+Betrag werden kopiert). Zusätzlich optional
// eine freie Strafe.
type assignReq struct {
	PlayerIDs  []uuid.UUID `json:"playerIds" binding:"required,min=1"`
	PenaltyIDs []uuid.UUID `json:"penaltyIds"`
	FreeLabel  string      `json:"freeLabel" binding:"max=120"`
	FreeAmount int         `json:"freeAmount" binding:"min=0,max=100000"`
}

func (a *API) AssignPenalty(c *gin.Context) {
	var req assignReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Mindestens einen Spieler wählen"})
		return
	}
	hasFree := req.FreeLabel != "" && req.FreeAmount > 0
	if len(req.PenaltyIDs) == 0 && !hasFree {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Mindestens ein Vergehen wählen (oder freie Strafe angeben)"})
		return
	}
	perPlayer := len(req.PenaltyIDs)
	if hasFree {
		perPlayer++
	}
	if len(req.PlayerIDs)*perPlayer > 60 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Zu viele Strafen auf einmal (max. 60)"})
		return
	}

	var squad []models.Player
	a.db.Find(&squad)
	squadIDs := map[uuid.UUID]bool{}
	for _, p := range squad {
		squadIDs[p.ID] = true
	}
	var katalog []models.Penalty
	a.db.Find(&katalog)
	byID := map[uuid.UUID]models.Penalty{}
	for _, k := range katalog {
		byID[k.ID] = k
	}

	rows := []models.PlayerPenalty{}
	for _, pid := range req.PlayerIDs {
		if !squadIDs[pid] {
			c.JSON(http.StatusNotFound, gin.H{"error": "Spieler nicht gefunden"})
			return
		}
		for _, penID := range req.PenaltyIDs {
			pen, ok := byID[penID]
			if !ok {
				c.JSON(http.StatusNotFound, gin.H{"error": "Vergehen nicht gefunden"})
				return
			}
			rows = append(rows, models.PlayerPenalty{PlayerID: pid, Label: pen.Label, Amount: pen.Amount})
		}
		if hasFree {
			rows = append(rows, models.PlayerPenalty{PlayerID: pid, Label: req.FreeLabel, Amount: req.FreeAmount})
		}
	}
	if err := a.db.Create(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Strafen konnten nicht gespeichert werden"})
		return
	}

	names := a.playerNames(req.PlayerIDs)
	entries := make([]logEntry, 0, len(rows))
	for i := range rows {
		pid := rows[i].PlayerID
		entries = append(entries, logEntry{
			Action: models.PenaltyActionAssign, PlayerID: &pid, PlayerName: names[pid],
			Label: rows[i].Label, Amount: rows[i].Amount,
		})
	}
	a.writePenaltyLog(c, entries...)

	c.JSON(http.StatusCreated, rows)
}

// penaltyEntries baut Protokollzeilen aus bestehenden Strafen.
func penaltyEntries(a *API, rows []models.PlayerPenalty, action string) []logEntry {
	ids := make([]uuid.UUID, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.PlayerID)
	}
	names := a.playerNames(ids)
	entries := make([]logEntry, 0, len(rows))
	for i := range rows {
		pid := rows[i].PlayerID
		entries = append(entries, logEntry{
			Action: action, PlayerID: &pid, PlayerName: names[pid],
			Label: rows[i].Label, Amount: rows[i].Amount,
		})
	}
	return entries
}

type bulkPaidReq struct {
	IDs  []uuid.UUID `json:"ids" binding:"required,min=1"`
	Paid bool        `json:"paid"`
}

// SetPenaltiesPaid — mehrere Strafen auf bezahlt/offen setzen.
func (a *API) SetPenaltiesPaid(c *gin.Context) {
	var req bulkPaidReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Keine Strafe ausgewählt"})
		return
	}
	// Vorher lesen: nur so weiß das Protokoll, um welche Strafen es geht.
	var rows []models.PlayerPenalty
	a.db.Where("id IN ?", req.IDs).Find(&rows)
	a.db.Model(&models.PlayerPenalty{}).Where("id IN ?", req.IDs).Update("paid", req.Paid)

	action := models.PenaltyActionPaid
	if !req.Paid {
		action = models.PenaltyActionUnpaid
	}
	a.writePenaltyLog(c, penaltyEntries(a, rows, action)...)

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type bulkIDsReq struct {
	IDs []uuid.UUID `json:"ids" binding:"required,min=1"`
}

// DeletePlayerPenalties — mehrere Strafen löschen.
func (a *API) DeletePlayerPenalties(c *gin.Context) {
	var req bulkIDsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Keine Strafe ausgewählt"})
		return
	}
	var rows []models.PlayerPenalty
	a.db.Where("id IN ?", req.IDs).Find(&rows)
	a.db.Delete(&models.PlayerPenalty{}, "id IN ?", req.IDs)
	a.writePenaltyLog(c, penaltyEntries(a, rows, models.PenaltyActionDelete)...)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
