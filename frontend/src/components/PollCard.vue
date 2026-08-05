<script setup lang="ts">
// Eine Abstimmung. Vor der eigenen Stimme bleiben die Zahlen verborgen, damit
// die ersten Stimmen den Rest nicht in eine Richtung ziehen.
import { ref, computed } from 'vue';
import { api, apiError } from '../services/api';
import { Check, Clock, Users, Lock, Trash2 } from 'lucide-vue-next';
import type { Poll } from '../types';

const props = defineProps<{ poll: Poll; canManage?: boolean }>();
const emit = defineEmits<{ changed: [Poll] }>();

const busy = ref(false);
const error = ref('');
const showVoters = ref(false);

const voted = computed(() => props.poll.myVotes.length > 0);
/** Ergebnis erst nach eigener Stimme — oder wenn die Abstimmung vorbei ist. */
const showResult = computed(() => voted.value || !props.poll.running);

const maxCount = computed(() => Math.max(...props.poll.counts, 1));

const deadline = computed(() => {
	if (!props.poll.endsAt) return '';
	return new Date(props.poll.endsAt).toLocaleString('de-DE', {
		weekday: 'short', day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit'
	});
});

function pct(n: number) {
	return props.poll.total ? Math.round((n / props.poll.total) * 100) : 0;
}

async function vote(idx: number) {
	if (!props.poll.running || busy.value) return;
	busy.value = true;
	error.value = '';

	// Einfachauswahl ersetzt, Mehrfachauswahl schaltet um.
	let next: number[];
	if (props.poll.multiChoice) {
		next = props.poll.myVotes.includes(idx)
			? props.poll.myVotes.filter((i) => i !== idx)
			: [...props.poll.myVotes, idx];
	} else {
		next = props.poll.myVotes.includes(idx) ? [] : [idx];
	}

	try {
		const { data } = await api.post<Poll>(`/polls/${props.poll.id}/vote`, { options: next });
		emit('changed', data);
	} catch (e) {
		error.value = apiError(e);
	} finally {
		busy.value = false;
	}
}

async function close() {
	if (!window.confirm('Abstimmung jetzt beenden? Das Ergebnis bleibt sichtbar.')) return;
	const { data } = await api.post<Poll>(`/polls/${props.poll.id}/close`, {});
	emit('changed', data);
}

async function remove() {
	if (!window.confirm(`„${props.poll.question}" samt aller Stimmen löschen?`)) return;
	await api.delete(`/polls/${props.poll.id}`);
	emit('changed', { ...props.poll, id: '' }); // leere ID = entfernt
}
</script>

<template>
	<section class="card poll" :class="{ beendet: !poll.running }">
		<div class="card-body">
			<div class="head">
				<span class="chip" :class="poll.running ? 'training' : ''">
					{{ poll.running ? 'Abstimmung' : 'Beendet' }}
				</span>
				<span v-if="poll.multiChoice" class="chip">Mehrfachauswahl</span>
			</div>

			<h2 class="q">{{ poll.question }}</h2>

			<p class="meta">
				<Clock v-if="deadline" :size="12" />
				<span v-if="deadline">{{ poll.running ? 'bis' : 'endete' }} {{ deadline }} Uhr</span>
				<span v-else>ohne festes Ende</span>
				<span class="dot">·</span>
				<Users :size="12" /> <span class="mono">{{ poll.total }}</span> abgestimmt
			</p>

			<p v-if="error" class="form-error" role="alert">{{ error }}</p>

			<div class="opts">
				<button
					v-for="(opt, i) in poll.options"
					:key="i"
					type="button"
					class="opt"
					:class="{ on: poll.myVotes.includes(i), aus: !poll.running }"
					:disabled="busy || !poll.running"
					@click="vote(i)"
				>
					<span v-if="showResult" class="fill" :style="{ width: pct(poll.counts[i]) + '%' }" />
					<span class="mark"><Check v-if="poll.myVotes.includes(i)" :size="13" /></span>
					<span class="label">{{ opt }}</span>
					<span v-if="showResult" class="num mono">{{ poll.counts[i] }}</span>
				</button>
			</div>

			<p v-if="!showResult" class="hint">Stimme abgeben, um das Ergebnis zu sehen.</p>

			<template v-if="showResult && poll.total">
				<button type="button" class="btn sm ghost block" @click="showVoters = !showVoters">
					{{ showVoters ? 'Namen ausblenden' : 'Wer hat wie gestimmt?' }}
				</button>
				<div v-if="showVoters" class="voters">
					<div v-for="(opt, i) in poll.options" :key="i" class="vgroup">
						<div class="vopt">{{ opt }}</div>
						<div class="vnames">{{ poll.voters[i].length ? poll.voters[i].join(', ') : '—' }}</div>
					</div>
				</div>
			</template>

			<div v-if="canManage" class="admin">
				<button v-if="poll.running" class="btn sm" @click="close"><Lock :size="13" /> Beenden</button>
				<button class="btn sm danger" @click="remove"><Trash2 :size="13" /> Löschen</button>
			</div>
		</div>
	</section>
</template>

<style scoped>
.poll.beendet { opacity: 0.85; }
.head { display: flex; gap: 7px; flex-wrap: wrap; margin-bottom: 9px; }
.q {
	font-family: var(--font-display);
	font-size: 19px;
	font-weight: 700;
	line-height: 1.2;
	color: var(--ink);
}
.meta {
	display: flex;
	align-items: center;
	gap: 5px;
	flex-wrap: wrap;
	font-size: 12px;
	color: var(--ink-3);
	margin: 7px 0 13px;
}
.meta :deep(svg) { color: var(--gold); }
.meta .dot { padding: 0 3px; }
.meta .mono { font-family: var(--font-mono); color: var(--ink-2); }

.opts { display: grid; gap: 8px; }
.opt {
	position: relative;
	display: flex;
	align-items: center;
	gap: 10px;
	width: 100%;
	min-height: 46px;
	padding: 10px 13px;
	overflow: hidden;
	text-align: left;
	background: var(--surface-flat);
	border: 1px solid var(--line);
	border-radius: var(--radius-sm);
	transition: border-color var(--t-fast), background var(--t-fast);
}
.opt:active { transform: scale(0.99); }
.opt.on { border-color: var(--gold); }
.opt.aus { cursor: default; }
.fill {
	position: absolute;
	inset: 0 auto 0 0;
	background: var(--gold-bg);
	transition: width var(--t-med) var(--ease-out);
}
.mark {
	position: relative;
	flex-shrink: 0;
	width: 21px;
	height: 21px;
	display: grid;
	place-items: center;
	border: 1px solid var(--line-2);
	border-radius: 6px;
}
.opt.on .mark { background: var(--gold); border-color: var(--gold); }
.opt.on .mark :deep(svg) { color: var(--gold-ink); }
.label { position: relative; flex: 1; min-width: 0; font-size: 14px; color: var(--ink); }
.num { position: relative; font-size: 14px; font-weight: 700; color: var(--gold); }

.hint { font-size: 12px; color: var(--ink-3); margin-top: 10px; }
.voters { margin-top: 10px; display: grid; gap: 9px; }
.vgroup { font-size: 12.5px; }
.vopt { color: var(--gold); font-weight: 600; }
.vnames { color: var(--ink-3); line-height: 1.4; }
.admin { display: flex; gap: 8px; margin-top: 14px; }
</style>
