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

	// Strafenkatalog der Saison 2025/26, wie im Aushang beschlossen (editierbar).
	// Die Nummerierung im Original springt von 7 auf 9 — hier fortlaufend.
	catalog := []models.Penalty{
		{Label: "Unentschuldigtes Fehlen beim Spiel", Amount: 5000, SortOrder: 1},
		{Label: "Unentschuldigtes Fehlen beim Training", Amount: 2500,
			Unit: "Abmeldung bis 16 Uhr beim Trainerteam, nicht in der Gruppe", SortOrder: 2},
		{Label: "Verspätet zum Training", Amount: 50,
			Unit: "pro Minute · ohne Abmeldung, 19:15 Uhr auf dem Platz", SortOrder: 3},
		{Label: "Verspätet zum Treffpunkt", Amount: 50,
			Unit: "pro Minute · ohne Abmeldung, ab 5 Min nach Treffpunkt", SortOrder: 4},
		{Label: "Gelbe Karte wegen Meckern oder Beleidigung", Amount: 1000, SortOrder: 5},
		{Label: "Gelb-Rote Karte wegen Meckern oder Beleidigung", Amount: 2000, SortOrder: 6},
		{Label: "Rote Karte wegen unsportlichem Verhalten", Amount: 4000, SortOrder: 7},
		{Label: "Trainingsanzug zum Spiel nicht angezogen", Amount: 500, Unit: "wenn im Besitz", SortOrder: 8},
		{Label: "Zum Aufwärmen Zipper oder T-Shirt nicht an", Amount: 500, Unit: "wenn im Besitz", SortOrder: 9},
		{Label: "Rauchen oder Alkohol im Trikot", Amount: 500, SortOrder: 10},
		{Label: "Gegenstände vergessen", Amount: 200,
			Unit: "pro Stück · Fußballschuhe etc., Unterwäsche zählt nicht", SortOrder: 11},
		{Label: "Handy klingelt oder Spielen während der Besprechung", Amount: 500, SortOrder: 12},
		{Label: "Handynutzung während des Spiels", Amount: 1000, SortOrder: 13},
		{Label: "Ball über den Zaun geschossen", Amount: 3000,
			Unit: "alle Schützen zahlen, auch wenn der Ball unerreichbar ist", SortOrder: 14},
		{Label: "Kabine unsauber verlassen", Amount: 500,
			Unit: "je Spieler im Spielbericht", SortOrder: 15},
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
