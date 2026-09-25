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
			'Recht vergeben: Strafen aufschreiben (Katalog pflegen und Strafen zuweisen).',
			'IBAN, Instagram und die fussball.de-IDs stehen ebenfalls hier.'
		],
		perm: 'admin'
	}
];
