# Niedduty v2 – CLAUDE.md

> **Aktueller Arbeitsstand & Erkenntnisse (fussball.de-Decode, Setup, TODO): siehe [HANDOFF.md](HANDOFF.md).**

> Vereins-Schaltzentrale **nur für ASG Aramäer Ahlen** (kein Multi-Tenant).
> Fokus: Ligatabelle · Kalender · Trainingsbeteiligung · Strafenkatalog/Kasse.
> Bewusst gecuttet: Aufstellung, Taktik, Übungskatalog, Fan-Feature, News.

## Projekt-Basics

- **Sprache**: UI-Strings, Kommentare, Commits → **STRENG DEUTSCH**. Variablennamen → Englisch.
- **Inhaber**: Alessandro Nieddu (= ADMIN im System)
- **Nutzung**: hauptsächlich **mobil** — Mobile-First denken (Bottom-Nav < 900px).

## Tech-Stack

| Schicht | Tool |
|---|---|
| Backend | **Go 1.26** (gin + gorm), Layout wie `~/projects/arbeit/inventory-easym` |
| Datenbank | PostgreSQL via Podman (`podman start niedduty-postgres`), DB **`niedduty2`** |
| Frontend | **Vue 3** (`<script setup>` + Composition API), Vite, Pinia, vue-router, TypeScript |
| Animation | **anime.js v4** (`import { animate, stagger } from 'animejs'`) — Helfer in `src/lib/motion.ts` |
| Icons | lucide-vue-next |
| Fonts | Saira Condensed (Display) · Saira (Body) · Chivo Mono (Zahlen) via @fontsource |

Go liegt in `~/.local/go/bin` (ggf. `export PATH=$PATH:~/.local/go/bin`).

## Dev-Commands

```bash
podman start niedduty-postgres          # DB zuerst
go run ./cmd/server                     # Backend :8080 (migriert + seedet automatisch)
cd frontend && npm run dev              # Vite :5174, Proxy /api → :8080
cd frontend && npx vue-tsc -b           # Typecheck
go test ./...                           # Go-Tests (fussball.de-Tests brauchen Netz, -short überspringt sie)
```

**Produktion** (ein Binary, Frontend eingebettet):

```bash
cd frontend && npm run build            # baut nach internal/web/dist
go build -o niedduty ./cmd/server       # Binary mit Frontend
podman compose up -d --build            # oder: App + PostgreSQL per compose.yaml
```

Demo-Login: `admin@aramaeer-ahlen.de` / `demo1234!`

## Struktur

```
cmd/server/main.go            # Bootstrap (+ Sync-Loop, Push-Loop, Frontend-Mount)
internal/config/              # Env (DATABASE_URL, LISTEN_ADDR, PRODUCTION, COOKIE_SECURE, TRUSTED_PROXIES)
internal/models/models.go     # Alle gorm-Modelle + Rollen/Rechte-Konstanten
internal/store/               # Open+AutoMigrate, seed.go (Demo-Daten wenn DB leer)
internal/middleware/auth.go   # Session-Cookie (ndt_session), RequireAdmin, RequirePerm
internal/api/                 # Ein File pro Ressource, api.go = Routen
├── fussball.go               # Tabellen-Sync + Spiele (Widget-Daten)
├── scouting.go               # Gegner-Steckbrief + Kaderstatistik
└── push.go                   # Abos + Erinnerungs-Loop
internal/fussball/            # fussball.go = Widgets (next.fussball.de), classic.go = ajax.team.*
internal/push/                # VAPID-Schlüssel + Versand
internal/web/                 # Eingebettetes Frontend (dist wird von Vite befüllt)
frontend/src/
├── App.vue                   # Mobile-Shell: Top-Bar (Logo · Instagram · Menü) + Bottom-Tabbar
├── views/                    # Login, Dashboard(Start), Liga, Termine, Strafen(=Kasse), Kader, Beteiligung, Einstellungen(=Verwaltung)
├── components/               # OpponentCard (Gegner), MatchCard, ScoreBoard, AppModal (Bottom-Sheet)
├── stores/auth.ts            # user, club, can(perm)
├── services/api.ts           # axios /api
├── lib/motion.ts             # enterRows, countUp, growBars (respektiert reduced-motion)
├── lib/push.ts               # Service Worker + Push-Abo
└── styles/main.css           # KOMPLETTES Design-System „Flutlicht v2" (Tokens + Klassen)
frontend/public/sw.js         # Service Worker (nur Push, bewusst kein Offline-Cache)
```

Routen/Tabs: `/`=Start · `/liga` · `/termine` · `/strafen`(Label „Kasse"); über Menü: `/beteiligung` (Recht
`beteiligung`), `/kader`, `/verwaltung` (Admin).
Alte Pfade (`/tabelle`,`/kalender`,`/einstellungen`,`/training`) leiten weiter. **Nur mobil** — auf breiten
Screens zentrierte Handy-Spalte (max 480px), keine Desktop-Sidebar mehr.

## Rollen & Rechte

- **ADMIN** (Alessandro): alles — Einstellungen, Konten+Rechte, Kader, Tabelle.
- **MEMBER** + vergebbare Rechte (`users.permissions` jsonb):
  - `strafen` — Strafenkatalog pflegen + Strafen aufschreiben
  - `termine` — Termine anlegen/ändern/löschen, fremde Zu-/Absagen setzen
  - `beteiligung` — Trainingsbeteiligung aller sehen (Training-Seite)
- Spieler-Konto via `users.player_id` mit Kader-Eintrag verknüpfen → darf eigene Zu-/Absage setzen.
- **Immer serverseitig prüfen** (`middleware.RequireAdmin` / `RequirePerm`), Frontend-Gating ist nur Komfort.

## Fachliches

- **Termine**: `events.date` als `YYYY-MM-DD`-Text; Wiederholung weekly/biweekly. API expandiert zu Occurrences; `eventKey` = `ID` bzw. `ID_YYYY-MM-DD`. Attendance hängt am `eventKey`.
- **Beteiligung**: `GET /api/attendance/stats?from&to` zählt nur Trainings-Vorkommen bis heute.
- **Strafen**: Beträge in **Cent**; beim Zuweisen wird Label+Betrag kopiert (Katalog-Änderungen verfälschen alte Strafen nicht).
  Katalog-Einträge mit `perUnit` (+ `unitLabel`, z. B. „Minuten") fragen beim Aufschreiben eine Menge ab —
  Betrag × Menge, Menge steht im kopierten Label.
- **Kassen-Ausgaben**: `models.Expense` (`/api/expenses`) = Geld raus (Bälle, Mannschaftsabend). Kassenstand =
  bezahlte Strafen − Ausgaben, kommt aus `/api/player-penalties/summary`. Ausgaben sind in der UI **gelb**
  (`--warn`), nicht grün/rot; jede Bewegung landet im `penalty_log`.
- **Gegner-Scouting**: `GET /api/fussball/scouting` → nächstes Spiel + Gegner (Tabellenzeile, Formkurve aus
  dessen Spielplan, frühere Duelle, ein Satz Klartext). 30-Min-Cache.
- **Spielerstatistik**: `GET /api/fussball/squad-stats?saison=2526` → Einsätze/Einsatzminuten/Tore aus
  `ajax.team.squad`. Namen sind font-verschleiert (dekodiert), die Zahlen kommen im Klartext. 6-h-Cache.
- **Push**: VAPID-Paar liegt in `settings` (oder per `VAPID_*`-Env). Erinnerungen: Vorschau an alle *ohne*
  Rückmeldung (Spiel −24 h, Training −5 h), Kurz-vorher an alle (−3 h). `push_reminders` verhindert Doppler.
- **Liga/fussball.de**: drei Widget-URLs am Club — `fussballDeWidget` (Tabelle), `fussballDeUpcoming` (kommende Spiele), `fussballDeResults` (Ergebnisse), alle mit Prefix `https://www.fussball.de/`. LigaView bettet sie per iframe ein (Segmented-Control). Tabelle zusätzlich manuell pflegbar (`PUT /api/table` ersetzt komplett) als Fallback.
- **Google-Kalender**: optionaler `club.googleCalendarUrl` (Prefix `https://calendar.google.com/`) → Button in Termine; zusätzlich pro Termin ein „Zu Google Kalender"-Deep-Link.
- **Trainingsbeteiligung**: `/beteiligung` (Recht `beteiligung`), Daten aus `/api/attendance/stats`.
- **Neuerungen-Notiz**: bei jedem Release `frontend/src/lib/changelog.ts` pflegen (Version = Datum + 2–3 Sätze
  für die Mannschaft). Gelesen wird am Konto gemerkt (`users.seen_changelog`).

## Design-System „Flutlicht v2"

Wappen-Farben: **Gold `#f4b125` · Rot `#d63a35` · warmes Schwarz `#0b0906`**. Logo: `frontend/public/logo.png`.

- Alle Tokens + Klassen in `styles/main.css` — **keine neuen Hex-Werte in Komponenten**, keine neuen Klassen erfinden wenn `.card/.btn/.tbl/.chip/.field/.board` reichen.
- Flächen-Tokens heißen historisch `--rasen-*` (sind inzwischen warmes Schwarz).
- **Ein Signature-Element pro Screen**: Start `.ticket` (Matchday-Ticket) + `.scoreboard` · Liga `.pos-card`
  mit `.formchip` und `.score-tiles` · Termine `.evcard`/`.datebox` · Kasse `.receipt` · Kader `.jersey`.
- Zahlen **immer** Chivo Mono + tabular, Count-up via anime.js.
- Motion: Enter 160–260ms ease-out, Listen-Stagger 40ms, alles über `lib/motion.ts` (reduced-motion beachtet).
- Schrift nie unter 11px.

## Sicherheits-Regeln

- Session: httpOnly-Cookie, 30 Tage, Token 32 Byte random.
- Alle Inputs validieren (gin binding + Whitelists: Position, Status, Event-Typ, Rechte).
- Externe URLs (Widget) nur mit `https://www.fussball.de/`-Prefix.
- Passwörter: bcrypt, min. 8 Zeichen.

## RELEASE.md

Bei jeder Änderung pflegen — pro Änderung eine Zeile mit Typ-Marker (✨ Feature · 🐛 Fix · 💄 UI · ⚙️ Technik).
