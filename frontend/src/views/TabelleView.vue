<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { api, apiError } from '../services/api';
import { useAuthStore } from '../stores/auth';
import { enterRows } from '../lib/motion';
import { Pencil, Plus, Trash2 } from 'lucide-vue-next';
import type { LeagueEntry } from '../types';

const auth = useAuthStore();
const table = ref<LeagueEntry[]>([]);
const editing = ref(false);
const draft = ref<LeagueEntry[]>([]);
const error = ref('');
const busy = ref(false);

async function load() {
	const { data } = await api.get<LeagueEntry[]>('/table');
	table.value = data;
	requestAnimationFrame(() => enterRows('.tbl-anim'));
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
		const { data } = await api.put<LeagueEntry[]>('/table', {
			entries: draft.value.filter((e) => e.teamName.trim() !== '')
		});
		table.value = data;
		editing.value = false;
	} catch (e) {
		error.value = apiError(e);
	} finally {
		busy.value = false;
	}
}

const diff = (e: LeagueEntry) => e.goalsFor - e.goalsAgainst;

onMounted(load);
</script>

<template>
	<div class="page-head">
		<h1>Ligatabelle</h1>
		<div style="display: flex; gap: 8px">
			<span class="sub">{{ auth.club?.liga }}</span>
			<button v-if="auth.isAdmin && !editing" class="btn sm" @click="startEdit"><Pencil :size="13" aria-hidden="true" /> Bearbeiten</button>
		</div>
	</div>

	<p v-if="error" class="form-error" role="alert">{{ error }}</p>

	<div v-if="auth.club?.fussballDeWidget && !editing" class="card" style="margin-bottom: 16px">
		<div class="card-head"><h2>Live von fussball.de</h2></div>
		<iframe
			:src="auth.club.fussballDeWidget"
			title="Ligatabelle von fussball.de"
			style="width: 100%; height: 640px; border: 0; background: #fff"
			loading="lazy"
		/>
	</div>

	<div v-if="!editing" class="card">
		<div class="card-body flush" style="overflow-x: auto">
			<table class="tbl">
				<thead>
					<tr>
						<th style="width: 40px">#</th><th>Mannschaft</th>
						<th class="num">Sp</th><th class="num">S</th><th class="num">U</th><th class="num">N</th>
						<th class="num">Tore</th><th class="num">Diff</th><th class="num">Pkt</th>
					</tr>
				</thead>
				<tbody>
					<tr v-for="(e, i) in table" :key="e.teamName" class="tbl-anim" :class="{ own: e.isOwn }">
						<td class="num">{{ i + 1 }}</td>
						<td>{{ e.teamName }}</td>
						<td class="num">{{ e.played }}</td>
						<td class="num">{{ e.won }}</td>
						<td class="num">{{ e.drawn }}</td>
						<td class="num">{{ e.lost }}</td>
						<td class="num">{{ e.goalsFor }}:{{ e.goalsAgainst }}</td>
						<td class="num" :style="{ color: diff(e) > 0 ? 'var(--gruen)' : diff(e) < 0 ? 'var(--bad)' : undefined }">
							{{ diff(e) > 0 ? '+' : '' }}{{ diff(e) }}
						</td>
						<td class="num" style="color: var(--gold); font-weight: 600">{{ e.points }}</td>
					</tr>
				</tbody>
			</table>
		</div>
	</div>

	<div v-else class="card">
		<div class="card-head">
			<h2>Tabelle bearbeiten</h2>
			<div style="display: flex; gap: 8px">
				<button class="btn sm" @click="addRow"><Plus :size="13" aria-hidden="true" /> Team</button>
				<button class="btn sm" @click="editing = false">Abbrechen</button>
				<button class="btn sm gold" :disabled="busy" @click="save">{{ busy ? 'Speichern …' : 'Speichern' }}</button>
			</div>
		</div>
		<div class="card-body flush" style="overflow-x: auto">
			<table class="tbl edit-tbl">
				<thead>
					<tr>
						<th>Mannschaft</th><th>Wir</th>
						<th class="num">Sp</th><th class="num">S</th><th class="num">U</th><th class="num">N</th>
						<th class="num">T+</th><th class="num">T-</th><th class="num">Pkt</th><th></th>
					</tr>
				</thead>
				<tbody>
					<tr v-for="(e, i) in draft" :key="i">
						<td><input v-model="e.teamName" class="input" aria-label="Mannschaftsname" /></td>
						<td style="text-align: center"><input v-model="e.isOwn" type="checkbox" aria-label="Eigenes Team" /></td>
						<td><input v-model.number="e.played" type="number" min="0" class="input num-in" aria-label="Spiele" /></td>
						<td><input v-model.number="e.won" type="number" min="0" class="input num-in" aria-label="Siege" /></td>
						<td><input v-model.number="e.drawn" type="number" min="0" class="input num-in" aria-label="Unentschieden" /></td>
						<td><input v-model.number="e.lost" type="number" min="0" class="input num-in" aria-label="Niederlagen" /></td>
						<td><input v-model.number="e.goalsFor" type="number" min="0" class="input num-in" aria-label="Tore" /></td>
						<td><input v-model.number="e.goalsAgainst" type="number" min="0" class="input num-in" aria-label="Gegentore" /></td>
						<td><input v-model.number="e.points" type="number" min="0" class="input num-in" aria-label="Punkte" /></td>
						<td><button class="btn sm danger" aria-label="Zeile löschen" @click="removeRow(i)"><Trash2 :size="13" /></button></td>
					</tr>
				</tbody>
			</table>
		</div>
	</div>
</template>

<style scoped>
.edit-tbl td { padding: 4px 6px; }
.edit-tbl .input { min-height: 32px; padding: 4px 8px; }
.num-in { width: 58px; font-family: var(--font-mono); }
</style>
