<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { api, apiError } from '../services/api';
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

async function loadMatches() {
	if (matches.value) return;
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

onMounted(loadTable);
</script>

<template>
	<div class="page-head">
		<h1>Liga</h1>
		<span class="sub">{{ auth.club?.liga }}</span>
	</div>

	<div class="segmented">
		<button :class="{ active: tab === 'tabelle' }" @click="tab = 'tabelle'; editing = false">Tabelle</button>
		<button :class="{ active: tab === 'spiele' }" @click="openSpiele">Spiele</button>
	</div>

	<p v-if="error" class="form-error" role="alert">{{ error }}</p>

	<!-- ── TABELLE (nativ, von fussball.de gesynct) ── -->
	<template v-if="tab === 'tabelle'">
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
.spin { animation: spin 0.9s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
</style>
