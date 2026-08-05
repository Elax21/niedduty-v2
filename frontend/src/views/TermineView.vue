<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { api, apiError } from '../services/api';
import { useRefresh } from '../lib/refresh';
import { useAuthStore } from '../stores/auth';
import { enterRows } from '../lib/motion';
import { Plus, Pencil, Trash2, Check, X, MapPin, Clock, CalendarPlus, CalendarDays, ExternalLink, Repeat, StickyNote, ChevronLeft, ChevronRight } from 'lucide-vue-next';
import AppModal from '../components/AppModal.vue';
import VenueBar from '../components/VenueBar.vue';
import type { Attendance, Occurrence, Match, Matches, TrainingSchedule, Venue } from '../types';

const auth = useAuthStore();
const scope = ref<'kommend' | 'vergangen'>('kommend');
const events = ref<Occurrence[]>([]);
const matchList = ref<Match[]>([]);
/* Nur für fussball.de-Spiele nötig: eigene Termine bringen ihre Zähler schon mit. */
const matchAttendance = ref<Record<string, Attendance[]>>({});
const error = ref('');
const rsvpBusy = ref('');

const today = new Date().toISOString().slice(0, 10);
const canManage = computed(() => auth.can('termine'));
const myPlayerId = computed(() => auth.user?.playerId ?? null);

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
	occNote: string;    // Notiz nur für diesen einen Termin
	isMatch: boolean;
	side?: string;      // Heim | Auswärts
	result?: string;    // "3:1" bei gespielten Spielen
	url?: string;       // fussball.de-Spielseite
	venue?: Venue;      // Spielstätte (nur fussball.de-Spiele)
	occ?: Occurrence;   // Original (nur eigene Termine, für Bearbeiten)
}

function eventItem(e: Occurrence): Item {
	return { key: e.eventKey, isoDate: e.occDate, time: e.startTime, type: e.type, title: e.title, location: e.location, notes: e.notes, occNote: e.occNote, isMatch: false, occ: e };
}
function matchItem(m: Match): Item {
	const opp = m.home.isOwn ? m.guest.name : m.home.name;
	return {
		key: 'fdm_' + m.id, isoDate: m.isoDate, time: m.time, type: 'spiel',
		title: opp, location: '', notes: '', occNote: '', isMatch: true,
		side: m.home.isOwn ? 'Heim' : 'Auswärts',
		result: m.played ? `${m.homeGoals}:${m.guestGoals}` : undefined,
		url: m.url,
		venue: m.venue
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
	matchList.value = m
		? (scope.value === 'kommend'
			? m.next.filter((x) => x.isoDate && x.isoDate >= today)
			: m.previous.filter((x) => x.isoDate && x.isoDate < today))
		: [];
	matchAttendance.value = {};
	matchList.value.slice(0, 12).forEach((x) => loadMatchAttendance('fdm_' + x.id));
	requestAnimationFrame(() => enterRows('.ev-anim'));
}
watch(scope, load);

async function loadMatchAttendance(key: string) {
	const { data } = await api.get<Attendance[]>('/attendance', { params: { eventKey: key } });
	matchAttendance.value = { ...matchAttendance.value, [key]: data };
}

/** Zähler und eigener Status — je nach Herkunft aus dem Termin oder nachgeladen. */
function counts(it: Item) {
	if (it.occ) return { yes: it.occ.attending, no: it.occ.declined, open: it.occ.open };
	const list = matchAttendance.value[it.key] ?? [];
	return {
		yes: list.filter((a) => a.status === 'attending').length,
		no: list.filter((a) => a.status === 'declined').length,
		open: 0
	};
}
function myStatus(it: Item): string {
	if (!myPlayerId.value) return '';
	if (it.occ) return it.occ.myStatus;
	return matchAttendance.value[it.key]?.find((a) => a.playerId === myPlayerId.value)?.status ?? '';
}

async function rsvp(it: Item, status: 'attending' | 'declined') {
	if (!myPlayerId.value || rsvpBusy.value) return;
	rsvpBusy.value = it.key;
	try {
		await api.put('/attendance', { eventKey: it.key, playerId: myPlayerId.value, status });
		if (it.occ) await load();
		else await loadMatchAttendance(it.key);
	} catch (e) {
		error.value = apiError(e);
	} finally {
		rsvpBusy.value = '';
	}
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

function weekday(d: string) {
	return new Date(d + 'T12:00').toLocaleDateString('de-DE', { weekday: 'short' }).replace('.', '');
}
function dayNum(d: string) {
	return new Date(d + 'T12:00').toLocaleDateString('de-DE', { day: '2-digit' });
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

// ── Kalender: ganzer Monat auf einen Blick, beliebig weit vorspulen ──
const mode = ref<'liste' | 'kalender'>('liste');
const calMonth = ref(startOfMonth(new Date()));
const calEvents = ref<Occurrence[]>([]);
const calMatches = ref<Match[]>([]);
const calSelected = ref(today);

function startOfMonth(d: Date) { return new Date(d.getFullYear(), d.getMonth(), 1); }

const calLabel = computed(() =>
	calMonth.value.toLocaleDateString('de-DE', { month: 'long', year: 'numeric' })
);

function shiftMonth(delta: number) {
	const d = calMonth.value;
	calMonth.value = new Date(d.getFullYear(), d.getMonth() + delta, 1);
}
function jumpToday() {
	calMonth.value = startOfMonth(new Date());
	calSelected.value = today;
}

async function loadCalendar() {
	const first = calMonth.value;
	const last = new Date(first.getFullYear(), first.getMonth() + 1, 0);
	// Ränder mitladen, damit die angeschnittenen Wochen auch Punkte zeigen.
	const from = iso(new Date(first.getFullYear(), first.getMonth(), -7));
	const to = iso(new Date(last.getFullYear(), last.getMonth(), last.getDate() + 7));
	const [ev, mt] = await Promise.all([
		api.get<Occurrence[]>('/events', { params: { from, to } }),
		api.get<Matches>('/fussball/matches').catch(() => ({ data: null as Matches | null }))
	]);
	calEvents.value = ev.data;
	calMatches.value = mt.data ? [...mt.data.previous, ...mt.data.next] : [];
}
watch(calMonth, loadCalendar);
watch(mode, (m) => { if (m === 'kalender' && !calEvents.value.length) loadCalendar(); });

/** Alle Termine des Kalender-Zeitraums, nach Datum gebündelt. */
const calByDate = computed(() => {
	const map = new Map<string, Item[]>();
	const add = (it: Item) => {
		if (!it.isoDate) return;
		if (!map.has(it.isoDate)) map.set(it.isoDate, []);
		map.get(it.isoDate)!.push(it);
	};
	calEvents.value.map(eventItem).forEach(add);
	calMatches.value.map(matchItem).forEach(add);
	for (const list of map.values()) list.sort((a, b) => a.time.localeCompare(b.time));
	return map;
});

/** Sechs Wochenzeilen ab Montag — feste Höhe, kein Springen beim Blättern. */
const calWeeks = computed(() => {
	const first = calMonth.value;
	const start = new Date(first);
	start.setDate(1 - ((first.getDay() + 6) % 7)); // zurück auf Montag
	const weeks: { iso: string; day: number; inMonth: boolean; items: Item[] }[][] = [];
	const cur = new Date(start);
	for (let w = 0; w < 6; w++) {
		const row = [];
		for (let d = 0; d < 7; d++) {
			const key = iso(cur);
			row.push({
				iso: key,
				day: cur.getDate(),
				inMonth: cur.getMonth() === first.getMonth(),
				items: calByDate.value.get(key) ?? []
			});
			cur.setDate(cur.getDate() + 1);
		}
		weeks.push(row);
	}
	return weeks;
});

const calDayItems = computed(() => calByDate.value.get(calSelected.value) ?? []);
const calDayLabel = computed(() =>
	new Date(calSelected.value + 'T12:00').toLocaleDateString('de-DE', {
		weekday: 'long', day: '2-digit', month: 'long'
	})
);

// ── Trainingsplan (feste Wochentage) ────────────────────────────
const weekdayNames = [
	{ nr: 1, kurz: 'Mo' }, { nr: 2, kurz: 'Di' }, { nr: 3, kurz: 'Mi' }, { nr: 4, kurz: 'Do' },
	{ nr: 5, kurz: 'Fr' }, { nr: 6, kurz: 'Sa' }, { nr: 7, kurz: 'So' }
];
const showSchedule = ref(false);
const scheduleBusy = ref(false);
const scheduleError = ref('');
const schedule = ref<TrainingSchedule>({
	weekdays: [], title: 'Training', startTime: '19:15', endTime: '21:00',
	location: '', notes: '', recurrenceEnd: ''
});

async function openSchedule() {
	scheduleError.value = '';
	try {
		const { data } = await api.get<TrainingSchedule>('/training-schedule');
		schedule.value = {
			...data,
			title: data.title || 'Training',
			startTime: data.startTime || '19:15',
			endTime: data.endTime || '21:00'
		};
	} catch (e) {
		scheduleError.value = apiError(e);
	}
	showSchedule.value = true;
}

function toggleWeekday(nr: number) {
	const set = new Set(schedule.value.weekdays);
	set.has(nr) ? set.delete(nr) : set.add(nr);
	schedule.value.weekdays = [...set].sort((a, b) => a - b);
}

async function saveSchedule() {
	scheduleBusy.value = true;
	scheduleError.value = '';
	try {
		await api.put('/training-schedule', schedule.value);
		showSchedule.value = false;
		await load();
	} catch (e) {
		scheduleError.value = apiError(e);
	} finally {
		scheduleBusy.value = false;
	}
}

// ── Notiz an einer einzelnen Einheit ────────────────────────────
const noteFor = ref<Item | null>(null);
const noteText = ref('');
const noteBusy = ref(false);

function openNote(it: Item) {
	noteFor.value = it;
	noteText.value = it.occNote || '';
}
async function saveNote() {
	if (!noteFor.value) return;
	noteBusy.value = true;
	try {
		await api.put('/event-notes', { eventKey: noteFor.value.key, text: noteText.value.trim() });
		noteFor.value = null;
		await load();
	} catch (e) {
		error.value = apiError(e);
	} finally {
		noteBusy.value = false;
	}
}

onMounted(load);
useRefresh(load);
</script>

<template>
	<div class="page-head">
		<h1>Termine</h1>
		<a v-if="auth.club?.googleCalendarUrl" :href="auth.club.googleCalendarUrl" target="_blank" rel="noopener" class="sub-mono">
			Google ›
		</a>
	</div>

	<div class="segmented">
		<button :class="{ active: mode === 'liste' && scope === 'kommend' }" @click="mode = 'liste'; scope = 'kommend'">Kommend</button>
		<button :class="{ active: mode === 'liste' && scope === 'vergangen' }" @click="mode = 'liste'; scope = 'vergangen'">Vergangen</button>
		<button :class="{ active: mode === 'kalender' }" @click="mode = 'kalender'">Kalender</button>
	</div>

	<button v-if="canManage" class="btn block sched-btn" @click="openSchedule">
		<Repeat :size="16" /> Trainingszeiten
	</button>

	<!-- ── KALENDER ── -->
	<template v-if="mode === 'kalender'">
		<div class="calbar">
			<button class="btn sm icon ghost" aria-label="Voriger Monat" @click="shiftMonth(-1)"><ChevronLeft :size="16" /></button>
			<strong class="calmonth">{{ calLabel }}</strong>
			<button class="btn sm icon ghost" aria-label="Nächster Monat" @click="shiftMonth(1)"><ChevronRight :size="16" /></button>
			<button class="btn sm ghost heute-btn" @click="jumpToday">Heute</button>
		</div>

		<div class="calgrid card">
			<div v-for="w in ['Mo', 'Di', 'Mi', 'Do', 'Fr', 'Sa', 'So']" :key="w" class="calhead">{{ w }}</div>
			<template v-for="(week, wi) in calWeeks" :key="wi">
				<button
					v-for="cell in week"
					:key="cell.iso"
					type="button"
					class="calcell"
					:class="{ out: !cell.inMonth, today: cell.iso === today, sel: cell.iso === calSelected }"
					@click="calSelected = cell.iso"
				>
					<span class="n">{{ cell.day }}</span>
					<span class="dots">
						<i v-for="it in cell.items.slice(0, 3)" :key="it.key" :class="it.type" />
					</span>
				</button>
			</template>
		</div>

		<div class="band"><span class="lbl">{{ calDayLabel }}</span><span class="rule" /></div>
		<div v-if="calDayItems.length" class="stack">
			<article v-for="it in calDayItems" :key="it.key" class="card calrow">
				<span class="chip" :class="it.type">{{ typeLabels[it.type] }}</span>
				<div class="grow">
					<div class="t">
						<template v-if="it.isMatch">{{ it.side === 'Heim' ? 'vs' : 'bei' }} {{ it.title }}</template>
						<template v-else>{{ it.title }}</template>
					</div>
					<div class="m">
						<template v-if="it.time">{{ it.time }} Uhr</template>
						<template v-if="it.location"> · {{ it.location }}</template>
					</div>
					<p v-if="it.occNote" class="occnote"><StickyNote :size="12" /> {{ it.occNote }}</p>
				</div>
				<span v-if="it.result" class="mono res">{{ it.result }}</span>
			</article>
		</div>
		<div v-else class="card"><div class="empty">Nichts an diesem Tag.</div></div>
	</template>

	<template v-else>

	<p v-if="error && !showForm" class="form-error" role="alert">{{ error }}</p>

	<template v-if="groups.length">
		<div v-for="g in groups" :key="g.label">
			<div class="band">
				<span class="lbl">{{ g.label }}</span>
				<span class="rule" />
			</div>
			<article
				v-for="it in g.items"
				:key="it.key"
				class="evcard ev-anim"
				:class="[it.type, { heute: it.isoDate === today }]"
			>
				<div class="datebox">
					<div class="wd">{{ weekday(it.isoDate) }}</div>
					<div class="d">{{ dayNum(it.isoDate) }}</div>
				</div>
				<div class="grow">
					<div class="head-line">
						<span class="chip" :class="it.type">{{ typeLabels[it.type] }}</span>
						<span v-if="it.isoDate === today" class="chip" style="color: var(--gold); border-color: var(--line-2)">Heute</span>
						<span v-if="it.result" class="mono res">{{ it.result }}</span>
					</div>
					<div class="t">
						<template v-if="it.isMatch">{{ it.side === 'Heim' ? 'vs' : 'bei' }} {{ it.title }}</template>
						<template v-else>{{ it.title }}</template>
					</div>
					<div class="m">
						<template v-if="it.time"><Clock :size="12" style="vertical-align: -2px" /> {{ it.time }} Uhr</template>
						<template v-if="it.side"> · {{ it.side }}</template>
						<template v-if="it.location"> · <MapPin :size="12" style="vertical-align: -2px" /> {{ it.location }}</template>
					</div>
					<p v-if="it.notes" class="m">{{ it.notes }}</p>
					<p v-if="it.occNote" class="occnote"><StickyNote :size="12" /> {{ it.occNote }}</p>
					<VenueBar
						v-if="it.isMatch && !it.result && (it.venue || it.time)"
						:venue="it.venue"
						:kickoff="it.time"
						style="margin-top: 9px"
					/>

					<div class="actions">
						<template v-if="myPlayerId && it.isoDate >= today">
							<button class="rsvp yes sm" :class="{ on: myStatus(it) === 'attending' }" :disabled="rsvpBusy === it.key" @click="rsvp(it, 'attending')">
								<Check :size="14" /> Zusage
							</button>
							<button class="rsvp no sm" :class="{ on: myStatus(it) === 'declined' }" :disabled="rsvpBusy === it.key" @click="rsvp(it, 'declined')">
								<X :size="14" /> Absage
							</button>
						</template>
						<a v-if="it.isMatch && it.url" :href="it.url" target="_blank" rel="noopener" class="btn sm icon ghost" aria-label="Auf fussball.de öffnen"><ExternalLink :size="14" /></a>
						<a :href="googleUrl(it)" target="_blank" rel="noopener" class="btn sm icon ghost" aria-label="Zu Google Kalender"><CalendarPlus :size="14" /></a>
						<template v-if="canManage && !it.isMatch && it.occ">
							<button class="btn sm icon ghost" aria-label="Notiz zu diesem Termin" @click="openNote(it)"><StickyNote :size="13" /></button>
							<button class="btn sm icon ghost" aria-label="Bearbeiten" @click="openEdit(it.occ)"><Pencil :size="13" /></button>
							<button class="btn sm icon danger" aria-label="Löschen" @click="removeEvent(it.occ)"><Trash2 :size="13" /></button>
						</template>
						<span class="tally">
							<span class="count-yes">{{ counts(it).yes }}</span>
							<span class="count-open"> / </span>
							<span class="count-no">{{ counts(it).no }}</span>
						</span>
					</div>
				</div>
			</article>
		</div>
	</template>
	<div v-else class="card">
		<div class="empty"><CalendarDays :size="30" class="ic" /><br />{{ scope === 'kommend' ? 'Keine anstehenden Termine.' : 'Keine vergangenen Termine.' }}</div>
	</div>
	</template>

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

	<!-- Feste Trainingszeiten -->
	<AppModal v-if="showSchedule" title="Trainingszeiten" @close="showSchedule = false">
		<form @submit.prevent="saveSchedule">
			<p v-if="scheduleError" class="form-error" role="alert">{{ scheduleError }}</p>
			<p class="hint">
				Wochentage antippen — für jeden ausgewählten Tag läuft eine wöchentliche Serie.
				Abgewählte Tage werden samt Rückmeldungen entfernt.
			</p>
			<div class="wdays">
				<button
					v-for="w in weekdayNames"
					:key="w.nr"
					type="button"
					class="wday"
					:class="{ on: schedule.weekdays.includes(w.nr) }"
					:aria-pressed="schedule.weekdays.includes(w.nr)"
					@click="toggleWeekday(w.nr)"
				>{{ w.kurz }}</button>
			</div>
			<div class="row2">
				<div class="field">
					<label for="ts-start">Beginn</label>
					<input id="ts-start" v-model="schedule.startTime" type="time" />
				</div>
				<div class="field">
					<label for="ts-end">Ende</label>
					<input id="ts-end" v-model="schedule.endTime" type="time" />
				</div>
			</div>
			<div class="field">
				<label for="ts-title">Titel</label>
				<input id="ts-title" v-model="schedule.title" maxlength="120" placeholder="Training" />
			</div>
			<div class="field">
				<label for="ts-loc">Ort</label>
				<input id="ts-loc" v-model="schedule.location" maxlength="120" placeholder="Kunstrasen Sportpark Nord" />
			</div>
			<div class="field">
				<label for="ts-notes">Notiz für alle Einheiten</label>
				<textarea id="ts-notes" v-model="schedule.notes" rows="2" maxlength="2000" />
			</div>
			<div class="field">
				<label for="ts-end-date">Serie bis (optional)</label>
				<input id="ts-end-date" v-model="schedule.recurrenceEnd" type="date" />
			</div>
			<button class="btn primary block" style="margin-top: 6px" :disabled="scheduleBusy">Speichern</button>
		</form>
	</AppModal>

	<!-- Notiz an einer einzelnen Einheit -->
	<AppModal v-if="noteFor" title="Notiz zum Termin" @close="noteFor = null">
		<form @submit.prevent="saveNote">
			<p class="hint">Gilt nur für diesen einen Termin — der Rest der Serie bleibt unberührt.</p>
			<div class="field">
				<label for="occ-note">Notiz</label>
				<textarea id="occ-note" v-model="noteText" rows="3" maxlength="2000" placeholder="z. B. Torwarttraining, Schienbeinschoner mitbringen" />
			</div>
			<button class="btn primary block" style="margin-top: 6px" :disabled="noteBusy">Speichern</button>
		</form>
	</AppModal>
</template>

<style scoped>
.head-line { display: flex; align-items: center; gap: 7px; flex-wrap: wrap; margin-bottom: 6px; }
.res { margin-left: auto; font-size: 16px; font-weight: 700; color: var(--gold); }
.check-line { display: flex; align-items: center; gap: 9px; font-size: 15px; color: var(--ink); cursor: pointer; }
.check-line input { width: 20px; height: 20px; accent-color: var(--gold); }

.sched-btn { margin-bottom: 14px; gap: 8px; }
.sched-btn :deep(svg) { color: var(--gold); }
.hint { font-size: 12.5px; line-height: 1.45; color: var(--ink-3); margin-bottom: 12px; }

.wdays { display: grid; grid-template-columns: repeat(7, 1fr); gap: 6px; margin-bottom: 14px; }
.wday {
	min-height: 42px;
	font-family: var(--font-display);
	font-size: 14px;
	font-weight: 700;
	text-transform: uppercase;
	color: var(--ink-2);
	background: var(--surface-flat);
	border: 1px solid var(--line);
	border-radius: var(--radius-sm);
	transition: background var(--t-fast), color var(--t-fast);
}
.wday.on {
	color: var(--gold-ink);
	background: var(--gold);
	border-color: var(--gold);
	box-shadow: var(--shadow-gold);
}

.occnote {
	display: flex;
	align-items: center;
	gap: 6px;
	margin-top: 4px;
	font-size: 12.5px;
	line-height: 1.4;
	color: var(--gold-soft);
}
.occnote :deep(svg) { flex-shrink: 0; }

/* ── Kalender ── */
.calbar { display: flex; align-items: center; gap: 8px; margin-bottom: 10px; }
.calmonth {
	font-family: var(--font-display);
	font-size: 17px;
	text-transform: uppercase;
	letter-spacing: 0.02em;
	color: var(--ink);
	min-width: 132px;
	text-align: center;
}
.heute-btn { margin-left: auto; }

.calgrid {
	display: grid;
	grid-template-columns: repeat(7, 1fr);
	gap: 2px;
	padding: 8px;
	margin-bottom: 14px;
}
.calhead {
	font-family: var(--font-display);
	font-size: 11px;
	text-transform: uppercase;
	color: var(--ink-3);
	text-align: center;
	padding-bottom: 4px;
}
.calcell {
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	gap: 3px;
	min-height: 42px;
	border-radius: 9px;
	border: 1px solid transparent;
	transition: background var(--t-fast);
}
.calcell .n { font-family: var(--font-mono); font-size: 13px; color: var(--ink); font-variant-numeric: tabular-nums; }
.calcell.out .n { color: var(--ink-3); }
.calcell.today { border-color: var(--line-2); }
.calcell.today .n { color: var(--gold); font-weight: 700; }
.calcell.sel { background: var(--surface-3); }
.calcell:active { transform: scale(0.94); }
.dots { display: flex; gap: 3px; height: 5px; }
.dots i { width: 5px; height: 5px; border-radius: 50%; background: var(--ink-3); }
.dots i.training { background: var(--gruen); }
.dots i.spiel { background: var(--rot); }
.dots i.mannschaftsabend { background: var(--gold); }

.calrow { display: flex; align-items: center; gap: 11px; padding: 12px 13px; }
.calrow .t { font-family: var(--font-display); font-size: 15px; font-weight: 600; color: var(--ink); }
.calrow .m { font-size: 12.5px; color: var(--ink-3); }
</style>
