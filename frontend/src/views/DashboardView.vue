<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { api } from '../services/api';
import { useRefresh } from '../lib/refresh';
import { useAuthStore } from '../stores/auth';
import { enterRows, countUp } from '../lib/motion';
import { MapPin, Clock, ChevronRight, CalendarDays, Check, X } from 'lucide-vue-next';
import OpponentCard from '../components/OpponentCard.vue';
import PollCard from '../components/PollCard.vue';
import type { LeagueEntry, Occurrence, Scouting, Poll } from '../types';

const auth = useAuthStore();

const table = ref<LeagueEntry[]>([]);
const events = ref<Occurrence[]>([]);
const openSum = ref(0);
const openShown = ref(0);
const pointsShown = ref(0);
const scouting = ref<Scouting | null>(null);
const polls = ref<Poll[]>([]);

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
const upcoming = computed(() => events.value.filter((e) => e.occDate >= today).slice(0, 5));
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

const greeting = computed(() => {
	const h = new Date().getHours();
	const first = (auth.user?.name || '').split(' ')[0];
	const time = h < 11 ? 'Guten Morgen' : h < 18 ? 'Servus' : 'Guten Abend';
	return `${time}, ${first}`;
});

async function loadEvents() {
	const from = new Date(Date.now() - 86400000).toISOString().slice(0, 10);
	const to = new Date(Date.now() + 60 * 86400000).toISOString().slice(0, 10);
	const { data } = await api.get<Occurrence[]>('/events', { params: { from, to } });
	events.value = data;
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
		<RouterLink v-else to="/termine" class="ticket-foot">
			<span class="rsvp">Alle Termine</span>
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
