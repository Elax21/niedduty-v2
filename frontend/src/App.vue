<script setup lang="ts">
import { ref, computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useAuthStore } from './stores/auth';
import {
	Home, Trophy, CalendarDays, Gavel,
	Users, Settings, LogOut, MoreVertical, X
} from 'lucide-vue-next';

const route = useRoute();
const router = useRouter();
const auth = useAuthStore();

const showShell = computed(() => route.name !== 'login' && !!auth.user);

const tabs = [
	{ to: '/', name: 'dashboard', label: 'Start', icon: Home },
	{ to: '/liga', name: 'liga', label: 'Liga', icon: Trophy },
	{ to: '/termine', name: 'termine', label: 'Termine', icon: CalendarDays },
	{ to: '/strafen', name: 'strafen', label: 'Strafen', icon: Gavel }
];

const menuOpen = ref(false);

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
	<div v-if="showShell" class="app">
		<header class="appbar">
			<div class="crest"><img src="/logo.png" alt="Wappen ASG Aramäer Ahlen" /></div>
			<div class="titles">
				<div class="name">{{ auth.club?.short || 'Aramäer' }} Ahlen</div>
				<div class="sub">{{ auth.club?.liga || 'Vereins-Schaltzentrale' }}</div>
			</div>
			<button class="bar-btn" aria-label="Menü" @click="menuOpen = true">
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
						<button v-if="auth.isAdmin" class="menu-link" @click="go('/kader')">
							<Users :size="19" /> <span>Kader verwalten</span>
						</button>
						<button v-if="auth.isAdmin" class="menu-link" @click="go('/verwaltung')">
							<Settings :size="19" /> <span>Verwaltung &amp; Einbettungen</span>
						</button>
						<div class="chalk-divider" />
						<button class="menu-link danger" @click="logout">
							<LogOut :size="19" /> <span>Abmelden</span>
						</button>
					</div>
				</div>
			</div>
		</Transition>
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
.menu-link.danger { color: var(--bad); }
.menu-link :deep(svg) { color: var(--gold); flex-shrink: 0; }
.menu-link.danger :deep(svg) { color: var(--bad); }
.fade-enter-active, .fade-leave-active { transition: opacity var(--t-fast); }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
