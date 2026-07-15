import { createRouter, createWebHistory } from 'vue-router';
import { useAuthStore } from '../stores/auth';

const router = createRouter({
	history: createWebHistory(),
	routes: [
		{ path: '/login', name: 'login', component: () => import('../views/LoginView.vue') },
		{ path: '/', name: 'dashboard', component: () => import('../views/DashboardView.vue') },
		{ path: '/tabelle', name: 'tabelle', component: () => import('../views/TabelleView.vue') },
		{ path: '/kalender', name: 'kalender', component: () => import('../views/KalenderView.vue') },
		{
			path: '/training',
			name: 'training',
			component: () => import('../views/TrainingView.vue'),
			meta: { perm: 'beteiligung' }
		},
		{ path: '/strafen', name: 'strafen', component: () => import('../views/StrafenView.vue') },
		{ path: '/kader', name: 'kader', component: () => import('../views/KaderView.vue'), meta: { admin: true } },
		{
			path: '/einstellungen',
			name: 'einstellungen',
			component: () => import('../views/EinstellungenView.vue'),
			meta: { admin: true }
		}
	]
});

router.beforeEach(async (to) => {
	const auth = useAuthStore();
	if (!auth.loaded) await auth.fetchMe();
	if (to.name !== 'login' && !auth.user) return { name: 'login' };
	if (to.name === 'login' && auth.user) return { name: 'dashboard' };
	if (to.meta.admin && !auth.isAdmin) return { name: 'dashboard' };
	if (to.meta.perm && !auth.can(to.meta.perm as 'beteiligung')) return { name: 'dashboard' };
});

export default router;
