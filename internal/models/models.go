package models

import (
	"time"

	"github.com/google/uuid"
)

// Rollen — Ein-Vereins-Setup (Aramäer Ahlen).
// ADMIN = Alessandro (alles, inkl. Einstellungen + Rechtevergabe).
// MEMBER = normales Mitglied; Zusatzrechte über Permissions.
const (
	RoleAdmin  = "ADMIN"
	RoleMember = "MEMBER"
)

// Vergebbare Rechte für MEMBER-Konten.
const (
	PermStrafen     = "strafen"     // Strafen aufschreiben (Katalog + Zuweisung)
	PermTermine     = "termine"     // Termine anlegen/ändern/löschen
	PermBeteiligung = "beteiligung" // Trainingsbeteiligung aller ansehen
)

// User — Login-Konto. Login primär über Alias; E-Mail optional (nur Admin/Kontakt).
type User struct {
	ID           uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Alias        string     `gorm:"uniqueIndex;not null" json:"alias"` // Login-Handle, klein geschrieben
	Email        *string    `gorm:"uniqueIndex" json:"email"`          // optional
	Name         string     `gorm:"not null" json:"name"`
	PasswordHash string     `gorm:"not null" json:"-"`
	Role         string     `gorm:"not null;default:MEMBER" json:"role"`
	Permissions  []string   `gorm:"serializer:json;type:jsonb;not null;default:'[]'" json:"permissions"`
	PlayerID     *uuid.UUID `gorm:"type:uuid" json:"playerId"`
	CreatedAt    time.Time  `json:"createdAt"`
}

// Invite — teilbarer Einladungslink zur Selbstregistrierung.
// MaxUses = 0 → unbegrenzt. Active=false → Link ungültig (z.B. neu erzeugt).
type Invite struct {
	Token     string     `gorm:"primaryKey" json:"token"`
	Active    bool       `gorm:"not null;default:true" json:"active"`
	MaxUses   int        `gorm:"not null;default:0" json:"maxUses"`
	UseCount  int        `gorm:"not null;default:0" json:"useCount"`
	ExpiresAt *time.Time `json:"expiresAt"`
	CreatedAt time.Time  `json:"createdAt"`
}

// Can prüft ein Recht; Admin darf alles.
func (u *User) Can(perm string) bool {
	if u.Role == RoleAdmin {
		return true
	}
	for _, p := range u.Permissions {
		if p == perm {
			return true
		}
	}
	return false
}

// Session — Server-seitige Login-Session (Token im httpOnly-Cookie).
type Session struct {
	Token     string    `gorm:"primaryKey" json:"-"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
	ExpiresAt time.Time `gorm:"not null" json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
}

// Club — genau eine Zeile: Vereins-Stammdaten Aramäer Ahlen.
type Club struct {
	ID             int    `gorm:"primaryKey" json:"id"` // immer 1
	Name           string `gorm:"not null" json:"name"`
	Short          string `json:"short"` // z.B. "ARA"
	PrimaryColor   string `json:"primaryColor"`
	SecondaryColor string `json:"secondaryColor"`
	KasseIban      string `json:"kasseIban"`
	KasseInhaber   string `json:"kasseInhaber"`
	Liga           string `json:"liga"`
	// fussball.de-Widget-IDs (next.fussball.de). Leer = Bereich zeigt Platzhalter
	// bzw. manuelle Tabellenpflege. Eingebunden via widgets.js + data-type.
	FussballTableId     string `json:"fussballTableId"`     // data-type=table
	FussballMatchesId   string `json:"fussballMatchesId"`   // data-type=team-matches (letzte + nächste)
	FussballNextMatchId string `json:"fussballNextMatchId"` // data-type=next-match
	// Optionaler Link auf den geteilten Google-Team-Kalender.
	GoogleCalendarUrl string `json:"googleCalendarUrl"`
}

// Player — Kader-Eintrag (schlank: Basis für Beteiligung + Strafen).
type Player struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Number    *int      `json:"number"`
	Position  string    `gorm:"not null;default:MF" json:"position"` // TW | AB | MF | ST
	Status    string    `gorm:"not null;default:fit" json:"status"`  // fit | verletzt | gesperrt | krank
	CreatedAt time.Time `json:"createdAt"`
}

// LeagueEntry — eine Zeile der Ligatabelle (manuell gepflegt).
type LeagueEntry struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	TeamName     string    `gorm:"not null" json:"teamName"`
	IsOwn        bool      `gorm:"not null;default:false" json:"isOwn"`
	Played       int       `gorm:"not null;default:0" json:"played"`
	Won          int       `gorm:"not null;default:0" json:"won"`
	Drawn        int       `gorm:"not null;default:0" json:"drawn"`
	Lost         int       `gorm:"not null;default:0" json:"lost"`
	GoalsFor     int       `gorm:"not null;default:0" json:"goalsFor"`
	GoalsAgainst int       `gorm:"not null;default:0" json:"goalsAgainst"`
	Points       int       `gorm:"not null;default:0" json:"points"`
}

// Penalty — Katalog-Eintrag (z.B. "Verspätung Training", 100 Cent pro Minute).
type Penalty struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Label     string    `gorm:"not null" json:"label"`
	Amount    int       `gorm:"not null" json:"amount"` // Cent
	Unit      string    `json:"unit"`                   // z.B. "pro Minute", "pro Vorfall"
	SortOrder int       `gorm:"not null;default:0" json:"sortOrder"`
	CreatedAt time.Time `json:"createdAt"`
}

// PlayerPenalty — zugewiesene Strafe. Betrag wird kopiert, damit spätere
// Katalog-Änderungen alte Strafen nicht verfälschen.
type PlayerPenalty struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	PlayerID  uuid.UUID `gorm:"type:uuid;not null;index" json:"playerId"`
	Label     string    `gorm:"not null" json:"label"`
	Amount    int       `gorm:"not null" json:"amount"` // Cent
	Paid      bool      `gorm:"not null;default:false" json:"paid"`
	CreatedAt time.Time `json:"createdAt"`
}

// Event — Kalender-Termin. Wiederholung: weekly | biweekly bis RecurrenceEnd.
type Event struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Title          string    `gorm:"not null" json:"title"`
	Type           string    `gorm:"not null;default:sonstiges" json:"type"` // training | spiel | mannschaftsabend | sonstiges
	Date           string    `gorm:"not null" json:"date"`                   // YYYY-MM-DD
	EndDate        string    `json:"endDate"`                                // mehrtägig, optional
	StartTime      string    `json:"startTime"`                              // HH:MM
	EndTime        string    `json:"endTime"`
	Location       string    `json:"location"`
	Notes          string    `json:"notes"`
	Recurring      bool      `gorm:"not null;default:false" json:"recurring"`
	RecurrenceType string    `json:"recurrenceType"` // weekly | biweekly
	RecurrenceEnd  string    `json:"recurrenceEnd"`  // YYYY-MM-DD
	CreatedAt      time.Time `json:"createdAt"`
}

// Occurrence — expandiertes Vorkommen eines Termins für einen Datumsbereich.
// EventKey = ID bei Einzelterminen, "ID_YYYY-MM-DD" bei Wiederholungen.
type Occurrence struct {
	Event
	EventKey string `json:"eventKey"`
	OccDate  string `json:"occDate"`
}

// EventAttendance — Zu-/Absage eines Spielers pro Termin-Vorkommen.
type EventAttendance struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	EventKey  string    `gorm:"not null;uniqueIndex:idx_att_event_player" json:"eventKey"`
	PlayerID  uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_att_event_player" json:"playerId"`
	Status    string    `gorm:"not null;default:attending" json:"status"` // attending | declined
	Reason    string    `json:"reason"`
	UpdatedAt time.Time `json:"updatedAt"`
}
