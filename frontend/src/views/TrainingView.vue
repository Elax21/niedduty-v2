<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { api } from '../services/api';
import { enterRows, growBars } from '../lib/motion';
import ScoreBoard from '../components/ScoreBoard.vue';
import type { StatsResponse } from '../types';

const stats = ref<StatsResponse | null>(null);
const months = ref(3);

async function load() {
	const to = new Date();
	const from = new Date();
	from.setMonth(from.getMonth() - months.value);
	const { data } = await api.get<StatsResponse>('/attendance/stats', {
		params: { from: from.toISOString().slice(0, 10), to: to.toISOString().slice(0, 10) }
	});
	stats.value = data;
	requestAnimationFrame(() => {
		enterRows('.tr-anim');
		growBars('.quote-bar .fill');
	});
}

const sorted = computed(() =>
	[...(stats.value?.players ?? [])].sort((a, b) => b.quotePct - a.quotePct)
);
const avg = computed(() => {
	const ps = stats.value?.players ?? [];
	return ps.length ? Math.round(ps.reduce((s, p) => s + p.quotePct, 0) / ps.length) : 0;
});

function barClass(pct: number) {
	return pct >= 75 ? 'gut' : pct >= 50 ? 'mittel' : 'schlecht';
}

watch(months, load);
onMounted(load);
</script>

<template>
	<div class="page-head">
		<h1>Trainingsbeteiligung</h1>
		<div class="field" style="margin: 0; min-width: 160px">
			<label for="tr-range" class="sr-only" style="position: absolute; left: -9999px">Zeitraum</label>
			<select id="tr-range" v-model.number="months">
				<option :value="1">Letzter Monat</option>
				<option :value="3">Letzte 3 Monate</option>
				<option :value="6">Letzte 6 Monate</option>
				<option :value="12">Letzte 12 Monate</option>
			</select>
		</div>
	</div>

	<div v-if="stats" class="card" style="margin-bottom: 16px">
		<div class="card-body" style="display: flex; gap: 28px; flex-wrap: wrap">
			<ScoreBoard :value="stats.trainings" label="Trainings im Zeitraum" />
			<ScoreBoard :value="avg" suffix="%" label="Beteiligung Ø" />
		</div>
	</div>

	<div class="card">
		<div class="card-head">
			<h2>Beteiligung pro Spieler</h2>
			<span class="meta">{{ stats?.from }} – {{ stats?.to }}</span>
		</div>
		<div class="card-body flush" style="overflow-x: auto">
			<table class="tbl">
				<thead>
					<tr>
						<th style="width: 44px">Nr.</th>
						<th>Spieler</th>
						<th style="min-width: 130px">Quote</th>
						<th class="num">%</th>
						<th class="num">Da</th>
						<th class="num">Abgesagt</th>
						<th class="num">Offen</th>
					</tr>
				</thead>
				<tbody>
					<tr v-for="p in sorted" :key="p.playerId" class="tr-anim">
						<td class="num">{{ p.number ?? '–' }}</td>
						<td>{{ p.name }}</td>
						<td>
							<div class="quote-bar">
								<div class="fill" :class="barClass(p.quotePct)" :style="{ width: p.quotePct + '%' }" />
							</div>
						</td>
						<td class="num" style="font-weight: 600">{{ p.quotePct }}</td>
						<td class="num" style="color: var(--gruen)">{{ p.attended }}</td>
						<td class="num" style="color: var(--bad)">{{ p.declined }}</td>
						<td class="num" style="color: var(--kreide-45)">{{ p.noAnswer }}</td>
					</tr>
				</tbody>
			</table>
			<p v-if="stats && !stats.trainings" class="empty">
				Keine Trainings im Zeitraum. Lege Trainings im Kalender an — die Zu-/Absagen der Spieler landen hier.
			</p>
		</div>
	</div>
</template>
