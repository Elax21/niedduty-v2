<script setup lang="ts">
import { ExternalLink } from 'lucide-vue-next';
import type { Match } from '../types';

defineProps<{ match: Match }>();
</script>

<template>
	<component :is="match.url ? 'a' : 'div'" class="mc" :class="{ link: match.url }" :href="match.url || undefined" :target="match.url ? '_blank' : undefined" rel="noopener">
		<div class="mc-when">
			<span>{{ match.date }}</span>
			<span v-if="match.time"> · {{ match.time }} Uhr</span>
			<ExternalLink v-if="match.url" :size="11" class="mc-ext" />
		</div>
		<div class="mc-body">
			<div class="mc-team" :class="{ own: match.home.isOwn }">
				<img v-if="match.home.logoUrl" :src="match.home.logoUrl" alt="" class="mc-logo" loading="lazy" />
				<span class="mc-name">{{ match.home.name }}</span>
			</div>
			<div class="mc-score">
				<template v-if="match.played">
					<span class="mono">{{ match.homeGoals }}</span><span class="sep">:</span><span class="mono">{{ match.guestGoals }}</span>
				</template>
				<span v-else class="mc-vs">vs</span>
			</div>
			<div class="mc-team guest" :class="{ own: match.guest.isOwn }">
				<span class="mc-name">{{ match.guest.name }}</span>
				<img v-if="match.guest.logoUrl" :src="match.guest.logoUrl" alt="" class="mc-logo" loading="lazy" />
			</div>
		</div>
	</component>
</template>

<style scoped>
.mc { display: block; padding: 12px 14px; border-bottom: 1px solid var(--line); color: inherit; transition: background var(--t-fast); }
.mc:last-child { border-bottom: none; }
.mc.link { cursor: pointer; }
.mc.link:hover { background: var(--gold-bg); }
.mc-ext { color: var(--ink-3); vertical-align: -1px; margin-left: 4px; }
.mc-when {
	font-size: 11.5px;
	color: var(--ink-3);
	font-family: var(--font-display);
	text-transform: uppercase;
	letter-spacing: 0.04em;
	text-align: center;
	margin-bottom: 8px;
}
.mc-body { display: flex; align-items: center; gap: 8px; }
.mc-team { flex: 1; min-width: 0; display: flex; align-items: center; gap: 8px; }
.mc-team.guest { justify-content: flex-end; }
.mc-logo { width: 26px; height: 26px; object-fit: contain; flex-shrink: 0; }
.mc-name {
	font-size: 13.5px;
	font-weight: 600;
	white-space: nowrap;
	overflow: hidden;
	text-overflow: ellipsis;
}
.mc-team.guest .mc-name { text-align: right; }
.mc-team.own .mc-name { color: var(--gold); }
.mc-score {
	flex-shrink: 0;
	min-width: 58px;
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 2px;
	font-size: 20px;
	font-weight: 600;
	color: var(--gold);
}
.mc-score .sep { color: var(--ink-3); }
.mc-vs {
	font-family: var(--font-display);
	font-size: 12px;
	color: var(--ink-3);
	text-transform: uppercase;
	letter-spacing: 0.06em;
	background: var(--surface-3);
	border: 1px solid var(--line);
	border-radius: 999px;
	padding: 3px 10px;
}
</style>
