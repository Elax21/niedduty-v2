<script setup lang="ts">
// Flutlicht-Hinweis: „Pack die Kabine auf deinen Startbildschirm."
// Chrome bekommt den echten Install-Dialog, iOS die Teilen-Anleitung.
import { ref } from 'vue';
import { Share, PlusSquare, Download, X } from 'lucide-vue-next';
import {
	installHintOpen, closeInstallHint, canPromptInstall, promptInstall, isIos, isIosChrome
} from '../lib/install';

const busy = ref(false);
const ios = isIos();
// Safari hat „Teilen" unten in der Leiste, Chrome/Firefox auf iOS oben rechts.
const shareWhere = isIosChrome() ? 'Oben rechts' : 'Unten in der Leiste';

async function install() {
	busy.value = true;
	try {
		await promptInstall();
	} finally {
		busy.value = false;
		closeInstallHint();
	}
}
</script>

<template>
	<Transition name="fade">
		<div v-if="installHintOpen" class="modal-backdrop" @click.self="closeInstallHint()">
			<div class="modal install" role="dialog" aria-modal="true" aria-label="App installieren">
				<div class="grabber" />

				<!-- Signature: zwei Flutlicht-Kegel auf das Wappen -->
				<div class="stage" aria-hidden="true">
					<span class="beam left" />
					<span class="beam right" />
					<span class="pool" />
					<div class="badge"><img src="/logo.png" alt="" /></div>
				</div>

				<div class="card-body">
					<h2 class="head">Kabine aufs Handy</h2>
					<p class="lead">
						Als App auf dem Startbildschirm: schneller drin, ohne Browserleiste — und
						Benachrichtigungen zu Spielen und Training kommen zuverlässig an.
					</p>

					<ol v-if="ios" class="steps">
						<li>
							<span class="num">1</span>
							<span>{{ shareWhere }} auf <Share :size="15" /> <b>Teilen</b> tippen.</span>
						</li>
						<li>
							<span class="num">2</span>
							<span><PlusSquare :size="15" /> <b>Zum Home-Bildschirm</b> wählen.</span>
						</li>
						<li>
							<span class="num">3</span>
							<span>Mit <b>Hinzufügen</b> bestätigen — fertig.</span>
						</li>
					</ol>

					<button
						v-else-if="canPromptInstall"
						class="btn gold block"
						:disabled="busy"
						@click="install"
					>
						<Download :size="17" /> Jetzt hinzufügen
					</button>

					<p v-else class="lead">
						Im Browser-Menü <b>„Zum Startbildschirm hinzufügen"</b> wählen.
					</p>

					<button class="btn ghost block later" @click="closeInstallHint()">
						<X :size="15" /> Später
					</button>
				</div>
			</div>
		</div>
	</Transition>
</template>

<style scoped>
.install .card-body { padding-top: 4px; }

/* ── Flutlicht-Bühne ───────────────────────────────────────────── */
.stage {
	position: relative;
	height: 150px;
	margin: 2px 0 4px;
	overflow: hidden;
	display: grid;
	place-items: end center;
	padding-bottom: 14px;
}
.beam {
	position: absolute;
	top: -30px;
	width: 130px;
	height: 210px;
	background: linear-gradient(to bottom, var(--gold-bg), transparent 72%);
	filter: blur(7px);
	opacity: 0.9;
	animation: beam-flicker 5s var(--ease-out) infinite;
}
.beam.left {
	left: 8%;
	clip-path: polygon(38% 0, 62% 0, 100% 100%, 0 100%);
	transform: rotate(13deg);
}
.beam.right {
	right: 8%;
	clip-path: polygon(38% 0, 62% 0, 100% 100%, 0 100%);
	transform: rotate(-13deg);
	animation-delay: -2.2s;
}
.pool {
	position: absolute;
	bottom: 4px;
	width: 200px;
	height: 54px;
	border-radius: 50%;
	background: radial-gradient(ellipse at center, var(--gold-bg), transparent 70%);
	filter: blur(3px);
}
.badge {
	position: relative;
	width: 84px;
	height: 84px;
	display: grid;
	place-items: center;
	border-radius: 21px;
	background: var(--surface-2);
	border: 1px solid var(--line-2);
	box-shadow: var(--shadow-gold);
	animation: badge-in var(--t-med) var(--ease-out) both;
}
.badge img { width: 60px; height: 60px; object-fit: contain; }

@keyframes beam-flicker {
	0%, 100% { opacity: 0.9; }
	45% { opacity: 0.62; }
	52% { opacity: 1; }
}
@keyframes badge-in {
	from { opacity: 0; transform: translateY(10px) scale(0.94); }
	to { opacity: 1; transform: none; }
}

/* ── Text & Schritte ───────────────────────────────────────────── */
.head {
	font-family: var(--font-display);
	font-size: 24px;
	font-weight: 700;
	text-transform: uppercase;
	letter-spacing: 0.02em;
	text-align: center;
	color: var(--ink);
}
.lead {
	margin: 8px 0 16px;
	font-size: 13.5px;
	line-height: 1.5;
	text-align: center;
	color: var(--ink-2);
}
.steps { display: grid; gap: 10px; margin-bottom: 16px; }
.steps li {
	display: flex;
	align-items: center;
	gap: 11px;
	padding: 11px 12px;
	font-size: 13.5px;
	color: var(--ink-2);
	background: var(--surface-flat);
	border: 1px solid var(--line);
	border-radius: var(--radius-sm);
}
.steps b { color: var(--ink); }
.steps :deep(svg) { color: var(--gold); vertical-align: -3px; }
.num {
	flex-shrink: 0;
	width: 24px;
	height: 24px;
	display: grid;
	place-items: center;
	font-family: var(--font-mono);
	font-size: 12px;
	color: var(--gold-ink);
	background: var(--gold);
	border-radius: 50%;
}
.later { margin-top: 10px; }

@media (prefers-reduced-motion: reduce) {
	.beam, .badge { animation: none; }
}
.fade-enter-active, .fade-leave-active { transition: opacity var(--t-fast); }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
