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
	PermUmfragen    = "umfragen"    // Abstimmungen starten und beenden
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
	// FussballTeamId — dauerhafte Mannschafts-ID von www.fussball.de. Wird beim
	// Spiele-Abruf automatisch erkannt und dient als Schlüssel für Kaderstatistik
	// und Gegner-Vergleich (klassische fussball.de-Seiten).
	FussballTeamId string `json:"fussballTeamId"`
	// Optionaler Link auf den geteilten Google-Team-Kalender.
	GoogleCalendarUrl string `json:"googleCalendarUrl"`
	// Optionaler Instagram-Auftritt des Vereins (Kopfleiste + Menü).
	InstagramUrl string `json:"instagramUrl"`
}

// Player — Kader-Eintrag (schlank: Basis für Beteiligung + Strafen).
type Player struct {
	ID       uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Name     string    `gorm:"not null" json:"name"`
	Number   *int      `json:"number"`
	Position string    `gorm:"not null;default:MF" json:"position"` // TW | AB | MF | ST
	Status   string    `gorm:"not null;default:fit" json:"status"`  // fit | verletzt | gesperrt | krank
	// Birthday als "YYYY-MM-DD"-Text (wie events.date) — für die
	// Geburtstagserinnerung zählt nur Tag und Monat.
	Birthday  string    `json:"birthday"`
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
	// Series kennzeichnet automatisch gepflegte Serien. "training" = kommt aus
	// dem Trainingsplan (Wochentage) und wird von dort ersetzt.
	Series    string    `gorm:"index" json:"series"`
	CreatedAt time.Time `json:"createdAt"`
}

// PenaltyLog — fälschungssicheres Protokoll der Kassenbewegungen.
//
// Jede Zuweisung, Löschung und Bezahlt-Änderung landet hier. Die Einträge
// lassen sich über keine Route ändern oder löschen; zusätzlich hängt jeder
// Eintrag per Hash am vorherigen (Kette), sodass auch ein Eingriff direkt in
// der Datenbank auffliegt — `GET /api/penalty-log/verify` prüft das.
type PenaltyLog struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Seq       int64     `gorm:"autoIncrement;uniqueIndex" json:"seq"`
	CreatedAt time.Time `json:"createdAt"`
	// Wer — Name und Alias werden mitgeschrieben, damit das Protokoll auch
	// nach einer Umbenennung oder Kontolöschung lesbar bleibt.
	ActorID    uuid.UUID  `gorm:"type:uuid" json:"actorId"`
	ActorName  string     `json:"actorName"`
	ActorAlias string     `json:"actorAlias"`
	Action     string     `gorm:"not null" json:"action"` // siehe PenaltyAction*
	PlayerID   *uuid.UUID `gorm:"type:uuid" json:"playerId"`
	PlayerName string     `json:"playerName"`
	Label      string     `json:"label"`
	Amount     int        `json:"amount"` // Cent
	PrevHash   string     `json:"prevHash"`
	Hash       string     `json:"hash"`
}

// Aktionen im Kassen-Protokoll.
const (
	PenaltyActionAssign   = "aufgeschrieben"
	PenaltyActionDelete   = "geloescht"
	PenaltyActionPaid     = "bezahlt"
	PenaltyActionUnpaid   = "wieder_offen"
	PenaltyActionCatalog  = "katalog_geaendert"
	PenaltyActionCatalogX = "katalog_geloescht"
)

// Poll — Abstimmung für die Mannschaft. Läuft bis EndsAt und erscheint bis
// dahin auf der Startseite.
type Poll struct {
	ID       uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Question string    `gorm:"not null" json:"question"`
	// Options als JSON-Liste. Der Index ist der Schlüssel der Stimme —
	// deshalb dürfen Optionen nach dem Start nicht mehr umsortiert werden.
	Options []string `gorm:"serializer:json;type:jsonb;not null;default:'[]'" json:"options"`
	// MultiChoice erlaubt mehrere Kreuze je Person.
	MultiChoice bool `gorm:"not null;default:false" json:"multiChoice"`
	// EndsAt — danach zählt keine Stimme mehr. Leer = läuft bis zum Schließen.
	EndsAt      *time.Time `json:"endsAt"`
	ClosedAt    *time.Time `json:"closedAt"`
	CreatedBy   uuid.UUID  `gorm:"type:uuid" json:"createdBy"`
	CreatorName string     `json:"creatorName"`
	CreatedAt   time.Time  `json:"createdAt"`
}

// PollVote — eine Stimme. Bei Einfachauswahl genau eine Zeile je Konto,
// bei Mehrfachauswahl eine je angekreuzter Option.
type PollVote struct {
	PollID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"pollId"`
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"userId"`
	OptionIdx int       `gorm:"primaryKey" json:"optionIdx"`
	VoterName string    `json:"voterName"`
	CreatedAt time.Time `json:"createdAt"`
}

// EventNote — Notiz an einem einzelnen Vorkommen (z. B. „heute Torwarttraining").
// Hängt am EventKey, gilt also nur für diesen einen Termin der Serie.
type EventNote struct {
	EventKey  string    `gorm:"primaryKey" json:"eventKey"`
	Text      string    `gorm:"not null" json:"text"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Occurrence — expandiertes Vorkommen eines Termins für einen Datumsbereich.
// EventKey = ID bei Einzelterminen, "ID_YYYY-MM-DD" bei Wiederholungen.
type Occurrence struct {
	Event
	EventKey string `json:"eventKey"`
	OccDate  string `json:"occDate"`
	// OccNote — Notiz nur für dieses Vorkommen (überschreibt nichts am Termin).
	OccNote string `json:"occNote"`
	// Rückmeldungs-Zähler, damit die Liste ohne Extra-Requests auskommt.
	Attending int    `json:"attending"`
	Declined  int    `json:"declined"`
	Open      int    `json:"open"`
	MyStatus  string `json:"myStatus"` // "", "attending", "declined"
	MyReason  string `json:"myReason"`
}

// Setting — schlichter Schlüssel/Wert-Speicher für Dinge, die den Neustart
// überleben müssen (aktuell: VAPID-Schlüsselpaar für Web-Push).
type Setting struct {
	Key   string `gorm:"primaryKey" json:"key"`
	Value string `gorm:"not null" json:"value"`
}

// PushSubscription — ein Gerät, das Push-Benachrichtigungen empfängt.
// Ein Konto kann mehrere Geräte haben (Handy, Tablet).
type PushSubscription struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
	Endpoint  string    `gorm:"uniqueIndex;not null" json:"endpoint"`
	P256dh    string    `gorm:"not null" json:"-"`
	Auth      string    `gorm:"not null" json:"-"`
	UserAgent string    `json:"userAgent"`
	Failures  int       `gorm:"not null;default:0" json:"failures"`
	CreatedAt time.Time `json:"createdAt"`
}

// PushDelivery — verhindert doppelte Erinnerungen. Weil jeder seine eigenen
// Vorlaufzeiten hat, gehört das Konto mit in den Schlüssel.
type PushDelivery struct {
	EventKey string    `gorm:"primaryKey" json:"eventKey"`
	Kind     string    `gorm:"primaryKey" json:"kind"` // vorschau | treffpunkt | kurzvorher | geburtstag
	UserID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"userId"`
	SentAt   time.Time `json:"sentAt"`
}

// PushSetting — persönliche Vorlaufzeiten für Erinnerungen. Ohne Eintrag
// gelten die Vorgaben aus defaultPushSetting.
type PushSetting struct {
	UserID uuid.UUID `gorm:"type:uuid;primaryKey" json:"userId"`
	// Minuten vor Anpfiff bzw. Trainingsbeginn. 0 = diese Erinnerung aus.
	TrainingLeadMin int `gorm:"not null;default:60" json:"trainingLeadMin"`
	MatchLeadMin    int `gorm:"not null;default:180" json:"matchLeadMin"`
	// MeetLeadMin — Minuten vor dem Treffpunkt (Treffpunkt selbst liegt
	// MeetBeforeMin vor dem Anpfiff).
	MeetLeadMin int `gorm:"not null;default:30" json:"meetLeadMin"`
	// Vorschau = Bitte um Rückmeldung, wenn noch keine da ist.
	VorschauSpiel    int  `gorm:"not null;default:1440" json:"vorschauSpiel"`
	VorschauTraining int  `gorm:"not null;default:300" json:"vorschauTraining"`
	Birthdays        bool `gorm:"not null;default:true" json:"birthdays"`
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
