export interface User {
	id: string;
	alias: string;
	email: string | null;
	name: string;
	role: 'ADMIN' | 'MEMBER';
	permissions: string[];
	playerId: string | null;
}

export interface MatchTeam {
	name: string;
	logoUrl: string;
	isOwn: boolean;
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
	googleCalendarUrl: string;
}

export interface Player {
	id: string;
	name: string;
	number: number | null;
	position: 'TW' | 'AB' | 'MF' | 'ST';
	status: 'fit' | 'verletzt' | 'gesperrt' | 'krank';
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

export interface StatsResponse {
	from: string;
	to: string;
	trainings: number;
	players: PlayerStats[];
}
