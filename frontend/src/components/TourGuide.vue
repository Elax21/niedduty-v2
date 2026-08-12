<script setup lang="ts">
// Geführter Rundgang: dunkelt die App ab, lässt genau ein Bedienelement frei
// und erklärt es daneben. Antippen bringt einen weiter — so lernt man die
// Wege, statt eine Bildergeschichte zu lesen.
//
// Ankerpunkte sind `data-tour`-Attribute in App.vue. Fehlt eines (Element für
// dieses Konto nicht vorhanden), wird der Schritt übersprungen.
import { ref, computed, onMounted, onBeforeUnmount, watch, nextTick } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { ChevronLeft, ChevronRight, Check, Hand } from 'lucide-vue-next';
import { useAuthStore } from '../stores/auth';
import { tourSteps, stepContent, type TourStep } from '../lib/tour';

const props = defineProps<{ menuOpen: boolean }>();
const emit = defineEmits<{ 'update:menuOpen': [boolean]; close: [] }>();

const auth = useAuthStore();
const router = useRouter();
const route = useRoute();

/** Nur Schritte, die dieses Konto überhaupt sehen kann. */
const steps = computed(() =>
	tourSteps.filter((s) => {
		if (!s.perm) return true;
		if (s.perm === 'admin') return auth.isAdmin;
		return auth.can(s.perm);
	})
);

const idx = ref(0);
const step = computed<TourStep>(() => steps.value[idx.value]);
const content = computed(() => stepContent(step.value));
const isLast = computed(() => idx.value >= steps.value.length - 1);

// ── Position des hervorgehobenen Elements ───────────────────────
const box = ref<{ top: number; left: number; width: number; height: number } | null>(null);
const PAD = 6; // Luft um das Element, damit der Rahmen nicht klebt

function measure() {
	const sel = step.value?.target;
	if (!sel) {
		box.value = null;
		return;
	}
	const el = document.querySelector(sel);
	if (!el) {
		box.value = null;
		return;
	}
	const r = el.getBoundingClientRect();
	if (r.width === 0 && r.height === 0) {
		box.value = null;
		return;
	}
	box.value = {
		top: Math.max(0, r.top - PAD),
		left: Math.max(0, r.left - PAD),
		width: r.width + PAD * 2,
		height: r.height + PAD * 2
	};
}

/** Nach Navigation und Sheet-Animation sitzt das Element erst später richtig. */
async function applyStep() {
	const s = step.value;
	if (!s) return;
	if (s.menu !== undefined && s.menu !== props.menuOpen) emit('update:menuOpen', s.menu);
	if (s.route && route.path !== s.route) await router.push(s.route);
	await nextTick();
	measure();
	setTimeout(measure, 120);
	setTimeout(measure, 380);
}

let poll: number | undefined;
onMounted(() => {
	applyStep();
	// Layout wandert (Sheet öffnet, Bilder laden, Tastatur) — billig nachmessen.
	poll = window.setInterval(measure, 300);
	window.addEventListener('resize', measure);
	window.addEventListener('scroll', measure, true);
	document.addEventListener('click', onDocClick, true);
	document.body.classList.add('tour-active');
});
onBeforeUnmount(() => {
	if (poll) clearInterval(poll);
	window.removeEventListener('resize', measure);
	window.removeEventListener('scroll', measure, true);
	document.removeEventListener('click', onDocClick, true);
	document.body.classList.remove('tour-active');
});

watch(idx, applyStep);
watch(() => props.menuOpen, measure);

// ── Vor / zurück ────────────────────────────────────────────────
function next() {
	if (isLast.value) finish();
	else idx.value++;
}
function back() {
	if (idx.value > 0) idx.value--;
}
function finish() {
	emit('update:menuOpen', false);
	emit('close');
}

/** Tippt der Nutzer selbst auf das markierte Element, geht es weiter. */
function onDocClick(ev: MouseEvent) {
	const s = step.value;
	if (!s?.tap || !s.target) return;
	const t = ev.target as Element | null;
	if (t && t.closest(s.target)) {
		// Erst die Navigation der App laufen lassen, dann weiterrücken.
		setTimeout(next, 60);
	}
}

// ── Karte oben oder unten, je nachdem wo Platz ist ──────────────
const cardStyle = computed(() => {
	const b = box.value;
	if (!b) return { top: '50%', transform: 'translateY(-50%)' };
	const below = window.innerHeight - (b.top + b.height);
	if (below > 220) return { top: `${b.top + b.height + 12}px` };
	return { bottom: `${window.innerHeight - b.top + 12}px` };
});
</script>

<template>
	<div class="tour" role="dialog" aria-modal="true" aria-label="Rundgang">
		<!-- Abdunkeln: vier Flächen um das Element, damit die Mitte antippbar bleibt -->
		<template v-if="box">
			<div class="shade" :style="{ top: 0, left: 0, right: 0, height: box.top + 'px' }" />
			<div class="shade" :style="{ top: box.top + box.height + 'px', left: 0, right: 0, bottom: 0 }" />
			<div class="shade" :style="{ top: box.top + 'px', left: 0, width: box.left + 'px', height: box.height + 'px' }" />
			<div
				class="shade"
				:style="{ top: box.top + 'px', left: box.left + box.width + 'px', right: 0, height: box.height + 'px' }"
			/>
			<!-- Ohne Tipp-Aufgabe soll das Element nicht anklickbar sein -->
			<div
				v-if="!step.tap"
				class="guard"
				:style="{ top: box.top + 'px', left: box.left + 'px', width: box.width + 'px', height: box.height + 'px' }"
			/>
			<div
				class="ring"
				:class="{ tapme: step.tap }"
				:style="{ top: box.top + 'px', left: box.left + 'px', width: box.width + 'px', height: box.height + 'px' }"
			/>
		</template>
		<div v-else class="shade full" />

		<div class="tourcard" :style="cardStyle">
			<div class="head">
				<span class="count">{{ idx + 1 }}/{{ steps.length }}</span>
				<h3>{{ content.title }}</h3>
			</div>
			<p class="lead">{{ content.lead }}</p>
			<ul v-if="content.points.length">
				<li v-for="(pt, i) in content.points" :key="i"><Check :size="13" /><span>{{ pt }}</span></li>
			</ul>

			<p v-if="step.tap && box" class="tap-hint"><Hand :size="14" /> Tippe auf das markierte Feld</p>

			<div class="nav">
				<button v-if="idx > 0" type="button" class="btn sm" @click="back"><ChevronLeft :size="14" /> Zurück</button>
				<button type="button" class="btn sm gold grow" @click="next">
					{{ isLast ? 'Fertig' : 'Weiter' }}<ChevronRight v-if="!isLast" :size="14" />
				</button>
			</div>
			<button type="button" class="btn ghost sm block skip" @click="finish">Rundgang überspringen</button>
		</div>
	</div>
</template>

<style scoped>
.tour { position: fixed; inset: 0; z-index: 200; }
.shade {
	position: fixed;
	background: rgba(4, 3, 2, 0.78);
	backdrop-filter: blur(1px);
}
.shade.full { inset: 0; }
.guard { position: fixed; }
.ring {
	position: fixed;
	border-radius: 13px;
	border: 2px solid var(--gold);
	box-shadow: 0 0 0 3px var(--gold-bg), 0 0 22px rgba(244, 177, 37, 0.45);
	pointer-events: none;
}
.ring.tapme { animation: tourpulse 1.5s ease-in-out infinite; }
@keyframes tourpulse {
	0%, 100% { box-shadow: 0 0 0 3px var(--gold-bg), 0 0 20px rgba(244, 177, 37, 0.4); }
	50% { box-shadow: 0 0 0 8px var(--gold-bg), 0 0 30px rgba(244, 177, 37, 0.6); }
}
@media (prefers-reduced-motion: reduce) {
	.ring.tapme { animation: none; }
}

.tourcard {
	position: fixed;
	left: 12px;
	right: 12px;
	max-width: 440px;
	margin: 0 auto;
	padding: 14px 15px 12px;
	background: var(--surface-2);
	border: 1px solid var(--line-2);
	border-radius: 16px;
	box-shadow: var(--shadow-2, 0 12px 40px rgba(0, 0, 0, 0.55));
}
.head { display: flex; align-items: center; gap: 9px; }
.count {
	font-family: var(--font-mono);
	font-size: 11px;
	font-variant-numeric: tabular-nums;
	color: var(--gold);
	background: var(--gold-bg);
	padding: 2px 7px;
	border-radius: 999px;
	flex-shrink: 0;
}
h3 {
	font-family: var(--font-display);
	font-size: 17px;
	font-weight: 700;
	text-transform: uppercase;
	letter-spacing: 0.02em;
	color: var(--ink);
}
.lead { margin: 8px 0 0; font-size: 13px; line-height: 1.5; color: var(--ink-2); }
ul { display: grid; gap: 7px; margin-top: 10px; }
li { display: flex; gap: 8px; font-size: 12.5px; line-height: 1.4; color: var(--ink-2); }
li :deep(svg) { color: var(--gold); flex-shrink: 0; margin-top: 3px; }
.tap-hint {
	display: flex;
	align-items: center;
	gap: 7px;
	margin-top: 10px;
	font-size: 12px;
	color: var(--gold);
}
.nav { display: flex; gap: 8px; margin-top: 12px; }
.nav .grow { flex: 1; justify-content: center; }
.skip { margin-top: 7px; }
</style>
