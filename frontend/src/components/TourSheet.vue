<script setup lang="ts">
// Zwei Auftritte, ein Inhalt (lib/help.ts):
//   mode="tour"  — Rundgang nach der Registrierung, Schritt für Schritt.
//   mode="hilfe" — Nachschlagewerk aus dem Menü, alles untereinander.
import { ref, computed } from 'vue';
import {
	Home, Trophy, CalendarDays, Wallet, Users, Settings, Bell,
	BarChart3, Repeat, ShieldCheck, Vote, ChevronLeft, ChevronRight, Check
} from 'lucide-vue-next';
import { useAuthStore } from '../stores/auth';
import { helpChapters } from '../lib/help';
import AppModal from './AppModal.vue';

const props = defineProps<{ mode: 'tour' | 'hilfe' }>();
const emit = defineEmits<{ close: [] }>();

const auth = useAuthStore();

const icons: Record<string, unknown> = {
	Home, Trophy, CalendarDays, Wallet, Users, Settings, Bell, BarChart3, Repeat, ShieldCheck, Vote
};

/** Nur zeigen, was für dieses Konto überhaupt erreichbar ist. */
const chapters = computed(() =>
	helpChapters.filter((ch) => {
		if (!ch.perm) return true;
		if (ch.perm === 'admin') return auth.isAdmin;
		return auth.can(ch.perm);
	})
);

const step = ref(0);
const current = computed(() => chapters.value[step.value]);
const isLast = computed(() => step.value >= chapters.value.length - 1);

function next() {
	if (isLast.value) emit('close');
	else step.value++;
}
function back() {
	if (step.value > 0) step.value--;
}
</script>

<template>
	<AppModal :title="props.mode === 'tour' ? 'Willkommen in der Kabine' : 'Hilfe'" @close="emit('close')">
		<!-- Rundgang: ein Kapitel nach dem anderen -->
		<template v-if="props.mode === 'tour'">
			<div v-if="current" class="chapter">
				<div class="badge"><component :is="icons[current.icon]" :size="22" /></div>
				<h3>{{ current.title }}</h3>
				<p class="lead">{{ current.lead }}</p>
				<ul>
					<li v-for="(pt, i) in current.points" :key="i">
						<Check :size="14" /><span>{{ pt }}</span>
					</li>
				</ul>
			</div>

			<div class="dots" aria-hidden="true">
				<i v-for="(ch, i) in chapters" :key="ch.key" :class="{ on: i === step, done: i < step }" />
			</div>

			<div class="nav">
				<button type="button" class="btn" :disabled="step === 0" @click="back">
					<ChevronLeft :size="15" /> Zurück
				</button>
				<button type="button" class="btn gold grow" @click="next">
					{{ isLast ? 'Los geht’s' : 'Weiter' }}
					<ChevronRight v-if="!isLast" :size="15" />
				</button>
			</div>
			<button type="button" class="btn ghost block skip" @click="emit('close')">Überspringen</button>
		</template>

		<!-- Hilfe: alles auf einmal zum Nachschlagen -->
		<template v-else>
			<section v-for="ch in chapters" :key="ch.key" class="entry">
				<h3><component :is="icons[ch.icon]" :size="17" /> {{ ch.title }}</h3>
				<p class="lead">{{ ch.lead }}</p>
				<ul>
					<li v-for="(pt, i) in ch.points" :key="i">
						<Check :size="14" /><span>{{ pt }}</span>
					</li>
				</ul>
			</section>
		</template>
	</AppModal>
</template>

<style scoped>
.chapter { text-align: center; }
.badge {
	width: 54px;
	height: 54px;
	margin: 2px auto 12px;
	display: grid;
	place-items: center;
	border-radius: 16px;
	background: var(--surface-2);
	border: 1px solid var(--line-2);
	box-shadow: var(--shadow-gold);
}
.badge :deep(svg) { color: var(--gold); }

h3 {
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 8px;
	font-family: var(--font-display);
	font-size: 19px;
	font-weight: 700;
	text-transform: uppercase;
	letter-spacing: 0.02em;
	color: var(--ink);
}
.entry h3 { justify-content: flex-start; font-size: 16px; }
.entry h3 :deep(svg) { color: var(--gold); }
.entry { padding: 14px 0; border-bottom: 1px solid var(--line); }
.entry:last-of-type { border-bottom: none; }

.lead { margin: 7px 0 12px; font-size: 13.5px; line-height: 1.5; color: var(--ink-2); }
.chapter .lead { text-align: center; }

ul { display: grid; gap: 9px; text-align: left; }
li {
	display: flex;
	gap: 9px;
	font-size: 13px;
	line-height: 1.45;
	color: var(--ink-2);
}
li :deep(svg) { color: var(--gold); flex-shrink: 0; margin-top: 2px; }

.dots { display: flex; justify-content: center; gap: 6px; margin: 18px 0 14px; }
.dots i { width: 6px; height: 6px; border-radius: 50%; background: var(--surface-4); }
.dots i.done { background: var(--gold-deep); }
.dots i.on { background: var(--gold); width: 18px; border-radius: 3px; }

.nav { display: flex; gap: 9px; }
.nav .grow { flex: 1; }
.skip { margin-top: 9px; }
</style>
