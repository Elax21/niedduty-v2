// Drehbuch für den geführten Rundgang beim ersten Einloggen.
//
// Der Rundgang hebt echte Bedienelemente hervor (Tabs, Menü) und lässt sie
// antippen — erklärt wird an der Stelle, an der man später auch tippt. Die
// Texte kommen aus lib/help.ts, damit Hilfe und Rundgang nicht auseinander
// laufen; nur die kurzen Zurufe stehen hier.
//
// `target` ist ein `[data-tour="…"]`-Selektor. Die Attribute stehen in App.vue —
// keine Klassennamen als Anker, die fliegen beim nächsten Umbau raus.

import { helpChapters } from './help';

export type TourPerm = 'strafen' | 'termine' | 'beteiligung' | 'umfragen' | 'admin';

export interface TourStep {
	key: string;
	/** Element, das hervorgehoben wird. Ohne target: mittige Karte. */
	target?: string;
	/** Vor dem Schritt dorthin navigieren. */
	route?: string;
	/** Menü-Sheet muss für diesen Schritt offen (true) bzw. zu (false) sein. */
	menu?: boolean;
	/** Nutzer muss das Element antippen, sonst geht es per „Weiter". */
	tap?: boolean;
	/** Kapitel aus help.ts — liefert Titel, Einleitung und Stichpunkte. */
	chapter?: string;
	/** Überschreibt den Kapitel-Titel/-Text. */
	title?: string;
	lead?: string;
	/** Nur zeigen, wenn das Konto das Recht hat. */
	perm?: TourPerm;
	/** Wie viele Stichpunkte aus dem Kapitel gezeigt werden (mobil: wenige). */
	maxPoints?: number;
}

export const tourSteps: TourStep[] = [
	{
		key: 'intro',
		title: 'Willkommen in der Kabine',
		lead: 'Kurzer Rundgang: Ich leuchte an, was wichtig ist — du tippst drauf. Dauert eine Minute, überspringen geht jederzeit.'
	},
	{
		key: 'kasse',
		route: '/',
		menu: false,
		chapter: 'kasse',
		lead: 'Die App ist die „Kasse": Strafenkatalog und offene Beträge — mehr braucht es nicht.',
		maxPoints: 3
	},
	{
		key: 'menu',
		target: '[data-tour="menu"]',
		menu: false,
		tap: true,
		title: 'Das Menü',
		lead: 'Alles Weitere steckt hier oben rechts. Tippe drauf.'
	},
	{
		key: 'kader',
		target: '[data-tour="menu-kader"]',
		menu: true,
		chapter: 'kader',
		perm: 'admin',
		maxPoints: 2
	},
	{
		key: 'verwaltung',
		target: '[data-tour="menu-verwaltung"]',
		menu: true,
		chapter: 'verwaltung',
		perm: 'admin',
		maxPoints: 3
	},
	{
		key: 'push',
		target: '[data-tour="menu-push"], [data-tour="menu-erinnerungen"]',
		menu: true,
		chapter: 'push',
		maxPoints: 2
	},
	{
		key: 'hilfe',
		target: '[data-tour="menu-hilfe"]',
		menu: true,
		title: 'Und wenn was unklar ist',
		lead: 'Unter „Hilfe & Erklärung" steht alles noch einmal ausführlich — auch der Rundgang lässt sich dort neu starten.'
	}
];

/** Text eines Schritts, zusammengesetzt aus Drehbuch und Hilfe-Kapitel. */
export function stepContent(step: TourStep): { title: string; lead: string; points: string[] } {
	const ch = step.chapter ? helpChapters.find((c) => c.key === step.chapter) : undefined;
	const points = ch ? ch.points.slice(0, step.maxPoints ?? 2) : [];
	return {
		title: step.title ?? ch?.title ?? '',
		lead: step.lead ?? ch?.lead ?? '',
		points
	};
}
