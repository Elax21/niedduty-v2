<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { api, apiError } from '../services/api';
import { useAuthStore } from '../stores/auth';
import { enterRows } from '../lib/motion';
import { Plus, Pencil, Trash2, Check, X, MapPin, Clock, CalendarPlus, CalendarDays, ExternalLink } from 'lucide-vue-next';
import AppModal from '../components/AppModal.vue';
import type { Attendance, Occurrence, Match, Matches } from '../types';

const auth = useAuthStore();
const scope = ref<'kommend' | 'vergangen'>('kommend');
const events = ref<Occurrence[]>([]);
const matchList = ref<Match[]>([]);
const attendance = ref<Record<string, Attendance[]>>({});
const error = ref('');

const today = new Date().toISOString().slice(0, 10);
const canManage = computed(() => auth.can('termine'));

const typeLabels: Record<string, string> = {
	training: 'Training', spiel: 'Spiel', mannschaftsabend: 'Abend', sonstiges: 'Termin'
};

function iso(d: Date) { return d.toISOString().slice(0, 10); }

// ── Vereinheitlichte Terminliste (eigene Termine + fussball.de-Spiele) ──
interface Item {
	key: string;        // eventKey für Zu-/Absage
	isoDate: string;
	time: string;
	type: string;
	title: string;
	location: string;
	notes: string;
	isMatch: boolean;
	side?: string;      // Heim | Auswärts
	result?: string;    // "3:1" bei gespielten Spielen
	url?: string;       // fussball.de-Spielseite
	occ?: Occurrence;   // Original (nur eigene Termine, für Bearbeiten)
}

function eventItem(e: Occurrence): Item {
	return { key: e.eventKey, isoDate: e.occDate, time: e.startTime, type: e.type, title: e.title, location: e.location, notes: e.notes, isMatch: false, occ: e };
}
function matchItem(m: Match): Item {
	const opp = m.home.isOwn ? m.guest.name : m.home.name;
	return {
		key: 'fdm_' + m.id, isoDate: m.isoDate, time: m.time, type: 'spiel',
		title: opp, location: '', notes: '', isMatch: true,
		side: m.home.isOwn ? 'Heim' : 'Auswärts',
		result: m.played ? `${m.homeGoals}:${m.guestGoals}` : undefined,
		url: m.url
	};
}

const items = computed<Item[]>(() => {
	const list = [...events.value.map(eventItem), ...matchList.value.map(matchItem)];
	list.sort((a, b) => (a.isoDate + a.time).localeCompare(b.isoDate + b.time));
	return scope.value === 'vergangen' ? list.reverse() : list;
});

const groups = computed(() => {
	const map = new Map<string, { label: string; items: Item[] }>();
	for (const it of items.value) {
		if (!it.isoDate) continue;
		const key = it.isoDate.slice(0, 7);
		if (!map.has(key)) {
			map.set(key, { label: new Date(it.isoDate + 'T12:00').toLocaleDateString('de-DE', { month: 'long', year: 'numeric' }), items: [] });
		}
		map.get(key)!.items.push(it);
	}
	return [...map.values()];
});

async function load() {
	const from = scope.value === 'kommend' ? today : iso(new Date(Date.now() - 90 * 86400000));
	const to = scope.value === 'kommend' ? iso(new Date(Date.now() + 120 * 86400000)) : today;
	const [ev, mt] = await Promise.all([
		api.get<Occurrence[]>('/events', { params: { from, to } }),
		api.get<Matches>('/fussball/matches').catch(() => ({ data: null as Matches | null }))
	]);
	events.value = scope.value === 'kommend'
		? ev.data.filter((e) => e.occDate >= today)
		: ev.data.filter((e) => e.occDate < today);
	const m = mt.data;
	if (m) {
		matchList.value = scope.value === 'kommend'
			? m.next.filter((x) => x.isoDate && x.isoDate >= today)
			: m.previous.filter((x) => x.isoDate && x.isoDate < today);
	} else {
		matchList.value = [];
	}
	attendance.value = {};
	items.value.slice(0, 30).forEach((it) => loadAttendance(it.key));
	requestAnimationFrame(() => enterRows('.ev-anim'));
}
watch(scope, load);

async function loadAttendance(key: string) {
	const { data } = await api.get<Attendance[]>('/attendance', { params: { eventKey: key } });
	attendance.value = { ...attendance.value, [key]: data };
}
function counts(key: string) {
	const list = attendance.value[key] ?? [];
	return { yes: list.filter((a) => a.status === 'attending').length, no: list.filter((a) => a.status === 'declined').length };
}
function myStatus(key: string): string | null {
	if (!auth.user?.playerId) return null;
	return attendance.value[key]?.find((a) => a.playerId === auth.user!.playerId)?.status ?? null;
}
async function rsvp(key: string, status: 'attending' | 'declined') {
	let reason = '';
	if (status === 'declined') reason = window.prompt('Grund der Absage (optional):') ?? '';
	await api.put('/attendance', { eventKey: key, playerId: auth.user!.playerId, status, reason });
	await loadAttendance(key);
}

function googleUrl(it: Item): string {
	const base = 'https://calendar.google.com/calendar/render?action=TEMPLATE';
	const d = it.isoDate.replace(/-/g, '');
	let dates: string;
	if (it.time) {
		const st = it.time.replace(':', '') + '00';
		dates = `${d}T${st}/${d}T${st}`;
	} else {
		const next = new Date(it.isoDate + 'T12:00');
		next.setDate(next.getDate() + 1);
		dates = `${d}/${next.toISOString().slice(0, 10).replace(/-/g, '')}`;
	}
	const title = it.isMatch ? `${typeLabels[it.type]}: ${it.title}` : it.title;
	const p = new URLSearchParams({ text: title, dates, location: it.location || '', details: it.notes || '' });
	return `${base}&${p.toString()}`;
}

// ── Anlegen / Bearbeiten (nur eigene Termine) ──
const showForm = ref(false);
const editId = ref<string | null>(null);
const form = ref({
	title: '', type: 'training', date: '', startTime: '', endTime: '',
	location: '', notes: '', recurring: false, recurrenceType: 'weekly', recurrenceEnd: ''
});

function openCreate() {
	editId.value = null;
	form.value = { title: '', type: 'training', date: today, startTime: '19:15', endTime: '', location: '', notes: '', recurring: false, recurrenceType: 'weekly', recurrenceEnd: '' };
	error.value = '';
	showForm.value = true;
}
function openEdit(e: Occurrence) {
	editId.value = e.id;
	form.value = {
		title: e.title, type: e.type, date: e.date, startTime: e.startTime, endTime: e.endTime,
		location: e.location, notes: e.notes, recurring: e.recurring, recurrenceType: e.recurrenceType || 'weekly', recurrenceEnd: e.recurrenceEnd
	};
	error.value = '';
	showForm.value = true;
}
async function submitForm() {
	error.value = '';
	const payload: Record<string, unknown> = { ...form.value };
	if (!form.value.recurring) { payload.recurrenceType = ''; payload.recurrenceEnd = ''; }
	try {
		if (editId.value) await api.put(`/events/${editId.value}`, payload);
		else await api.post('/events', payload);
		showForm.value = false;
		await load();
	} catch (e) {
		error.value = apiError(e);
	}
}
async function removeEvent(e: Occurrence) {
	if (!window.confirm(`„${e.title}" löschen? Serientermine werden komplett entfernt.`)) return;
	await api.delete(`/events/${e.id}`);
	await load();
}

onMounted(load);
</script>

<template>
	<div class="page-head">
		<h1>Termine</h1>
		<a v-if="auth.club?.googleCalendarUrl" :href="auth.club.googleCalendarUrl" target="_blank" rel="noopener" class="btn sm ghost">
			<CalendarDays :size="14" /> Google
		</a>
	</div>

	<div class="segmented">
		<button :class="{ active: scope === 'kommend' }" @click="scope = 'kommend'">Kommend</button>
		<button :class="{ active: scope === 'vergangen' }" @click="scope = 'vergangen'">Vergangen</button>
	</div>

	<p v-if="error && !showForm" class="form-error" role="alert">{{ error }}</p>

	<template v-if="groups.length">
		<div v-for="g in groups" :key="g.label" class="month-group">
			<div class="month-label">{{ g.label }}</div>
			<div class="card">
				<div class="card-body flush">
					<div v-for="it in g.items" :key="it.key" class="ev ev-anim" :class="{ today: it.isoDate === today, match: it.isMatch }">
						<div class="ev-date">
							<span class="dow">{{ new Date(it.isoDate + 'T12:00').toLocaleDateString('de-DE', { weekday: 'short' }) }}</span>
							<span class="dnum">{{ new Date(it.isoDate + 'T12:00').getDate() }}</span>
						</div>
						<div class="ev-body">
							<div class="ev-top">
								<span class="chip" :class="it.type">{{ typeLabels[it.type] }}</span>
								<strong class="ev-title">
									<template v-if="it.isMatch">{{ it.side === 'Heim' ? 'vs' : 'bei' }} {{ it.title }}</template>
									<template v-else>{{ it.title }}</template>
								</strong>
								<span v-if="it.result" class="ev-result mono">{{ it.result }}</span>
							</div>
							<div class="ev-meta">
								<span v-if="it.time"><Clock :size="12" /> {{ it.time }}</span>
								<span v-if="it.side">{{ it.side }}</span>
								<span v-if="it.location"><MapPin :size="12" /> {{ it.location }}</span>
							</div>
							<p v-if="it.notes" class="ev-notes">{{ it.notes }}</p>

							<div class="ev-actions">
								<span class="rsvp-count"><Check :size="13" /> {{ counts(it.key).yes }}</span>
								<span class="rsvp-count no"><X :size="13" /> {{ counts(it.key).no }}</span>

								<template v-if="auth.user?.playerId && it.isoDate >= today">
									<button class="btn sm" :class="{ gold: myStatus(it.key) === 'attending' }" @click="rsvp(it.key, 'attending')">Zusage</button>
									<button class="btn sm" :class="{ danger: myStatus(it.key) === 'declined' }" @click="rsvp(it.key, 'declined')">Absage</button>
								</template>

								<a v-if="it.isMatch && it.url" :href="it.url" target="_blank" rel="noopener" class="btn sm icon ghost" aria-label="Auf fussball.de öffnen" title="Auf fussball.de"><ExternalLink :size="14" /></a>
								<a :href="googleUrl(it)" target="_blank" rel="noopener" class="btn sm icon ghost" aria-label="Zu Google Kalender" title="Zu Google Kalender"><CalendarPlus :size="14" /></a>
								<template v-if="canManage && !it.isMatch && it.occ">
									<button class="btn sm icon ghost" aria-label="Bearbeiten" @click="openEdit(it.occ)"><Pencil :size="13" /></button>
									<button class="btn sm icon danger" aria-label="Löschen" @click="removeEvent(it.occ)"><Trash2 :size="13" /></button>
								</template>
							</div>
						</div>
					</div>
				</div>
			</div>
		</div>
	</template>
	<div v-else class="card">
		<div class="empty"><CalendarDays :size="30" class="ic" /><br />{{ scope === 'kommend' ? 'Keine anstehenden Termine.' : 'Keine vergangenen Termine.' }}</div>
	</div>

	<button v-if="canManage" class="fab" aria-label="Termin anlegen" @click="openCreate"><Plus :size="24" /></button>

	<AppModal v-if="showForm" :title="editId ? 'Termin bearbeiten' : 'Neuer Termin'" @close="showForm = false">
		<form @submit.prevent="submitForm">
			<p v-if="error" class="form-error" role="alert">{{ error }}</p>
			<div class="field">
				<label for="ev-title">Titel</label>
				<input id="ev-title" v-model="form.title" required maxlength="120" />
			</div>
			<div class="field">
				<label for="ev-type">Typ</label>
				<select id="ev-type" v-model="form.type">
					<option value="training">Training</option>
					<option value="spiel">Spiel</option>
					<option value="mannschaftsabend">Mannschaftsabend</option>
					<option value="sonstiges">Sonstiges</option>
				</select>
			</div>
			<div class="field">
				<label for="ev-date">Datum</label>
				<input id="ev-date" v-model="form.date" type="date" required />
			</div>
			<div class="row2">
				<div class="field">
					<label for="ev-start">Beginn</label>
					<input id="ev-start" v-model="form.startTime" type="time" />
				</div>
				<div class="field">
					<label for="ev-end">Ende</label>
					<input id="ev-end" v-model="form.endTime" type="time" />
				</div>
			</div>
			<div class="field">
				<label for="ev-loc">Ort</label>
				<input id="ev-loc" v-model="form.location" maxlength="120" />
			</div>
			<div class="field">
				<label for="ev-notes">Notiz</label>
				<textarea id="ev-notes" v-model="form.notes" rows="2" maxlength="2000" />
			</div>
			<label class="check-line">
				<input v-model="form.recurring" type="checkbox" /> <span>Wiederholen</span>
			</label>
			<div v-if="form.recurring" class="row2" style="margin-top: 12px">
				<div class="field">
					<label for="ev-rtype">Rhythmus</label>
					<select id="ev-rtype" v-model="form.recurrenceType">
						<option value="weekly">Wöchentlich</option>
						<option value="biweekly">Alle 2 Wochen</option>
					</select>
				</div>
				<div class="field">
					<label for="ev-rend">Serie bis</label>
					<input id="ev-rend" v-model="form.recurrenceEnd" type="date" />
				</div>
			</div>
			<button class="btn primary block" style="margin-top: 6px">{{ editId ? 'Speichern' : 'Anlegen' }}</button>
		</form>
	</AppModal>
</template>

<style scoped>
.month-group + .month-group { margin-top: 18px; }
.month-label {
	font-family: var(--font-display);
	font-size: 13px;
	font-weight: 600;
	text-transform: uppercase;
	letter-spacing: 0.06em;
	color: var(--gold);
	margin: 0 4px 8px;
}
.ev { display: flex; gap: 12px; padding: 14px 15px; border-bottom: 1px solid var(--line); }
.ev:last-child { border-bottom: none; }
.ev.today { background: linear-gradient(90deg, var(--gold-bg), transparent 70%); box-shadow: inset 3px 0 0 var(--gold); }
.ev.match { background: linear-gradient(90deg, var(--rot-bg), transparent 78%); }
.ev.match.today { box-shadow: inset 3px 0 0 var(--rot); }
.ev-date {
	flex-shrink: 0;
	width: 46px;
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	gap: 1px;
	background: var(--surface-3);
	border: 1px solid var(--line);
	border-radius: 11px;
	padding: 6px 0;
	align-self: flex-start;
}
.ev-date .dow { font-family: var(--font-display); font-size: 11px; text-transform: uppercase; color: var(--ink-3); }
.ev-date .dnum { font-family: var(--font-mono); font-size: 22px; font-weight: 600; color: var(--gold); line-height: 1; }
.ev-body { flex: 1; min-width: 0; }
.ev-top { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.ev-title { font-size: 16px; }
.ev-result { margin-left: auto; font-size: 16px; font-weight: 700; color: var(--gold); }
.ev-meta { display: flex; flex-wrap: wrap; gap: 4px 14px; font-size: 12.5px; color: var(--ink-2); margin-top: 5px; }
.ev-meta span { display: inline-flex; align-items: center; gap: 4px; }
.ev-meta :deep(svg) { color: var(--gold); }
.ev-notes { font-size: 12.5px; color: var(--ink-3); margin-top: 5px; }
.ev-actions { display: flex; align-items: center; gap: 7px; flex-wrap: wrap; margin-top: 10px; }
.rsvp-count { display: inline-flex; align-items: center; gap: 3px; font-family: var(--font-mono); font-size: 13px; color: var(--gruen); }
.rsvp-count.no { color: var(--bad); }
.check-line { display: flex; align-items: center; gap: 9px; font-size: 15px; color: var(--ink); cursor: pointer; }
.check-line input { width: 20px; height: 20px; accent-color: var(--gold); }
</style>
