<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { api, apiError } from '../services/api';
import { useAuthStore } from '../stores/auth';
import { ChevronLeft, ChevronRight, Plus, Pencil, Trash2, Check, X } from 'lucide-vue-next';
import AppModal from '../components/AppModal.vue';
import type { Attendance, Occurrence } from '../types';

const auth = useAuthStore();

const now = new Date();
const year = ref(now.getFullYear());
const month = ref(now.getMonth()); // 0-basiert
const events = ref<Occurrence[]>([]);
const selectedDay = ref(now.toISOString().slice(0, 10));
const attendance = ref<Record<string, Attendance[]>>({});
const error = ref('');

const typeLabels: Record<string, string> = {
	training: 'Training', spiel: 'Spiel', mannschaftsabend: 'Abend', sonstiges: 'Termin'
};
const monthName = computed(() =>
	new Date(year.value, month.value, 1).toLocaleDateString('de-DE', { month: 'long', year: 'numeric' })
);

function iso(y: number, m: number, d: number) {
	return `${y}-${String(m + 1).padStart(2, '0')}-${String(d).padStart(2, '0')}`;
}

// Wochen-Raster (Mo–So)
const weeks = computed(() => {
	const first = new Date(year.value, month.value, 1);
	const offset = (first.getDay() + 6) % 7;
	const daysInMonth = new Date(year.value, month.value + 1, 0).getDate();
	const cells: ({ date: string; day: number } | null)[] = Array(offset).fill(null);
	for (let d = 1; d <= daysInMonth; d++) cells.push({ date: iso(year.value, month.value, d), day: d });
	while (cells.length % 7 !== 0) cells.push(null);
	const out = [];
	for (let i = 0; i < cells.length; i += 7) out.push(cells.slice(i, i + 7));
	return out;
});

const byDay = computed(() => {
	const map: Record<string, Occurrence[]> = {};
	for (const e of events.value) (map[e.occDate] ??= []).push(e);
	return map;
});
const dayEvents = computed(() => byDay.value[selectedDay.value] ?? []);
const todayIso = new Date().toISOString().slice(0, 10);

async function load() {
	const from = iso(year.value, month.value, 1);
	const to = iso(year.value, month.value, new Date(year.value, month.value + 1, 0).getDate());
	const { data } = await api.get<Occurrence[]>('/events', { params: { from, to } });
	events.value = data;
}

function shiftMonth(d: number) {
	const m = new Date(year.value, month.value + d, 1);
	year.value = m.getFullYear();
	month.value = m.getMonth();
}
watch([year, month], load);

async function loadAttendance(key: string) {
	const { data } = await api.get<Attendance[]>('/attendance', { params: { eventKey: key } });
	attendance.value[key] = data;
}
watch(dayEvents, (list) => list.forEach((e) => loadAttendance(e.eventKey)), { immediate: false });

function counts(key: string) {
	const list = attendance.value[key] ?? [];
	return {
		yes: list.filter((a) => a.status === 'attending').length,
		no: list.filter((a) => a.status === 'declined').length
	};
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

// ── Termin anlegen / bearbeiten ──
const showForm = ref(false);
const editId = ref<string | null>(null);
const form = ref({
	title: '', type: 'training', date: '', endDate: '', startTime: '', endTime: '',
	location: '', notes: '', recurring: false, recurrenceType: 'weekly', recurrenceEnd: ''
});

function openCreate() {
	editId.value = null;
	form.value = {
		title: '', type: 'training', date: selectedDay.value, endDate: '', startTime: '19:30', endTime: '',
		location: '', notes: '', recurring: false, recurrenceType: 'weekly', recurrenceEnd: ''
	};
	showForm.value = true;
}
function openEdit(e: Occurrence) {
	editId.value = e.id;
	form.value = {
		title: e.title, type: e.type, date: e.date, endDate: e.endDate, startTime: e.startTime,
		endTime: e.endTime, location: e.location, notes: e.notes,
		recurring: e.recurring, recurrenceType: e.recurrenceType || 'weekly', recurrenceEnd: e.recurrenceEnd
	};
	showForm.value = true;
}
async function submitForm() {
	error.value = '';
	const payload = { ...form.value };
	if (!payload.recurring) {
		payload.recurrenceType = '';
		payload.recurrenceEnd = '';
	}
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
	if (!window.confirm(`„${e.title}" wirklich löschen? Bei Serienterminen wird die ganze Serie entfernt.`)) return;
	await api.delete(`/events/${e.id}`);
	await load();
}

onMounted(load);
</script>

<template>
	<div class="page-head">
		<h1>Kalender</h1>
		<button v-if="auth.can('termine')" class="btn primary sm" @click="openCreate"><Plus :size="14" aria-hidden="true" /> Termin</button>
	</div>

	<div class="grid kal-grid">
		<div class="card">
			<div class="card-head">
				<button class="btn sm" aria-label="Voriger Monat" @click="shiftMonth(-1)"><ChevronLeft :size="15" /></button>
				<h2>{{ monthName }}</h2>
				<button class="btn sm" aria-label="Nächster Monat" @click="shiftMonth(1)"><ChevronRight :size="15" /></button>
			</div>
			<div class="card-body">
				<div class="kal-week kal-head-row">
					<span v-for="d in ['Mo', 'Di', 'Mi', 'Do', 'Fr', 'Sa', 'So']" :key="d">{{ d }}</span>
				</div>
				<div v-for="(week, wi) in weeks" :key="wi" class="kal-week">
					<button
						v-for="(cell, ci) in week"
						:key="ci"
						class="kal-day"
						:class="{
							empty: !cell,
							selected: cell?.date === selectedDay,
							today: cell?.date === todayIso
						}"
						:disabled="!cell"
						@click="cell && (selectedDay = cell.date)"
					>
						<template v-if="cell">
							<span class="kal-num">{{ cell.day }}</span>
							<span class="kal-dots">
								<i v-for="e in (byDay[cell.date] ?? []).slice(0, 3)" :key="e.eventKey" class="dot" :class="e.type" />
							</span>
						</template>
					</button>
				</div>
			</div>
		</div>

		<div class="card">
			<div class="card-head">
				<h2>{{ new Date(selectedDay + 'T12:00').toLocaleDateString('de-DE', { weekday: 'long', day: '2-digit', month: 'long' }) }}</h2>
			</div>
			<div class="card-body flush">
				<p v-if="!dayEvents.length" class="empty">Kein Termin an diesem Tag.</p>
				<div v-for="e in dayEvents" :key="e.eventKey" class="day-event">
					<div class="day-event-top">
						<span class="chip" :class="e.type">{{ typeLabels[e.type] }}</span>
						<strong>{{ e.title }}</strong>
						<span v-if="auth.can('termine')" style="margin-left: auto; display: flex; gap: 4px">
							<button class="btn sm" aria-label="Bearbeiten" @click="openEdit(e)"><Pencil :size="12" /></button>
							<button class="btn sm danger" aria-label="Löschen" @click="removeEvent(e)"><Trash2 :size="12" /></button>
						</span>
					</div>
					<div class="day-event-meta">
						<template v-if="e.startTime">{{ e.startTime }}<template v-if="e.endTime">–{{ e.endTime }}</template> Uhr</template>
						<template v-if="e.location"> · {{ e.location }}</template>
					</div>
					<p v-if="e.notes" class="day-event-notes">{{ e.notes }}</p>
					<div class="day-event-rsvp">
						<span class="rsvp-count"><Check :size="13" aria-hidden="true" /> {{ counts(e.eventKey).yes }}</span>
						<span class="rsvp-count declined"><X :size="13" aria-hidden="true" /> {{ counts(e.eventKey).no }}</span>
						<template v-if="auth.user?.playerId">
							<button
								class="btn sm"
								:class="{ gold: myStatus(e.eventKey) === 'attending' }"
								@click="rsvp(e.eventKey, 'attending')"
							>Zusagen</button>
							<button
								class="btn sm"
								:class="{ danger: myStatus(e.eventKey) === 'declined' }"
								@click="rsvp(e.eventKey, 'declined')"
							>Absagen</button>
						</template>
					</div>
				</div>
			</div>
		</div>
	</div>

	<AppModal v-if="showForm" :title="editId ? 'Termin bearbeiten' : 'Neuer Termin'" @close="showForm = false">
		<form @submit.prevent="submitForm">
			<p v-if="error" class="form-error" role="alert">{{ error }}</p>
			<div class="field">
				<label for="ev-title">Titel</label>
				<input id="ev-title" v-model="form.title" required maxlength="120" />
			</div>
			<div class="grid cols-2">
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
				<label for="ev-notes">Notizen</label>
				<textarea id="ev-notes" v-model="form.notes" rows="2" maxlength="2000" />
			</div>
			<div class="field" style="flex-direction: row; align-items: center; gap: 8px">
				<input id="ev-rec" v-model="form.recurring" type="checkbox" style="width: auto; min-height: auto" />
				<label for="ev-rec" style="margin: 0">Wiederholen</label>
			</div>
			<div v-if="form.recurring" class="grid cols-2">
				<div class="field">
					<label for="ev-rtype">Rhythmus</label>
					<select id="ev-rtype" v-model="form.recurrenceType">
						<option value="weekly">Wöchentlich</option>
						<option value="biweekly">Alle zwei Wochen</option>
					</select>
				</div>
				<div class="field">
					<label for="ev-rend">Ende der Serie</label>
					<input id="ev-rend" v-model="form.recurrenceEnd" type="date" />
				</div>
			</div>
			<button class="btn primary" style="width: 100%; justify-content: center">
				{{ editId ? 'Speichern' : 'Anlegen' }}
			</button>
		</form>
	</AppModal>
</template>

<style scoped>
.kal-grid { grid-template-columns: 3fr 2fr; }
@media (max-width: 900px) { .kal-grid { grid-template-columns: 1fr; } }

.kal-week { display: grid; grid-template-columns: repeat(7, 1fr); gap: 3px; margin-bottom: 3px; }
.kal-head-row span {
	font-family: var(--font-display);
	font-size: 11.5px;
	font-weight: 600;
	text-transform: uppercase;
	color: var(--kreide-45);
	text-align: center;
	padding: 2px 0 6px;
}
.kal-day {
	position: relative;
	aspect-ratio: 1.15;
	min-height: 44px;
	border: 1px solid var(--hair);
	border-radius: 4px;
	background: rgba(242, 240, 230, 0.02);
	display: flex;
	flex-direction: column;
	align-items: flex-start;
	padding: 4px 6px;
	transition: background var(--t-fast), border-color var(--t-fast);
}
.kal-day:not(.empty):hover { background: rgba(242, 240, 230, 0.07); }
.kal-day.empty { border-color: transparent; background: none; cursor: default; }
.kal-day.selected { border-color: var(--gold); background: var(--gold-bg); }
.kal-day.today .kal-num { color: var(--gold); font-weight: 600; }
.kal-num { font-family: var(--font-mono); font-size: 12.5px; }
.kal-dots { display: flex; gap: 3px; margin-top: auto; }
.dot { width: 6px; height: 6px; border-radius: 50%; background: var(--kreide-45); }
.dot.training { background: var(--gruen); }
.dot.spiel { background: var(--rot-hell); }
.dot.mannschaftsabend { background: var(--gold); }

.day-event { padding: 12px 14px; border-bottom: 1px solid var(--hair); }
.day-event:last-child { border-bottom: none; }
.day-event-top { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.day-event-meta { font-size: 12.5px; color: var(--kreide-70); margin-top: 4px; }
.day-event-notes { font-size: 12.5px; color: var(--kreide-45); margin-top: 4px; }
.day-event-rsvp { display: flex; align-items: center; gap: 8px; margin-top: 8px; flex-wrap: wrap; }
.rsvp-count { display: inline-flex; align-items: center; gap: 3px; font-family: var(--font-mono); font-size: 13px; color: var(--gruen); }
.rsvp-count.declined { color: var(--bad); }
</style>
