package api

import (
	"github.com/alessandro/niedduty/internal/middleware"
	"github.com/alessandro/niedduty/internal/models"
	"github.com/alessandro/niedduty/internal/web"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// API bündelt DB und Laufzeit-Einstellungen für alle Handler.
type API struct {
	db *gorm.DB
	// secureCookies — Session-Cookie nur über HTTPS ausliefern (Produktion).
	secureCookies bool
}

func New(db *gorm.DB, secureCookies bool) *API {
	return &API{db: db, secureCookies: secureCookies}
}

// Routes registriert alle HTTP-Routen.
func (a *API) Routes(r *gin.Engine) {
	r.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	// Build-Kennung des Frontends. Die installierte App vergleicht sie und
	// lädt sich neu, wenn ein neues Binary läuft.
	r.GET("/api/version", func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.JSON(200, gin.H{"version": web.Version()})
	})

	r.POST("/api/auth/login", a.Login)
	r.POST("/api/auth/logout", a.Logout)
	r.POST("/api/auth/register", a.Register)
	r.GET("/api/invite/:token", a.CheckInvite)

	auth := r.Group("/api", middleware.Auth(a.db))
	{
		auth.GET("/auth/me", a.Me)
		auth.PUT("/auth/me/tutorial", a.SetTutorialDone)
		auth.GET("/club", a.GetClub)
		auth.GET("/players", a.ListPlayers)
		auth.GET("/table", a.GetTable)
		auth.GET("/penalties", a.ListPenalties)
		auth.GET("/player-penalties", a.ListPlayerPenalties)
		auth.GET("/player-penalties/summary", a.PenaltiesSummary)
		auth.GET("/events", a.ListEvents)
		auth.GET("/training-schedule", a.GetTrainingSchedule)
		auth.GET("/stats/overview", a.StatsOverview)
		auth.GET("/polls", a.ListPolls)
		auth.GET("/polls/running", a.ListRunningPolls)
		auth.POST("/polls/:id/vote", a.Vote)

		// Abstimmungs-Starter: Admin oder Recht "umfragen".
		umfragen := auth.Group("", middleware.RequirePerm(models.PermUmfragen))
		{
			umfragen.POST("/polls", a.CreatePoll)
			umfragen.POST("/polls/:id/close", a.ClosePoll)
			umfragen.DELETE("/polls/:id", a.DeletePoll)
		}
		auth.GET("/fussball/matches", a.GetMatches)
		auth.GET("/fussball/scouting", a.GetScouting)
		auth.GET("/fussball/squad-stats", a.GetSquadStats)
		auth.GET("/attendance", a.ListAttendance)
		auth.GET("/push/key", a.GetPushKey)
		auth.POST("/push/subscribe", a.Subscribe)
		auth.POST("/push/unsubscribe", a.Unsubscribe)
		auth.POST("/push/test", a.TestPush)
		auth.GET("/push/settings", a.GetPushSettings)
		auth.PUT("/push/settings", a.PutPushSettings)
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
			strafen.GET("/penalty-log", a.ListPenaltyLog)
			strafen.GET("/penalty-log/verify", a.VerifyPenaltyLog)
		}

		// Termin-Einsteller: Admin oder Recht "termine".
		termine := auth.Group("", middleware.RequirePerm(models.PermTermine))
		{
			termine.POST("/events", a.CreateEvent)
			termine.PUT("/events/:id", a.UpdateEvent)
			termine.DELETE("/events/:id", a.DeleteEvent)
			termine.PUT("/training-schedule", a.PutTrainingSchedule)
			termine.PUT("/event-notes", a.PutEventNote)
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
