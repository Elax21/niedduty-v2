# Release-Notes

## 29.07.2026 — fussball.de nativ, Einladungs-Login, Strafen-Privatsphäre

- ✨ **fussball.de nativ**: Tabelle + Spiele werden serverseitig von `next.fussball.de` geholt, die per Custom-Font verschleierten Daten dekodiert (Glyphennamen) und **im eigenen Design** gerendert — kein iframe, keine Domain-Sperre. Team-Logos inklusive.
- ⚙️ Tabellen-Sync in die DB (Start + alle 30 Min, `POST /api/table/sync` als Admin-Button), Spiele live über `GET /api/fussball/matches` (10-Min-Cache). Neues Paket `internal/fussball`. Kein externer Cron nötig.
- ✨ **Einladungslink + Selbstregistrierung**: Admin erzeugt Link (Verwaltung), Mitglieder registrieren sich mit Vorname/Nachname/**Alias**/Passwort → automatisch Kader-Eintrag. **Login per Alias**. `models.Invite`, `User.Alias`, `POST /auth/register`.
- ✨ **Strafen-Privatsphäre**: Spieler sehen nur eigene Strafen + Team-Gesamtsumme; nur Recht `strafen`/Admin sehen alles + WhatsApp-Status. `GET /player-penalties/summary`.
- 🗑️ Sauberer Seed: nur Club (mit fussball.de-Widget-IDs), Admin, Strafenkatalog, Mi/Fr-Trainings (19:15 Sportpark Nord). Kein Demo-Kader/-Tabelle mehr.

## 29.07.2026 — Design-Overhaul „Matchday" (mobile-only)

- 💄 Komplett neues, eigenständiges Design-System (`styles/main.css`): warmes Schwarz, kräftigeres Wappen-Gold/Rot, Trikot-Diagonale als Signatur, größere Schrift (Basis 15px, nichts < 12px), große Tap-Ziele
- 💄 Reines Mobile-Layout: Desktop-Sidebar entfernt, stattdessen sticky Top-Bar (Logo + Menü) und Bottom-Tabbar (Start · Liga · Termine · Strafen); auf breiten Screens läuft die App als zentrierte „Handy-Spalte"
- ✨ **Liga**-Seite mit Segmented-Control: Tabelle · Nächste Spiele · Ergebnisse — jeweils als fussball.de-Einbettung, manuelle Tabelle als Fallback
- ⚙️ Verein bekommt drei fussball.de-Widget-URLs (Tabelle, kommende Spiele, Ergebnisse) + optionalen Google-Team-Kalender-Link; alle mit Prefix-Validierung
- ✨ **Termine** neu als mobile Monats-Liste (statt Kalendergrid) mit Zu-/Absage, Anlegen/Bearbeiten (inkl. Serien) und „Zu Google Kalender"-Link pro Termin
- 💄 **Start**-Seite als Matchday-Dashboard: Hero mit nächstem Termin + Countdown, KPI-Kacheln (Platz, Punkte, offene Kasse), Vorschau Termine/Tabelle
- 💄 **Strafen/Kasse** mobil überarbeitet (Segmented Kasse/Katalog, Strichliste, WhatsApp-Status, Bulk-Aktionen bleiben), **Kader** & **Verwaltung** mobil überarbeitet, Formulare als Bottom-Sheets
- 🗑️ Trainingsbeteiligung (Quote/Statistik) aus der Oberfläche entfernt — Recht `beteiligung` bleibt serverseitig, ist aktuell ohne UI

## 14.07.2026 — Strafen v2 (Features aus Niedduty 1 übernommen)

- ✨ Strafe aufschreiben: **mehrere Spieler × mehrere Vergehen** auf einmal (max. 60), plus optionale freie Strafe — wie in v1
- ✨ Bulk-Aktionen in der Kasse: Strafen per Checkbox auswählen → gesammelt bezahlt/offen setzen oder löschen
- ✨ **WhatsApp-Status**: Kasse als Bild (1080×1920, Vereinsfarben, Logo, offene Beträge, Summe, IBAN) — mobil direkt über Share-Sheet, Desktop als Download
- ⚙️ API: `POST /player-penalties` nimmt jetzt `playerIds[]` + `penaltyIds[]` + freie Strafe; neue Bulk-Endpoints `/player-penalties/paid` und `/player-penalties/delete`

## 14.07.2026 — Neustart als Go + Vue 3 („Niedduty v2")

- ⚙️ Komplett-Rewrite: Go-Backend (gin/gorm, Architektur wie inventory-easym) + Vue-3-Frontend (Vite, Pinia, anime.js)
- ⚙️ Nur noch ein Verein: ASG Aramäer Ahlen — Multi-Tenancy entfernt
- ✨ Rechtesystem: Admin (Alessandro) vergibt `strafen` / `termine` / `beteiligung` an Mitglieds-Konten (Verwaltung)
- ✨ Ligatabelle: manuelle Pflege im Bearbeiten-Modus **oder** fussball.de-Widget-Einbettung
- ✨ Kalender: Monatsansicht, Serientermine (wöchentlich/14-tägig), Zu-/Absagen mit Grund
- ✨ Trainingsbeteiligung: Quote pro Spieler (Zeitraum wählbar), zählt nur Trainings bis heute
- ✨ Strafenkatalog & Kasse: Katalog, Strafen aufschreiben (auch frei), bezahlt-Toggle, Summen offen/bezahlt
- ✨ Kader: schlanke Spielerliste (Nummer, Position, Status) als Basis für Beteiligung + Strafen
- 💄 Design „Flutlicht in Vereinsfarben": warmes Schwarz, Wappen-Gold, Aramäer-Rot, Anzeigetafel-Zahlen mit Count-up, Vereinslogo
- 💄 Mobile-First: Bottom-Navigation, kompakte Anzeigetafel, Tabellen scrollen horizontal
- 🗑️ Gecuttet: Aufstellung, Taktiktafel, Standards, Übungskatalog/Trainingsplanung, Fan-Feature, News
