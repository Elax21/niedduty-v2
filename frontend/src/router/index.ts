import { createRouter, createWebHistory } from 'vue-router';
import { useAuthStore } from '../stores/auth';

const router = createRouter({
	history: createWebHistory(),
	routes: [
		{ path: '/login', name: 'login', component: () => import('../views/LoginView.vue') },
		{ path: '/register/:token', name: 'register', component: () => import('../views/RegisterView.vue') },
		{ path: '/', name: 'dashboard', component: () => import('../views/DashboardView.vue') },
		{ path: '/liga', name: 'liga', component: () => import('../views/LigaView.vue') },
		{ path: '/termine', name: 'termine', component: () => import('../views/TermineView.vue') },
		{ path: '/strafen', name: 'strafen', component: () => import('../views/StrafenView.vue') },
		{ path: '/kader', name: 'kader', component: () => import('../views/KaderView.vue'), meta: { admin: true } },
		{ path: '/verwaltung', name: 'verwaltung', component: () => import('../views/EinstellungenView.vue'), meta: { admin: true } },
		// Alte Pfade umleiten
		{ path: '/tabelle', redirect: '/liga' },
		{ path: '/kalender', redirect: '/termine' },
		{ path: '/einstellungen', redirect: '/verwaltung' },
		{ path: '/training', redirect: '/' },
		{ path: '/:pathMatch(.*)*', redirect: '/' }
	]
});

const publicRoutes = new Set(['login', 'register']);

router.beforeEach(async (to) => {
	const auth = useAuthStore();
	if (!auth.loaded) await auth.fetchMe();
	const isPublic = publicRoutes.has(String(to.name));
	if (!isPublic && !auth.user) return { name: 'login' };
	if (to.name === 'login' && auth.user) return { name: 'dashboard' };
	if (to.meta.admin && !auth.isAdmin) return { name: 'dashboard' };
});

export default router;
