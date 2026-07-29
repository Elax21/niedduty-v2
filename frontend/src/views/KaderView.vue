<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { api, apiError } from '../services/api';
import { enterRows } from '../lib/motion';
import { Plus, Pencil, Trash2 } from 'lucide-vue-next';
import AppModal from '../components/AppModal.vue';
import type { Player } from '../types';

const players = ref<Player[]>([]);
const error = ref('');

const positions = [
	{ key: 'TW', label: 'Tor' },
	{ key: 'AB', label: 'Abwehr' },
	{ key: 'MF', label: 'Mittelfeld' },
	{ key: 'ST', label: 'Sturm' }
] as const;

const grouped = computed(() =>
	positions.map((pos) => ({ ...pos, players: players.value.filter((p) => p.position === pos.key) }))
);

async function load() {
	const { data } = await api.get<Player[]>('/players');
	players.value = data;
	requestAnimationFrame(() => enterRows('.kad-anim'));
}

const showForm = ref(false);
const editId = ref<string | null>(null);
const form = ref({ name: '', number: '' as string | number, position: 'MF', status: 'fit' });

function openCreate() {
	editId.value = null;
	form.value = { name: '', number: '', position: 'MF', status: 'fit' };
	error.value = '';
	showForm.value = true;
}
function openEdit(p: Player) {
	editId.value = p.id;
	form.value = { name: p.name, number: p.number ?? '', position: p.position, status: p.status };
	error.value = '';
	showForm.value = true;
}
async function submit() {
	error.value = '';
	const payload = {
		name: form.value.name.trim(),
		number: form.value.number === '' ? null : Number(form.value.number),
		position: form.value.position,
		status: form.value.status
	};
	try {
		if (editId.value) await api.put(`/players/${editId.value}`, payload);
		else await api.post('/players', payload);
		showForm.value = false;
		await load();
	} catch (e) {
		error.value = apiError(e);
	}
}
async function remove(p: Player) {
	if (!window.confirm(`${p.name} aus dem Kader löschen? Strafen und Rückmeldungen werden mit entfernt.`)) return;
	await api.delete(`/players/${p.id}`);
	await load();
}

onMounted(load);
</script>

<template>
	<div class="page-head">
		<h1>Kader</h1>
		<span class="sub">{{ players.length }} Spieler</span>
	</div>

	<div class="stack">
		<div v-for="group in grouped" :key="group.key" class="card kad-anim">
			<div class="card-head"><h2>{{ group.label }}</h2><span class="meta">{{ group.players.length }}</span></div>
			<div class="card-body flush">
				<div v-for="p in group.players" :key="p.id" class="lrow">
					<span class="pnum mono">{{ p.number ?? '–' }}</span>
					<span class="grow t">{{ p.name }}</span>
					<span class="chip" :class="p.status">{{ p.status }}</span>
					<button class="btn sm icon ghost" aria-label="Bearbeiten" @click="openEdit(p)"><Pencil :size="13" /></button>
					<button class="btn sm icon danger" aria-label="Löschen" @click="remove(p)"><Trash2 :size="13" /></button>
				</div>
				<p v-if="!group.players.length" class="empty">Keine Spieler.</p>
			</div>
		</div>
	</div>

	<button class="fab" aria-label="Spieler anlegen" @click="openCreate"><Plus :size="24" /></button>

	<AppModal v-if="showForm" :title="editId ? 'Spieler bearbeiten' : 'Neuer Spieler'" @close="showForm = false">
		<form @submit.prevent="submit">
			<p v-if="error" class="form-error" role="alert">{{ error }}</p>
			<div class="field">
				<label for="pl-name">Name</label>
				<input id="pl-name" v-model="form.name" required maxlength="100" />
			</div>
			<div class="row2">
				<div class="field">
					<label for="pl-number">Nummer</label>
					<input id="pl-number" v-model="form.number" type="number" min="1" max="99" inputmode="numeric" />
				</div>
				<div class="field">
					<label for="pl-pos">Position</label>
					<select id="pl-pos" v-model="form.position">
						<option value="TW">Torwart</option>
						<option value="AB">Abwehr</option>
						<option value="MF">Mittelfeld</option>
						<option value="ST">Sturm</option>
					</select>
				</div>
			</div>
			<div class="field">
				<label for="pl-status">Status</label>
				<select id="pl-status" v-model="form.status">
					<option value="fit">Fit</option>
					<option value="verletzt">Verletzt</option>
					<option value="gesperrt">Gesperrt</option>
					<option value="krank">Krank</option>
				</select>
			</div>
			<button class="btn primary block">Speichern</button>
		</form>
	</AppModal>
</template>

<style scoped>
.pnum {
	width: 34px; height: 34px;
	display: grid; place-items: center;
	flex-shrink: 0;
	font-size: 15px; font-weight: 600;
	color: var(--gold);
	background: var(--surface-3);
	border: 1px solid var(--line);
	border-radius: 9px;
}
</style>
