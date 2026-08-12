<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useAuthStore } from './stores/auth';
import { pushActive, pushSubscribe, pushUnsubscribe, pushSupported, pushTest } from './lib/push';
import { watchInstall, openInstallHint, isStandalone } from './lib/install';
import { refreshAll, refreshing, watchVersion } from './lib/refresh';
import InstallHint from './components/InstallHint.vue';
import PushSettingsSheet from './components/PushSettingsSheet.vue';
import TourSheet from './components/TourSheet.vue';
import TourGuide from './components/TourGuide.vue';
import {
	Home, Trophy, CalendarDays, Wallet,
	Users, Settings, LogOut, MoreVertical, X,
	Instagram, Bell, BellOff, BarChart3, Smartphone, RefreshCw, SlidersHorizontal, HelpCircle, Vote
} from 'lucide-vue-next';

const route = useRoute();
const router = useRouter();
const auth = useAuthStore();

const showShell = computed(() => route.name !== 'login' && !!auth.user);

const tabs = [
	{ to: '/', name: 'dashboard', label: 'Start', icon: Home },
	{ to: '/liga', name: 'liga', label: 'Liga', icon: Trophy },
	{ to: '/termine', name: 'termine', label: 'Termine', icon: CalendarDays },
	{ to: '/strafen', name: 'strafen', label: 'Kasse', icon: Wallet }
];

const menuOpen = ref(false);

// ── Benachrichtigungen ──────────────────────────────────────────
const pushOn = ref(false);
const pushBusy = ref(false);
const pushMsg = ref('');
const canPush = pushSupported();

async function refreshPush() {
	if (canPush) pushOn.value = await pushActive();
}
onMounted(refreshPush);
watch(() => auth.user, refreshPush);

// ── App-Installation ────────────────────────────────────────────
const installed = isStandalone();
onMounted(watchInstall);
onMounted(() => watchVersion());

function showInstall() {
	menuOpen.value = false;
	openInstallHint();
}

// ── Erklärungen ─────────────────────────────────────────────────
// Beim ersten Einloggen läuft der geführte Rundgang durch die echte
// Bedienung; derselbe Inhalt liegt danach als Hilfe im Menü. Der Merker hängt
// am Konto (`user.tutorialDone`), nicht am Gerät.
const showPushSettings = ref(false);
const tourMode = ref<'hilfe' | null>(null);
const guideOn = ref(false);

watch(
	() => auth.user?.id,
	(id) => {
		if (!id || auth.user?.tutorialDone) return;
		setTimeout(() => {
			if (auth.user && !auth.user.tutorialDone) guideOn.value = true;
		}, 600);
	},
	{ immediate: true }
);

function endGuide() {
	guideOn.value = false;
	auth.setTutorialDone(true);
}
function restartGuide() {
	tourMode.value = null;
	menuOpen.value = false;
	setTimeout(() => (guideOn.value = true), 250);
}

function openHelp() {
	menuOpen.value = false;
	tourMode.value = 'hilfe';
}
function openPushSettings() {
	menuOpen.value = false;
	showPushSettings.value = true;
}

async function togglePush() {
	pushBusy.value = true;
	pushMsg.value = '';
	try {
		if (pushOn.value) {
			await pushUnsubscribe();
			pushMsg.value = 'Benachrichtigungen aus.';
		} else {
			await pushSubscribe();
			await pushTest();
			pushMsg.value = 'Eingeschaltet — Probenachricht ist unterwegs.';
		}
		await refreshPush();
	} catch (e) {
		pushMsg.value = e instanceof Error ? e.message : 'Hat nicht geklappt.';
	} finally {
		pushBusy.value = false;
	}
}

async function logout() {
	menuOpen.value = false;
	await auth.logout();
	router.push('/login');
}
function go(to: string) {
	menuOpen.value = false;
	router.push(to);
}
</script>

<template>
	<div v-if="showShell" class="app" :data-screen="String(route.name || '')">
		<header class="appbar">
			<div class="crest"><img src="/logo.png" alt="Wappen ASG Aramäer Ahlen" /></div>
			<div class="titles">
				<div class="name">{{ auth.club?.short || 'Aramäer' }} Ahlen</div>
				<div class="sub">{{ auth.club?.liga || 'Vereins-Schaltzentrale' }}</div>
			</div>
			<button
				class="bar-btn"
				:class="{ spinning: refreshing }"
				aria-label="Daten neu laden"
				:disabled="refreshing"
				@click="refreshAll"
			>
				<RefreshCw :size="18" />
			</button>
			<button class="bar-btn" data-tour="menu" aria-label="Menü" @click="menuOpen = true">
				<MoreVertical :size="20" />
			</button>
		</header>

		<main class="appmain">
			<RouterView v-slot="{ Component }">
				<component :is="Component" />
			</RouterView>
		</main>

		<nav class="tabbar" aria-label="Hauptnavigation">
			<RouterLink
				v-for="t in tabs"
				:key="t.to"
				:to="t.to"
				:data-tour="'tab-' + t.name"
				class="tabitem"
				:class="{ 'router-link-active': route.name === t.name }"
			>
				<span class="ic"><component :is="t.icon" :size="22" aria-hidden="true" /></span>
				<span>{{ t.label }}</span>
			</RouterLink>
		</nav>

		<!-- Menü-Sheet -->
		<Transition name="fade">
			<div v-if="menuOpen" class="modal-backdrop" @click.self="menuOpen = false">
				<div class="modal" role="dialog" aria-modal="true" aria-label="Menü">
					<div class="grabber" />
					<div class="card-head" style="border: none">
						<h2>{{ auth.user?.name }}</h2>
						<button class="btn sm icon ghost" aria-label="Schließen" @click="menuOpen = false"><X :size="16" /></button>
					</div>
					<div class="card-body" style="padding-top: 0">
						<button v-if="canPush" class="menu-link" data-tour="menu-push" :disabled="pushBusy" @click="togglePush">
							<component :is="pushOn ? Bell : BellOff" :size="19" />
							<span>{{ pushOn ? 'Benachrichtigungen an' : 'Benachrichtigungen aus' }}</span>
						</button>
						<p v-if="pushMsg" class="push-hint">{{ pushMsg }}</p>

						<button class="menu-link" data-tour="menu-erinnerungen" @click="openPushSettings">
							<SlidersHorizontal :size="19" /> <span>Erinnerungen einstellen</span>
						</button>

						<button v-if="!installed" class="menu-link" @click="showInstall">
							<Smartphone :size="19" /> <span>Auf den Startbildschirm</span>
						</button>

						<button class="menu-link" data-tour="menu-abstimmungen" @click="go('/abstimmungen')">
							<Vote :size="19" /> <span>Abstimmungen</span>
						</button>
						<button v-if="auth.can('beteiligung')" class="menu-link" data-tour="menu-beteiligung" @click="go('/beteiligung')">
							<BarChart3 :size="19" /> <span>Trainingsbeteiligung</span>
						</button>
						<button v-if="auth.isAdmin" class="menu-link" data-tour="menu-kader" @click="go('/kader')">
							<Users :size="19" /> <span>Kader &amp; Statistik</span>
						</button>
						<button v-if="auth.isAdmin" class="menu-link" data-tour="menu-verwaltung" @click="go('/verwaltung')">
							<Settings :size="19" /> <span>Verwaltung &amp; Einbettungen</span>
						</button>
						<a
							v-if="auth.club?.instagramUrl"
							class="menu-link"
							:href="auth.club.instagramUrl"
							target="_blank"
							rel="noopener"
						>
							<Instagram :size="19" /> <span>Instagram</span>
						</a>
						<button class="menu-link" data-tour="menu-hilfe" @click="openHelp">
							<HelpCircle :size="19" /> <span>Hilfe &amp; Erklärung</span>
						</button>
						<div class="chalk-divider" />
						<button class="menu-link danger" @click="logout">
							<LogOut :size="19" /> <span>Abmelden</span>
						</button>
					</div>
				</div>
			</div>
		</Transition>

		<InstallHint />
		<PushSettingsSheet v-if="showPushSettings" @close="showPushSettings = false" />
		<TourSheet v-if="tourMode" @close="tourMode = null" @restart="restartGuide" />
		<TourGuide v-if="guideOn" v-model:menu-open="menuOpen" @close="endGuide" />
	</div>
	<RouterView v-else />
</template>

<style scoped>
.menu-link {
	display: flex;
	align-items: center;
	gap: 13px;
	width: 100%;
	padding: 14px 6px;
	font-family: var(--font-display);
	font-size: 16px;
	font-weight: 600;
	text-transform: uppercase;
	letter-spacing: 0.03em;
	color: var(--ink);
	border-radius: 11px;
	transition: background var(--t-fast);
}
.menu-link:hover { background: var(--surface-3); }
.menu-link:disabled { opacity: 0.5; }
.menu-link.danger { color: var(--bad); }
.menu-link :deep(svg) { color: var(--gold); flex-shrink: 0; }
.menu-link.danger :deep(svg) { color: var(--bad); }
.bar-btn.spinning :deep(svg) { animation: spin 700ms linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
@media (prefers-reduced-motion: reduce) {
	.bar-btn.spinning :deep(svg) { animation: none; opacity: 0.5; }
}
.push-hint {
	font-size: 12.5px;
	color: var(--ink-3);
	padding: 0 6px 10px;
	line-height: 1.4;
}
.fade-enter-active, .fade-leave-active { transition: opacity var(--t-fast); }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
