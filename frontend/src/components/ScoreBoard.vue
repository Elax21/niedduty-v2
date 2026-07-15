<script setup lang="ts">
import { ref, watch, onMounted } from 'vue';
import { countUp } from '../lib/motion';

// Anzeigetafel-Kachel: zählt Zahlen hoch, Text bleibt statisch.
const props = defineProps<{
	value: number | string;
	label: string;
	suffix?: string;
}>();

const shown = ref<string>('0');

function run() {
	if (typeof props.value === 'number') {
		countUp(props.value, (v) => (shown.value = String(v)));
	} else {
		shown.value = props.value;
	}
}

onMounted(run);
watch(() => props.value, run);
</script>

<template>
	<div>
		<div class="board" role="img" :aria-label="`${label}: ${value}${suffix ?? ''}`">
			<span v-for="(ch, i) in shown" :key="i" class="tile" :class="{ sep: !/[0-9]/.test(ch) }" aria-hidden="true">{{ ch === ' ' ? '' : ch }}</span>
			<span v-if="suffix" class="tile sep" aria-hidden="true">{{ suffix }}</span>
		</div>
		<div class="board-label">{{ label }}</div>
	</div>
</template>
