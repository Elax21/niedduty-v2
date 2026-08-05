package main

import (
	"log"
	"time"

	"github.com/alessandro/niedduty/internal/api"
	"github.com/alessandro/niedduty/internal/config"
	"github.com/alessandro/niedduty/internal/store"
	"github.com/alessandro/niedduty/internal/web"
	"github.com/gin-gonic/gin"
)

// Niedduty-Server: API für Ligatabelle, Kalender, Trainingsbeteiligung
// und Strafenkatalog von Aramäer Ahlen. In Produktion liefert dasselbe
// Binary auch das eingebettete Frontend aus.
func main() {
	cfg := config.Load()
	if cfg.Production {
		gin.SetMode(gin.ReleaseMode)
	}

	db, err := store.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	store.Seed(db)

	// Tabelle beim Start + alle 30 Min von fussball.de holen (kein Cron nötig).
	api.StartTableSyncLoop(db, 30*time.Minute)
	// Termin-Erinnerungen per Web-Push (ebenfalls ohne externen Cron).
	api.StartReminderLoop(db, 5*time.Minute)

	r := gin.Default()
	if err := r.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		log.Fatalf("trusted proxies: %v", err)
	}
	api.New(db, cfg.SecureCookies).Routes(r)
	if web.Mount(r) {
		log.Print("Frontend aus dem Binary eingebunden")
	}

	log.Printf("niedduty-server läuft auf %s", cfg.ListenAddr)
	if err := r.Run(cfg.ListenAddr); err != nil {
		log.Fatal(err)
	}
}
