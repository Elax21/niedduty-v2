<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { api, apiError } from '../services/api';
import { useRefresh } from '../lib/refresh';
import { useAuthStore } from '../stores/auth';
import { enterRows, countUp } from '../lib/motion';
import { shareKasseImage } from '../lib/shareImage';
import { Plus, Trash2, Check, Share2, Pencil, RotateCcw, ShieldCheck, ShieldAlert } from 'lucide-vue-next';
import AppModal from '../components/AppModal.vue';
import type { Penalty, Player, PlayerPenalty, PenaltyLogEntry, PenaltyLogCheck } from '../types';

const auth = useAuthStore();
const tab = ref<'kasse' | 'katalog' | 'protokoll'>('kasse');
const catalog = ref<Penalty[]>([]);
const assigned = ref<PlayerPenalty[]>([]);
const players = ref<Player[]>([]);
const error = ref('');
const shareMsg = ref('');
const openShown = ref(0);
const paidShown = ref(0);
const teamOpen = ref(0);
const teamOpenShown = ref(0);

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
	return [...map.values()].filter((e) => e.items.length).sort((a, b) => b.open - a.open);
});

async function load() {
	const [c, a, p, s] = await Promise.all([
		api.get<Penalty[]>('/penalties'),
		api.get<PlayerPenalty[]>('/player-penalties'),
		api.get<Player[]>('/players'),
		api.get<{ totalOpen: number; totalPaid: number }>('/player-penalties/summary')
	]);
	catalog.value = c.data;
	assigned.value = a.data;
	players.value = p.data;
	teamOpen.value = s.data.totalOpen;
	selected.value = new Set();
	countUp(openSum.value, (v) => (openShown.value = v));
	countUp(paidSum.value, (v) => (paidShown.value = v));
	countUp(teamOpen.value, (v) => (teamOpenShown.value = v));
	requestAnimationFrame(() => enterRows('.str-anim'));
}

// ── Strafe(n) aufschreiben ──
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

// ── Auswahl + Bulk ──
const selected = ref(new Set<string>());
function toggleSelect(id: string) {
	const s = new Set(selected.value);
	s.has(id) ? s.delete(id) : s.add(id);
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
/** Einzelne Strafe löschen — für Fehleinträge, ohne Umweg über die Auswahl. */
async function removeAssigned(pp: PlayerPenalty) {
	if (!window.confirm(`„${pp.label}" (${euro(pp.amount)}) wirklich löschen?`)) return;
	await api.post('/player-penalties/delete', { ids: [pp.id] });
	await load();
}

async function togglePaid(pp: PlayerPenalty) {
	await api.post('/player-penalties/paid', { ids: [pp.id], paid: !pp.paid });
	await load();
}

// ── Protokoll ──────────────────────────────────────────────────
// Wer Strafen aufschreiben darf, darf auch löschen. Damit das niemand still
// tut, steht jede Bewegung hier — inklusive Prüfung der Hash-Kette.
const log = ref<PenaltyLogEntry[]>([]);
const logCheck = ref<PenaltyLogCheck | null>(null);

const actionLabels: Record<string, string> = {
	aufgeschrieben: 'aufgeschrieben',
	geloescht: 'gelöscht',
	bezahlt: 'auf bezahlt gesetzt',
	wieder_offen: 'wieder geöffnet',
	katalog_geaendert: 'Katalog geändert',
	katalog_geloescht: 'Katalog-Eintrag gelöscht'
};

async function openLog() {
	tab.value = 'protokoll';
	const [l, v] = await Promise.all([
		api.get<PenaltyLogEntry[]>('/penalty-log'),
		api.get<PenaltyLogCheck>('/penalty-log/verify')
	]);
	log.value = l.data;
	logCheck.value = v.data;
}

function logTime(iso: string) {
	return new Date(iso).toLocaleString('de-DE', {
		day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit'
	});
}

// ── WhatsApp-Status ──
const sharing = ref(false);
async function shareStatus() {
	sharing.value = true;
	shareMsg.value = '';
	const result = await shareKasseImage({
		clubName: auth.club?.name ?? 'Aramäer Ahlen',
		rows: byPlayer.value.filter((e) => e.open > 0).map((e) => ({ name: e.player.name, open: e.open, count: e.items.filter((i) => !i.paid).length })),
		totalOpen: openSum.value,
		iban: auth.club?.kasseIban || undefined,
		inhaber: auth.club?.kasseInhaber || undefined
	});
	sharing.value = false;
	if (result === 'downloaded') shareMsg.value = 'Bild heruntergeladen — direkt in WhatsApp posten.';
	if (result === 'failed') shareMsg.value = 'Bild konnte nicht erstellt werden.';
}

// ── Katalog ──
const showCatalogForm = ref(false);
const editCatalogId = ref<string | null>(null);
const catalogForm = ref({ label: '', amountEuro: '', unit: '' });

function openCatalogCreate() {
	editCatalogId.value = null;
	catalogForm.value = { label: '', amountEuro: '', unit: '' };
	error.value = '';
	showCatalogForm.value = true;
}
function openCatalogEdit(p: Penalty) {
	editCatalogId.value = p.id;
	catalogForm.value = { label: p.label, amountEuro: (p.amount / 100).toFixed(2).replace('.', ','), unit: p.unit };
	error.value = '';
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
useRefresh(load);
</script>

<template>
	<div class="page-head">
		<h1>{{ canWrite ? 'Kasse' : 'Meine Strafen' }}</h1>
		<button v-if="canWrite" class="btn sm gold" :disabled="sharing" @click="shareStatus">
			<Share2 :size="14" /> {{ sharing ? '…' : 'Status' }}
		</button>
	</div>

	<p v-if="shareMsg" class="form-ok" role="status">{{ shareMsg }}</p>

	<!-- Signature: Kassenzettel -->
	<section class="receipt str-anim">
		<div class="halves">
			<div class="half offen">
				<div class="overline">{{ canWrite ? 'Offen' : 'Deine Strafen' }}</div>
				<div class="amount">{{ euro(openShown) }}</div>
			</div>
			<div class="half bezahlt">
				<div class="overline">{{ canWrite ? 'Bezahlt' : 'Kasse gesamt' }}</div>
				<div class="amount">{{ canWrite ? euro(paidShown) : euro(teamOpenShown) }}</div>
			</div>
		</div>
		<template v-if="auth.club?.kasseIban">
			<div class="tear" />
			<div class="foot">
				<div class="overline">Mannschaftskasse</div>
				<div class="iban">{{ auth.club.kasseIban }}</div>
				<div class="owner">{{ auth.club.kasseInhaber }}</div>
			</div>
		</template>
	</section>

	<div class="segmented" style="margin-top: 16px">
		<button :class="{ active: tab === 'kasse' }" @click="tab = 'kasse'">Kasse</button>
		<button :class="{ active: tab === 'katalog' }" @click="tab = 'katalog'">Katalog</button>
		<button v-if="canWrite" :class="{ active: tab === 'protokoll' }" @click="openLog">Protokoll</button>
	</div>

	<!-- Bulk-Leiste -->
	<div v-if="canWrite && selected.size" class="bulk-bar" role="toolbar">
		<strong class="mono">{{ selected.size }}</strong>
		<button class="btn sm gold" @click="bulkPaid(true)"><Check :size="13" /> Bezahlt</button>
		<button class="btn sm" @click="bulkPaid(false)">Offen</button>
		<button class="btn sm danger" @click="bulkDelete"><Trash2 :size="13" /></button>
		<button class="btn sm ghost" style="margin-left: auto" @click="selected = new Set()">×</button>
	</div>

	<!-- ── KASSE ── -->
	<template v-if="tab === 'kasse'">
		<div v-if="byPlayer.length" class="stack">
			<!-- Nach offenem Betrag sortiert = Rangliste der Kabinen-Sünder -->
			<div v-for="(entry, i) in byPlayer" :key="entry.player.id" class="card str-anim" :class="entry.open ? 'hat-offen' : 'ist-bezahlt'">
				<div class="kasse-head">
					<span class="rang" :class="{ top: i === 0 && entry.open > 0 }">{{ entry.player.number ?? i + 1 }}</span>
					<strong>{{ entry.player.name }}</strong>
					<span class="tally-marks" aria-hidden="true">
						<i v-for="n in Math.min(entry.items.filter(i2 => !i2.paid).length, 12)" :key="n" :class="{ fifth: n % 5 === 0 }" />
					</span>
					<span class="mono kasse-sum" :style="{ color: entry.open ? 'var(--bad)' : 'var(--gruen)' }">{{ euro(entry.open) }}</span>
				</div>
				<div class="card-body flush">
					<div v-for="pp in entry.items" :key="pp.id" class="kasse-item" :class="{ paid: pp.paid }">
						<input v-if="canWrite" type="checkbox" class="kasse-check" :checked="selected.has(pp.id)" :aria-label="`${pp.label} auswählen`" @change="toggleSelect(pp.id)" />
						<span class="grow">{{ pp.label }}</span>
						<span class="mono">{{ euro(pp.amount) }}</span>
						<button v-if="canWrite" class="btn sm icon" :class="{ gold: pp.paid }" :aria-label="pp.paid ? 'Wieder öffnen' : 'Als bezahlt'" @click="togglePaid(pp)">
							<component :is="pp.paid ? RotateCcw : Check" :size="13" />
						</button>
						<button v-if="canWrite" class="btn sm icon danger" aria-label="Strafe löschen" @click="removeAssigned(pp)"><Trash2 :size="13" /></button>
					</div>
				</div>
			</div>
		</div>
		<div v-else class="card"><div class="empty">Noch keine Strafen aufgeschrieben. Die Kasse dankt trotzdem.</div></div>
	</template>

	<!-- ── PROTOKOLL ── -->
	<template v-else-if="tab === 'protokoll'">
		<div v-if="logCheck" class="card chain" :class="{ broken: !logCheck.ok }">
			<component :is="logCheck.ok ? ShieldCheck : ShieldAlert" :size="20" />
			<div class="grow">
				<div class="t">{{ logCheck.message }}</div>
				<div class="s">{{ logCheck.count }} Einträge · Hash-Kette geprüft</div>
			</div>
		</div>

		<div class="card">
			<div class="card-body flush">
				<div v-for="e in log" :key="e.id" class="logrow" :class="e.action">
					<span class="mono when">{{ logTime(e.createdAt) }}</span>
					<span class="grow">
						<span class="t">{{ e.actorName }} hat {{ actionLabels[e.action] ?? e.action }}</span>
						<span class="s">
							<template v-if="e.playerName">{{ e.playerName }} · </template>{{ e.label }}
						</span>
					</span>
					<span class="mono amt">{{ euro(e.amount) }}</span>
				</div>
				<p v-if="!log.length" class="empty">Noch keine Bewegungen.</p>
			</div>
		</div>
	</template>

	<!-- ── KATALOG ── -->
	<template v-else>
		<div class="card">
			<div class="card-head">
				<h2>Strafenkatalog</h2>
				<button v-if="canWrite" class="btn sm" @click="openCatalogCreate"><Plus :size="13" /> Neu</button>
			</div>
			<div class="card-body flush">
				<div v-for="p in catalog" :key="p.id" class="lrow">
					<span class="grow">
						<span class="t">{{ p.label }}</span>
						<span v-if="p.unit" class="s">{{ p.unit }}</span>
					</span>
					<span class="mono" style="color: var(--gold); font-weight: 600">{{ euro(p.amount) }}</span>
					<template v-if="canWrite">
						<button class="btn sm icon ghost" aria-label="Bearbeiten" @click="openCatalogEdit(p)"><Pencil :size="13" /></button>
						<button class="btn sm icon danger" aria-label="Löschen" @click="removeCatalog(p)"><Trash2 :size="13" /></button>
					</template>
				</div>
				<p v-if="!catalog.length" class="empty">Noch kein Katalog. Lege den ersten Eintrag an.</p>
			</div>
		</div>
	</template>

	<button v-if="canWrite" class="fab" aria-label="Strafe aufschreiben" @click="openAssign"><Plus :size="24" /></button>

	<!-- Aufschreiben -->
	<AppModal v-if="showAssign" title="Strafe aufschreiben" @close="showAssign = false">
		<form @submit.prevent="submitAssign">
			<p v-if="error" class="form-error" role="alert">{{ error }}</p>
			<div class="field">
				<label>Spieler ({{ assignPlayers.length }})
					<button type="button" class="btn sm ghost" style="margin-left: 8px; min-height: 28px; padding: 3px 10px" @click="toggleAll">
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
						<span>{{ c.label }} <em class="mono" style="color: var(--gold); font-style: normal">{{ euro(c.amount) }}</em></span>
					</label>
				</div>
			</div>
			<div class="row2">
				<div class="field">
					<label for="as-free">Freie Strafe</label>
					<input id="as-free" v-model="freeLabel" maxlength="120" placeholder="z.B. Bierkasten" />
				</div>
				<div class="field">
					<label for="as-free-amount">Betrag (€)</label>
					<input id="as-free-amount" v-model="freeAmount" inputmode="decimal" placeholder="5,00" />
				</div>
			</div>
			<button class="btn primary block" :disabled="!assignPlayers.length">Aufschreiben</button>
		</form>
	</AppModal>

	<!-- Katalog-Eintrag -->
	<AppModal v-if="showCatalogForm" :title="editCatalogId ? 'Eintrag bearbeiten' : 'Neuer Eintrag'" @close="showCatalogForm = false">
		<form @submit.prevent="submitCatalog">
			<p v-if="error" class="form-error" role="alert">{{ error }}</p>
			<div class="field">
				<label for="cat-label">Bezeichnung</label>
				<input id="cat-label" v-model="catalogForm.label" required maxlength="120" />
			</div>
			<div class="row2">
				<div class="field">
					<label for="cat-amount">Betrag (€)</label>
					<input id="cat-amount" v-model="catalogForm.amountEuro" inputmode="decimal" placeholder="5,00" required />
				</div>
				<div class="field">
					<label for="cat-unit">Einheit</label>
					<input id="cat-unit" v-model="catalogForm.unit" maxlength="40" placeholder="pro Minute" />
				</div>
			</div>
			<button class="btn primary block">Speichern</button>
		</form>
	</AppModal>
</template>

<style scoped>
.bulk-bar {
	display: flex;
	align-items: center;
	gap: 7px;
	flex-wrap: wrap;
	padding: 10px 12px;
	margin-bottom: 14px;
	background: var(--gold-bg);
	border: 1px solid var(--line-2);
	border-radius: 13px;
	position: sticky;
	top: calc(var(--appbar-h) + var(--safe-top));
	z-index: 20;
}

.card.hat-offen { border-color: rgba(239, 90, 79, 0.24); }
.card.ist-bezahlt { border-color: rgba(87, 192, 125, 0.2); }
.kasse-head { display: flex; align-items: center; gap: 10px; padding: 12px 15px; border-bottom: 1px solid var(--line); }
.kasse-head strong { font-size: 15.5px; }
.kasse-sum { margin-left: auto; font-weight: 700; font-size: 15px; }
/* Nummern-Badge: Rückennummer, sonst Platz in der Rangliste */
.rang {
	width: 32px; height: 32px;
	flex-shrink: 0;
	display: grid; place-items: center;
	border-radius: 10px;
	background: var(--surface-3);
	font-family: var(--font-mono);
	font-variant-numeric: tabular-nums;
	font-size: 13px;
	font-weight: 700;
	color: var(--gold);
}
.rang.top { background: linear-gradient(180deg, var(--gold-soft), var(--gold)); color: var(--gold-ink); }

.kasse-item { display: flex; align-items: center; gap: 10px; font-size: 14px; padding: 10px 15px; color: var(--ink-2); border-bottom: 1px solid var(--line); }
.kasse-item:last-child { border-bottom: none; }
.kasse-item.paid { opacity: 0.62; }

/* ── Protokoll ── */
.chain { display: flex; align-items: center; gap: 12px; padding: 13px 15px; margin-bottom: 12px; }
.chain :deep(svg) { color: var(--gruen); flex-shrink: 0; }
.chain.broken :deep(svg) { color: var(--bad); }
.chain .t { font-family: var(--font-display); font-size: 14.5px; font-weight: 600; color: var(--ink); }
.chain .s { font-size: 12px; color: var(--ink-3); }
.chain.broken .t { color: var(--bad); }

.logrow { display: flex; align-items: center; gap: 11px; padding: 11px 14px; border-bottom: 1px solid var(--line); }
.logrow:last-child { border-bottom: none; }
.logrow .when { font-size: 11.5px; color: var(--ink-3); flex-shrink: 0; }
.logrow .t { display: block; font-size: 13.5px; color: var(--ink); }
.logrow .s { display: block; font-size: 12px; color: var(--ink-3); }
.logrow .amt { font-size: 13px; color: var(--ink-2); }
.logrow.geloescht .t { color: var(--bad); }
.logrow.bezahlt .t { color: var(--gruen); }
.logrow.wieder_offen .t { color: var(--warn); }
.kasse-item.paid .grow { text-decoration: line-through; }
.kasse-check { accent-color: var(--gold); width: 20px; height: 20px; flex-shrink: 0; }

.pick-list { max-height: 210px; overflow-y: auto; border: 1px solid var(--line-2); border-radius: 11px; background: var(--bg); padding: 4px; }
.pick-item { display: flex; align-items: center; gap: 10px; padding: 10px 10px; font-size: 15px; border-radius: 9px; cursor: pointer; }
.pick-item:hover { background: var(--gold-bg); }
.pick-item input { accent-color: var(--gold); width: 20px; height: 20px; flex-shrink: 0; }
.pick-item span { display: flex; justify-content: space-between; width: 100%; gap: 8px; }
</style>
