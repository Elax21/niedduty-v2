<script setup lang="ts">
import { computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useAuthStore } from './stores/auth';
import {
	LayoutDashboard, ListOrdered, CalendarDays, ClipboardCheck,
	Gavel, Users, Settings, LogOut
} from 'lucide-vue-next';

const route = useRoute();
const router = useRouter();
const auth = useAuthStore();

const showShell = computed(() => route.name !== 'login' && !!auth.user);

const navItems = computed(() => {
	const items = [
		{ to: '/', label: 'Zentrale', icon: LayoutDashboard, show: true },
		{ to: '/tabelle', label: 'Tabelle', icon: ListOrdered, show: true },
		{ to: '/kalender', label: 'Kalender', icon: CalendarDays, show: true },
		{ to: '/training', label: 'Training', icon: ClipboardCheck, show: auth.can('beteiligung') },
		{ to: '/strafen', label: 'Strafen', icon: Gavel, show: true },
		{ to: '/kader', label: 'Kader', icon: Users, show: auth.isAdmin },
		{ to: '/einstellungen', label: 'Verwaltung', icon: Settings, show: auth.isAdmin }
	];
	return items.filter((i) => i.show);
});

async function logout() {
	await auth.logout();
	router.push('/login');
}
</script>

<template>
	<div v-if="showShell" class="shell">
		<aside class="sidebar">
			<div class="club-block">
				<div class="crest">
					<div class="crest-mark"><img src="/logo.png" alt="" /></div>
					<div>
						<div class="name">{{ auth.club?.name || 'Aramäer Ahlen' }}</div>
						<div class="liga">{{ auth.club?.liga || '' }}</div>
					</div>
				</div>
			</div>
			<nav class="nav" aria-label="Hauptnavigation">
				<RouterLink v-for="item in navItems" :key="item.to" :to="item.to" class="navitem">
					<component :is="item.icon" :size="17" aria-hidden="true" />
					<span>{{ item.label }}</span>
				</RouterLink>
			</nav>
			<div class="sidebar-foot">
				<div>{{ auth.user?.name }}</div>
				<button class="btn sm" style="margin-top: 8px" @click="logout">
					<LogOut :size="13" aria-hidden="true" /> Abmelden
				</button>
			</div>
		</aside>
		<main class="main">
			<RouterView />
		</main>
	</div>
	<RouterView v-else />
</template>
