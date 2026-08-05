<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { api, apiError } from '../services/api';
import { useRefresh } from '../lib/refresh';
import { enterRows } from '../lib/motion';
import { Plus, Pencil, Trash2, ExternalLink, BarChart3 } from 'lucide-vue-next';
import AppModal from '../components/AppModal.vue';
import type { Player, SquadStatsResponse, SquadStat } from '../types';

const players = ref<Player[]>([]);
const error = ref('');
const tab = ref<'kader' | 'statistik'>('kader');

const positions = [
	{ key: 'TW', label: 'Tor' },
	{ key: 'AB', label: 'Abwehr' },
	{ key: 'MF', label: 'Mittelfeld' },
	{ key: 'ST', label: 'Sturm' }
] as const;

type PosKey = (typeof positions)[number]['key'];
const filter = ref<PosKey | null>(null);

const grouped = computed(() =>
	positions
		.filter((pos) => !filter.value || filter.value === pos.key)
		.map((pos) => ({ ...pos, players: players.value.filter((p) => p.position === pos.key) }))
);

function togglePos(key: PosKey) {
	filter.value = filter.value === key ? null : key;
}
function countIn(key: PosKey) {
	return players.value.filter((p) => p.position === key).length;
}

async function load() {
	const { data } = await api.get<Player[]>('/players');
	players.value = data;
	requestAnimationFrame(() => enterRows('.kad-anim'));
}

// ── fussball.de-Kaderstatistik ──────────────────────────────────
const stats = ref<SquadStat[]>([]);
const statSeason = ref('');
const statsLoading = ref(false);
const statsError = ref('');
const sortBy = ref<'minutes' | 'goals' | 'matches'>('minutes');

/** Aktuelle und vorige Saison als fussball.de-Kennung ("2627"). */
const seasons = computed(() => {
	const now = new Date();
	const y = now.getMonth() + 1 < 7 ? now.getFullYear() - 1 : now.getFullYear();
	const key = (start: number) => `${String(start % 100).padStart(2, '0')}${String((start + 1) % 100).padStart(2, '0')}`;
	return [
		{ key: key(y), label: `${y}/${y + 1}` },
		{ key: key(y - 1), label: `${y - 1}/${y}` }
	];
});

const sortedStats = computed(() => [...stats.value].sort((a, b) => b[sortBy.value] - a[sortBy.value]));

async function loadStats(season?: string) {
	statsLoading.value = true;
	statsError.value = '';
	try {
		const { data } = await api.get<SquadStatsResponse>('/fussball/squad-stats', {
			params: season ? { saison: season } : {}
		});
		stats.value = data.players;
		statSeason.value = data.season;
	} catch (e) {
		statsError.value = apiError(e);
	} finally {
		statsLoading.value = false;
	}
	requestAnimationFrame(() => enterRows('.stat-anim'));
}

function openStats() {
	tab.value = 'statistik';
	if (!stats.value.length) loadStats();
}

function minutesText(m: number) {
	return m >= 1000 ? `${Math.round(m / 90)} Sp.-Äq.` : `${m} Min`;
}

// ── Kader pflegen ───────────────────────────────────────────────
const showForm = ref(false);
const editId = ref<string | null>(null);
const form = ref({ name: '', number: '' as string | number, position: 'MF', status: 'fit', birthday: '' });

function openCreate() {
	editId.value = null;
	form.value = { name: '', number: '', position: 'MF', status: 'fit', birthday: '' };
	error.value = '';
	showForm.value = true;
}
function openEdit(p: Player) {
	editId.value = p.id;
	form.value = { name: p.name, number: p.number ?? '', position: p.position, status: p.status, birthday: p.birthday ?? '' };
	error.value = '';
	showForm.value = true;
}
async function submit() {
	error.value = '';
	const payload = {
		name: form.value.name.trim(),
		number: form.value.number === '' ? null : Number(form.value.number),
		position: form.value.position,
		status: form.value.status,
		birthday: form.value.birthday
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
useRefresh(load);
</script>

<template>
	<div class="page-head">
		<h1>Kader</h1>
		<span class="sub-mono">{{ players.length }} Spieler</span>
	</div>

	<div class="segmented">
		<button :class="{ active: tab === 'kader' }" @click="tab = 'kader'">Kader</button>
		<button :class="{ active: tab === 'statistik' }" @click="openStats">Statistik</button>
	</div>

	<!-- ── KADER ── -->
	<template v-if="tab === 'kader'">
		<!-- Positions-Übersicht, zugleich Filter -->
		<div class="posgrid" style="margin-bottom: 16px">
			<button
				v-for="pos in positions"
				:key="pos.key"
				class="postile"
				:class="{ on: filter === pos.key }"
				@click="togglePos(pos.key)"
			>
				<div class="n">{{ countIn(pos.key) }}</div>
				<div class="l">{{ pos.label }}</div>
			</button>
		</div>

		<div v-for="group in grouped" :key="group.key">
			<div class="band">
				<span class="lbl">{{ group.label }}</span>
				<span class="rule" />
			</div>
			<div v-for="p in group.players" :key="p.id" class="prow kad-anim">
				<span class="jersey">{{ p.number ?? '–' }}</span>
				<span class="grow">
					<span class="n">{{ p.name }}</span>
					<span class="r">{{ group.label }}</span>
				</span>
				<span class="chip" :class="p.status">{{ p.status }}</span>
				<button class="btn sm icon ghost" aria-label="Bearbeiten" @click="openEdit(p)"><Pencil :size="13" /></button>
				<button class="btn sm icon danger" aria-label="Löschen" @click="remove(p)"><Trash2 :size="13" /></button>
			</div>
			<p v-if="!group.players.length" class="empty">Keine Spieler.</p>
		</div>

		<button class="fab" aria-label="Spieler anlegen" @click="openCreate"><Plus :size="24" /></button>
	</template>

	<!-- ── STATISTIK (direkt von fussball.de) ── -->
	<template v-else>
		<div class="segmented">
			<button v-for="s in seasons" :key="s.key" :class="{ active: statSeason === s.key }" @click="loadStats(s.key)">
				{{ s.label }}
			</button>
		</div>

		<p v-if="statsError" class="form-error" role="alert">{{ statsError }}</p>

		<div class="card">
			<div class="card-head">
				<h2>Einsätze</h2>
				<div class="sortbtns">
					<button :class="{ on: sortBy === 'minutes' }" @click="sortBy = 'minutes'">Min</button>
					<button :class="{ on: sortBy === 'goals' }" @click="sortBy = 'goals'">Tore</button>
					<button :class="{ on: sortBy === 'matches' }" @click="sortBy = 'matches'">Spiele</button>
				</div>
			</div>
			<div class="card-body flush">
				<p v-if="statsLoading" class="empty">Hole Daten von fussball.de …</p>
				<p v-else-if="!sortedStats.length" class="empty">
					<BarChart3 :size="30" class="ic" /><br />
					Für diese Saison gibt es noch keine Einsätze.
				</p>
				<template v-else>
					<div class="statrow head">
						<span class="rank" />
						<span class="nm overline">Spieler</span>
						<span class="v overline">Sp</span>
						<span class="v overline">Min</span>
						<span class="v overline">Tore</span>
					</div>
					<component
						:is="p.profileUrl ? 'a' : 'div'"
						v-for="(p, i) in sortedStats"
						:key="p.name"
						class="statrow stat-anim"
						:href="p.profileUrl || undefined"
						:target="p.profileUrl ? '_blank' : undefined"
						rel="noopener"
					>
						<span class="rank">{{ i + 1 }}</span>
						<span class="nm">{{ p.name }}</span>
						<span class="v dim">{{ p.matches }}</span>
						<span class="v">{{ p.minutes }}</span>
						<span class="v gold">{{ p.goals }}</span>
					</component>
				</template>
			</div>
		</div>
		<p class="quelle">
			Quelle: fussball.de · Einsatzminuten wie dort ausgewiesen
			<ExternalLink :size="11" style="vertical-align: -1px" />
		</p>
	</template>

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
				<label for="pl-bday">Geburtstag (für die Erinnerung)</label>
				<input id="pl-bday" v-model="form.birthday" type="date" />
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
.statrow.head { border-bottom: 1px solid var(--line-2); padding-top: 8px; padding-bottom: 8px; }
.statrow.head .v, .statrow.head .nm { color: var(--ink-3); }
.sortbtns { display: flex; gap: 4px; }
.sortbtns button {
	font-family: var(--font-display);
	font-size: 11px;
	font-weight: 700;
	text-transform: uppercase;
	letter-spacing: 0.08em;
	color: var(--ink-3);
	padding: 5px 9px;
	border-radius: 8px;
	border: 1px solid transparent;
}
.sortbtns button.on { color: var(--gold); border-color: var(--line-2); background: var(--gold-bg); }
.quelle { font-size: 11.5px; color: var(--ink-3); text-align: center; margin-top: 12px; }
</style>
