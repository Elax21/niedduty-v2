package store

import (
	"log"
	"time"

	"github.com/alessandro/niedduty/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Seed legt Demo-Daten für Aramäer Ahlen an, wenn die DB leer ist.
func Seed(db *gorm.DB) {
	var count int64
	db.Model(&models.User{}).Count(&count)
	if count > 0 {
		return
	}
	log.Println("Seed: lege Demo-Daten für Aramäer Ahlen an …")

	db.Create(&models.Club{
		ID: 1, Name: "Aramäer Ahlen", Short: "ARA",
		PrimaryColor: "#C8342B", SecondaryColor: "#E3B341",
		Liga: "Kreisliga A Beckum",
	})

	hash, _ := bcrypt.GenerateFromPassword([]byte("demo1234!"), bcrypt.DefaultCost)
	db.Create(&models.User{
		Email: "admin@aramaeer-ahlen.de", Name: "Alessandro Nieddu",
		PasswordHash: string(hash), Role: models.RoleAdmin, Permissions: []string{},
	})

	num := func(n int) *int { return &n }
	players := []models.Player{
		{Name: "Gabriel Aydin", Number: num(1), Position: "TW"},
		{Name: "Johannes Gabriel", Number: num(12), Position: "TW"},
		{Name: "Aziz Demir", Number: num(2), Position: "AB"},
		{Name: "Simon Chabo", Number: num(3), Position: "AB"},
		{Name: "Daniel Barsom", Number: num(4), Position: "AB"},
		{Name: "Jakob Malki", Number: num(5), Position: "AB"},
		{Name: "Petrus Younan", Number: num(15), Position: "AB"},
		{Name: "Lukas Hanna", Number: num(6), Position: "MF"},
		{Name: "Matay Gürsel", Number: num(8), Position: "MF"},
		{Name: "Elias Sahin", Number: num(10), Position: "MF"},
		{Name: "Josef Tuma", Number: num(14), Position: "MF"},
		{Name: "Markus Iskender", Number: num(16), Position: "MF"},
		{Name: "Sargon Akay", Number: num(18), Position: "MF"},
		{Name: "David Oez", Number: num(7), Position: "ST"},
		{Name: "Tuma Karat", Number: num(9), Position: "ST"},
		{Name: "Ninos Aslan", Number: num(11), Position: "ST"},
		{Name: "Afrem Bulut", Number: num(19), Position: "ST", Status: "verletzt"},
		{Name: "Isa Gabriel", Number: num(20), Position: "ST"},
	}
	for i := range players {
		if players[i].Status == "" {
			players[i].Status = "fit"
		}
		db.Create(&players[i])
	}

	table := []models.LeagueEntry{
		{TeamName: "SC DJK Everswinkel", Played: 30, Won: 22, Drawn: 4, Lost: 4, GoalsFor: 81, GoalsAgainst: 32, Points: 70},
		{TeamName: "Aramäer Ahlen", IsOwn: true, Played: 30, Won: 20, Drawn: 5, Lost: 5, GoalsFor: 74, GoalsAgainst: 35, Points: 65},
		{TeamName: "Fortuna Walstedde", Played: 30, Won: 18, Drawn: 6, Lost: 6, GoalsFor: 66, GoalsAgainst: 38, Points: 60},
		{TeamName: "SG Sendenhorst", Played: 30, Won: 16, Drawn: 7, Lost: 7, GoalsFor: 61, GoalsAgainst: 41, Points: 55},
		{TeamName: "TuS Freckenhorst", Played: 30, Won: 15, Drawn: 6, Lost: 9, GoalsFor: 58, GoalsAgainst: 44, Points: 51},
		{TeamName: "Vorwärts Ahlen", Played: 30, Won: 14, Drawn: 5, Lost: 11, GoalsFor: 55, GoalsAgainst: 48, Points: 47},
		{TeamName: "SW Liesborn", Played: 30, Won: 12, Drawn: 8, Lost: 10, GoalsFor: 49, GoalsAgainst: 45, Points: 44},
		{TeamName: "GW Amelsbüren", Played: 30, Won: 11, Drawn: 7, Lost: 12, GoalsFor: 46, GoalsAgainst: 50, Points: 40},
		{TeamName: "BSV Ostbevern", Played: 30, Won: 10, Drawn: 8, Lost: 12, GoalsFor: 44, GoalsAgainst: 51, Points: 38},
		{TeamName: "SC Hoetmar", Played: 30, Won: 9, Drawn: 9, Lost: 12, GoalsFor: 41, GoalsAgainst: 49, Points: 36},
		{TeamName: "Ahlener SG II", Played: 30, Won: 8, Drawn: 8, Lost: 14, GoalsFor: 39, GoalsAgainst: 55, Points: 32},
		{TeamName: "TuS Wadersloh", Played: 30, Won: 7, Drawn: 9, Lost: 14, GoalsFor: 36, GoalsAgainst: 54, Points: 30},
		{TeamName: "SV Drensteinfurt", Played: 30, Won: 6, Drawn: 8, Lost: 16, GoalsFor: 33, GoalsAgainst: 61, Points: 26},
		{TeamName: "Westfalia Vorhelm", Played: 30, Won: 5, Drawn: 7, Lost: 18, GoalsFor: 29, GoalsAgainst: 66, Points: 22},
		{TeamName: "SG Telgte II", Played: 30, Won: 4, Drawn: 6, Lost: 20, GoalsFor: 27, GoalsAgainst: 74, Points: 18},
		{TeamName: "SC Füchtorf", Played: 30, Won: 2, Drawn: 5, Lost: 23, GoalsFor: 21, GoalsAgainst: 87, Points: 11},
	}
	for i := range table {
		db.Create(&table[i])
	}

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

	// Wiederkehrende Trainings: Di + Do 19:30 seit Anfang Juni.
	trainings := []models.Event{
		{Title: "Training", Type: "training", Date: "2026-06-02", StartTime: "19:30", EndTime: "21:00",
			Location: "Sportplatz Ahlen-Süd", Recurring: true, RecurrenceType: "weekly"},
		{Title: "Training", Type: "training", Date: "2026-06-04", StartTime: "19:30", EndTime: "21:00",
			Location: "Sportplatz Ahlen-Süd", Recurring: true, RecurrenceType: "weekly"},
	}
	for i := range trainings {
		db.Create(&trainings[i])
	}
	extras := []models.Event{
		{Title: "Testspiel vs. Fortuna Walstedde", Type: "spiel", Date: "2026-07-19", StartTime: "15:00", Location: "Sportplatz Walstedde"},
		{Title: "Saisoneröffnung Grillen", Type: "mannschaftsabend", Date: "2026-07-24", StartTime: "18:30", Location: "Vereinsheim"},
		{Title: "Trainingsauftakt Vorbereitung", Type: "sonstiges", Date: "2026-07-07", StartTime: "19:00", Location: "Vereinsheim"},
	}
	for i := range extras {
		db.Create(&extras[i])
	}

	// Beteiligung für vergangene Trainings (deterministisch gemischt).
	today := time.Now()
	for _, tr := range trainings {
		start, _ := time.Parse("2006-01-02", tr.Date)
		occ := 0
		for d := start; d.Before(today); d = d.AddDate(0, 0, 7) {
			key := tr.ID.String() + "_" + d.Format("2006-01-02")
			for i, p := range players {
				m := (i*7 + occ*3 + i*occ) % 10
				switch {
				case m < 7:
					db.Create(&models.EventAttendance{EventKey: key, PlayerID: p.ID, Status: "attending", UpdatedAt: d})
				case m < 9:
					reason := "Arbeit"
					if m == 8 {
						reason = "Krank"
					}
					db.Create(&models.EventAttendance{EventKey: key, PlayerID: p.ID, Status: "declined", Reason: reason, UpdatedAt: d})
				}
			}
			occ++
		}
	}
	log.Println("Seed: fertig.")
}
