<script setup lang="ts">
// Jeder stellt selbst ein, wie früh er Bescheid bekommt. Vorlaufzeiten in
// Minuten, 0 = diese Erinnerung aus.
import { ref, onMounted } from 'vue';
import { api, apiError } from '../services/api';
import { pushSupported } from '../lib/push';
import AppModal from './AppModal.vue';
import type { PushSettings } from '../types';

const emit = defineEmits<{ close: [] }>();

const s = ref<PushSettings>({
	trainingLeadMin: 60,
	matchLeadMin: 180,
	meetLeadMin: 30,
	vorschauSpiel: 1440,
	vorschauTraining: 300,
	birthdays: true
});
const meetBefore = ref(90);
const busy = ref(false);
const error = ref('');
const saved = ref(false);
// Auf dem iPhone gibt es Push erst, wenn die App auf dem Startbildschirm liegt.
// Die Zeiten lassen sich trotzdem schon setzen — sie hängen am Konto.
const canPush = pushSupported();

/** Auswahl in Minuten — bewusst grob, damit es auf dem Handy schnell geht. */
const kurzOptionen = [
	{ v: 0, l: 'aus' },
	{ v: 30, l: '30 Min' },
	{ v: 60, l: '1 Std' },
	{ v: 120, l: '2 Std' },
	{ v: 180, l: '3 Std' }
];
const vorschauOptionen = [
	{ v: 0, l: 'aus' },
	{ v: 180, l: '3 Std' },
	{ v: 300, l: '5 Std' },
	{ v: 720, l: '12 Std' },
	{ v: 1440, l: '1 Tag' },
	{ v: 2880, l: '2 Tage' }
];
const treffOptionen = [
	{ v: 0, l: 'aus' },
	{ v: 15, l: '15 Min' },
	{ v: 30, l: '30 Min' },
	{ v: 60, l: '1 Std' }
];

async function load() {
	try {
		const { data } = await api.get<{ settings: PushSettings; meetBeforeMatchMin: number }>('/push/settings');
		s.value = data.settings;
		meetBefore.value = data.meetBeforeMatchMin;
	} catch (e) {
		error.value = apiError(e);
	}
}

async function save() {
	busy.value = true;
	error.value = '';
	try {
		await api.put('/push/settings', s.value);
		saved.value = true;
		setTimeout(() => emit('close'), 600);
	} catch (e) {
		error.value = apiError(e);
	} finally {
		busy.value = false;
	}
}

onMounted(load);
</script>

<template>
	<AppModal title="Benachrichtigungen" @close="emit('close')">
		<form @submit.prevent="save">
			<p v-if="error" class="form-error" role="alert">{{ error }}</p>
			<p v-if="!canPush" class="hint" style="margin-bottom: 14px">
				Dieses Gerät kann noch keine Benachrichtigungen zustellen. Auf dem iPhone geht das
				erst, wenn die App über „Teilen → Zum Home-Bildschirm" installiert ist. Die Zeiten
				kannst du trotzdem schon festlegen — sie gelten für dein Konto, nicht für das Gerät.
			</p>

			<div class="field">
				<label for="ps-training">Vor dem Training erinnern</label>
				<select id="ps-training" v-model.number="s.trainingLeadMin">
					<option v-for="o in kurzOptionen" :key="o.v" :value="o.v">{{ o.l }}</option>
				</select>
			</div>

			<div class="field">
				<label for="ps-spiel">Vor dem Spiel erinnern</label>
				<select id="ps-spiel" v-model.number="s.matchLeadMin">
					<option v-for="o in kurzOptionen" :key="o.v" :value="o.v">{{ o.l }}</option>
				</select>
			</div>

			<div class="field">
				<label for="ps-treff">Vor dem Treffpunkt erinnern</label>
				<select id="ps-treff" v-model.number="s.meetLeadMin">
					<option v-for="o in treffOptionen" :key="o.v" :value="o.v">{{ o.l }}</option>
				</select>
				<p class="hint">Treffpunkt ist immer {{ meetBefore }} Minuten vor Anpfiff.</p>
			</div>

			<div class="chalk-divider" />

			<div class="field">
				<label for="ps-vs">Rückmeldung erbitten — Spiel</label>
				<select id="ps-vs" v-model.number="s.vorschauSpiel">
					<option v-for="o in vorschauOptionen" :key="o.v" :value="o.v">{{ o.l }}</option>
				</select>
			</div>

			<div class="field">
				<label for="ps-vt">Rückmeldung erbitten — Training</label>
				<select id="ps-vt" v-model.number="s.vorschauTraining">
					<option v-for="o in vorschauOptionen" :key="o.v" :value="o.v">{{ o.l }}</option>
				</select>
				<p class="hint">Kommt nur, solange du weder zu- noch abgesagt hast.</p>
			</div>

			<label class="check-line">
				<input v-model="s.birthdays" type="checkbox" />
				<span>Geburtstage der Mannschaft</span>
			</label>

			<button class="btn primary block" style="margin-top: 14px" :disabled="busy">
				{{ saved ? 'Gespeichert' : 'Speichern' }}
			</button>
		</form>
	</AppModal>
</template>

<style scoped>
.hint { font-size: 12px; color: var(--ink-3); margin-top: 5px; line-height: 1.4; }
.check-line { display: flex; align-items: center; gap: 9px; font-size: 15px; color: var(--ink); cursor: pointer; }
.check-line input { width: 20px; height: 20px; accent-color: var(--gold); }
</style>
