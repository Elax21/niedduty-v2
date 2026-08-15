<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { api, apiError } from '../services/api';
import { useRefresh } from '../lib/refresh';
import { useAuthStore } from '../stores/auth';
import { enterRows, countUp } from '../lib/motion';
import { shareKasseImage } from '../lib/shareImage';
import { Plus, Trash2, Check, Share2, Pencil, RotateCcw, ShieldCheck, ShieldAlert, ShoppingBag } from 'lucide-vue-next';
import AppModal from '../components/AppModal.vue';
import type { Penalty, Player, PlayerPenalty, PenaltyLogEntry, PenaltyLogCheck, Expense } from '../types';

const auth = useAuthStore();
const tab = ref<'kasse' | 'ausgaben' | 'katalog' | 'protokoll'>('kasse');
const catalog = ref<Penalty[]>([]);
const assigned = ref<PlayerPenalty[]>([]);
const expenses = ref<Expense[]>([]);
const players = ref<Player[]>([]);
const error = ref('');
const shareMsg = ref('');
const openShown = ref(0);
const paidShown = ref(0);
const teamOpen = ref(0);
const teamOpenShown = ref(0);
// Ausgaben + Kassenstand kommen aus der Summary, damit auch Spieler ohne
// Einblick in fremde Strafen den Stand der Kasse sehen.
const teamSpent = ref(0);
const teamSpentShown = ref(0);
const balanceShown = ref(0);

const canWrite = computed(() => auth.can('strafen'));

function euro(cents: number) {
	return (cents / 100).toLocaleString('de-DE', { style: 'currency', currency: 'EUR' });
}

const openSum = computed(() => assigned.value.filter((p) => !p.paid).reduce((s, p) => s + p.amount, 0));
const paidSum = computed(() => assigned.value.filter((p) => p.paid).reduce((s, p) => s + p.amount, 0));

// Filter für Aufschreiber: bei 20 Spielern ist die Liste sonst lang, wenn man
// nur wissen will, was einer offen hat.
const filterPlayer = ref('');
const filterOpenOnly = ref(false);

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

/** Angezeigte Liste — nach Spieler und „nur offen" gefiltert. */
const shownPlayers = computed(() =>
	byPlayer.value.filter(
		(e) => (!filterPlayer.value || e.player.id === filterPlayer.value) && (!filterOpenOnly.value || e.open > 0)
	)
);

async function load() {
	const [c, a, p, s, ex] = await Promise.all([
		api.get<Penalty[]>('/penalties'),
		api.get<PlayerPenalty[]>('/player-penalties'),
		api.get<Player[]>('/players'),
		api.get<{ totalOpen: number; totalPaid: number; totalSpent: number; balance: number }>('/player-penalties/summary'),
		api.get<Expense[]>('/expenses')
	]);
	catalog.value = c.data;
	assigned.value = a.data;
	players.value = p.data;
	expenses.value = ex.data;
	teamOpen.value = s.data.totalOpen;
	teamSpent.value = s.data.totalSpent;
	selected.value = new Set();
	countUp(openSum.value, (v) => (openShown.value = v));
	countUp(paidSum.value, (v) => (paidShown.value = v));
	countUp(teamOpen.value, (v) => (teamOpenShown.value = v));
	countUp(teamSpent.value, (v) => (teamSpentShown.value = v));
	countUp(s.data.balance, (v) => (balanceShown.value = v));
	requestAnimationFrame(() => enterRows('.str-anim'));
}

// ── Strafe(n) aufschreiben ──
const showAssign = ref(false);
const assignPlayers = ref<string[]>([]);
const assignPenalties = ref<string[]>([]);
const freeLabel = ref('');
const freeAmount = ref('');
// Menge je Katalog-Eintrag (nur bei perUnit) — z.B. 7 Minuten Verspätung.
const assignQty = ref<Record<string, number>>({});

/** Was der gewählte Eintrag mit der eingestellten Menge kostet. */
function penaltyCost(p: Penalty) {
	return p.perUnit ? p.amount * Math.max(1, assignQty.value[p.id] ?? 1) : p.amount;
}
/** Summe der Zuweisung pro Spieler (ohne freie Strafe). */
const assignSum = computed(() =>
	catalog.value.filter((c) => assignPenalties.value.includes(c.id)).reduce((s, c) => s + penaltyCost(c), 0)
);

function openAssign() {
	assignPlayers.value = [];
	assignPenalties.value = [];
	assignQty.value = {};
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
	const quantities: Record<string, number> = {};
	for (const id of assignPenalties.value) {
		const pen = catalog.value.find((c) => c.id === id);
		if (pen?.perUnit) quantities[id] = Math.max(1, assignQty.value[id] ?? 1);
	}
	try {
		await api.post('/player-penalties', {
			playerIds: assignPlayers.value,
			penaltyIds: assignPenalties.value,
			quantities,
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

// ── Ausgaben ───────────────────────────────────────────────────
// Was die Kasse verlässt: Bälle, Mannschaftsabend, Essen. Wird vom
// bezahlten Geld abgezogen und steht ebenfalls im Protokoll.
const showExpense = ref(false);
const expenseForm = ref({ label: '', amountEuro: '', date: '' });

function heute() {
	return new Date().toISOString().slice(0, 10);
}
function openExpense() {
	expenseForm.value = { label: '', amountEuro: '', date: heute() };
	error.value = '';
	showExpense.value = true;
}
async function submitExpense() {
	error.value = '';
	const amount = Math.round(parseFloat(expenseForm.value.amountEuro.replace(',', '.')) * 100);
	if (!expenseForm.value.label.trim() || !amount || amount <= 0) {
		error.value = 'Grund und gültiger Betrag nötig';
		return;
	}
	try {
		await api.post('/expenses', {
			label: expenseForm.value.label.trim(),
			amount,
			date: expenseForm.value.date || heute()
		});
		showExpense.value = false;
		tab.value = 'ausgaben';
		await load();
	} catch (e) {
		error.value = apiError(e);
	}
}
async function removeExpense(e: Expense) {
	if (!window.confirm(`Ausgabe „${e.label}" (${euro(e.amount)}) wirklich löschen?`)) return;
	await api.delete(`/expenses/${e.id}`);
	await load();
}
function tagKurz(d: string) {
	if (!d) return '';
	const [y, m, day] = d.split('-');
	return `${day}.${m}.${y.slice(2)}`;
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
	katalog_geloescht: 'Katalog-Eintrag gelöscht',
	ausgabe: 'Geld ausgegeben',
	ausgabe_geloescht: 'Ausgabe gelöscht'
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
const catalogForm = ref({ label: '', amountEuro: '', unit: '', perUnit: false, unitLabel: '' });

function openCatalogCreate() {
	editCatalogId.value = null;
	catalogForm.value = { label: '', amountEuro: '', unit: '', perUnit: false, unitLabel: '' };
	error.value = '';
	showCatalogForm.value = true;
}
function openCatalogEdit(p: Penalty) {
	editCatalogId.value = p.id;
	catalogForm.value = {
		label: p.label, amountEuro: (p.amount / 100).toFixed(2).replace('.', ','), unit: p.unit,
		perUnit: p.perUnit, unitLabel: p.unitLabel || 'Minuten'
	};
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
		perUnit: catalogForm.value.perUnit,
		unitLabel: catalogForm.value.perUnit ? catalogForm.value.unitLabel.trim() || 'Minuten' : '',
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
		<!-- Ausgaben + Kassenstand: was rausging und was übrig ist -->
		<div class="bilanz">
			<div class="b-item">
				<span class="l">Ausgaben</span>
				<span class="v raus mono">−{{ euro(teamSpentShown) }}</span>
			</div>
			<div class="b-item">
				<span class="l">Kassenstand</span>
				<span class="v mono" :class="{ minus: balanceShown < 0 }">{{ euro(balanceShown) }}</span>
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
		<button :class="{ active: tab === 'ausgaben' }" @click="tab = 'ausgaben'">Ausgaben</button>
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
		<!-- Filter: wer? und nur die, die noch was offen haben -->
		<div v-if="canWrite && byPlayer.length" class="kasse-filter">
			<select v-model="filterPlayer" class="input" aria-label="Nach Spieler filtern">
				<option value="">Alle Spieler ({{ byPlayer.length }})</option>
				<option v-for="e in byPlayer" :key="e.player.id" :value="e.player.id">
					{{ e.player.name }} — {{ euro(e.open) }} offen
				</option>
			</select>
			<button class="btn sm" :class="{ gold: filterOpenOnly }" @click="filterOpenOnly = !filterOpenOnly">
				Nur offen
			</button>
		</div>

		<div v-if="shownPlayers.length" class="stack">
			<!-- Nach offenem Betrag sortiert = Rangliste der Kabinen-Sünder -->
			<div v-for="(entry, i) in shownPlayers" :key="entry.player.id" class="card str-anim" :class="entry.open ? 'hat-offen' : 'ist-bezahlt'">
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
		<div v-else class="card">
			<div class="empty">
				{{ byPlayer.length ? 'Kein Treffer — Filter zurücksetzen?' : 'Noch keine Strafen aufgeschrieben. Die Kasse dankt trotzdem.' }}
			</div>
		</div>
	</template>

	<!-- ── AUSGABEN ── -->
	<template v-else-if="tab === 'ausgaben'">
		<div class="card ist-ausgabe">
			<div class="card-head">
				<h2>Aus der Kasse bezahlt</h2>
				<button v-if="canWrite" class="btn sm warn" @click="openExpense"><Plus :size="13" /> Ausgabe</button>
			</div>
			<div class="card-body flush">
				<div v-for="e in expenses" :key="e.id" class="exrow str-anim">
					<span class="ex-icon"><ShoppingBag :size="15" /></span>
					<span class="grow">
						<span class="t">{{ e.label }}</span>
						<span class="s">{{ tagKurz(e.date) }} · {{ e.creatorName }}</span>
					</span>
					<span class="mono amt">−{{ euro(e.amount) }}</span>
					<button v-if="canWrite" class="btn sm icon danger" aria-label="Ausgabe löschen" @click="removeExpense(e)"><Trash2 :size="13" /></button>
				</div>
				<p v-if="!expenses.length" class="empty">Noch nichts ausgegeben. Die Kasse ist voll.</p>
			</div>
		</div>
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
					<span class="mono kat-betrag">
						{{ euro(p.amount) }}
						<span v-if="p.perUnit" class="je">× {{ p.unitLabel || 'Einheit' }}</span>
					</span>
					<template v-if="canWrite">
						<button class="btn sm icon ghost" aria-label="Bearbeiten" @click="openCatalogEdit(p)"><Pencil :size="13" /></button>
						<button class="btn sm icon danger" aria-label="Löschen" @click="removeCatalog(p)"><Trash2 :size="13" /></button>
					</template>
				</div>
				<p v-if="!catalog.length" class="empty">Noch kein Katalog. Lege den ersten Eintrag an.</p>
			</div>
		</div>
	</template>

	<!-- FAB folgt dem Tab: auf „Ausgaben" wird Geld ausgegeben, sonst aufgeschrieben -->
	<button
		v-if="canWrite"
		class="fab"
		:class="{ warn: tab === 'ausgaben' }"
		:aria-label="tab === 'ausgaben' ? 'Ausgabe eintragen' : 'Strafe aufschreiben'"
		@click="tab === 'ausgaben' ? openExpense() : openAssign()"
	>
		<Plus :size="24" />
	</button>

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
					<div v-for="c in catalog" :key="c.id" class="pick-wrap">
						<label class="pick-item">
							<input v-model="assignPenalties" type="checkbox" :value="c.id" @change="assignQty[c.id] = assignQty[c.id] ?? 1" />
							<span>
								{{ c.label }}
								<em class="mono" style="color: var(--gold); font-style: normal">
									{{ euro(penaltyCost(c)) }}<template v-if="c.perUnit"><span class="je">{{ Math.max(1, assignQty[c.id] ?? 1) }} × {{ euro(c.amount) }}</span></template>
								</em>
							</span>
						</label>
						<!-- Mengen-Strafe: z.B. 7 Minuten zu spät = 7 × 50 ct -->
						<div v-if="c.perUnit && assignPenalties.includes(c.id)" class="qty">
							<span class="ql">{{ c.unitLabel || 'Menge' }}</span>
							<button type="button" class="btn sm icon" aria-label="Weniger" @click="assignQty[c.id] = Math.max(1, (assignQty[c.id] ?? 1) - 1)">−</button>
							<input
								v-model.number="assignQty[c.id]"
								type="number"
								class="qty-input mono"
								min="1"
								max="500"
								inputmode="numeric"
								:aria-label="`${c.label}: ${c.unitLabel || 'Menge'}`"
							/>
							<button type="button" class="btn sm icon" aria-label="Mehr" @click="assignQty[c.id] = Math.min(500, (assignQty[c.id] ?? 1) + 1)">+</button>
						</div>
					</div>
				</div>
				<p v-if="assignSum" class="hint">Je Spieler: <strong class="mono">{{ euro(assignSum) }}</strong></p>
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
					<label for="cat-unit">Zusatz</label>
					<input id="cat-unit" v-model="catalogForm.unit" maxlength="120" placeholder="pro Minute" />
				</div>
			</div>
			<label class="switch-row">
				<input v-model="catalogForm.perUnit" type="checkbox" />
				<span>
					<span class="t">Betrag je Einheit</span>
					<span class="s">Beim Aufschreiben wird eine Menge abgefragt (z.B. Minuten Verspätung) und multipliziert.</span>
				</span>
			</label>
			<div v-if="catalogForm.perUnit" class="field">
				<label for="cat-unitlabel">Einheit (Mehrzahl)</label>
				<input id="cat-unitlabel" v-model="catalogForm.unitLabel" maxlength="40" placeholder="Minuten" />
			</div>
			<button class="btn primary block">Speichern</button>
		</form>
	</AppModal>

	<!-- Ausgabe aus der Kasse -->
	<AppModal v-if="showExpense" title="Geld ausgeben" @close="showExpense = false">
		<form @submit.prevent="submitExpense">
			<p v-if="error" class="form-error" role="alert">{{ error }}</p>
			<div class="field">
				<label for="ex-label">Grund</label>
				<input id="ex-label" v-model="expenseForm.label" required maxlength="120" placeholder="z.B. Bälle gekauft" />
			</div>
			<div class="row2">
				<div class="field">
					<label for="ex-amount">Betrag (€)</label>
					<input id="ex-amount" v-model="expenseForm.amountEuro" inputmode="decimal" placeholder="50,00" required />
				</div>
				<div class="field">
					<label for="ex-date">Datum</label>
					<input id="ex-date" v-model="expenseForm.date" type="date" />
				</div>
			</div>
			<button class="btn primary block">Aus der Kasse abziehen</button>
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

.kasse-filter { display: flex; gap: 8px; margin-bottom: 12px; }
.kasse-filter .input { flex: 1; min-width: 0; min-height: 36px; padding: 6px 10px; font-size: 14px; }
.kasse-filter .btn { flex-shrink: 0; }

.hint { font-size: 12.5px; color: var(--ink-3); margin-top: 8px; }
.hint strong { color: var(--gold); }

/* ── Ausgaben (Geld raus = gelb, nicht rot) ── */
.card.ist-ausgabe { border-color: rgba(231, 187, 70, 0.28); }
.exrow { display: flex; align-items: center; gap: 11px; padding: 11px 14px; border-bottom: 1px solid var(--line); }
.exrow:last-child { border-bottom: none; }
.exrow .ex-icon {
	width: 32px; height: 32px;
	flex-shrink: 0;
	display: grid; place-items: center;
	border-radius: 10px;
	background: var(--warn-bg);
	color: var(--warn);
}
.exrow .t { display: block; font-size: 14px; color: var(--ink); }
.exrow .s { display: block; font-size: 12px; color: var(--ink-3); }
.exrow .amt { font-size: 14.5px; font-weight: 700; color: var(--warn); }
.logrow.ausgabe .t { color: var(--warn); }
.logrow.ausgabe_geloescht .t { color: var(--ink-3); }

/* ── Mengen-Strafe (z.B. Minuten Verspätung) ── */
.kat-betrag { color: var(--gold); font-weight: 600; text-align: right; }
.je { display: block; font-family: var(--font-body); font-size: 11px; font-weight: 400; color: var(--ink-3); }
.pick-wrap + .pick-wrap { border-top: 1px solid var(--line); }
.qty { display: flex; align-items: center; gap: 8px; padding: 0 10px 10px 40px; }
.qty .ql { font-size: 12px; color: var(--ink-3); margin-right: auto; }
.qty-input {
	width: 64px;
	text-align: center;
	padding: 6px 4px;
	min-height: 36px;
	font-size: 15px;
	font-variant-numeric: tabular-nums;
	color: var(--ink);
	background: var(--surface-3);
	border: 1px solid var(--line-2);
	border-radius: 9px;
	-moz-appearance: textfield;
}
.qty-input::-webkit-outer-spin-button, .qty-input::-webkit-inner-spin-button { -webkit-appearance: none; margin: 0; }

.switch-row { display: flex; align-items: flex-start; gap: 11px; padding: 11px 0 4px; cursor: pointer; }
.switch-row input { accent-color: var(--gold); width: 20px; height: 20px; flex-shrink: 0; margin-top: 2px; }
.switch-row .t { display: block; font-size: 14.5px; color: var(--ink); }
.switch-row .s { display: block; font-size: 12px; color: var(--ink-3); margin-top: 2px; }

.pick-list { max-height: 210px; overflow-y: auto; border: 1px solid var(--line-2); border-radius: 11px; background: var(--bg); padding: 4px; }
.pick-item { display: flex; align-items: center; gap: 10px; padding: 10px 10px; font-size: 15px; border-radius: 9px; cursor: pointer; }
.pick-item:hover { background: var(--gold-bg); }
.pick-item input { accent-color: var(--gold); width: 20px; height: 20px; flex-shrink: 0; }
.pick-item span { display: flex; justify-content: space-between; width: 100%; gap: 8px; }
</style>
