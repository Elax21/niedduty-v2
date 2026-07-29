<script setup lang="ts">
import { X } from 'lucide-vue-next';

defineProps<{ title: string }>();
const emit = defineEmits<{ close: [] }>();
</script>

<template>
	<div class="modal-backdrop" @click.self="emit('close')" @keydown.esc="emit('close')">
		<div class="modal" role="dialog" aria-modal="true" :aria-label="title">
			<div class="grabber" />
			<div class="card-head" style="border-bottom: 1px solid var(--line)">
				<h2>{{ title }}</h2>
				<button class="btn sm icon ghost" aria-label="Schließen" @click="emit('close')"><X :size="16" /></button>
			</div>
			<div class="card-body">
				<slot />
			</div>
		</div>
	</div>
</template>
