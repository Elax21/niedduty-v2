// Was ist neu — kurze Notiz beim Öffnen der App, einmal je Version.
//
// `changelogVersion` ist das Datum der Neuerungen. Steht am Konto ein anderer
// (oder gar kein) Wert, zeigt die App die Punkte unten einmal an. Nur die
// Sätze, die für die Mannschaft zählen — Technik bleibt in RELEASE.md.
//
// Beim nächsten Release: Version hochsetzen, Punkte austauschen. Fertig.

export interface ChangelogPoint {
	title: string;
	text: string;
}

export const changelogVersion = '2026-08-15b';

export const changelogTitle = 'Neu in der Kabine';

export const changelogLead = 'Das ist dazugekommen:';

export const changelogPoints: ChangelogPoint[] = [
	{
		title: 'Spiele stehen jetzt oben',
		text: 'Die Startseite zeigt immer den Termin, der wirklich als nächstes kommt — also auch das Spiel vom Sonntag und nicht nur das Training. Spiele stehen außerdem in „Nächste Termine".'
	},
	{
		title: 'Geburtstage',
		text: 'Wer heute Geburtstag hat, steht ganz oben auf der Startseite. Trag den Geburtstag im Kader ein, dann klappt das auch für dich.'
	},
	{
		title: 'Strafen nach Minuten',
		text: 'Beim Aufschreiben lässt sich jetzt eine Menge wählen — 7 Minuten zu spät sind 7 × 0,50 €. Welche Vergehen so gezählt werden, stellst du im Katalog ein („Betrag je Einheit").'
	},
	{
		title: 'Geld ausgeben',
		text: 'Neuer Tab „Ausgaben" in der Kasse: Grund und Betrag eintragen (z. B. „Bälle gekauft" 50 €), und der Kassenstand oben zieht das ab. Ausgaben sind gelb und stehen wie alles andere im Protokoll.'
	},
	{
		title: 'Kasse filtern',
		text: 'Wer Strafen aufschreibt, kann in der Kasse nach einem Spieler filtern oder sich nur die anzeigen lassen, die noch etwas offen haben.'
	}
];
