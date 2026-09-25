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

export const changelogVersion = '2026-09-25';

export const changelogTitle = 'Neu in der Kabine';

export const changelogLead = 'Das ist dazugekommen:';

export const changelogPoints: ChangelogPoint[] = [
	{
		title: 'Alles dreht sich um die Kasse',
		text: 'Die App ist jetzt schlank auf den Strafenkatalog und die Kasse zugeschnitten. Start, Liga, Termine und Abstimmungen sind raus — beim Öffnen landest du direkt in der Kasse.'
	},
	{
		title: 'Kompletter Kader ist drin',
		text: 'Der ganze Kader wurde einmal von fussball.de übernommen. So kann jeder eine Strafe bekommen — auch wer sich nie selbst angemeldet hat. Namen, Nummern und Positionen kann der Admin im Kader nachpflegen.'
	}
];
