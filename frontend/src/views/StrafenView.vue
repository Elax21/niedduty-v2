<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { api, apiError } from '../services/api';
import { useAuthStore } from '../stores/auth';
import { enterRows } from '../lib/motion';
import { shareKasseImage } from '../lib/shareImage';
import { Plus, Trash2, Check, Share2, Pencil } from 'lucide-vue-next';
import AppModal from '../components/AppModal.vue';
import ScoreBoard from '../components/ScoreBoard.vue';
import type { Penalty, Player, PlayerPenalty } from '../types';

const auth = useAuthStore();
const catalog = ref<Penalty[]>([]);
const assigned = ref<PlayerPenalty[]>([]);
const players = ref<Player[]>([]);
const error = ref('');
const shareMsg = ref('');

const canWrite = computed(() => auth.can('strafen'));

function euro(cents: number) {
	return (cents / 100).toLocaleString('de-DE', { style: 'currency', currency: 'EUR' });
}

const openSum = computed(() => assigned.value.filter((p) => !p.paid).reduce((s, p) => s + p.amount, 0));
const paidSum = computed(() => assigned.value.filter((p) => p.paid).reduce((s, p) => s + p.amount, 0));

const byPlayer = computed(() => {
	const map = new Map<string, { player: Player; items: PlayerPenalty[]; open: number }>();
	for (const p of players.value) map.set(p.id, { player: p, items: [], open: 0 });
	for (const pp of assigned.value) {
		const entry = map.get(pp.playerId);
		if (!entry) continue;
		entry.items.push(pp);
		if (!pp.paid) entry.open += pp.amount;
	}
	return [...map.values()]
		.filter((e) => e.items.length)
		.sort((a, b) => b.open - a.open);
});

async function load() {
	const [c, a, p] = await Promise.all([
		api.get<Penalty[]>('/penalties'),
		api.get<PlayerPenalty[]>('/player-penalties'),
		api.get<Player[]>('/players')
	]);
	catalog.value = c.data;
	assigned.value = a.data;
	players.value = p.data;
	selected.value.clear();
	requestAnimationFrame(() => enterRows('.str-anim'));
}

// ── Strafe(n) aufschreiben: mehrere Spieler × mehrere Vergehen ──
const showAssign = ref(false);
const assignPlayers = ref<string[]>([]);
const assignPenalties = ref<string[]>([]);
const freeLabel = ref('');
const freeAmount = ref('');

function openAssign() {
	assignPlayers.value = [];
	assignPenalties.value = [];
	freeLabel.value = '';
	freeAmount.value = '';
	error.value = '';
	showAssign.value = true;
}
function toggleAll() {
	assignPlayers.value = assignPlayers.value.length === players.value.length ? [] : players.value.map((p) => p.id);
}
async function submitAssign() {
	error.value = '';
	const amount = freeAmount.value ? Math.round(parseFloat(freeAmount.value.replace(',', '.')) * 100) : 0;
	if (freeLabel.value.trim() && (!amount || amount <= 0)) {
		error.value = 'Freie Strafe: gültigen Betrag angeben';
		return;
	}
	try {
		await api.post('/player-penalties', {
			playerIds: assignPlayers.value,
			penaltyIds: assignPenalties.value,
			freeLabel: freeLabel.value.trim(),
			freeAmount: amount
		});
		showAssign.value = false;
		await load();
	} catch (e) {
		error.value = apiError(e);
	}
}

// ── Auswahl + Bulk-Aktionen ──
const selected = ref(new Set<string>());
function toggleSelect(id: string) {
	const s = new Set(selected.value);
	if (s.has(id)) s.delete(id);
	else s.add(id);
	selected.value = s;
}
async function bulkPaid(paid: boolean) {
	await api.post('/player-penalties/paid', { ids: [...selected.value], paid });
	await load();
}
async function bulkDelete() {
	if (!window.confirm(`${selected.value.size} Strafe(n) wirklich löschen?`)) return;
	await api.post('/player-penalties/delete', { ids: [...selected.value] });
	await load();
}
async function togglePaid(pp: PlayerPenalty) {
	await api.post('/player-penalties/paid', { ids: [pp.id], paid: !pp.paid });
	await load();
}

// ── WhatsApp-Status ──
const sharing = ref(false);
async function shareStatus() {
	sharing.value = true;
	shareMsg.value = '';
	const result = await shareKasseImage({
		clubName: auth.club?.name ?? 'Aramäer Ahlen',
		rows: byPlayer.value
			.filter((e) => e.open > 0)
			.map((e) => ({ name: e.player.name, open: e.open, count: e.items.filter((i) => !i.paid).length })),
		totalOpen: openSum.value,
		iban: auth.club?.kasseIban || undefined,
		inhaber: auth.club?.kasseInhaber || undefined
	});
	sharing.value = false;
	if (result === 'downloaded') shareMsg.value = 'Bild heruntergeladen — direkt in WhatsApp posten.';
	if (result === 'failed') shareMsg.value = 'Bild konnte nicht erstellt werden.';
}

// ── Katalog pflegen ──
const showCatalogForm = ref(false);
const editCatalogId = ref<string | null>(null);
const catalogForm = ref({ label: '', amountEuro: '', unit: '' });

function openCatalogCreate() {
	editCatalogId.value = null;
	catalogForm.value = { label: '', amountEuro: '', unit: '' };
	showCatalogForm.value = true;
}
function openCatalogEdit(p: Penalty) {
	editCatalogId.value = p.id;
	catalogForm.value = { label: p.label, amountEuro: (p.amount / 100).toFixed(2).replace('.', ','), unit: p.unit };
	showCatalogForm.value = true;
}
async function submitCatalog() {
	error.value = '';
	const amount = Math.round(parseFloat(catalogForm.value.amountEuro.replace(',', '.')) * 100);
	if (!catalogForm.value.label.trim() || !amount || amount <= 0) {
		error.value = 'Bezeichnung und gültiger Betrag nötig';
		return;
	}
	const payload = {
		label: catalogForm.value.label.trim(),
		amount,
		unit: catalogForm.value.unit.trim(),
		sortOrder: editCatalogId.value ? (catalog.value.find((c) => c.id === editCatalogId.value)?.sortOrder ?? 0) : catalog.value.length + 1
	};
	try {
		if (editCatalogId.value) await api.put(`/penalties/${editCatalogId.value}`, payload);
		else await api.post('/penalties', payload);
		showCatalogForm.value = false;
		await load();
	} catch (e) {
		error.value = apiError(e);
	}
}
async function removeCatalog(p: Penalty) {
	if (!window.confirm(`Katalog-Eintrag „${p.label}" löschen?`)) return;
	await api.delete(`/penalties/${p.id}`);
	await load();
}

onMounted(load);
</script>

<template>
	<div class="page-head">
		<h1>Strafenkatalog &amp; Kasse</h1>
		<div style="display: flex; gap: 8px; flex-wrap: wrap">
			<button class="btn gold sm" :disabled="sharing" @click="shareStatus">
				<Share2 :size="14" aria-hidden="true" /> {{ sharing ? 'Erstelle …' : 'WhatsApp-Status' }}
			</button>
			<button v-if="canWrite" class="btn primary sm" @click="openAssign"><Plus :size="14" aria-hidden="true" /> Strafe aufschreiben</button>
		</div>
	</div>

	<p v-if="shareMsg" class="form-error" style="color: var(--gold); border-color: var(--hair-strong); background: var(--gold-bg)" role="status">{{ shareMsg }}</p>

	<div class="card str-anim" style="margin-bottom: 16px">
		<div class="card-body" style="display: flex; gap: 28px; flex-wrap: wrap">
			<ScoreBoard :value="euro(openSum)" label="Offen" />
			<ScoreBoard :value="euro(paidSum)" label="Bezahlt" />
			<div v-if="auth.club?.kasseIban" style="margin-left: auto; font-size: 12.5px; color: var(--kreide-70)">
				<div class="board-label" style="margin: 0 0 4px">Mannschaftskasse</div>
				<div style="font-family: var(--font-mono)">{{ auth.club.kasseIban }}</div>
				<div>{{ auth.club.kasseInhaber }}</div>
			</div>
		</div>
	</div>

	<div v-if="canWrite && selected.size" class="bulk-bar" role="toolbar" aria-label="Aktionen für ausgewählte Strafen">
		<strong class="num">{{ selected.size }} ausgewählt</strong>
		<button class="btn sm gold" @click="bulkPaid(true)"><Check :size="13" aria-hidden="true" /> Bezahlt</button>
		<button class="btn sm" @click="bulkPaid(false)">Offen</button>
		<button class="btn sm danger" @click="bulkDelete"><Trash2 :size="13" aria-hidden="true" /> Löschen</button>
		<button class="btn sm" style="margin-left: auto" @click="selected = new Set()">Auswahl aufheben</button>
	</div>

	<div class="grid cols-2">
		<div class="card str-anim">
			<div class="card-head">
				<h2>Katalog</h2>
				<button v-if="canWrite" class="btn sm" @click="openCatalogCreate"><Plus :size="13" aria-hidden="true" /> Eintrag</button>
			</div>
			<div class="card-body flush">
				<table class="tbl">
					<tbody>
						<tr v-for="p in catalog" :key="p.id">
							<td>
								{{ p.label }}
								<span v-if="p.unit" style="color: var(--kreide-45); font-size: 12px"> {{ p.unit }}</span>
							</td>
							<td class="num" style="color: var(--gold); font-weight: 600; width: 90px">{{ euro(p.amount) }}</td>
							<td v-if="canWrite" style="width: 76px; text-align: right; white-space: nowrap">
								<button class="btn sm" aria-label="Bearbeiten" @click="openCatalogEdit(p)"><Pencil :size="12" /></button>
								<button class="btn sm danger" aria-label="Löschen" @click="removeCatalog(p)"><Trash2 :size="12" /></button>
							</td>
						</tr>
					</tbody>
				</table>
				<p v-if="!catalog.length" class="empty">Noch kein Katalog. Lege den ersten Eintrag an.</p>
			</div>
		</div>

		<div class="card str-anim">
			<div class="card-head"><h2>Kasse nach Spieler</h2></div>
			<div class="card-body flush">
				<div v-for="entry in byPlayer" :key="entry.player.id" class="kasse-player">
					<div class="kasse-head">
						<strong>{{ entry.player.name }}</strong>
						<span class="tally" aria-hidden="true">
							<i v-for="n in Math.min(entry.items.filter(i => !i.paid).length, 12)" :key="n" class="stroke" :class="{ five: n % 5 === 0 }" />
						</span>
						<span class="num kasse-sum" :style="{ color: entry.open ? 'var(--bad)' : 'var(--gruen)' }">
							{{ euro(entry.open) }}
						</span>
					</div>
					<div v-for="pp in entry.items" :key="pp.id" class="kasse-item" :class="{ paid: pp.paid }">
						<input
							v-if="canWrite"
							type="checkbox"
							class="kasse-check"
							:checked="selected.has(pp.id)"
							:aria-label="`${pp.label} auswählen`"
							@change="toggleSelect(pp.id)"
						/>
						<span>{{ pp.label }}</span>
						<span class="num">{{ euro(pp.amount) }}</span>
						<button v-if="canWrite" class="btn sm" :class="{ gold: pp.paid }" :aria-label="pp.paid ? 'Als offen markieren' : 'Als bezahlt markieren'" @click="togglePaid(pp)">
							<Check :size="12" />
						</button>
					</div>
				</div>
				<p v-if="!byPlayer.length" class="empty">Noch keine Strafen aufgeschrieben. Die Kasse dankt trotzdem.</p>
			</div>
		</div>
	</div>

	<AppModal v-if="showAssign" title="Strafe aufschreiben" @close="showAssign = false">
		<form @submit.prevent="submitAssign">
			<p v-if="error" class="form-error" role="alert">{{ error }}</p>

			<div class="field">
				<label>Spieler ({{ assignPlayers.length }})
					<button type="button" class="btn sm" style="margin-left: 8px" @click="toggleAll">
						{{ assignPlayers.length === players.length ? 'Keinen' : 'Alle' }}
					</button>
				</label>
				<div class="pick-list">
					<label v-for="p in players" :key="p.id" class="pick-item">
						<input v-model="assignPlayers" type="checkbox" :value="p.id" />
						<span>{{ p.number ? p.number + ' · ' : '' }}{{ p.name }}</span>
					</label>
				</div>
			</div>

			<div class="field">
				<label>Vergehen ({{ assignPenalties.length }})</label>
				<div class="pick-list">
					<label v-for="c in catalog" :key="c.id" class="pick-item">
						<input v-model="assignPenalties" type="checkbox" :value="c.id" />
						<span>{{ c.label }} <em class="num" style="color: var(--gold); font-style: normal">{{ euro(c.amount) }}</em></span>
					</label>
				</div>
			</div>

			<div class="grid cols-2">
				<div class="field">
					<label for="as-free">Freie Strafe (optional)</label>
					<input id="as-free" v-model="freeLabel" maxlength="120" placeholder="z.B. Bierkasten" />
				</div>
				<div class="field">
					<label for="as-free-amount">Betrag (€)</label>
					<input id="as-free-amount" v-model="freeAmount" inputmode="decimal" placeholder="5,00" />
				</div>
			</div>

			<button class="btn primary" style="width: 100%; justify-content: center" :disabled="!assignPlayers.length">
				Aufschreiben
			</button>
		</form>
	</AppModal>

	<AppModal v-if="showCatalogForm" :title="editCatalogId ? 'Katalog-Eintrag bearbeiten' : 'Neuer Katalog-Eintrag'" @close="showCatalogForm = false">
		<form @submit.prevent="submitCatalog">
			<p v-if="error" class="form-error" role="alert">{{ error }}</p>
			<div class="field">
				<label for="cat-label">Bezeichnung</label>
				<input id="cat-label" v-model="catalogForm.label" required maxlength="120" />
			</div>
			<div class="grid cols-2">
				<div class="field">
					<label for="cat-amount">Betrag (€)</label>
					<input id="cat-amount" v-model="catalogForm.amountEuro" inputmode="decimal" placeholder="5,00" required />
				</div>
				<div class="field">
					<label for="cat-unit">Einheit (optional)</label>
					<input id="cat-unit" v-model="catalogForm.unit" maxlength="40" placeholder="pro Minute" />
				</div>
			</div>
			<button class="btn primary" style="width: 100%; justify-content: center">Speichern</button>
		</form>
	</AppModal>
</template>

<style scoped>
.bulk-bar {
	display: flex;
	align-items: center;
	gap: 8px;
	flex-wrap: wrap;
	padding: 10px 14px;
	margin-bottom: 16px;
	background: var(--gold-bg);
	border: 1px solid var(--hair-strong);
	border-radius: var(--radius);
}
.kasse-player { padding: 10px 14px; border-bottom: 1px solid var(--hair); }
.kasse-player:last-child { border-bottom: none; }
.kasse-head { display: flex; align-items: center; gap: 10px; }
.kasse-sum { margin-left: auto; font-weight: 600; }
.tally { display: inline-flex; gap: 3px; align-items: flex-end; height: 14px; }
.stroke { width: 2px; height: 13px; background: var(--kreide-70); border-radius: 1px; display: inline-block; }
.stroke.five { transform: rotate(-58deg) translateY(-2px); margin-left: -14px; }
.kasse-item {
	display: flex;
	align-items: center;
	gap: 8px;
	font-size: 13px;
	padding: 5px 0 5px 12px;
	color: var(--kreide-70);
}
.kasse-item .num { margin-left: auto; }
.kasse-item.paid { opacity: 0.45; }
.kasse-item.paid > span { text-decoration: line-through; }
.kasse-check { accent-color: var(--gold); width: 16px; height: 16px; flex-shrink: 0; }
.pick-list {
	max-height: 180px;
	overflow-y: auto;
	border: 1px solid var(--hair-strong);
	border-radius: 4px;
	background: var(--rasen-950);
	padding: 4px;
}
.pick-item {
	display: flex;
	align-items: center;
	gap: 8px;
	padding: 7px 8px;
	font-size: 13.5px;
	border-radius: 3px;
	cursor: pointer;
}
.pick-item:hover { background: rgba(240, 168, 28, 0.08); }
.pick-item input { accent-color: var(--gold); width: 16px; height: 16px; }
.pick-item span { display: flex; justify-content: space-between; width: 100%; gap: 8px; }
</style>
