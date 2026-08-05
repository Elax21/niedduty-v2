<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { api, apiError } from '../services/api';
import { useRefresh } from '../lib/refresh';
import { useAuthStore } from '../stores/auth';
import { enterRows } from '../lib/motion';
import { Pencil, Plus, Trash2, RefreshCw, CalendarClock } from 'lucide-vue-next';
import MatchCard from '../components/MatchCard.vue';
import type { LeagueEntry, Matches } from '../types';

const auth = useAuthStore();
const tab = ref<'tabelle' | 'spiele'>('tabelle');

const table = ref<LeagueEntry[]>([]);
const editing = ref(false);
const draft = ref<LeagueEntry[]>([]);
const error = ref('');
const busy = ref(false);
const syncing = ref(false);

const matches = ref<Matches | null>(null);
const matchesLoading = ref(false);

async function loadTable() {
	const { data } = await api.get<LeagueEntry[]>('/table');
	table.value = data;
	requestAnimationFrame(() => enterRows('.tbl-anim'));
}

async function loadMatches(force = false) {
	if (matches.value && !force) return;
	matchesLoading.value = true;
	try {
		const { data } = await api.get<Matches>('/fussball/matches');
		matches.value = data;
	} catch (e) {
		error.value = apiError(e);
	} finally {
		matchesLoading.value = false;
	}
}

function openSpiele() {
	tab.value = 'spiele';
	editing.value = false;
	loadMatches();
}

async function sync() {
	error.value = '';
	syncing.value = true;
	try {
		const { data } = await api.post<LeagueEntry[]>('/table/sync');
		if (Array.isArray(data)) table.value = data;
		else await loadTable();
	} catch (e) {
		error.value = apiError(e);
	} finally {
		syncing.value = false;
	}
}

function startEdit() {
	draft.value = table.value.map((e) => ({ ...e }));
	editing.value = true;
}
function addRow() {
	draft.value.push({ teamName: '', isOwn: false, played: 0, won: 0, drawn: 0, lost: 0, goalsFor: 0, goalsAgainst: 0, points: 0 });
}
function removeRow(i: number) {
	draft.value.splice(i, 1);
}
async function save() {
	error.value = '';
	busy.value = true;
	try {
		const { data } = await api.put<LeagueEntry[]>('/table', { entries: draft.value.filter((e) => e.teamName.trim() !== '') });
		table.value = data;
		editing.value = false;
	} catch (e) {
		error.value = apiError(e);
	} finally {
		busy.value = false;
	}
}

const diff = (e: LeagueEntry) => e.goalsFor - e.goalsAgainst;
const nextMatches = computed(() => matches.value?.next ?? []);
const prevMatches = computed(() => matches.value?.previous ?? []);

// ── Eigene Position + Formkurve (aus den letzten Spielen abgeleitet) ──
const ownIndex = computed(() => table.value.findIndex((e) => e.isOwn));
const ownEntry = computed(() => (ownIndex.value < 0 ? null : table.value[ownIndex.value]));

/** Letzte fünf gespielte Partien als S/U/N. */
const ownForm = computed(() => {
	const out: { result: 'S' | 'U' | 'N'; score: string; opponent: string }[] = [];
	for (const m of prevMatches.value) {
		if (!m.played || m.homeGoals === null || m.guestGoals === null) continue;
		const home = m.home.isOwn;
		if (!home && !m.guest.isOwn) continue;
		const own = home ? m.homeGoals : m.guestGoals;
		const other = home ? m.guestGoals : m.homeGoals;
		out.push({
			result: own > other ? 'S' : own < other ? 'N' : 'U',
			score: `${own}:${other}`,
			opponent: home ? m.guest.name : m.home.name
		});
		if (out.length === 5) break;
	}
	return out;
});

/** Jüngstes gespieltes Spiel — bekommt die Score-Tiles-Behandlung. */
const lastResult = computed(() => prevMatches.value.find((m) => m.played) ?? null);

useRefresh(() => Promise.all([loadTable(), loadMatches(true)]));

onMounted(() => {
	loadTable();
	loadMatches();
});
</script>

<template>
	<div class="page-head">
		<h1>Liga</h1>
		<span class="sub-mono">{{ auth.club?.liga }}</span>
	</div>

	<div class="segmented">
		<button :class="{ active: tab === 'tabelle' }" @click="tab = 'tabelle'; editing = false">Tabelle</button>
		<button :class="{ active: tab === 'spiele' }" @click="openSpiele">Spiele</button>
	</div>

	<p v-if="error" class="form-error" role="alert">{{ error }}</p>

	<!-- ── TABELLE (nativ, von fussball.de gesynct) ── -->
	<template v-if="tab === 'tabelle'">
		<!-- Signature: eigene Position groß, darunter Formkurve -->
		<section v-if="ownEntry && !editing" class="pos-card" style="margin-bottom: 14px">
			<div class="rank">{{ ownIndex + 1 }}.</div>
			<div class="mid">
				<div class="team">{{ ownEntry.teamName }}</div>
				<div v-if="ownForm.length" class="formrow">
					<span
						v-for="(f, i) in ownForm"
						:key="i"
						class="formchip"
						:class="f.result"
						:title="`${f.score} gegen ${f.opponent}`"
					>{{ f.result }}</span>
					<span class="overline cap">Form</span>
				</div>
			</div>
			<div class="pts">
				<div class="n" style="color: var(--gold)">{{ ownEntry.points }}</div>
				<div class="overline" style="margin-top: 4px">Pkt</div>
			</div>
		</section>

		<!-- Letztes Ergebnis als Beleg mit Score-Tiles -->
		<section v-if="lastResult && !editing" class="card" style="margin-bottom: 14px; background: var(--surface-inset)">
			<div class="card-body">
				<div class="overline" style="text-align: center">{{ lastResult.date }}</div>
				<div class="score-tiles" style="margin: 12px 0 10px">
					<span class="t">{{ lastResult.homeGoals }}</span>
					<span class="sep">:</span>
					<span class="t">{{ lastResult.guestGoals }}</span>
				</div>
				<div class="last-teams">
					<span :class="{ own: lastResult.home.isOwn }">{{ lastResult.home.name }}</span>
					<span class="dash">–</span>
					<span :class="{ own: lastResult.guest.isOwn }">{{ lastResult.guest.name }}</span>
				</div>
			</div>
		</section>

		<div v-if="!editing" class="card">
			<div class="card-head">
				<h2>Tabelle</h2>
				<div style="display: flex; gap: 6px">
					<button class="btn sm ghost" :disabled="syncing" title="Von fussball.de aktualisieren" @click="sync">
						<RefreshCw :size="13" :class="{ spin: syncing }" /> {{ syncing ? 'Sync …' : 'Sync' }}
					</button>
					<button v-if="auth.isAdmin" class="btn sm ghost" aria-label="Manuell bearbeiten" @click="startEdit"><Pencil :size="13" /></button>
				</div>
			</div>
			<div class="card-body flush" style="overflow-x: auto">
				<table class="tbl">
					<thead>
						<tr><th style="width: 34px">#</th><th>Team</th><th class="num">Sp</th><th class="num">Diff</th><th class="num">Pkt</th></tr>
					</thead>
					<tbody>
						<tr v-for="(e, i) in table" :key="e.teamName" class="tbl-anim" :class="{ own: e.isOwn }">
							<td class="num" style="color: var(--gold)">{{ i + 1 }}</td>
							<td>{{ e.teamName }}</td>
							<td class="num" style="color: var(--ink-3)">{{ e.played }}</td>
							<td class="num" :style="{ color: diff(e) > 0 ? 'var(--gruen)' : diff(e) < 0 ? 'var(--bad)' : 'var(--ink-2)' }">{{ diff(e) > 0 ? '+' : '' }}{{ diff(e) }}</td>
							<td class="num" style="color: var(--gold); font-weight: 700">{{ e.points }}</td>
						</tr>
					</tbody>
				</table>
				<p v-if="!table.length" class="empty">Noch keine Tabelle. „Sync" holt sie von fussball.de.</p>
			</div>
		</div>

		<!-- Manuell bearbeiten (Fallback) -->
		<div v-else class="card">
			<div class="card-head">
				<h2>Manuell bearbeiten</h2>
				<div style="display: flex; gap: 6px">
					<button class="btn sm" @click="addRow"><Plus :size="13" /></button>
					<button class="btn sm ghost" @click="editing = false">Abbr.</button>
					<button class="btn sm gold" :disabled="busy" @click="save">{{ busy ? '…' : 'Speichern' }}</button>
				</div>
			</div>
			<div class="card-body flush" style="overflow-x: auto">
				<table class="tbl edit-tbl">
					<thead>
						<tr><th>Team</th><th>Wir</th><th class="num">Sp</th><th class="num">S</th><th class="num">U</th><th class="num">N</th><th class="num">T+</th><th class="num">T-</th><th class="num">Pkt</th><th></th></tr>
					</thead>
					<tbody>
						<tr v-for="(e, i) in draft" :key="i">
							<td><input v-model="e.teamName" class="input" aria-label="Team" /></td>
							<td style="text-align: center"><input v-model="e.isOwn" type="checkbox" aria-label="Eigenes Team" /></td>
							<td><input v-model.number="e.played" type="number" min="0" class="input num-in" aria-label="Spiele" /></td>
							<td><input v-model.number="e.won" type="number" min="0" class="input num-in" aria-label="Siege" /></td>
							<td><input v-model.number="e.drawn" type="number" min="0" class="input num-in" aria-label="Unentschieden" /></td>
							<td><input v-model.number="e.lost" type="number" min="0" class="input num-in" aria-label="Niederlagen" /></td>
							<td><input v-model.number="e.goalsFor" type="number" min="0" class="input num-in" aria-label="Tore" /></td>
							<td><input v-model.number="e.goalsAgainst" type="number" min="0" class="input num-in" aria-label="Gegentore" /></td>
							<td><input v-model.number="e.points" type="number" min="0" class="input num-in" aria-label="Punkte" /></td>
							<td><button class="btn sm danger icon" aria-label="Zeile löschen" @click="removeRow(i)"><Trash2 :size="13" /></button></td>
						</tr>
					</tbody>
				</table>
			</div>
		</div>
	</template>

	<!-- ── SPIELE (nativ von fussball.de) ── -->
	<template v-else>
		<p v-if="matchesLoading" class="empty">Lade Spiele …</p>
		<template v-else>
			<div v-if="nextMatches.length" class="card" style="margin-bottom: 14px">
				<div class="card-head"><h2>Kommende Spiele</h2></div>
				<div class="card-body flush">
					<MatchCard v-for="(m, i) in nextMatches" :key="'n' + i" :match="m" />
				</div>
			</div>
			<div v-if="prevMatches.length" class="card">
				<div class="card-head"><h2>Letzte Ergebnisse</h2></div>
				<div class="card-body flush">
					<MatchCard v-for="(m, i) in prevMatches" :key="'p' + i" :match="m" />
				</div>
			</div>
			<div v-if="!nextMatches.length && !prevMatches.length" class="card">
				<div class="empty"><CalendarClock :size="30" class="ic" /><br />Keine Spiele von fussball.de gefunden.</div>
			</div>
		</template>
	</template>
</template>

<style scoped>
.edit-tbl td { padding: 4px 5px; }
.edit-tbl .input { min-height: 36px; padding: 5px 7px; border-radius: 8px; }
.num-in { width: 52px; font-family: var(--font-mono); text-align: center; }
.last-teams {
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 8px;
	font-size: 13px;
	color: var(--ink-2);
	text-align: center;
}
.last-teams .own { color: var(--gold); font-weight: 600; }
.last-teams .dash { color: var(--ink-3); }
.spin { animation: spin 0.9s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
</style>
