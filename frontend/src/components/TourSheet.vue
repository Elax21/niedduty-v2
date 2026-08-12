<script setup lang="ts">
// Nachschlagewerk aus dem Menü — derselbe Inhalt wie im geführten Rundgang
// (lib/help.ts), hier aber alles untereinander zum Blättern.
import { computed } from 'vue';
import {
	Home, Trophy, CalendarDays, Wallet, Users, Settings, Bell,
	BarChart3, Repeat, ShieldCheck, Vote, Check, Compass
} from 'lucide-vue-next';
import { useAuthStore } from '../stores/auth';
import { helpChapters } from '../lib/help';
import AppModal from './AppModal.vue';

const emit = defineEmits<{ close: []; restart: [] }>();

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
</script>

<template>
	<AppModal title="Hilfe" @close="emit('close')">
		<button type="button" class="btn gold block restart" @click="emit('restart')">
			<Compass :size="15" /> Rundgang neu starten
		</button>

		<section v-for="ch in chapters" :key="ch.key" class="entry">
			<h3><component :is="icons[ch.icon]" :size="17" /> {{ ch.title }}</h3>
			<p class="lead">{{ ch.lead }}</p>
			<ul>
				<li v-for="(pt, i) in ch.points" :key="i">
					<Check :size="14" /><span>{{ pt }}</span>
				</li>
			</ul>
		</section>
	</AppModal>
</template>

<style scoped>
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

ul { display: grid; gap: 9px; text-align: left; }
li {
	display: flex;
	gap: 9px;
	font-size: 13px;
	line-height: 1.45;
	color: var(--ink-2);
}
li :deep(svg) { color: var(--gold); flex-shrink: 0; margin-top: 2px; }

.restart { margin-bottom: 4px; justify-content: center; }
</style>
