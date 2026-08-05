<script setup lang="ts">
/* Trainingsbeteiligung: wer war wie oft da. Zählt nur Trainings-Vorkommen
   bis heute — kommende Termine gelten nicht als verpasst. */
import { ref, computed, onMounted } from 'vue';
import { api } from '../services/api';
import { useRefresh } from '../lib/refresh';
import { enterRows, growBars } from '../lib/motion';
import { BarChart3 } from 'lucide-vue-next';
import type { StatsResponse, StatsOverview } from '../types';

const data = ref<StatsResponse | null>(null);
const overview = ref<StatsOverview | null>(null);
const range = ref<'30' | '90' | '365'>('90');
const loading = ref(true);

const ranges = [
	{ key: '30', label: '30 Tage' },
	{ key: '90', label: '3 Monate' },
	{ key: '365', label: 'Saison' }
] as const;

const rows = computed(() =>
	[...(data.value?.players ?? [])].sort((a, b) => b.quotePct - a.quotePct || b.attended - a.attended)
);
const schnitt = computed(() => {
	const list = data.value?.players ?? [];
	if (!list.length) return 0;
	return Math.round(list.reduce((s, p) => s + p.quotePct, 0) / list.length);
});

function quoteClass(pct: number) {
	if (pct >= 75) return 'gut';
	if (pct < 40) return 'schwach';
	return '';
}

async function load() {
	loading.value = true;
	const days = Number(range.value);
	const from = new Date(Date.now() - days * 86400000).toISOString().slice(0, 10);
	const to = new Date().toISOString().slice(0, 10);
	try {
		const months = days <= 30 ? 3 : days <= 90 ? 6 : 12;
		const [res, ov] = await Promise.all([
			api.get<StatsResponse>('/attendance/stats', { params: { from, to } }),
			api.get<StatsOverview>('/stats/overview', { params: { months } })
		]);
		data.value = res.data;
		overview.value = ov.data;
	} finally {
		loading.value = false;
	}
	requestAnimationFrame(() => {
		enterRows('.bt-row');
		growBars('.bt-row .quota > i');
		growBars('.chart .bar > i');
	});
}

/** Balkenhöhe in Prozent — relativ zum größten Wert der Reihe. */
function relative(value: number, values: number[]): number {
	const max = Math.max(...values, 1);
	return Math.round((value / max) * 100);
}

const monthQuotes = computed(() => (overview.value?.months ?? []).map((m) => m.quotePct));
const monthUnits = computed(() => (overview.value?.months ?? []).map((m) => m.trainings + m.matches));
const kasseValues = computed(() => (overview.value?.kasse ?? []).map((k) => k.open + k.paid));
const hatKasse = computed(() => kasseValues.value.some((v) => v > 0));

function euro(cents: number) {
	return (cents / 100).toLocaleString('de-DE', { style: 'currency', currency: 'EUR', maximumFractionDigits: 0 });
}

function pick(key: '30' | '90' | '365') {
	range.value = key;
	load();
}

onMounted(load);
useRefresh(load);
</script>

<template>
	<div class="page-head">
		<h1>Beteiligung</h1>
		<span class="sub-mono">{{ data?.trainings ?? 0 }} Trainings · Ø {{ schnitt }} %</span>
	</div>

	<div class="segmented">
		<button v-for="r in ranges" :key="r.key" :class="{ active: range === r.key }" @click="pick(r.key)">
			{{ r.label }}
		</button>
	</div>

	<!-- Verlauf über die Monate -->
	<section v-if="overview?.months.length" class="card chart">
		<div class="card-head"><h2>Beteiligung je Monat</h2></div>
		<div class="card-body">
			<div class="bars">
				<div v-for="m in overview.months" :key="m.month" class="bar">
					<span class="val mono">{{ m.quotePct }}%</span>
					<i :style="{ height: relative(m.quotePct, monthQuotes) + '%' }" :class="quoteClass(m.quotePct)" />
					<span class="lbl">{{ m.label }}</span>
				</div>
			</div>
			<p class="legend">Zusagen im Verhältnis zu allen möglichen Teilnahmen ({{ overview.squadSize }} im Kader).</p>
		</div>
	</section>

	<!-- Wie viele Einheiten lagen im Monat -->
	<section v-if="overview?.months.length" class="card chart">
		<div class="card-head"><h2>Einheiten je Monat</h2></div>
		<div class="card-body">
			<div class="bars">
				<div v-for="m in overview.months" :key="m.month" class="bar">
					<span class="val mono">{{ m.trainings + m.matches }}</span>
					<i class="split" :style="{ height: relative(m.trainings + m.matches, monthUnits) + '%' }">
						<u v-if="m.matches" class="matches" :style="{ flexBasis: (m.matches / Math.max(m.trainings + m.matches, 1)) * 100 + '%' }" />
					</i>
					<span class="lbl">{{ m.label }}</span>
				</div>
			</div>
			<p class="legend"><span class="dot training" /> Training <span class="dot spiel" /> Spiel</p>
		</div>
	</section>

	<!-- Welcher Wochentag zieht -->
	<section v-if="overview?.weekdays.length" class="card chart">
		<div class="card-head"><h2>Bester Wochentag</h2></div>
		<div class="card-body">
			<div class="bars">
				<div v-for="w in overview.weekdays" :key="w.weekday" class="bar">
					<span class="val mono">{{ w.quotePct }}%</span>
					<i :style="{ height: relative(w.quotePct, overview.weekdays.map((x) => x.quotePct)) + '%' }" :class="quoteClass(w.quotePct)" />
					<span class="lbl">{{ w.label }}</span>
				</div>
			</div>
			<p class="legend">Über alle Einheiten des Zeitraums.</p>
		</div>
	</section>

	<!-- Kassenverlauf -->
	<section v-if="hatKasse" class="card chart">
		<div class="card-head"><h2>Strafen je Monat</h2></div>
		<div class="card-body">
			<div class="bars">
				<div v-for="k in overview!.kasse" :key="k.month" class="bar">
					<span class="val mono">{{ euro(k.open + k.paid) }}</span>
					<i class="split kasse" :style="{ height: relative(k.open + k.paid, kasseValues) + '%' }">
						<u v-if="k.open" class="offen" :style="{ flexBasis: (k.open / Math.max(k.open + k.paid, 1)) * 100 + '%' }" />
					</i>
					<span class="lbl">{{ k.label }}</span>
				</div>
			</div>
			<p class="legend"><span class="dot bezahlt" /> bezahlt <span class="dot offen" /> offen</p>
		</div>
	</section>

	<div class="card">
		<div class="card-body flush">
			<p v-if="loading" class="empty">Wird geladen …</p>
			<p v-else-if="!rows.length" class="empty">
				<BarChart3 :size="30" class="ic" /><br />
				<template v-if="!data?.trainings">Im Zeitraum lag kein Training.</template>
				<template v-else>Noch niemand im Kader — Spieler kommen über den Einladungslink.</template>
			</p>
			<div v-for="(p, i) in rows" v-else :key="p.playerId" class="bt-row">
				<div class="line">
					<span class="rank mono">{{ i + 1 }}</span>
					<span class="nm">{{ p.name }}</span>
					<span class="pct mono" :class="quoteClass(p.quotePct)">{{ p.quotePct }} %</span>
				</div>
				<div class="quota"><i :class="quoteClass(p.quotePct)" :style="{ width: p.quotePct + '%' }" /></div>
				<div class="detail">
					<span class="count-yes">{{ p.attended }} da</span>
					<span class="count-no">{{ p.declined }} ab</span>
					<span class="count-open">{{ p.noAnswer }} ohne Rückmeldung</span>
				</div>
			</div>
		</div>
	</div>
</template>

<style scoped>
.bt-row { padding: 12px 15px; border-bottom: 1px solid var(--line); }
.bt-row:last-child { border-bottom: none; }
.line { display: flex; align-items: baseline; gap: 10px; }
.rank { width: 20px; font-size: 12px; color: var(--ink-3); text-align: right; }
.nm { flex: 1; min-width: 0; font-size: 15px; font-weight: 600; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.pct { font-size: 14px; font-weight: 700; color: var(--gold); }
.pct.gut { color: var(--gruen); }
.pct.schwach { color: var(--bad); }
.detail { display: flex; gap: 12px; font-size: 11.5px; margin-top: 6px; }

/* ── Diagramme ── */
.chart { margin-bottom: 14px; }
.bars {
	display: flex;
	align-items: flex-end;
	gap: 6px;
	height: 132px;
	padding-top: 16px;
}
.bar { flex: 1; min-width: 0; display: flex; flex-direction: column; align-items: center; height: 100%; }
.bar .val {
	font-size: 10.5px;
	color: var(--ink-3);
	margin-bottom: 3px;
	white-space: nowrap;
}
.bar > i {
	width: 100%;
	min-height: 3px;
	margin-top: auto;
	border-radius: 5px 5px 0 0;
	background: var(--gold);
	transform-origin: bottom;
}
.bar > i.gut { background: var(--gruen); }
.bar > i.schwach { background: var(--bad); }
.bar > i.split { display: flex; flex-direction: column; background: var(--gruen); overflow: hidden; }
.bar > i.split u { display: block; width: 100%; background: var(--rot); flex-grow: 0; flex-shrink: 0; }
.bar > i.split.kasse { background: var(--gruen); }
.bar > i.split.kasse u.offen { background: var(--bad); }
.bar .lbl {
	font-family: var(--font-display);
	font-size: 10.5px;
	text-transform: uppercase;
	color: var(--ink-3);
	margin-top: 5px;
	white-space: nowrap;
}
.legend { display: flex; align-items: center; gap: 6px; font-size: 11.5px; color: var(--ink-3); margin-top: 11px; }
.legend .dot { width: 8px; height: 8px; border-radius: 50%; background: var(--gruen); display: inline-block; }
.legend .dot.spiel, .legend .dot.offen { background: var(--rot); }
.legend .dot.offen { background: var(--bad); }
</style>
