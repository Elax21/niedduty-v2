import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { api } from '../services/api';
import type { Club, User } from '../types';

export const useAuthStore = defineStore('auth', () => {
	const user = ref<User | null>(null);
	const club = ref<Club | null>(null);
	const loaded = ref(false);

	const isAdmin = computed(() => user.value?.role === 'ADMIN');

	function can(perm: 'strafen' | 'termine' | 'beteiligung'): boolean {
		if (!user.value) return false;
		return user.value.role === 'ADMIN' || user.value.permissions.includes(perm);
	}

	async function fetchMe(): Promise<boolean> {
		try {
			const { data } = await api.get<User>('/auth/me');
			user.value = data;
			if (!club.value) {
				const res = await api.get<Club>('/club');
				club.value = res.data;
			}
			return true;
		} catch {
			user.value = null;
			return false;
		} finally {
			loaded.value = true;
		}
	}

	async function login(login: string, password: string) {
		const { data } = await api.post<User>('/auth/login', { login, password });
		user.value = data;
		const res = await api.get<Club>('/club');
		club.value = res.data;
	}

	async function register(payload: {
		token: string;
		firstName: string;
		lastName: string;
		alias: string;
		password: string;
	}) {
		const { data } = await api.post<User>('/auth/register', payload);
		user.value = data;
		const res = await api.get<Club>('/club');
		club.value = res.data;
	}

	async function logout() {
		await api.post('/auth/logout');
		user.value = null;
	}

	return { user, club, loaded, isAdmin, can, fetchMe, login, register, logout };
});
