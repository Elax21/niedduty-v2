package api

import (
	"github.com/alessandro/niedduty/internal/middleware"
	"github.com/alessandro/niedduty/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// API bündelt die DB für alle Handler.
type API struct {
	db *gorm.DB
}

func New(db *gorm.DB) *API {
	return &API{db: db}
}

// Routes registriert alle HTTP-Routen.
func (a *API) Routes(r *gin.Engine) {
	r.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	r.POST("/api/auth/login", a.Login)
	r.POST("/api/auth/logout", a.Logout)
	r.POST("/api/auth/register", a.Register)
	r.GET("/api/invite/:token", a.CheckInvite)

	auth := r.Group("/api", middleware.Auth(a.db))
	{
		auth.GET("/auth/me", a.Me)
		auth.GET("/club", a.GetClub)
		auth.GET("/players", a.ListPlayers)
		auth.GET("/table", a.GetTable)
		auth.GET("/penalties", a.ListPenalties)
		auth.GET("/player-penalties", a.ListPlayerPenalties)
		auth.GET("/player-penalties/summary", a.PenaltiesSummary)
		auth.GET("/events", a.ListEvents)
		auth.GET("/fussball/matches", a.GetMatches)
		auth.GET("/attendance", a.ListAttendance)
		// Spieler dürfen die eigene Zu-/Absage setzen; Handler prüft Ownership.
		auth.PUT("/attendance", a.SetAttendance)

		// Beteiligung aller sehen: Admin oder Recht "beteiligung".
		auth.GET("/attendance/stats", middleware.RequirePerm(models.PermBeteiligung), a.AttendanceStats)

		// Strafenaufschreiber: Admin oder Recht "strafen".
		strafen := auth.Group("", middleware.RequirePerm(models.PermStrafen))
		{
			strafen.POST("/penalties", a.CreatePenalty)
			strafen.PUT("/penalties/:id", a.UpdatePenalty)
			strafen.DELETE("/penalties/:id", a.DeletePenalty)
			strafen.POST("/player-penalties", a.AssignPenalty)
			strafen.POST("/player-penalties/paid", a.SetPenaltiesPaid)
			strafen.POST("/player-penalties/delete", a.DeletePlayerPenalties)
		}

		// Termin-Einsteller: Admin oder Recht "termine".
		termine := auth.Group("", middleware.RequirePerm(models.PermTermine))
		{
			termine.POST("/events", a.CreateEvent)
			termine.PUT("/events/:id", a.UpdateEvent)
			termine.DELETE("/events/:id", a.DeleteEvent)
		}

		// Nur Admin: Einstellungen, Benutzer + Rechte, Kader, Tabelle.
		admin := auth.Group("", middleware.RequireAdmin())
		{
			admin.PUT("/club", a.UpdateClub)
		admin.GET("/invite", a.GetInvite)
		admin.POST("/invite", a.CreateInvite)
		admin.DELETE("/invite", a.DeactivateInvite)
			admin.GET("/users", a.ListUsers)
			admin.POST("/users", a.CreateUser)
			admin.PUT("/users/:id", a.UpdateUser)
			admin.DELETE("/users/:id", a.DeleteUser)
			admin.POST("/players", a.CreatePlayer)
			admin.PUT("/players/:id", a.UpdatePlayer)
			admin.DELETE("/players/:id", a.DeletePlayer)
			admin.PUT("/table", a.ReplaceTable)
		admin.POST("/table/sync", a.SyncTable)
		}
	}
}
