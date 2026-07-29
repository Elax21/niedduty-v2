package store

import (
	"log"

	"github.com/alessandro/niedduty/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Seed legt die Grundkonfiguration für Aramäer Ahlen an, wenn die DB leer ist:
// Verein (inkl. fussball.de-Widget-IDs), Admin-Konto, Strafenkatalog und die
// wiederkehrenden Trainings. Kader + Tabelle bleiben leer — Spieler kommen über
// Selbstregistrierung, die Tabelle wird von fussball.de gesynct.
func Seed(db *gorm.DB) {
	var count int64
	db.Model(&models.User{}).Count(&count)
	if count > 0 {
		return
	}
	log.Println("Seed: lege Grundkonfiguration für Aramäer Ahlen an …")

	db.Create(&models.Club{
		ID: 1, Name: "Aramäer Ahlen", Short: "ARA",
		PrimaryColor: "#C8342B", SecondaryColor: "#E3B341",
		Liga:                "Kreisliga A Beckum",
		FussballTableId:     "85bc8155-cd18-449f-b5d5-db1ef7277ab9",
		FussballMatchesId:   "82539853-64c3-4562-8a70-23e05606df0f",
		FussballNextMatchId: "aab8a3a1-12c9-4a0a-bd06-da2911b780ea",
	})

	hash, _ := bcrypt.GenerateFromPassword([]byte("demo1234!"), bcrypt.DefaultCost)
	adminEmail := "admin@aramaeer-ahlen.de"
	db.Create(&models.User{
		Alias: "admin", Email: &adminEmail, Name: "Alessandro Nieddu",
		PasswordHash: string(hash), Role: models.RoleAdmin, Permissions: []string{},
	})

	// Strafenkatalog — sinnvolle Startwerte (editierbar).
	catalog := []models.Penalty{
		{Label: "Verspätung Training", Amount: 100, Unit: "pro Minute", SortOrder: 1},
		{Label: "Unentschuldigtes Fehlen Training", Amount: 1000, SortOrder: 2},
		{Label: "Verspätung Spieltag", Amount: 500, SortOrder: 3},
		{Label: "Unentschuldigtes Fehlen Spiel", Amount: 2500, SortOrder: 4},
		{Label: "Gelbe Karte (Meckern)", Amount: 1000, SortOrder: 5},
		{Label: "Rote Karte (Unsportlichkeit)", Amount: 2500, SortOrder: 6},
		{Label: "Handy in der Kabinenbesprechung", Amount: 500, SortOrder: 7},
		{Label: "Falsches Trikot / Ausrüstung vergessen", Amount: 500, SortOrder: 8},
		{Label: "Kiste Bier vergessen (Geburtstag)", Amount: 2000, SortOrder: 9},
	}
	for i := range catalog {
		db.Create(&catalog[i])
	}

	// Wiederkehrende Trainings: Mi + Fr 19:15 am Sportpark Nord.
	trainings := []models.Event{
		{Title: "Training", Type: "training", Date: "2026-06-03", StartTime: "19:15", EndTime: "21:00",
			Location: "Sportpark Nord", Recurring: true, RecurrenceType: "weekly"},
		{Title: "Training", Type: "training", Date: "2026-06-05", StartTime: "19:15", EndTime: "21:00",
			Location: "Sportpark Nord", Recurring: true, RecurrenceType: "weekly"},
	}
	for i := range trainings {
		db.Create(&trainings[i])
	}
	log.Println("Seed: fertig.")
}
