<script setup lang="ts">
// Kurze Notiz beim Öffnen der App, wenn es Neuerungen gibt. Inhalt und
// Version stehen in lib/changelog.ts; gelesen wird am Konto gemerkt.
import { onMounted } from 'vue';
import { Sparkles } from 'lucide-vue-next';
import AppModal from './AppModal.vue';
import { changelogTitle, changelogLead, changelogPoints } from '../lib/changelog';
import { enterRows } from '../lib/motion';

const emit = defineEmits<{ close: [] }>();

onMounted(() => requestAnimationFrame(() => enterRows('.cn-row')));
</script>

<template>
	<AppModal :title="changelogTitle" @close="emit('close')">
		<p class="cn-lead">{{ changelogLead }}</p>
		<div class="cn-list">
			<div v-for="p in changelogPoints" :key="p.title" class="cn-row">
				<span class="cn-icon"><Sparkles :size="15" /></span>
				<span>
					<span class="t">{{ p.title }}</span>
					<span class="s">{{ p.text }}</span>
				</span>
			</div>
		</div>
		<button class="btn primary block" style="margin-top: 16px" @click="emit('close')">Alles klar</button>
	</AppModal>
</template>

<style scoped>
.cn-lead { font-size: 13.5px; color: var(--ink-2); margin-bottom: 12px; }
.cn-list { display: flex; flex-direction: column; gap: 10px; }
.cn-row {
	display: flex;
	align-items: flex-start;
	gap: 11px;
	padding: 12px 13px;
	border: 1px solid var(--line-2);
	border-radius: 13px;
	background: var(--surface-flat);
}
.cn-icon {
	width: 30px; height: 30px;
	flex-shrink: 0;
	display: grid; place-items: center;
	border-radius: 9px;
	background: var(--gold-bg);
	color: var(--gold);
}
.cn-row .t { display: block; font-family: var(--font-display); font-size: 14.5px; font-weight: 600; color: var(--ink); }
.cn-row .s { display: block; font-size: 12.5px; line-height: 1.45; color: var(--ink-3); margin-top: 3px; }
</style>
