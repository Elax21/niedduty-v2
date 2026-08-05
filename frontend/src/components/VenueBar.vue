<script setup lang="ts">
// Treffpunkt und Anfahrt zu einem Spiel. Die Adresse kommt von der
// fussball.de-Spielseite; der Treffpunkt liegt fest 1:30 vor Anpfiff.
import { computed } from 'vue';
import { MapPin, Navigation, Clock } from 'lucide-vue-next';
import { googleMapsUrl, appleMapsUrl, showAppleMaps, meetingTime, MEET_BEFORE_MATCH_MIN } from '../lib/maps';
import type { Venue } from '../types';

const props = defineProps<{ venue?: Venue | null; kickoff?: string }>();

const meet = computed(() => (props.kickoff ? meetingTime(props.kickoff) : ''));
const address = computed(() => props.venue?.address ?? '');
const apple = showAppleMaps();
</script>

<template>
	<div v-if="address || meet" class="venuebar">
		<div v-if="meet" class="line">
			<Clock :size="13" />
			<span>Treffpunkt <b class="mono">{{ meet }}</b> Uhr <span class="dim">({{ MEET_BEFORE_MATCH_MIN }} Min vor Anpfiff)</span></span>
		</div>
		<div v-if="venue?.name" class="line">
			<MapPin :size="13" />
			<span>{{ venue.name }}</span>
		</div>
		<div v-if="address" class="navi">
			<a class="btn sm" :href="googleMapsUrl(address)" target="_blank" rel="noopener">
				<Navigation :size="13" /> Google Maps
			</a>
			<a v-if="apple" class="btn sm" :href="appleMapsUrl(address)" target="_blank" rel="noopener">
				<Navigation :size="13" /> Apple Karten
			</a>
		</div>
	</div>
</template>

<style scoped>
.venuebar {
	display: grid;
	gap: 7px;
	padding: 11px 12px;
	background: var(--surface-flat);
	border: 1px solid var(--line);
	border-radius: var(--radius-sm);
}
.line { display: flex; align-items: center; gap: 8px; font-size: 12.5px; color: var(--ink-2); line-height: 1.35; }
.line :deep(svg) { color: var(--gold); flex-shrink: 0; }
.line b { color: var(--ink); }
.dim { color: var(--ink-3); }
/* auto-fit statt Flex-Wrap: nie schmaler als 132px, nie breiter als der
   Container — ein einzelner Knopf (Android ohne Apple Karten) füllt die Zeile. */
.navi { display: grid; grid-template-columns: repeat(auto-fit, minmax(132px, 1fr)); gap: 8px; }
.navi .btn { min-width: 0; min-height: 40px; white-space: nowrap; }
</style>
