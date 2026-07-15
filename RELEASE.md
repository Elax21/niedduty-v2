# Release-Notes

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
