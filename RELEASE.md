# Release-Notes

## 15.08.2026 — Startseite: Spiele, Geburtstage · Kasse filtern

- ✨ **Oben steht der Termin, der wirklich als nächstes kommt**: Die Startseite mischt die fussball.de-Spiele unter die eigenen Termine (Schlüssel `fdm_<id>` wie in der Termine-Seite, damit Zu-/Absagen nicht in zwei Töpfen landen). Heutige Termine fallen aus dem Ticket, sobald sie vorbei sind (Ende, sonst Beginn) — vorher stand das Training noch abends um elf oben.
- ✨ **Geburtstagskarte** auf der Startseite: „Markus hat heute Geburtstag — wird 27." Tag und Monat entscheiden, das Jahr liefert nur das Alter (viele Kader-Einträge haben keins). Mehrere an einem Tag stehen in einer Zeile.
- ✨ **Kasse filtern** (Recht `strafen`): Auswahl nach Spieler samt offenem Betrag, dazu ein Schalter „Nur offen".
- 💄 **Ticket ohne Kader-Verknüpfung**: Konten ohne verknüpften Spieler sahen nur „Alle Termine" und keinen Grund dafür. Admins kommen jetzt von dort direkt in die Verwaltung.

## 15.08.2026 — Minuten-Strafen und Ausgaben aus der Kasse

- ✨ **Strafen je Minute** (`penalties.per_unit` + `unit_label`): Katalog-Einträge lassen sich als „Betrag je Einheit" markieren. Beim Aufschreiben erscheint dann ein Mengen-Zähler (z. B. Minuten Verspätung), der Betrag wird multipliziert und die Menge landet im Text der Strafe („Verspätung Training (7 Minuten)"). Bestehende Einträge mit Zusatz „pro Minute" werden beim ersten Start einmalig umgestellt (Merker `migration.perUnitMinuten` in `settings`).
- ✨ **Geld ausgeben** (`models.Expense`, `GET/POST/DELETE /api/expenses`, Tab „Ausgaben" in der Kasse): Ausgabe mit Grund, Betrag und Datum eintragen — z. B. „Bälle gekauft" 50 €. Sehen dürfen es alle, eintragen nur mit Recht `strafen`. Jede Ausgabe und jede Löschung steht im Kassen-Protokoll (Hash-Kette).
- ✨ **„Neu in der Kabine"** (`components/ChangeNotes.vue`, Texte + Version in `lib/changelog.ts`): kurze Notiz beim Öffnen der App, sobald es Neuerungen gibt — einmal je Version. Der Merker hängt wie beim Rundgang am Konto (`users.seen_changelog`, `PUT /api/auth/me/changelog`, nur die eigene Session), nicht am Gerät. Wer den Rundgang noch vor sich hat, sieht die Notiz nicht; für den ist die App ohnehin neu.
- 💄 **Kassenzettel zeigt jetzt Ausgaben und Kassenstand** (bezahlt − ausgegeben). Geld, das die Kasse verlässt, ist durchgehend **gelb** (`--warn`) statt grün/rot — Liste, Protokollzeile und der Plus-Button im Ausgaben-Tab; ein negativer Kassenstand bleibt rot.

## 12.08.2026 — Geführter Rundgang

- ✨ **Geführter Rundgang beim ersten Einloggen** (`components/TourGuide.vue`, Drehbuch in `lib/tour.ts`): die App dunkelt ab, hebt ein echtes Bedienelement hervor (Tab „Start", „Liga", „Termine", „Kasse", dann das Menü mit Abstimmungen, Beteiligung, Kader, Verwaltung, Benachrichtigungen, Hilfe) und erklärt es direkt daneben. Bei den Tabs muss man selbst tippen — so lernt man den Weg statt einer Bildergeschichte; „Weiter" bleibt als Ausweg. Schritte, für die das Konto kein Recht hat, fallen raus. Anker sind `data-tour`-Attribute, keine Klassennamen.
- ✨ **Merker am Konto statt am Gerät**: `users.tutorial_done` + `PUT /api/auth/me/tutorial` (nur die eigene Session, keine User-ID im Request). Neues Handy heißt nicht neuer Rundgang; der bisherige `localStorage`-Merker ist weg.
- 💄 **Hilfe & Erklärung** zeigt nur noch das Nachschlagewerk, oben mit Knopf „Rundgang neu starten". Texte kommen weiterhin aus `lib/help.ts` — eine Quelle für Rundgang und Hilfe.
- ⚙️ **Nächtliches Backup** (`deploy/doBackup.sh`, in CT 213 unter `/root/bin/doBackup.sh`, Cron 3:00): `pg_dump` von `niedduty2` plus `/etc/niedduty` und die systemd-Unit, gepackt nach `/var/bak/<Tag>/` und per SFTP auf `v2.hetzner.emserver.de` ins Konto `websitebackup` (Chroot `/var/backup/websitebackup/<Tag>/`, 31 Tage Rotation). Nach der Vorlage von CT 205/210, auf PostgreSQL gedreht. Kein Passwort im Script (`pg_dump` als `postgres` über peer-auth), eigener ed25519-Schlüssel mit `restrict` in `authorized_keys`; das Gegenkonto ist per `ForceCommand internal-sftp` schreibend-only (kein Lesen, Löschen, Auflisten). Leerer Dump bricht ab, statt eine gute Sicherung zu überschreiben.

## 05.08.2026 — Live auf niedduty.de, Trainingsplan, Kalender, PWA-Hinweis

- ✨ **Rückmeldungen sind jetzt vertraulich**: wer namentlich zu- oder abgesagt hat, sieht nur noch der Mannschaftsrat (Recht `beteiligung`). Alle anderen bekommen über `GET /api/attendance` ausschließlich die eigene Rückmeldung; die Zähler am Termin (zugesagt/abgesagt/offen) bleiben für alle sichtbar.
- ✨ **Begründung zur Rückmeldung**: nach einer Absage fragt die App freiwillig nach dem Grund, änderbar jederzeit über das Termin-Sheet. Der Mannschaftsrat sieht sie in der Rückmeldungs-Liste hinter dem Termin.
- ✨ **Abstimmungen** (`models.Poll`, `/api/polls`, Seite `/abstimmungen`): Frage mit bis zu zehn Antworten, wahlweise Mehrfachauswahl und festem Ende. Laufende Abstimmungen stehen auch auf der Startseite. Das Ergebnis wird erst nach der eigenen Stimme sichtbar, damit die ersten Stimmen den Rest nicht ziehen; wer wie gestimmt hat, lässt sich aufklappen. Beim Start geht eine Benachrichtigung an alle, 24 Stunden vor Ablauf noch eine an alle, die fehlen. Starten darf, wer das neue Recht **`umfragen`** hat (Mannschaftsrat) — abstimmen darf jeder.
- ✨ **Strafenkatalog 2025/26** aus dem Aushang übernommen (15 Einträge, von „Unentschuldigtes Fehlen beim Spiel" 50 € bis „Kabine unsauber verlassen" 5 €). Ersetzt den bisherigen Demo-Katalog, auch im Seed. Das Feld für den Zusatz (`unit`) fasst jetzt 120 statt 40 Zeichen — die Erläuterungen aus dem Aushang passten sonst nicht.
- 🐛 **„Erinnerungen einstellen" war auf dem iPhone unsichtbar**: der Menüpunkt hing an `pushSupported()`, und `PushManager` gibt es in Safari erst, wenn die App auf dem Startbildschirm liegt. Die Vorlaufzeiten hängen aber am Konto, nicht am Gerät — der Eintrag ist jetzt immer da, mit Hinweis, falls das Gerät noch nicht zustellen kann.

- ⚙️ **Produktiv-Deploy** auf `https://niedduty.de` (LXC 213 „niedduty-v2" auf v20: Go-Binary als systemd-Dienst auf Port 8213, PostgreSQL im selben Container, nginx-Reverse-Proxy in CT 200, Let's-Encrypt-Zertifikat für `niedduty.de` + `www`).
- ✨ **Trainingszeiten** (`GET/PUT /api/training-schedule`, Button in Termine): Wochentage antippen — je Tag läuft eine wöchentliche Serie mit gemeinsamer Zeit, Ort und Notiz. Abgewählte Tage fliegen samt Rückmeldungen raus. Serien tragen `events.series = "training"`.
- ✨ **Notiz je Einheit** (`PUT /api/event-notes`, `models.EventNote`): Notiz an einem einzelnen Vorkommen (`eventKey`), unabhängig vom Rest der Serie — erscheint in Liste und Kalender.
- ✨ **Kalender-Ansicht** in Termine: Monatsraster mit Punkten je Termin-Typ, beliebig vor- und zurückblättern, „Heute"-Sprung, Tagesliste darunter.
- ✨ **„Auf den Startbildschirm"** (`lib/install.ts`, `components/InstallHint.vue`): Flutlicht-Hinweis mit zwei Lichtkegeln aufs Wappen — Chromium bekommt den echten Install-Dialog über `beforeinstallprompt`, iOS die Teilen-Anleitung (Safari vs. Chrome-iOS unterschieden). Kommt genau einmal beim ersten Einloggen, danach nur über den Menüpunkt.
- 🐛 **Neue Version kam in der installierten App nie an**: als PWA wird die Seite nach dem Start nie wieder geladen, ein neues Binary blieb also unsichtbar. Der Server liefert jetzt unter `GET /api/version` einen Fingerabdruck des ausgelieferten Frontends (`web.Version()`); die App merkt sich den Wert beim Start und lädt sich neu, sobald er sich ändert — beim Auffrischen-Knopf, alle 15 Minuten und beim Zurückkehren aus dem Hintergrund.
- ✨ **Auffrischen-Knopf** in der Kopfleiste (`lib/refresh.ts`) — als installierte App gibt es keine Browserleiste; jede Seite meldet ihre Ladefunktion an.
- ✨ **Kasse**: bezahlte Strafen mit einem Tipp wieder öffnen (eigenes Icon) und einzelne Strafen direkt an der Zeile löschen.
- ✨ **Kassen-Protokoll** (`models.PenaltyLog`, `GET /api/penalty-log`, Reiter „Protokoll"): jede Zuweisung, Löschung, Bezahlt-Änderung und Katalog-Änderung wird mit Zeitpunkt, Konto (Name + Alias), Spieler, Bezeichnung und Betrag festgehalten. Einträge sind über keine Route änder- oder löschbar.
- ✨ **Manipulationsschutz**: jeder Protokolleintrag hängt per SHA-256 am Vorgänger (`prevHash`/`hash`). `GET /api/penalty-log/verify` rechnet die Kette nach und nennt den ersten veränderten Eintrag — auch ein Eingriff direkt in der Datenbank fliegt auf. Anhängen ist per Mutex serialisiert, damit die Kette nicht reißt.
- ✨ **Jedes Konto ist auch Spieler**: neu angelegte Konten bekommen automatisch einen Kader-Eintrag und können selbst zu- und absagen — auch Admins und Strafenaufschreiber. Für Trainer/Betreuer gibt es die Option „Kein Kader-Eintrag" (`noPlayer`).
- ✨ **Treffpunkt & Navigation**: die Spielstätte wird von der fussball.de-Spielseite gelesen (`internal/fussball/venue.go` — die Widgets liefern `venue: null`). An jedem kommenden Spiel steht jetzt Treffpunkt (**1:30 vor Anpfiff**), Platzname und ein Knopf für Google Maps bzw. Apple Karten (letzterer nur auf iOS).
- ✨ **Geburtstage**: Feld bei der Registrierung und am Kader-Eintrag (`players.birthday`). Am Tag selbst geht eine Push-Gratulation an alle, die das eingeschaltet haben — inklusive Alter, wenn das Jahr bekannt ist.
- ✨ **Erinnerungen selbst einstellen** (`GET/PUT /api/push/settings`, Menü → „Erinnerungen einstellen"): Vorlauf für Training, Spiel und Treffpunkt getrennt wählbar, jede Erinnerung einzeln abschaltbar, dazu die Rückmelde-Bitte und die Geburtstage. Der Erinnerungs-Loop rechnet jetzt **je Konto**; `push_deliveries` ersetzt `push_reminders` und nimmt das Konto in den Schlüssel auf.
- ✨ **Diagramme in der Beteiligung** (`GET /api/stats/overview`): Beteiligungsquote je Monat, Einheiten je Monat (Training/Spiel getrennt), bester Wochentag und Strafen je Monat (offen/bezahlt). Alles aus PostgreSQL gerechnet, kein zweiter Datenbestand.
- ✨ **Rundgang & Hilfe** (`lib/help.ts`, `components/TourSheet.vue`): nach dem ersten Login führt ein Rundgang einmalig durch alle Funktionen; derselbe Inhalt liegt dauerhaft als „Hilfe & Erklärung" im Menü. Kapitel, für die das Konto kein Recht hat, werden ausgeblendet.
- 🐛 **Bearbeiten, Notiz und Löschen im Termin-Sheet blieben wirkungslos**: das Sheet wurde geschlossen, bevor der Callback lief, und der griff dann auf den bereits geleerten Zustand zu. Der Termin wird jetzt vorher festgehalten. Die Non-null-Assertions im Template hatten den Fehler vor `vue-tsc` verborgen — sie sind raus.
- 💄 **Terminkarte entschlackt**: auf der Karte stehen nur noch Zusage, Absage, Anstoß, Treffpunkt und der Zähler. fussball.de, Notiz, Bearbeiten, Löschen und die Anfahrt liegen im Termin-Sheet — erreichbar über das Menü rechts der Rückmeldung oder durch Tippen auf die Karte. Damit gibt es je Karte genau ein sekundäres Ziel statt bis zu fünf.
- 🗑️ **Google Kalender entfernt**: weder der Deep-Link je Termin noch der „Google ›"-Link in der Kopfzeile. Das Feld `club.googleCalendarUrl` bleibt vorerst in der Verwaltung stehen, wird aber nirgends mehr benutzt.
- 💄 Enge Stellen entzerrt: Zu-/Absage in der Terminliste jetzt breite Primärzeile mit den Werkzeugen darunter (bis zu fünf Icons plus Zähler passten nicht in eine Zeile) · Instagram nur noch im Menü statt doppelt in der Kopfleiste · Navi-Knöpfe brechen über `auto-fit` sauber auf eine Spalte.
- 💄 Das Recht **`beteiligung`** ist jetzt auch in der Rechte-Liste der Verwaltung vergebbar (fehlte in der Oberfläche, obwohl es serverseitig längst geprüft wurde).

## 01.08.2026 — Redesign „Flutlicht v2", Gegner-Scouting, Push, Hosting

- 💄 **Redesign „Flutlicht v2"** (`styles/main.css`): neue Flächen-Tokens (dunkleres warmes Schwarz), `--surface-flat/-inset`, `--tile-black`, `--line-3`, `--gold-ink`. Pro Screen ein Signature-Element: **Matchday-Ticket** (Perforation, Trikot-Streifen, XXL-Countdown, Zu-/Absage direkt drin) · **Anzeigetafel-Leiste** (Platz/Punkte/Kasse als eine Einheit) · **Score-Tiles** · **Kassenzettel** · **Trikot-Nummer**.
- 💄 Liga: eigene Position als große Karte mit **Formkurve** (S/U/N aus den letzten fünf Spielen), letztes Ergebnis als Beleg mit Score-Tiles. Termine: Karten statt Listenzeilen mit farbcodiertem Datumsblock und „Heute"-Highlight. Kader: Positions-Raster als Filter + trikotförmige Rückennummern. Tabbar ohne Gold-Strich, vierter Tab heißt jetzt **Kasse**.
- ✨ **Gegner-Scouting** (`GET /api/fussball/scouting`): nächster Gegner mit Tabellenplatz, Toren, Formkurve, früheren Aufeinandertreffen und einem Satz Klartext — als Karte auf der Startseite.
- ✨ **Spielerstatistik von fussball.de** (`GET /api/fussball/squad-stats`): Einsätze, **Einsatzminuten** und Tore je Spieler, sortierbar, Saison wählbar — im Kader unter „Statistik". Damit muss man für Zahlen nicht mehr auf fussball.de.
- ⚙️ Neuer **Classic-Scraper** `internal/fussball/classic.go` für die klassischen `ajax.team.*`-Seiten (Kaderstatistik, Spielplan **beliebiger** Mannschaften → Gegner-Form). Verschleierte Texte werden schlüsselbezogen über `data-obfuscation` dekodiert. Die dauerhafte Mannschafts-ID wird beim Spiele-Abruf automatisch erkannt und am Verein gespeichert.
- ✨ **Push-Benachrichtigungen** (Web-Push/VAPID, ohne Fremddienst): Vorschau-Erinnerung an alle ohne Rückmeldung (Spiel 24 h, Training 5 h vorher) und Kurz-vorher-Info an alle (3 h). Schalter im Menü, Service Worker + PWA-Manifest, Abo-Verwaltung unter `/api/push/*`. Tote Abos werden automatisch entfernt.
- ✨ **Zusagen-Zähler** kommen direkt mit der Terminliste (`attending`/`declined`/`open`/`myStatus`) — kein Nachladen pro Termin mehr. Zu-/Absage jetzt auch mit einem Tipp vom Start-Ticket.
- ✨ **Trainingsbeteiligung** hat wieder eine Oberfläche (`/beteiligung`, Recht `beteiligung`): Quote je Spieler mit Balken, Zeitraum 30 Tage / 3 Monate / Saison.
- ✨ **Strafen-Rangliste**: Kasse sortiert nach offenem Betrag, Nummern-Badge und Strichliste; Spitzenreiter in Gold.
- ✨ Optionaler **Instagram-Link** am Verein (`club.instagramUrl`, Prefix-Whitelist) → Icon in der Kopfleiste und Eintrag im Menü.
- ⚙️ **Hosting**: Frontend wird ins Go-Binary eingebettet (`internal/web`, Vite baut nach `internal/web/dist`) — ein Artefakt, SPA-Routing inklusive. Dazu `Dockerfile` (Multi-Stage) und `compose.yaml` mit PostgreSQL. Neue Env-Schalter `PRODUCTION`, `COOKIE_SECURE`, `TRUSTED_PROXIES`; in Produktion ist `DATABASE_URL` Pflicht und das Session-Cookie bekommt `Secure`.

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
