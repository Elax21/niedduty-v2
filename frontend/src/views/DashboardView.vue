<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { api } from '../services/api';
import { useAuthStore } from '../stores/auth';
import { enterRows } from '../lib/motion';
import ScoreBoard from '../components/ScoreBoard.vue';
import type { LeagueEntry, Occurrence, PlayerPenalty, StatsResponse } from '../types';

const auth = useAuthStore();

const table = ref<LeagueEntry[]>([]);
const events = ref<Occurrence[]>([]);
const penalties = ref<PlayerPenalty[]>([]);
const stats = ref<StatsResponse | null>(null);

const today = new Date().toISOString().slice(0, 10);

const ownPos = computed(() => table.value.findIndex((e) => e.isOwn) + 1);
const ownEntry = computed(() => table.value.find((e) => e.isOwn));
const upcoming = computed(() => events.value.filter((e) => e.occDate >= today).slice(0, 6));
const nextEvent = computed(() => upcoming.value[0] ?? null);
const openSum = computed(() =>
	penalties.value.filter((p) => !p.paid).reduce((s, p) => s + p.amount, 0)
);
const avgQuote = computed(() => {
	const ps = stats.value?.players ?? [];
	if (!ps.length) return 0;
	return Math.round(ps.reduce((s, p) => s + p.quotePct, 0) / ps.length);
});
const tableSlice = computed(() => {
	const idx = table.value.findIndex((e) => e.isOwn);
	const start = Math.max(0, Math.min(idx - 2, table.value.length - 5));
	return table.value.slice(start, start + 5).map((e, i) => ({ ...e, pos: start + i + 1 }));
});

const typeLabels: Record<string, string> = {
	training: 'Training', spiel: 'Spiel', mannschaftsabend: 'Abend', sonstiges: 'Termin'
};

function euro(cents: number) {
	return (cents / 100).toLocaleString('de-DE', { style: 'currency', currency: 'EUR' });
}
function fmtDate(d: string) {
	return new Date(d + 'T12:00').toLocaleDateString('de-DE', { weekday: 'short', day: '2-digit', month: '2-digit' });
}

onMounted(async () => {
	const from = new Date(Date.now() - 86400000).toISOString().slice(0, 10);
	const to = new Date(Date.now() + 45 * 86400000).toISOString().slice(0, 10);
	const [t, ev, pp] = await Promise.all([
		api.get<LeagueEntry[]>('/table'),
		api.get<Occurrence[]>('/events', { params: { from, to } }),
		api.get<PlayerPenalty[]>('/player-penalties')
	]);
	table.value = t.data;
	events.value = ev.data;
	penalties.value = pp.data;
	if (auth.can('beteiligung')) {
		const s = await api.get<StatsResponse>('/attendance/stats');
		stats.value = s.data;
	}
	requestAnimationFrame(() => enterRows('.dash-anim'));
});
</script>

<template>
	<div class="page-head">
		<h1>Zentrale</h1>
		<span class="sub">{{ new Date().toLocaleDateString('de-DE', { weekday: 'long', day: '2-digit', month: 'long', year: 'numeric' }) }}</span>
	</div>

	<div class="card dash-anim" style="margin-bottom: 16px">
		<div class="card-body board-row">
			<ScoreBoard :value="ownPos || '–'" label="Tabellenplatz" />
			<ScoreBoard :value="ownEntry?.points ?? 0" label="Punkte" />
			<ScoreBoard :value="euro(openSum)" label="Offene Kasse" />
			<ScoreBoard v-if="stats" :value="avgQuote" suffix="%" label="Beteiligung Ø" />
			<div v-if="nextEvent" class="next-event">
				<div class="board-label" style="margin: 0 0 4px">Nächster Termin</div>
				<div class="next-title">{{ nextEvent.title }}</div>
				<div class="next-meta">
					<span class="chip" :class="nextEvent.type">{{ typeLabels[nextEvent.type] }}</span>
					{{ fmtDate(nextEvent.occDate) }}<template v-if="nextEvent.startTime"> · {{ nextEvent.startTime }} Uhr</template>
					<template v-if="nextEvent.location"> · {{ nextEvent.location }}</template>
				</div>
			</div>
		</div>
	</div>

	<div class="grid cols-2">
		<div class="card dash-anim">
			<div class="card-head"><h2>Nächste Termine</h2><RouterLink to="/kalender" class="meta">Kalender →</RouterLink></div>
			<div class="card-body flush">
				<table v-if="upcoming.length" class="tbl">
					<tbody>
						<tr v-for="e in upcoming" :key="e.eventKey">
							<td style="width: 90px" class="num">{{ fmtDate(e.occDate) }}</td>
							<td>{{ e.title }}<span v-if="e.location" style="color: var(--kreide-45)"> · {{ e.location }}</span></td>
							<td style="text-align: right"><span class="chip" :class="e.type">{{ typeLabels[e.type] }}</span></td>
						</tr>
					</tbody>
				</table>
				<p v-else class="empty">Keine anstehenden Termine. Lege im Kalender neue an.</p>
			</div>
		</div>

		<div class="card dash-anim">
			<div class="card-head"><h2>Tabelle</h2><RouterLink to="/tabelle" class="meta">Komplett →</RouterLink></div>
			<div class="card-body flush">
				<table class="tbl">
					<tbody>
						<tr v-for="e in tableSlice" :key="e.teamName" :class="{ own: e.isOwn }">
							<td class="num" style="width: 34px">{{ e.pos }}</td>
							<td>{{ e.teamName }}</td>
							<td class="num" style="width: 44px">{{ e.played }}</td>
							<td class="num" style="width: 54px; color: var(--gold); font-weight: 600">{{ e.points }}</td>
						</tr>
					</tbody>
				</table>
			</div>
		</div>
	</div>
</template>

<style scoped>
.board-row {
	display: flex;
	gap: 28px;
	flex-wrap: wrap;
	align-items: flex-start;
}
.next-event { margin-left: auto; min-width: 200px; }
.next-title { font-family: var(--font-display); font-size: 19px; font-weight: 700; text-transform: uppercase; }
.next-meta { font-size: 12.5px; color: var(--kreide-70); display: flex; gap: 6px; align-items: center; flex-wrap: wrap; margin-top: 4px; }
@media (max-width: 900px) {
	.board-row { gap: 16px; }
	.next-event { margin-left: 0; }
}
</style>
