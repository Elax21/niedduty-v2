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
```

Demo-Login: `admin@aramaeer-ahlen.de` / `demo1234!`

## Struktur

```
cmd/server/main.go            # Bootstrap
internal/config/              # Env (DATABASE_URL, LISTEN_ADDR)
internal/models/models.go     # Alle gorm-Modelle + Rollen/Rechte-Konstanten
internal/store/               # Open+AutoMigrate, seed.go (Demo-Daten wenn DB leer)
internal/middleware/auth.go   # Session-Cookie (ndt_session), RequireAdmin, RequirePerm
internal/api/                 # Ein File pro Ressource, api.go = Routen
frontend/src/
├── App.vue                   # Mobile-Shell: sticky Top-Bar (Logo + Menü) + Bottom-Tabbar
├── views/                    # Login, Dashboard(Start), Liga, Termine, Strafen, Kader, Einstellungen(=Verwaltung)
├── components/               # ScoreBoard (Anzeigetafel-Kachel), AppModal (Bottom-Sheet)
├── stores/auth.ts            # user, club, can(perm)
├── services/api.ts           # axios /api
├── lib/motion.ts             # enterRows, countUp, growBars (respektiert reduced-motion)
└── styles/main.css           # KOMPLETTES Design-System „Matchday" (Tokens + Klassen)
```

Routen/Tabs: `/`=Start · `/liga` · `/termine` · `/strafen`; Admin über Menü: `/kader`, `/verwaltung`.
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
- **Liga/fussball.de**: drei Widget-URLs am Club — `fussballDeWidget` (Tabelle), `fussballDeUpcoming` (kommende Spiele), `fussballDeResults` (Ergebnisse), alle mit Prefix `https://www.fussball.de/`. LigaView bettet sie per iframe ein (Segmented-Control). Tabelle zusätzlich manuell pflegbar (`PUT /api/table` ersetzt komplett) als Fallback.
- **Google-Kalender**: optionaler `club.googleCalendarUrl` (Prefix `https://calendar.google.com/`) → Button in Termine; zusätzlich pro Termin ein „Zu Google Kalender"-Deep-Link.
- **Trainingsbeteiligung**: aktuell **ohne UI** (Recht `beteiligung` + `/api/attendance/stats` existieren serverseitig weiter, sind aber nicht verlinkt).

## Design-System „Flutlicht in Vereinsfarben"

Wappen-Farben: **Gold `#f0a81c` · Rot `#c0272d` · warmes Schwarz**. Logo: `frontend/public/logo.png`.

- Alle Tokens + Klassen in `styles/main.css` — **keine neuen Hex-Werte in Komponenten**, keine neuen Klassen erfinden wenn `.card/.btn/.tbl/.chip/.field/.board` reichen.
- Flächen-Tokens heißen historisch `--rasen-*` (sind inzwischen warmes Schwarz).
- **Signature**: Anzeigetafel (`.board`/ScoreBoard.vue) — Zahlen immer Chivo Mono + tabular, Count-up via anime.js.
- Motion: Enter 160–260ms ease-out, Listen-Stagger 40ms, alles über `lib/motion.ts` (reduced-motion beachtet).
- Schrift nie unter 11px.

## Sicherheits-Regeln

- Session: httpOnly-Cookie, 30 Tage, Token 32 Byte random.
- Alle Inputs validieren (gin binding + Whitelists: Position, Status, Event-Typ, Rechte).
- Externe URLs (Widget) nur mit `https://www.fussball.de/`-Prefix.
- Passwörter: bcrypt, min. 8 Zeichen.

## RELEASE.md

Bei jeder Änderung pflegen — pro Änderung eine Zeile mit Typ-Marker (✨ Feature · 🐛 Fix · 💄 UI · ⚙️ Technik).
