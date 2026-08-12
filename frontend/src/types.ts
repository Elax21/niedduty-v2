export interface User {
	id: string;
	alias: string;
	email: string | null;
	name: string;
	role: 'ADMIN' | 'MEMBER';
	permissions: string[];
	playerId: string | null;
	tutorialDone: boolean;
}

export interface MatchTeam {
	name: string;
	logoUrl: string;
	isOwn: boolean;
	teamId: string;
	clubId: string;
}

export interface Match {
	id: string;
	date: string;
	isoDate: string;
	url: string;
	time: string;
	home: MatchTeam;
	guest: MatchTeam;
	homeGoals: number | null;
	guestGoals: number | null;
	played: boolean;
	venue?: Venue;
}

export interface Matches {
	previous: Match[];
	next: Match[];
}

export interface Invite {
	token: string;
	active: boolean;
	maxUses: number;
	useCount: number;
	expiresAt: string | null;
	createdAt: string;
}

export interface Club {
	id: number;
	name: string;
	short: string;
	primaryColor: string;
	secondaryColor: string;
	kasseIban: string;
	kasseInhaber: string;
	liga: string;
	fussballTableId: string;
	fussballMatchesId: string;
	fussballNextMatchId: string;
	fussballTeamId: string;
	googleCalendarUrl: string;
	instagramUrl: string;
}

export interface Player {
	id: string;
	name: string;
	number: number | null;
	position: 'TW' | 'AB' | 'MF' | 'ST';
	status: 'fit' | 'verletzt' | 'gesperrt' | 'krank';
	birthday: string; // "YYYY-MM-DD" oder ""
}

export interface LeagueEntry {
	id?: string;
	teamName: string;
	isOwn: boolean;
	played: number;
	won: number;
	drawn: number;
	lost: number;
	goalsFor: number;
	goalsAgainst: number;
	points: number;
}

export interface Penalty {
	id: string;
	label: string;
	amount: number; // Cent
	unit: string;
	sortOrder: number;
}

export interface PlayerPenalty {
	id: string;
	playerId: string;
	label: string;
	amount: number; // Cent
	paid: boolean;
	createdAt: string;
}

export interface Occurrence {
	id: string;
	eventKey: string;
	occDate: string;
	title: string;
	type: 'training' | 'spiel' | 'mannschaftsabend' | 'sonstiges';
	date: string;
	endDate: string;
	startTime: string;
	endTime: string;
	location: string;
	notes: string;
	recurring: boolean;
	recurrenceType: string;
	recurrenceEnd: string;
	series: string;
	occNote: string;
	attending: number;
	declined: number;
	open: number;
	myStatus: '' | 'attending' | 'declined';
	myReason: string;
}

/** Ein Eintrag im fälschungssicheren Kassen-Protokoll. */
export interface PenaltyLogEntry {
	id: string;
	seq: number;
	createdAt: string;
	actorName: string;
	actorAlias: string;
	action: string;
	playerName: string;
	label: string;
	amount: number;
}

export interface PenaltyLogCheck {
	ok: boolean;
	count: number;
	brokenAt?: number;
	message: string;
}

/** Feste Trainings-Wochentage (1 = Montag … 7 = Sonntag). */
export interface TrainingSchedule {
	weekdays: number[];
	title: string;
	startTime: string;
	endTime: string;
	location: string;
	notes: string;
	recurrenceEnd: string;
}

export interface Attendance {
	id: string;
	eventKey: string;
	playerId: string;
	status: 'attending' | 'declined';
	reason: string;
}

export interface PlayerStats {
	playerId: string;
	name: string;
	number: number | null;
	attended: number;
	declined: number;
	noAnswer: number;
	total: number;
	quotePct: number;
}

/** Ein Spiel der Formkurve (aus Sicht der jeweiligen Mannschaft). */
export interface FormEntry {
	result: 'S' | 'U' | 'N';
	opponent: string;
	score: string;
	date: string;
	home: boolean;
}

/** Früheres Aufeinandertreffen mit dem Gegner (aus unserer Sicht). */
export interface Meeting {
	date: string;
	score: string;
	result: 'S' | 'U' | 'N' | '';
	home: boolean;
	note: string;
}

/** Steckbrief des nächsten Gegners. */
export interface OpponentInfo {
	name: string;
	logoUrl: string;
	teamId: string;
	position: number;
	played: number;
	won: number;
	drawn: number;
	lost: number;
	goalsFor: number;
	goalsAgainst: number;
	points: number;
	inTable: boolean;
	form: FormEntry[];
	meetings: Meeting[];
	summary: string;
}

export interface Scouting {
	match: Match | null;
	opponent: OpponentInfo | null;
	ownForm: FormEntry[] | null;
	atHome: boolean;
}

/** Kaderstatistik von fussball.de. */
export interface SquadStat {
	name: string;
	matches: number;
	minutes: number;
	goals: number;
	profileUrl: string;
}

export interface SquadStatsResponse {
	season: string;
	players: SquadStat[];
}

export interface StatsResponse {
	from: string;
	to: string;
	trainings: number;
	players: PlayerStats[];
}

/** Spielstätte eines fussball.de-Spiels (von der Spielseite gelesen). */
export interface Venue {
	name: string;
	address: string;
}

/** Persönliche Vorlaufzeiten für Push-Erinnerungen (Minuten). */
export interface PushSettings {
	trainingLeadMin: number;
	matchLeadMin: number;
	meetLeadMin: number;
	vorschauSpiel: number;
	vorschauTraining: number;
	birthdays: boolean;
}

export interface MonthStat {
	month: string;
	label: string;
	trainings: number;
	matches: number;
	attending: number;
	declined: number;
	noAnswer: number;
	quotePct: number;
	avgAttending: number;
}

export interface WeekdayStat {
	weekday: number;
	label: string;
	count: number;
	quotePct: number;
}

export interface KasseMonth {
	month: string;
	label: string;
	open: number;
	paid: number;
	count: number;
}

export interface TopAttender {
	name: string;
	attended: number;
	total: number;
	quotePct: number;
}

export interface StatsOverview {
	months: MonthStat[];
	weekdays: WeekdayStat[];
	kasse: KasseMonth[];
	squadSize: number;
	topPlayers: TopAttender[];
}

/** Abstimmung samt Zählung. */
export interface Poll {
	id: string;
	question: string;
	options: string[];
	multiChoice: boolean;
	endsAt: string | null;
	closedAt: string | null;
	createdBy: string;
	creatorName: string;
	createdAt: string;
	counts: number[];
	voters: string[][];
	myVotes: number[];
	total: number;
	running: boolean;
}
