package main

import (
	"log"
	"time"

	"github.com/alessandro/niedduty/internal/api"
	"github.com/alessandro/niedduty/internal/config"
	"github.com/alessandro/niedduty/internal/store"
	"github.com/gin-gonic/gin"
)

// Niedduty-Server: API für Ligatabelle, Kalender, Trainingsbeteiligung
// und Strafenkatalog von Aramäer Ahlen.
func main() {
	cfg := config.Load()

	db, err := store.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	store.Seed(db)

	// Tabelle beim Start + alle 30 Min von fussball.de holen (kein Cron nötig).
	api.StartTableSyncLoop(db, 30*time.Minute)

	r := gin.Default()
	api.New(db).Routes(r)

	log.Printf("niedduty-server läuft auf %s", cfg.ListenAddr)
	if err := r.Run(cfg.ListenAddr); err != nil {
		log.Fatal(err)
	}
}
