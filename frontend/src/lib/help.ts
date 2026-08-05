// Inhalte für Tutorial und Hilfe kommen aus einer Quelle — sonst läuft die
// Erklärung irgendwann der App hinterher.
//
// `perm` blendet Kapitel aus, die für das Konto nicht gelten:
// leer = alle, sonst das nötige Recht bzw. "admin".

export interface HelpChapter {
	key: string;
	icon: string; // lucide-Name
	title: string;
	lead: string;
	points: string[];
	perm?: 'strafen' | 'termine' | 'beteiligung' | 'admin';
}

export const helpChapters: HelpChapter[] = [
	{
		key: 'start',
		icon: 'Home',
		title: 'Start',
		lead: 'Alles Wichtige auf einen Blick, sobald du die App öffnest.',
		points: [
			'Das Matchday-Ticket zeigt das nächste Pflichtspiel mit Countdown.',
			'„Bin dabei" oder „Absage" direkt vom Ticket — spart den Umweg über die Termine.',
			'Darunter Platz, Punkte und offene Kasse, dazu der Steckbrief des nächsten Gegners.'
		]
	},
	{
		key: 'liga',
		icon: 'Trophy',
		title: 'Liga',
		lead: 'Tabelle und Spiele kommen direkt von fussball.de.',
		points: [
			'Die Tabelle aktualisiert sich von selbst, alle 30 Minuten.',
			'Unter „Spiele" stehen die letzten Ergebnisse und die nächsten Ansetzungen.',
			'Ein Tipp auf ein Spiel öffnet die fussball.de-Seite mit allen Details.'
		]
	},
	{
		key: 'termine',
		icon: 'CalendarDays',
		title: 'Termine',
		lead: 'Training, Spiele und Mannschaftsabende an einer Stelle.',
		points: [
			'Zusage oder Absage mit einem Tipp — die Zähler daneben zeigen den Stand.',
			'Nach einer Absage fragt die App nach einem Grund. Freiwillig, hilft aber beim Planen.',
			'Wer namentlich zu- oder abgesagt hat, sieht nur der Mannschaftsrat. Alle anderen sehen nur die Zahlen.',
			'Bei Spielen stehen Treffpunkt (1:30 vor Anpfiff), Platz und Navi-Knöpfe für Google und Apple Karten.',
			'Der Reiter „Kalender" zeigt den ganzen Monat, du kannst beliebig weit vorspulen.',
			'Tippen auf einen Termin öffnet die Details: Anfahrt, Notiz, Bearbeiten und Löschen.'
		]
	},
	{
		key: 'trainingszeiten',
		icon: 'Repeat',
		title: 'Trainingszeiten',
		lead: 'Die festen Wochentage pflegst du an einer Stelle.',
		points: [
			'Wochentage antippen — für jeden läuft eine wöchentliche Serie.',
			'Zeit, Ort und eine Notiz gelten für alle Einheiten.',
			'Einzelne Einheit abweichend? Das Zettel-Symbol an der Terminzeile setzt eine Notiz nur für diesen Tag.',
			'Abgewählte Wochentage werden samt Rückmeldungen entfernt.'
		],
		perm: 'termine'
	},
	{
		key: 'abstimmungen',
		icon: 'Vote',
		title: 'Abstimmungen',
		lead: 'Fragen an die Mannschaft — laufende stehen auch auf der Startseite.',
		points: [
			'Antippen genügt; die eigene Stimme lässt sich bis zum Ende ändern.',
			'Das Ergebnis siehst du, sobald du selbst abgestimmt hast.',
			'Starten darf der Mannschaftsrat (Recht „Abstimmungen starten"). Alle bekommen sofort eine Benachrichtigung, 24 Stunden vor Ablauf gibt es eine Erinnerung für alle, die noch fehlen.'
		]
	},
	{
		key: 'kasse',
		icon: 'Wallet',
		title: 'Kasse',
		lead: 'Der Strafenkatalog und wer noch zahlen muss.',
		points: [
			'Ohne Sonderrecht siehst du nur deine eigenen Strafen und die Gesamtsumme der Mannschaft.',
			'Aufschreiber wählen mehrere Spieler und Vergehen auf einmal aus.',
			'Bezahlt setzen geht mit einem Tipp, versehentlich Bezahltes lässt sich wieder öffnen.',
			'„Status" erzeugt ein Bild für die WhatsApp-Gruppe.'
		]
	},
	{
		key: 'protokoll',
		icon: 'ShieldCheck',
		title: 'Kassen-Protokoll',
		lead: 'Jede Bewegung ist nachvollziehbar — auch für die Aufschreiber.',
		points: [
			'Aufschreiben, Löschen, Bezahlt-Setzen und Katalog-Änderungen landen mit Konto und Uhrzeit im Protokoll.',
			'Einträge lassen sich über die App nicht ändern oder löschen.',
			'Jeder Eintrag hängt per Prüfsumme am vorherigen. Oben steht, ob die Kette unversehrt ist.'
		],
		perm: 'strafen'
	},
	{
		key: 'beteiligung',
		icon: 'BarChart3',
		title: 'Beteiligung',
		lead: 'Wer kommt wie oft — mit Verlauf über die Monate.',
		points: [
			'Quote je Spieler für 30 Tage, 3 Monate oder die Saison.',
			'Die Diagramme zeigen den Monatsverlauf, den stärksten Wochentag und die Kasse über die Zeit.',
			'Gezählt werden nur Einheiten bis heute — kommende Termine gelten nicht als verpasst.'
		],
		perm: 'beteiligung'
	},
	{
		key: 'kader',
		icon: 'Users',
		title: 'Kader',
		lead: 'Mannschaft, Rückennummern und die Statistik von fussball.de.',
		points: [
			'Position, Status und Rückennummer pflegst du direkt am Spieler.',
			'Unter „Statistik" stehen Einsätze, Einsatzminuten und Tore je Saison.',
			'Der Geburtstag sorgt für die automatische Erinnerung an die Mannschaft.'
		],
		perm: 'admin'
	},
	{
		key: 'push',
		icon: 'Bell',
		title: 'Benachrichtigungen',
		lead: 'Du bestimmst, wann dein Handy sich meldet.',
		points: [
			'Im Menü einschalten — auf dem iPhone erst, wenn die App auf dem Startbildschirm liegt.',
			'Vorlaufzeiten für Training, Spiel und Treffpunkt stellst du selbst ein, jede lässt sich abschalten.',
			'„Rückmeldung erbitten" kommt nur, solange du weder zu- noch abgesagt hast.'
		]
	},
	{
		key: 'verwaltung',
		icon: 'Settings',
		title: 'Verwaltung',
		lead: 'Konten, Rechte und die Einbettungen von fussball.de.',
		points: [
			'Einladungslink erzeugen und teilen — jeder registriert sich selbst und bekommt einen Kader-Eintrag.',
			'Rechte vergeben: Strafen aufschreiben, Termine verwalten, Beteiligung sehen, Abstimmungen starten.',
			'IBAN, Instagram und die fussball.de-IDs stehen ebenfalls hier.'
		],
		perm: 'admin'
	}
];
