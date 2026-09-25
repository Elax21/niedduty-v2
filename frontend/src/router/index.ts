import { createRouter, createWebHistory } from 'vue-router';
import { useAuthStore } from '../stores/auth';

const router = createRouter({
	history: createWebHistory(),
	routes: [
		{ path: '/login', name: 'login', component: () => import('../views/LoginView.vue') },
		{ path: '/register/:token', name: 'register', component: () => import('../views/RegisterView.vue') },
		// Die App ist auf die Kasse (Strafenkatalog) reduziert — sie ist die Startseite.
		{ path: '/', name: 'strafen', component: () => import('../views/StrafenView.vue') },
		{ path: '/kader', name: 'kader', component: () => import('../views/KaderView.vue'), meta: { admin: true } },
		{ path: '/verwaltung', name: 'verwaltung', component: () => import('../views/EinstellungenView.vue'), meta: { admin: true } },
		// Alte/entfernte Pfade landen wieder auf der Kasse.
		{ path: '/:pathMatch(.*)*', redirect: '/' }
	]
});

const publicRoutes = new Set(['login', 'register']);

router.beforeEach(async (to) => {
	const auth = useAuthStore();
	if (!auth.loaded) await auth.fetchMe();
	const isPublic = publicRoutes.has(String(to.name));
	if (!isPublic && !auth.user) return { name: 'login' };
	if (to.name === 'login' && auth.user) return { path: '/' };
	if (to.meta.admin && !auth.isAdmin) return { path: '/' };
});

export default router;
