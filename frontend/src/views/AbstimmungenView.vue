<script setup lang="ts">
/* Abstimmungen für die Mannschaft. Starten darf, wer das Recht „umfragen" hat
   (Mannschaftsrat) — abstimmen darf jeder. */
import { ref, computed, onMounted } from 'vue';
import { api, apiError } from '../services/api';
import { useRefresh } from '../lib/refresh';
import { useAuthStore } from '../stores/auth';
import { enterRows } from '../lib/motion';
import { Plus, Vote, X } from 'lucide-vue-next';
import AppModal from '../components/AppModal.vue';
import PollCard from '../components/PollCard.vue';
import type { Poll } from '../types';

const auth = useAuthStore();
const polls = ref<Poll[]>([]);
const loading = ref(true);

const canCreate = computed(() => auth.can('umfragen'));
const running = computed(() => polls.value.filter((p) => p.running));
const past = computed(() => polls.value.filter((p) => !p.running));

async function load() {
	loading.value = true;
	try {
		const { data } = await api.get<Poll[]>('/polls');
		polls.value = data;
	} finally {
		loading.value = false;
	}
	requestAnimationFrame(() => enterRows('.poll'));
}

/** Karte meldet ihr neues Ergebnis; leere ID heißt gelöscht. */
function onChanged(p: Poll) {
	if (!p.id) {
		load();
		return;
	}
	polls.value = polls.value.map((x) => (x.id === p.id ? p : x));
}

// ── Neue Abstimmung ──
const showForm = ref(false);
const error = ref('');
const busy = ref(false);
const form = ref({ question: '', options: ['', ''], multiChoice: false, endsAt: '' });

function openCreate() {
	form.value = { question: '', options: ['', ''], multiChoice: false, endsAt: '' };
	error.value = '';
	showForm.value = true;
}
function addOption() {
	if (form.value.options.length < 10) form.value.options.push('');
}
function removeOption(i: number) {
	if (form.value.options.length > 2) form.value.options.splice(i, 1);
}

async function submit() {
	error.value = '';
	const options = form.value.options.map((o) => o.trim()).filter(Boolean);
	if (options.length < 2) {
		error.value = 'Mindestens zwei Antworten angeben';
		return;
	}
	busy.value = true;
	try {
		await api.post('/polls', {
			question: form.value.question.trim(),
			options,
			multiChoice: form.value.multiChoice,
			endsAt: form.value.endsAt
		});
		showForm.value = false;
		await load();
	} catch (e) {
		error.value = apiError(e);
	} finally {
		busy.value = false;
	}
}

onMounted(load);
useRefresh(load);
</script>

<template>
	<div class="page-head">
		<h1>Abstimmungen</h1>
		<span v-if="running.length" class="sub-mono">{{ running.length }} offen</span>
	</div>

	<p v-if="loading" class="card"><span class="empty">Wird geladen …</span></p>

	<template v-else>
		<PollCard
			v-for="p in running"
			:key="p.id"
			:poll="p"
			:can-manage="canCreate"
			style="margin-bottom: 14px"
			@changed="onChanged"
		/>

		<template v-if="past.length">
			<div class="band"><span class="lbl">Beendet</span><span class="rule" /></div>
			<PollCard
				v-for="p in past"
				:key="p.id"
				:poll="p"
				:can-manage="canCreate"
				style="margin-bottom: 14px"
				@changed="onChanged"
			/>
		</template>

		<div v-if="!polls.length" class="card">
			<div class="empty">
				<Vote :size="30" class="ic" /><br />
				<template v-if="canCreate">Noch keine Abstimmung. Leg die erste an.</template>
				<template v-else>Noch keine Abstimmung.</template>
			</div>
		</div>
	</template>

	<button v-if="canCreate" class="fab" aria-label="Abstimmung starten" @click="openCreate">
		<Plus :size="24" />
	</button>

	<AppModal v-if="showForm" title="Abstimmung starten" @close="showForm = false">
		<form @submit.prevent="submit">
			<p v-if="error" class="form-error" role="alert">{{ error }}</p>

			<div class="field">
				<label for="pl-q">Frage</label>
				<input id="pl-q" v-model="form.question" required maxlength="200" placeholder="z. B. Wann machen wir den Mannschaftsabend?" />
			</div>

			<div class="field">
				<label>Antworten</label>
				<div v-for="(_, i) in form.options" :key="i" class="optrow">
					<input v-model="form.options[i]" maxlength="120" :placeholder="`Antwort ${i + 1}`" />
					<button
						type="button"
						class="btn sm icon ghost"
						:disabled="form.options.length <= 2"
						aria-label="Antwort entfernen"
						@click="removeOption(i)"
					>
						<X :size="14" />
					</button>
				</div>
				<button type="button" class="btn sm ghost block" :disabled="form.options.length >= 10" @click="addOption">
					<Plus :size="13" /> Antwort hinzufügen
				</button>
			</div>

			<label class="check-line">
				<input v-model="form.multiChoice" type="checkbox" />
				<span>Mehrere Antworten erlaubt</span>
			</label>

			<div class="field" style="margin-top: 12px">
				<label for="pl-end">Ende (optional)</label>
				<input id="pl-end" v-model="form.endsAt" type="datetime-local" />
				<p class="hint">
					Alle bekommen sofort eine Benachrichtigung. Wer 24 Stunden vor Ablauf noch nicht
					abgestimmt hat, wird noch einmal erinnert.
				</p>
			</div>

			<button class="btn primary block" style="margin-top: 6px" :disabled="busy">Starten</button>
		</form>
	</AppModal>
</template>

<style scoped>
.optrow { display: flex; gap: 8px; align-items: center; margin-bottom: 8px; }
.optrow input { flex: 1; min-width: 0; }
.check-line { display: flex; align-items: center; gap: 9px; font-size: 15px; color: var(--ink); cursor: pointer; }
.check-line input { width: 20px; height: 20px; accent-color: var(--gold); }
.hint { font-size: 12px; color: var(--ink-3); margin-top: 6px; line-height: 1.4; }
</style>
