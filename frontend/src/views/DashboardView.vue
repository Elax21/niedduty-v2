<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { api } from '../services/api';
import { useRefresh } from '../lib/refresh';
import { useAuthStore } from '../stores/auth';
import { enterRows, countUp } from '../lib/motion';
import { MapPin, Clock, ChevronRight, CalendarDays, Check, X, Cake } from 'lucide-vue-next';
import OpponentCard from '../components/OpponentCard.vue';
import PollCard from '../components/PollCard.vue';
import type { LeagueEntry, Occurrence, Scouting, Poll, Player, Matches, Match, Attendance } from '../types';

const auth = useAuthStore();

const table = ref<LeagueEntry[]>([]);
const events = ref<Occurrence[]>([]);
const openSum = ref(0);
const openShown = ref(0);
const pointsShown = ref(0);
const scouting = ref<Scouting | null>(null);
const polls = ref<Poll[]>([]);
const players = ref<Player[]>([]);
const matchList = ref<Match[]>([]);
/** Rückmeldungen zu fussball.de-Spielen — die hängen an keinem Termin. */
const matchAttendance = ref<Record<string, Attendance[]>>({});

/** Karte meldet ihr neues Ergebnis zurück. */
function onPollChanged(p: Poll) {
	polls.value = p.id ? polls.value.map((x) => (x.id === p.id ? p : x)) : polls.value.filter((x) => x.id);
}
const rsvpBusy = ref(false);

const today = new Date().toISOString().slice(0, 10);

const ownPos = computed(() => {
	const i = table.value.findIndex((e) => e.isOwn);
	return i < 0 ? null : i + 1;
});
const ownEntry = computed(() => table.value.find((e) => e.isOwn) ?? null);
/** Jetzt als "HH:MM" — für die Frage, ob ein Termin von heute schon durch ist. */
function nowHM() {
	return new Date().toTimeString().slice(0, 5);
}

// Spiele stehen nicht in den Terminen, sie kommen von fussball.de. Damit oben
// wirklich der nächste Termin steht — meistens eben das Spiel und nicht das
// Training —, werden sie hier eingemischt. Schlüssel wie in der Termine-Seite
// (`fdm_<id>`), sonst landen Zu- und Absagen in zwei verschiedenen Töpfen.
function matchOccurrence(m: Match): Occurrence {
	const opp = m.home.isOwn ? m.guest.name : m.home.name;
	return {
		id: '', eventKey: 'fdm_' + m.id, occDate: m.isoDate, title: opp, type: 'spiel',
		date: m.isoDate, endDate: '', startTime: m.time, endTime: '',
		location: m.venue?.name || m.venue?.address || '',
		notes: m.home.isOwn ? 'Heimspiel' : 'Auswärtsspiel',
		recurring: false, recurrenceType: '', recurrenceEnd: '', series: 'fussball',
		occNote: '', attending: 0, declined: 0, open: 0, myStatus: '', myReason: ''
	};
}

/** Eigene Termine + Spiele, nach Datum und Uhrzeit sortiert. */
const allEvents = computed(() => {
	const list = [...events.value, ...matchList.value.map(matchOccurrence)];
	// Zähler der Spiele kommen nachgeladen dazu.
	for (const o of list) {
		const att = matchAttendance.value[o.eventKey];
		if (!att) continue;
		o.attending = att.filter((a) => a.status === 'attending').length;
		o.declined = att.filter((a) => a.status === 'declined').length;
		o.myStatus = att.find((a) => a.playerId === myPlayerId.value)?.status ?? '';
	}
	return list.sort((a, b) => (a.occDate + a.startTime).localeCompare(b.occDate + b.startTime));
});

// Oben steht immer der Termin, der wirklich als nächstes kommt: heutige
// Termine fallen raus, sobald sie vorbei sind (Ende, sonst Beginn).
const upcoming = computed(() =>
	allEvents.value
		.filter((e) => {
			if (e.occDate > today) return true;
			if (e.occDate < today) return false;
			const ende = e.endTime || e.startTime;
			return !ende || ende >= nowHM();
		})
		.slice(0, 5)
);
const nextEvent = computed(() => upcoming.value[0] ?? null);
/** Nur Spieler-Konten dürfen sich selbst zu- oder absagen. */
const myPlayerId = computed(() => auth.user?.playerId ?? null);

const typeLabels: Record<string, string> = {
	training: 'Training', spiel: 'Spiel', mannschaftsabend: 'Abend', sonstiges: 'Termin'
};

function euro(cents: number) {
	return (cents / 100).toLocaleString('de-DE', { style: 'currency', currency: 'EUR' });
}
function fmtLong(d: string) {
	return new Date(d + 'T12:00').toLocaleDateString('de-DE', { weekday: 'long', day: '2-digit', month: 'long' });
}
function dayNum(d: string) {
	return new Date(d + 'T12:00').toLocaleDateString('de-DE', { day: '2-digit' });
}
function weekdayShort(d: string) {
	return new Date(d + 'T12:00').toLocaleDateString('de-DE', { weekday: 'short' }).replace('.', '');
}

/** Tage bis zum Termin — zweistellig mit führender Null (Ticket-Optik). */
const daysLeft = computed(() => {
	if (!nextEvent.value) return 0;
	const diff = Math.round(
		(new Date(nextEvent.value.occDate + 'T12:00').getTime() - new Date(today + 'T12:00').getTime()) / 86400000
	);
	return Math.max(0, diff);
});
const daysLabel = computed(() => (daysLeft.value === 0 ? 'Heute' : daysLeft.value === 1 ? 'Tag' : 'Tage'));

// ── Geburtstage ────────────────────────────────────────────────
// Wer heute Geburtstag hat, steht ganz oben — Tag und Monat zählen, das Jahr
// nur fürs Alter (viele Kader-Einträge haben keins).
const birthdayToday = computed(() => players.value.filter((p) => p.birthday?.slice(5, 10) === today.slice(5, 10)));

function age(p: Player): number | null {
	const year = Number(p.birthday?.slice(0, 4));
	if (!year || year <= 1900) return null;
	return Number(today.slice(0, 4)) - year;
}

/** „Markus" · „Markus und Ali" · „Markus, Ali und Sam" */
const birthdayNames = computed(() => {
	const names = birthdayToday.value.map((p) => p.name.split(' ')[0]);
	if (names.length <= 1) return names[0] ?? '';
	return `${names.slice(0, -1).join(', ')} und ${names[names.length - 1]}`;
});
const birthdayText = computed(() => {
	const list = birthdayToday.value;
	if (!list.length) return '';
	const verb = list.length === 1 ? 'hat' : 'haben';
	const jahre = list.length === 1 ? age(list[0]) : null;
	const wird = jahre ? ` — wird ${jahre}` : '';
	return `${birthdayNames.value} ${verb} heute Geburtstag${wird}.`;
});

const greeting = computed(() => {
	const h = new Date().getHours();
	const first = (auth.user?.name || '').split(' ')[0];
	const time = h < 11 ? 'Guten Morgen' : h < 18 ? 'Servus' : 'Guten Abend';
	return `${time}, ${first}`;
});

async function loadEvents() {
	const from = new Date(Date.now() - 86400000).toISOString().slice(0, 10);
	const to = new Date(Date.now() + 60 * 86400000).toISOString().slice(0, 10);
	const [ev, mt] = await Promise.all([
		api.get<Occurrence[]>('/events', { params: { from, to } }),
		api.get<Matches>('/fussball/matches').catch(() => ({ data: null as Matches | null }))
	]);
	events.value = ev.data;
	matchList.value = (mt.data?.next ?? []).filter((m) => m.isoDate && m.isoDate >= today && m.isoDate <= to);
	// Nur die nächsten paar Spiele brauchen Zähler — der Rest steht ohnehin
	// nur klein in der Liste.
	matchAttendance.value = {};
	await Promise.all(matchList.value.slice(0, 3).map(loadMatchAttendance));
}

async function loadMatchAttendance(m: Match) {
	const key = 'fdm_' + m.id;
	try {
		const { data } = await api.get<Attendance[]>('/attendance', { params: { eventKey: key } });
		matchAttendance.value = { ...matchAttendance.value, [key]: data };
	} catch {
		/* Ohne Zähler ist die Karte immer noch brauchbar. */
	}
}

/** Zu-/Absage direkt vom Ticket — optimistisch, danach frisch laden. */
async function setRsvp(occ: Occurrence, status: 'attending' | 'declined') {
	if (!myPlayerId.value || rsvpBusy.value) return;
	rsvpBusy.value = true;
	const before = occ.myStatus;
	occ.myStatus = status;
	try {
		await api.put('/attendance', { eventKey: occ.eventKey, playerId: myPlayerId.value, status });
		await loadEvents();
	} catch {
		occ.myStatus = before;
	} finally {
		rsvpBusy.value = false;
	}
}

async function loadAll() {
	const [t, s] = await Promise.all([
		api.get<LeagueEntry[]>('/table'),
		api.get<{ totalOpen: number; totalPaid: number }>('/player-penalties/summary'),
		loadEvents()
	]);
	table.value = t.data;
	openSum.value = s.data.totalOpen;
	countUp(openSum.value, (v) => (openShown.value = v));
	countUp(ownEntry.value?.points ?? 0, (v) => (pointsShown.value = v));
	requestAnimationFrame(() => enterRows('.dash-anim'));

	// Kader nur für die Geburtstagskarte — darf ruhig nachtröpfeln.
	api.get<Player[]>('/players')
		.then((r) => { players.value = r.data; })
		.catch(() => {});

	// Laufende Abstimmungen — kurz, damit sie oben auffallen.
	api.get<Poll[]>('/polls/running')
		.then((r) => { polls.value = r.data; })
		.catch(() => {});

	// Gegner-Steckbrief nachladen — darf ruhig ein paar Hundert ms brauchen.
	api.get<Scouting>('/fussball/scouting')
		.then((r) => { scouting.value = r.data; })
		.catch(() => {});
}

onMounted(loadAll);
useRefresh(loadAll);
</script>

<template>
	<div class="hello dash-anim">
		<div class="hi">{{ greeting }}</div>
		<div class="date">{{ new Date().toLocaleDateString('de-DE', { weekday: 'long', day: '2-digit', month: 'long' }) }}</div>
	</div>

	<!-- Geburtstag: steht über allem, gilt nur heute -->
	<section v-if="birthdayToday.length" class="bday dash-anim">
		<span class="bday-icon"><Cake :size="20" /></span>
		<span class="grow">
			<span class="t">{{ birthdayText }}</span>
			<span class="s">Gratulieren nicht vergessen.</span>
		</span>
	</section>

	<!-- Signature: Matchday-Ticket -->
	<section v-if="nextEvent" class="ticket dash-anim" :class="nextEvent.type">
		<div class="ticket-top">
			<div class="ticket-line">
				<span class="chip" :class="nextEvent.type">{{ typeLabels[nextEvent.type] }}</span>
				<span class="ticket-meta">{{ fmtLong(nextEvent.occDate) }}</span>
			</div>
			<div class="ticket-body">
				<div style="min-width: 0">
					<div class="ticket-vs">{{ nextEvent.type === 'spiel' ? 'gegen' : 'als nächstes' }}</div>
					<h2 class="ticket-opp">{{ nextEvent.title }}</h2>
				</div>
				<div class="ticket-count">
					<div class="n">{{ daysLeft === 0 ? '––' : String(daysLeft).padStart(2, '0') }}</div>
					<div class="l">{{ daysLabel }}</div>
				</div>
			</div>
			<div class="ticket-facts">
				<span v-if="nextEvent.startTime" class="time"><Clock :size="13" style="vertical-align: -2px" /> {{ nextEvent.startTime }} Uhr</span>
				<span v-if="nextEvent.location"><MapPin :size="13" style="vertical-align: -2px" /> {{ nextEvent.location }}</span>
				<span class="count-yes">{{ nextEvent.attending }} zugesagt</span>
				<span v-if="nextEvent.declined" class="count-no">{{ nextEvent.declined }} ab</span>
				<span v-if="nextEvent.open" class="count-open">{{ nextEvent.open }} offen</span>
			</div>
		</div>
		<div class="perf" />
		<div v-if="myPlayerId" class="ticket-foot">
			<button class="rsvp yes" :class="{ on: nextEvent.myStatus === 'attending' }" :disabled="rsvpBusy" @click="setRsvp(nextEvent, 'attending')">
				<Check :size="16" /> Bin dabei
			</button>
			<button class="rsvp no" :class="{ on: nextEvent.myStatus === 'declined' }" :disabled="rsvpBusy" @click="setRsvp(nextEvent, 'declined')">
				<X :size="16" /> Absage
			</button>
		</div>
		<!-- Ohne verknüpften Kader-Eintrag geht keine Zu-/Absage — sagen statt schweigen. -->
		<RouterLink v-else :to="auth.isAdmin ? '/verwaltung' : '/termine'" class="ticket-foot">
			<span class="rsvp">{{ auth.isAdmin ? 'Konto mit Spieler verknüpfen' : 'Alle Termine' }}</span>
		</RouterLink>
	</section>
	<div v-else class="card dash-anim">
		<div class="empty"><CalendarDays :size="30" class="ic" /><br />Kein Termin geplant.</div>
	</div>

	<!-- Anzeigetafel: Platz · Punkte · Kasse -->
	<div class="scoreboard dash-anim" style="margin-top: 14px">
		<div class="cell platz">
			<div class="n">{{ ownPos ? ownPos + '.' : '–' }}</div>
			<div class="l">Platz</div>
		</div>
		<div class="cell punkte">
			<div class="n">{{ ownEntry ? pointsShown : '–' }}</div>
			<div class="l">Punkte</div>
		</div>
		<div class="cell kasse">
			<div class="n" style="font-size: 20px">{{ euro(openShown) }}</div>
			<div class="l">Offen</div>
		</div>
	</div>

	<!-- Laufende Abstimmungen — oben, damit niemand sie verpasst -->
	<PollCard
		v-for="p in polls"
		:key="p.id"
		:poll="p"
		class="dash-anim"
		style="margin-top: 14px"
		@changed="onPollChanged"
	/>

	<!-- Gegner-Steckbrief zum nächsten Pflichtspiel -->
	<OpponentCard v-if="scouting?.match" :scouting="scouting" class="dash-anim" style="margin-top: 14px" />

	<!-- Nächste Termine -->
	<div class="card dash-anim" style="margin-top: 14px">
		<div class="card-head">
			<h2>Nächste Termine</h2>
			<RouterLink to="/termine" class="meta">Alle <ChevronRight :size="13" style="vertical-align: -2px" /></RouterLink>
		</div>
		<div class="card-body flush">
			<template v-if="upcoming.length">
				<RouterLink v-for="e in upcoming" :key="e.eventKey" to="/termine" class="lrow">
					<span class="mini-date">
						<span class="d">{{ dayNum(e.occDate) }}</span>
						<span class="w">{{ weekdayShort(e.occDate) }}</span>
					</span>
					<span class="bar" :class="e.type" />
					<span class="grow">
						<span class="t">{{ e.title }}</span>
						<span class="s">{{ typeLabels[e.type] }}<template v-if="e.startTime"> · {{ e.startTime }} Uhr</template></span>
					</span>
					<span class="mono count-yes" style="font-size: 12.5px">{{ e.attending }}</span>
				</RouterLink>
			</template>
			<p v-else class="empty">Keine anstehenden Termine.</p>
		</div>
	</div>

	<!-- Mini-Tabelle -->
	<div v-if="table.length" class="card dash-anim" style="margin-top: 14px">
		<div class="card-head">
			<h2>Tabelle</h2>
			<RouterLink to="/liga" class="meta">Komplett <ChevronRight :size="13" style="vertical-align: -2px" /></RouterLink>
		</div>
		<div class="card-body flush">
			<table class="tbl">
				<tbody>
					<tr v-for="(e, i) in table.slice(0, 5)" :key="e.teamName" :class="{ own: e.isOwn }">
						<td class="num" style="width: 30px; color: var(--gold)">{{ i + 1 }}</td>
						<td>{{ e.teamName }}</td>
						<td class="num" style="width: 38px; color: var(--ink-3)">{{ e.played }}</td>
						<td class="num" style="width: 42px; color: var(--gold); font-weight: 700">{{ e.points }}</td>
					</tr>
				</tbody>
			</table>
		</div>
	</div>
</template>

<style scoped>
/* Geburtstagskarte — gold, aber ruhig; sie soll das Ticket nicht überstrahlen */
.bday {
	display: flex;
	align-items: center;
	gap: 12px;
	padding: 13px 15px;
	margin-bottom: 14px;
	border: 1px solid rgba(244, 177, 37, 0.34);
	border-radius: 16px;
	background: var(--gold-bg);
}
.bday-icon {
	width: 38px; height: 38px;
	flex-shrink: 0;
	display: grid; place-items: center;
	border-radius: 12px;
	background: linear-gradient(180deg, var(--gold-soft), var(--gold));
	color: var(--gold-ink);
}
.bday .t { display: block; font-family: var(--font-display); font-size: 15px; font-weight: 600; color: var(--ink); }
.bday .s { display: block; font-size: 12.5px; color: var(--ink-3); margin-top: 2px; }

.hello { margin: 2px 2px 16px; }
.hi { font-family: var(--font-display); font-size: 30px; font-weight: 800; text-transform: uppercase; line-height: 1; }
.date { color: var(--ink-3); font-size: 12.5px; margin-top: 5px; }

/* Kompakte Datumsspalte in der Termin-Vorschau */
.mini-date { width: 40px; flex-shrink: 0; text-align: center; }
.mini-date .d {
	display: block;
	font-family: var(--font-mono);
	font-variant-numeric: tabular-nums;
	font-size: 19px;
	font-weight: 700;
	color: var(--gold);
	line-height: 1;
}
.mini-date .w { display: block; font-size: 10px; color: var(--ink-3); text-transform: uppercase; margin-top: 3px; }
.bar { width: 3px; height: 30px; border-radius: 2px; background: var(--ink-3); flex-shrink: 0; }
.bar.spiel { background: var(--rot-soft); }
.bar.training { background: var(--gruen); }
.bar.mannschaftsabend { background: var(--gold); }

.ticket-foot { color: inherit; }
</style>
