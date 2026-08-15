package api

import (
	"net/http"
	"time"

	"github.com/alessandro/niedduty/internal/middleware"
	"github.com/alessandro/niedduty/internal/models"
	"github.com/gin-gonic/gin"
)

// Ausgaben der Mannschaftskasse — was rausgeht (Bälle, Mannschaftsabend,
// Essen). Sehen dürfen es alle, eintragen nur die Kassenwarte; jede Bewegung
// landet im Kassen-Protokoll.

type expenseReq struct {
	Label  string `json:"label" binding:"required,max=120"`
	Amount int    `json:"amount" binding:"required,min=1,max=10000000"` // Cent
	Date   string `json:"date" binding:"max=10"`                        // YYYY-MM-DD
}

func (a *API) ListExpenses(c *gin.Context) {
	var list []models.Expense
	a.db.Order("date desc, created_at desc").Find(&list)
	c.JSON(http.StatusOK, list)
}

func (a *API) CreateExpense(c *gin.Context) {
	var req expenseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Grund und Betrag (Cent) angeben"})
		return
	}
	if req.Date == "" {
		req.Date = time.Now().Format("2006-01-02")
	} else if _, err := time.Parse("2006-01-02", req.Date); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datum muss YYYY-MM-DD sein"})
		return
	}
	user := middleware.CurrentUser(c)
	e := models.Expense{
		Label: req.Label, Amount: req.Amount, Date: req.Date,
		CreatedBy: user.ID, CreatorName: user.Name,
	}
	if err := a.db.Create(&e).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ausgabe konnte nicht gespeichert werden"})
		return
	}
	a.writePenaltyLog(c, logEntry{Action: models.PenaltyActionExpense, Label: e.Label, Amount: e.Amount})
	c.JSON(http.StatusCreated, e)
}

func (a *API) DeleteExpense(c *gin.Context) {
	var e models.Expense
	if err := a.db.First(&e, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ausgabe nicht gefunden"})
		return
	}
	a.db.Delete(&models.Expense{}, "id = ?", e.ID)
	a.writePenaltyLog(c, logEntry{Action: models.PenaltyActionExpenseX, Label: e.Label, Amount: e.Amount})
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
