# Niedduty v2 – Handoff / Arbeitsstand

Stand: 29.07.2026. Diese Datei fasst alles zusammen, um an einem anderen Rechner weiterzuarbeiten. Ergänzt `CLAUDE.md` (Projekt-Basics).

## 1. Was ist passiert (großer Umbau)

- **Komplettes Mobile-Design „Matchday"** (`frontend/src/styles/main.css`): reines Handy-Layout, sticky Top-Bar + Bottom-Tabbar (Start · Liga · Termine · Strafen), Admin-Menü (☰) → Kader/Verwaltung. Desktop-Sidebar entfernt.
- **Einladungs-/Selbstregistrierung mit Alias-Login**: Admin erzeugt in der Verwaltung einen Einladungslink; jeder registriert sich mit Vorname/Nachname/**Alias**/Passwort und bekommt automatisch einen Kader-Eintrag. Login per Alias (Admin-E-Mail geht weiter).
- **Strafen-Privatsphäre**: Normale Spieler sehen serverseitig nur ihre **eigenen** Strafen + die Team-Gesamtsumme; nur Konten mit Recht `strafen` (oder Admin) sehen alles und haben den WhatsApp-Status-Button.
- **fussball.de nativ** (Kern-Feature, siehe §4): Tabelle + Spiele werden serverseitig von fussball.de geholt, dekodiert und **in unserem Design** gerendert (kein iframe, keine Domain-Sperre).
- **Trainings**: Mi + Fr 19:15, Sportpark Nord (im Seed).
- **Sauberer Start**: Seed legt nur Club, Admin (`admin` / `demo1234!`), Strafenkatalog und die Trainings an. **Kein Demo-Kader, keine Demo-Tabelle** — Spieler kommen per Registrierung, Tabelle per fussball.de-Sync.

## 2. Toolchain / Dev starten

**Auf DIESEM Mac** liegen portable Toolchains in `~/.local` (Go 1.26.5 `~/.local/go/bin`, Node 22 `~/.local/node/bin`, PostgreSQL 17 `~/.local/pgsql` mit Daten in `~/.local/niedduty-pgdata`). Auf einem **anderen Rechner** stattdessen Go 1.26 + Node + PostgreSQL installieren (brew: `brew install go node postgresql@17`), DB `niedduty2` anlegen, DSN siehe `internal/config/config.go` (`postgres://root:...@localhost:5432/niedduty2`).

```bash
# DB (dieser Mac)
~/.local/pgsql/bin/pg_ctl -D ~/.local/niedduty-pgdata -l /tmp/pg.log -o "-p 5432 -k /tmp" start
# Backend (:8080, migriert + seedet + startet fussball.de-Tabellensync)
PATH=~/.local/go/bin:$PATH go run ./cmd/server
# Frontend (:5174, Proxy /api -> :8080)
cd frontend && PATH=~/.local/node/bin:$PATH npm run dev
# Typecheck
cd frontend && npx vue-tsc -b
```

**Aufs Handy** (gleiches WLAN): `npm run dev -- --host`, dann `http://<Mac-LAN-IP>:5174/` im Handy-Browser (Safari → „Zum Home-Bildschirm" = App-Feeling). LAN-IP: `ipconfig getifaddr en0`.

## 3. Architektur (Go + Vue)

- Backend: gin + gorm + PostgreSQL. `internal/api/*` je Ressource, Routen in `api.go`. Auth via httpOnly-Cookie `ndt_session`.
- Auth/Invite: `internal/api/auth.go` (Login per Alias/E-Mail, `startSession`), `internal/api/invites.go` (Invite CRUD + `Register`). Modelle: `models.User` hat `Alias` (unique) + optionale `Email *string`; `models.Invite`.
- Frontend: Vue 3 `<script setup>`, Pinia (`stores/auth.ts`: `login`, `register`, `can`), Router mit öffentlichen Routen `login` + `register/:token`.

## 4. fussball.de-Integration (WICHTIG – hier steckt die Arbeit)

Die neuen fussball.de-Widgets (next.fussball.de) sind **iframe-Embeds mit Domain-Sperre** (funktionieren nur auf der konfigurierten Website `niedduty.com`, nicht localhost). Deshalb holen wir die **Daten selbst** und rendern nativ.

**Unsere Widget-IDs** (in `internal/store/seed.go` + Verwaltung hinterlegt):
- Tabelle (`table`): `85bc8155-cd18-449f-b5d5-db1ef7277ab9`
- Spiele letzte+nächste (`team-matches`): `82539853-64c3-4562-8a70-23e05606df0f`
- Nächstes Spiel (`next-match`): `aab8a3a1-12c9-4a0a-bd06-da2911b780ea` (aktuell ungenutzt; team-matches liefert „next" mit)

**Datenquelle**: `https://next.fussball.de/widget/<type>/<id>` → Next.js-Seite mit allen Daten im `<script id="__NEXT_DATA__">`-JSON unter `props.pageProps` (`table.entries` bzw. `previousMatches`/`nextMatches`). Server-Fetch liefert die Daten unabhängig von der Domain-Sperre (Referer `https://niedduty.com/` gesetzt, `invalidReferrer:false`).

**Obfuskation (Anti-Scraping)**: Zahlen/Namen sind mit Private-Use-Codepoints (U+E000–U+F8FF) verschleiert. `pageProps.obfuscatedFont` ist ein Font-Schlüssel; die Font unter `https://www.fussball.de/export.fontface/-/format/ttf/id/<key>/type/font` mappt jeden Codepoint auf eine Glyphe mit **echtem Namen** (z.B. `U+F04E → "zero"`). Dekodierung: Codepoint → Glyphenname (`golang.org/x/image/font/sfnt`, `GlyphName`) → echtes Zeichen (`glyphMap` in `internal/fussball/fussball.go`). **Font-Key wechselt pro Request** → Font immer mit der Antwort zusammen laden.

**Code**:
- `internal/fussball/fussball.go`: `FetchTable(id, ownName)` + `FetchMatches(id, ownName)` (Fetch, `__NEXT_DATA__`-Extraktion, Font-Decode).
- `internal/api/fussball.go`: `SyncTableFromFussball(db)` (schreibt in `league_entries`), `SyncTable` (Admin `POST /api/table/sync`), `StartTableSyncLoop` (Start + alle 30 Min, aus `cmd/server/main.go`), `GetMatches` (`GET /api/fussball/matches`, 10-Min-In-Memory-Cache).
- Frontend: `components/MatchCard.vue`, native Tabelle + Spiele in `views/LigaView.vue`, „Nächstes Pflichtspiel" in `DashboardView.vue`.

**Kein Cron nötig** — Server synct Tabelle selbst; Spiele werden live (mit Cache) geladen. Team-Logos kommen als echte fussball.de-Bild-URLs.

Verifiziert (Saison 2026/27 startet erst, daher 0 Punkte): Tabellensync = 16 echte Kreisliga-A-Beckum-Teams (Aramäer = `isOwn`); Spiele = letztes 11:4 + 5 kommende Termine.

## 4b. fussball.de klassisch (Kaderstatistik + Gegner) — Stand 01.08.2026

Die Widgets liefern nur Tabelle und eigene Spiele. Alles andere holen wir von den **klassischen**
Seiten (`internal/fussball/classic.go`) — die funktionieren mit **jeder** `team-id`:

| Zweck | URL |
|---|---|
| Kaderstatistik | `https://www.fussball.de/ajax.team.squad/-/mode/PAGE/order-by/1/saison/<2526>/show-filter/true/team-id/<id>` |
| Spielplan (gespielt) | `https://www.fussball.de/ajax.team.prev.games/-/mode/PAGE/show-token/false/team-id/<id>` |
| Spielplan (angesetzt) | `.../ajax.team.next.games/...` |

- Spalten der Kadertabelle: **Spieler · Einsätze · Einsatzminuten · Tore**. Nur die **Namen** sind
  verschleiert (`data-obfuscation="<key>"` am Element → gleiche Font-Mechanik wie bei den Widgets),
  die Zahlen kommen im Klartext. Bei den Spielplänen ist es umgekehrt: Namen im Klartext, **Ergebnisse**
  verschleiert. Deshalb dekodiert `keyedDecoder` schlüsselbezogen pro Teilbaum.
- Unsere Mannschafts-ID: `011MIC2EF8000000VTVG0001VTR8C1K7`. Sie steht **unverschleiert** in den
  Widget-Daten (`nextMatches[].homeTeam.teamPermanentId`) und wird beim ersten Spiele-Abruf automatisch
  als `club.fussballTeamId` gespeichert — dieselbe Quelle liefert auch die ID des **Gegners**.
- Saisonkennung: `CurrentSeason(t)` → "2627"; Wechsel am 1. Juli.
- Verifiziert am 01.08.2026: Kaderstatistik 25/26 mit 31 Spielern (Dülek 30 Sp./2700 Min., Sodic 27 Tore),
  Gegner-Form von Ahlener SG II über deren Spielplan.

## 4c. Push (Web-Push, ohne Fremddienst)

- VAPID-Paar wird beim ersten Start erzeugt und in `settings` abgelegt (überlebt Neustarts — sonst
  werden alle Abos ungültig). Vorgeben per `VAPID_PUBLIC_KEY` / `VAPID_PRIVATE_KEY`.
- Erinnerungs-Loop alle 5 Min in `internal/api/push.go`; `push_reminders` (eventKey + kind) verhindert Doppler.
- **Testen geht nur auf einem echten Gerät.** Auf dem iPhone erst über „Teilen → Zum Home-Bildschirm"
  installieren, sonst gibt Safari keine Push-Erlaubnis. Schalter: Menü → „Benachrichtigungen".

## 4d. Hosting

Ein Artefakt: `npm run build` schreibt nach `internal/web/dist`, `go build` bettet es ein, der Server
liefert SPA + API auf einem Port (`internal/web/web.go`, History-Routing inklusive).

```bash
POSTGRES_PASSWORD=… podman compose up -d --build   # App + PostgreSQL
```

Davor gehört ein Reverse-Proxy mit TLS (Caddy/nginx) — die App spricht http auf 8080 und ist in der
compose-Datei nur an `127.0.0.1` gebunden. Env: `PRODUCTION=true` (gin Release + `DATABASE_URL` Pflicht),
`COOKIE_SECURE=true`, `TRUSTED_PROXIES`.

**Vor dem ersten echten Deploy noch zu tun:** das Fallback-Passwort in `internal/config/config.go`
(`postgres://root:…`) steht im Repo — für Produktion zwingend `DATABASE_URL` setzen (macht `PRODUCTION=true`
auch verpflichtend) und das Admin-Passwort `demo1234!` aus dem Seed ändern.

## 5. Offen / Ideen

- Push auf dem eigenen Handy durchtesten (Einschalten → Probenachricht → echte Erinnerung vor dem Training).
- Optional: `fussballNextMatchId`-Feld/`next-match`-Widget entfernen (ungenutzt), da team-matches „next" mitliefert.
- Optional: Kaderstatistik mit den eigenen Kader-Einträgen verknüpfen (Namensabgleich), damit Strafen,
  Beteiligung und Einsatzminuten auf einer Spielerkarte zusammenlaufen.
- Optional: fussball.de-Widget-Farben (Erweiterte Einstellungen) — für den Fall, dass man doch iframes nutzen will; Speichern erfordert fussball.de-Login.
- Deploy: In Produktion auf `niedduty.com` würden auch die originalen fussball.de-iframes rendern — aktuell aber bewusst durch native Darstellung ersetzt.

## 6. Git

Zum Weiterarbeiten am anderen Rechner: committen + pushen (aktuell **nicht** committed). Branch `main`. RELEASE.md pflegen.
