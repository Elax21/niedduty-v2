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

export const changelogVersion = '2026-08-15';

export const changelogTitle = 'Neu in der Kabine';

export const changelogLead = 'Zwei Sachen in der Kasse sind dazugekommen:';

export const changelogPoints: ChangelogPoint[] = [
	{
		title: 'Strafen nach Minuten',
		text: 'Beim Aufschreiben lässt sich jetzt eine Menge wählen — 7 Minuten zu spät sind 7 × 0,50 €. Welche Vergehen so gezählt werden, stellst du im Katalog ein („Betrag je Einheit").'
	},
	{
		title: 'Geld ausgeben',
		text: 'Neuer Tab „Ausgaben" in der Kasse: Grund und Betrag eintragen (z. B. „Bälle gekauft" 50 €), und der Kassenstand oben zieht das ab. Ausgaben sind gelb und stehen wie alles andere im Protokoll.'
	}
];
