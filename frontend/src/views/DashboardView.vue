<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { api } from '../services/api';
import { useAuthStore } from '../stores/auth';
import { enterRows, countUp } from '../lib/motion';
import { MapPin, Clock, ChevronRight, CalendarDays, Wallet } from 'lucide-vue-next';
import MatchCard from '../components/MatchCard.vue';
import type { LeagueEntry, Occurrence, Match, Matches } from '../types';

const auth = useAuthStore();

const table = ref<LeagueEntry[]>([]);
const events = ref<Occurrence[]>([]);
const openSum = ref(0);
const openShown = ref(0);
const nextMatch = ref<Match | null>(null);

const today = new Date().toISOString().slice(0, 10);

const ownPos = computed(() => {
	const i = table.value.findIndex((e) => e.isOwn);
	return i < 0 ? null : i + 1;
});
const ownEntry = computed(() => table.value.find((e) => e.isOwn) ?? null);
const upcoming = computed(() => events.value.filter((e) => e.occDate >= today).slice(0, 5));
const nextEvent = computed(() => upcoming.value[0] ?? null);

const typeLabels: Record<string, string> = {
	training: 'Training', spiel: 'Spiel', mannschaftsabend: 'Abend', sonstiges: 'Termin'
};

function euro(cents: number) {
	return (cents / 100).toLocaleString('de-DE', { style: 'currency', currency: 'EUR' });
}
function fmtLong(d: string) {
	return new Date(d + 'T12:00').toLocaleDateString('de-DE', { weekday: 'long', day: '2-digit', month: 'long' });
}
function fmtShort(d: string) {
	return new Date(d + 'T12:00').toLocaleDateString('de-DE', { weekday: 'short', day: '2-digit', month: '2-digit' });
}
function countdown(d: string): string {
	const diff = Math.round((new Date(d + 'T12:00').getTime() - new Date(today + 'T12:00').getTime()) / 86400000);
	if (diff <= 0) return 'Heute';
	if (diff === 1) return 'Morgen';
	if (diff < 7) return `in ${diff} Tagen`;
	if (diff < 14) return 'nächste Woche';
	return `in ${Math.round(diff / 7)} Wochen`;
}

const tableSlice = computed(() => {
	const idx = table.value.findIndex((e) => e.isOwn);
	if (idx < 0) return table.value.slice(0, 5).map((e, i) => ({ ...e, pos: i + 1 }));
	const start = Math.max(0, Math.min(idx - 1, table.value.length - 4));
	return table.value.slice(start, start + 4).map((e, i) => ({ ...e, pos: start + i + 1 }));
});

const greeting = computed(() => {
	const h = new Date().getHours();
	const first = (auth.user?.name || '').split(' ')[0];
	const time = h < 11 ? 'Guten Morgen' : h < 18 ? 'Servus' : 'Guten Abend';
	return `${time}, ${first}`;
});

onMounted(async () => {
	const from = new Date(Date.now() - 86400000).toISOString().slice(0, 10);
	const to = new Date(Date.now() + 60 * 86400000).toISOString().slice(0, 10);
	const [t, ev, s] = await Promise.all([
		api.get<LeagueEntry[]>('/table'),
		api.get<Occurrence[]>('/events', { params: { from, to } }),
		api.get<{ totalOpen: number; totalPaid: number }>('/player-penalties/summary')
	]);
	table.value = t.data;
	events.value = ev.data;
	openSum.value = s.data.totalOpen;
	countUp(openSum.value, (v) => (openShown.value = v));
	requestAnimationFrame(() => enterRows('.dash-anim'));

	// Nächstes Pflichtspiel separat (live von fussball.de, blockiert nicht).
	api.get<Matches>('/fussball/matches')
		.then((r) => { nextMatch.value = r.data.next[0] ?? null; })
		.catch(() => {});
});
</script>

<template>
	<div class="hello dash-anim">
		<div class="hi">{{ greeting }}</div>
		<div class="date">{{ new Date().toLocaleDateString('de-DE', { weekday: 'long', day: '2-digit', month: 'long' }) }}</div>
	</div>

	<!-- Matchday-Hero -->
	<RouterLink v-if="nextEvent" to="/termine" class="hero dash-anim" :class="nextEvent.type">
		<div class="hero-top">
			<span class="chip" :class="nextEvent.type">{{ typeLabels[nextEvent.type] }}</span>
			<span class="cd">{{ countdown(nextEvent.occDate) }}</span>
		</div>
		<div class="hero-title">{{ nextEvent.title }}</div>
		<div class="hero-meta">
			<span>{{ fmtLong(nextEvent.occDate) }}</span>
			<span v-if="nextEvent.startTime"><Clock :size="13" /> {{ nextEvent.startTime }} Uhr</span>
			<span v-if="nextEvent.location"><MapPin :size="13" /> {{ nextEvent.location }}</span>
		</div>
	</RouterLink>
	<div v-else class="card dash-anim">
		<div class="empty"><CalendarDays :size="30" class="ic" /><br />Kein Termin geplant.</div>
	</div>

	<!-- KPI-Kacheln -->
	<div class="stat-row dash-anim" style="margin-top: 14px">
		<div class="stat">
			<div class="k">Tabellenplatz</div>
			<div class="v">{{ ownPos ? ownPos + '.' : '–' }}</div>
		</div>
		<div class="stat">
			<div class="k">Punkte</div>
			<div class="v">{{ ownEntry?.points ?? '–' }}</div>
		</div>
		<div class="stat rot">
			<div class="k">Offene Kasse</div>
			<div class="v" style="font-size: 20px">{{ euro(openShown) }}</div>
		</div>
	</div>

	<!-- Nächstes Pflichtspiel (nativ von fussball.de) -->
	<div v-if="nextMatch" class="card dash-anim" style="margin-top: 14px">
		<div class="card-head"><h2>Nächstes Pflichtspiel</h2><RouterLink to="/liga" class="meta">Liga <ChevronRight :size="13" style="vertical-align: -2px" /></RouterLink></div>
		<div class="card-body flush"><MatchCard :match="nextMatch" /></div>
	</div>

	<!-- Nächste Termine -->
	<div class="card dash-anim" style="margin-top: 14px">
		<div class="card-head">
			<h2>Nächste Termine</h2>
			<RouterLink to="/termine" class="meta">Alle <ChevronRight :size="13" style="vertical-align: -2px" /></RouterLink>
		</div>
		<div class="card-body flush">
			<template v-if="upcoming.length">
				<RouterLink v-for="e in upcoming" :key="e.eventKey" to="/termine" class="lrow">
					<span class="chip" :class="e.type">{{ typeLabels[e.type] }}</span>
					<span class="grow">
						<span class="t">{{ e.title }}</span>
						<span class="s">{{ fmtShort(e.occDate) }}<template v-if="e.startTime"> · {{ e.startTime }} Uhr</template></span>
					</span>
					<ChevronRight :size="17" style="color: var(--ink-3)" />
				</RouterLink>
			</template>
			<p v-else class="empty">Keine anstehenden Termine.</p>
		</div>
	</div>

	<!-- Mini-Tabelle -->
	<div v-if="tableSlice.length" class="card dash-anim" style="margin-top: 14px">
		<div class="card-head">
			<h2>Tabelle</h2>
			<RouterLink to="/liga" class="meta">Komplett <ChevronRight :size="13" style="vertical-align: -2px" /></RouterLink>
		</div>
		<div class="card-body flush">
			<table class="tbl">
				<tbody>
					<tr v-for="e in tableSlice" :key="e.teamName" :class="{ own: e.isOwn }">
						<td class="num" style="width: 30px; color: var(--gold)">{{ e.pos }}</td>
						<td>{{ e.teamName }}</td>
						<td class="num" style="width: 38px; color: var(--ink-3)">{{ e.played }}</td>
						<td class="num" style="width: 42px; color: var(--gold); font-weight: 700">{{ e.points }}</td>
					</tr>
				</tbody>
			</table>
		</div>
	</div>

	<RouterLink to="/strafen" class="card dash-anim kasse-cta" style="margin-top: 14px">
		<Wallet :size="20" />
		<span class="grow">Zur Mannschaftskasse</span>
		<ChevronRight :size="18" style="color: var(--ink-3)" />
	</RouterLink>
</template>

<style scoped>
.hello { margin: 2px 2px 16px; }
.hi { font-family: var(--font-display); font-size: 26px; font-weight: 700; text-transform: uppercase; line-height: 1; }
.date { color: var(--ink-3); font-size: 12.5px; margin-top: 3px; }

.hero {
	display: block;
	position: relative;
	overflow: hidden;
	border-radius: var(--radius);
	padding: 16px;
	color: var(--ink);
	background:
		linear-gradient(150deg, rgba(214, 58, 53, 0.22), transparent 60%),
		linear-gradient(180deg, var(--surface-3), var(--surface-1));
	border: 1px solid var(--line-2);
	box-shadow: var(--shadow-card);
}
.hero.spiel { background:
	linear-gradient(150deg, rgba(214, 58, 53, 0.34), transparent 62%),
	linear-gradient(180deg, var(--surface-3), var(--surface-1)); }
.hero.training { background:
	linear-gradient(150deg, rgba(87, 192, 125, 0.2), transparent 62%),
	linear-gradient(180deg, var(--surface-3), var(--surface-1)); }
.hero::after {
	content: '';
	position: absolute;
	right: -40px; top: -60px;
	width: 180px; height: 260px;
	transform: rotate(22deg);
	background: repeating-linear-gradient(90deg, rgba(244, 177, 37, 0.06) 0 14px, transparent 14px 28px);
	pointer-events: none;
}
.hero-top { display: flex; align-items: center; justify-content: space-between; }
.cd {
	font-family: var(--font-mono);
	font-size: 13px;
	color: var(--gold);
	background: var(--gold-bg);
	border: 1px solid var(--line-2);
	border-radius: 999px;
	padding: 3px 11px;
}
.hero-title { font-family: var(--font-display); font-size: 27px; font-weight: 700; text-transform: uppercase; line-height: 1.03; margin: 12px 0 8px; }
.hero-meta { display: flex; flex-wrap: wrap; gap: 6px 14px; font-size: 13px; color: var(--ink-2); }
.hero-meta span { display: inline-flex; align-items: center; gap: 4px; }
.hero-meta :deep(svg) { color: var(--gold); }

.kasse-cta {
	display: flex;
	align-items: center;
	gap: 12px;
	padding: 15px;
	font-family: var(--font-display);
	font-size: 16px;
	font-weight: 600;
	text-transform: uppercase;
	letter-spacing: 0.03em;
}
.kasse-cta :deep(svg):first-child { color: var(--gold); }
</style>
