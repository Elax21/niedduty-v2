<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { api, apiError } from '../services/api';
import { useAuthStore } from '../stores/auth';
import { Plus, Trash2 } from 'lucide-vue-next';
import AppModal from '../components/AppModal.vue';
import type { Club, Player, User } from '../types';

const auth = useAuthStore();
const users = ref<User[]>([]);
const players = ref<Player[]>([]);
const clubForm = ref<Club | null>(null);
const clubMsg = ref('');
const error = ref('');

const permLabels: Record<string, string> = {
	strafen: 'Strafen aufschreiben',
	termine: 'Termine verwalten',
	beteiligung: 'Beteiligung ansehen'
};

async function load() {
	const [u, p, c] = await Promise.all([
		api.get<User[]>('/users'),
		api.get<Player[]>('/players'),
		api.get<Club>('/club')
	]);
	users.value = u.data;
	players.value = p.data;
	clubForm.value = { ...c.data };
}

async function saveClub() {
	if (!clubForm.value) return;
	clubMsg.value = '';
	error.value = '';
	try {
		const { data } = await api.put<Club>('/club', clubForm.value);
		auth.club = data;
		clubMsg.value = 'Gespeichert.';
		setTimeout(() => (clubMsg.value = ''), 3000);
	} catch (e) {
		error.value = apiError(e);
	}
}

// ── Benutzer ──
const showUserForm = ref(false);
const userError = ref('');
const userForm = ref({ email: '', name: '', password: '', permissions: [] as string[], playerId: '' });

function openUserCreate() {
	userForm.value = { email: '', name: '', password: '', permissions: [], playerId: '' };
	userError.value = '';
	showUserForm.value = true;
}
async function submitUser() {
	userError.value = '';
	try {
		await api.post('/users', {
			...userForm.value,
			playerId: userForm.value.playerId || null
		});
		showUserForm.value = false;
		await load();
	} catch (e) {
		userError.value = apiError(e);
	}
}
async function togglePerm(u: User, perm: string) {
	const perms = u.permissions.includes(perm)
		? u.permissions.filter((p) => p !== perm)
		: [...u.permissions, perm];
	try {
		await api.put(`/users/${u.id}`, { name: u.name, permissions: perms, playerId: u.playerId });
		u.permissions = perms;
	} catch (e) {
		error.value = apiError(e);
	}
}
async function linkPlayer(u: User, ev: Event) {
	const val = (ev.target as HTMLSelectElement).value;
	try {
		await api.put(`/users/${u.id}`, { name: u.name, permissions: u.permissions, playerId: val || null });
		u.playerId = val || null;
	} catch (e) {
		error.value = apiError(e);
	}
}
async function removeUser(u: User) {
	if (!window.confirm(`Konto „${u.name}" (${u.email}) löschen?`)) return;
	await api.delete(`/users/${u.id}`);
	await load();
}

onMounted(load);
</script>

<template>
	<div class="page-head">
		<h1>Verwaltung</h1>
		<span class="sub">Nur für den Admin sichtbar</span>
	</div>

	<p v-if="error" class="form-error" role="alert">{{ error }}</p>

	<div class="grid cols-2" style="align-items: start">
		<div class="card">
			<div class="card-head"><h2>Verein</h2><span v-if="clubMsg" class="meta" style="color: var(--gruen)">{{ clubMsg }}</span></div>
			<div v-if="clubForm" class="card-body">
				<form @submit.prevent="saveClub">
					<div class="grid cols-2">
						<div class="field">
							<label for="cl-name">Name</label>
							<input id="cl-name" v-model="clubForm.name" required maxlength="100" />
						</div>
						<div class="field">
							<label for="cl-short">Kürzel</label>
							<input id="cl-short" v-model="clubForm.short" maxlength="5" />
						</div>
					</div>
					<div class="field">
						<label for="cl-liga">Liga</label>
						<input id="cl-liga" v-model="clubForm.liga" maxlength="60" />
					</div>
					<div class="field">
						<label for="cl-widget">fussball.de Widget-URL (Live-Tabelle)</label>
						<input id="cl-widget" v-model="clubForm.fussballDeWidget" maxlength="300" placeholder="https://www.fussball.de/widget2/…" />
					</div>
					<div class="grid cols-2">
						<div class="field">
							<label for="cl-iban">Kasse IBAN</label>
							<input id="cl-iban" v-model="clubForm.kasseIban" maxlength="40" />
						</div>
						<div class="field">
							<label for="cl-inhaber">Kasse Inhaber</label>
							<input id="cl-inhaber" v-model="clubForm.kasseInhaber" maxlength="100" />
						</div>
					</div>
					<button class="btn gold">Speichern</button>
				</form>
			</div>
		</div>

		<div class="card">
			<div class="card-head">
				<h2>Konten &amp; Rechte</h2>
				<button class="btn sm" @click="openUserCreate"><Plus :size="13" aria-hidden="true" /> Konto</button>
			</div>
			<div class="card-body flush">
				<div v-for="u in users" :key="u.id" class="user-row">
					<div class="user-top">
						<strong>{{ u.name }}</strong>
						<span class="chip" :class="{ mannschaftsabend: u.role === 'ADMIN' }">{{ u.role === 'ADMIN' ? 'Admin' : 'Mitglied' }}</span>
						<span style="color: var(--kreide-45); font-size: 12px">{{ u.email }}</span>
						<button v-if="u.role !== 'ADMIN'" class="btn sm danger" style="margin-left: auto" aria-label="Konto löschen" @click="removeUser(u)">
							<Trash2 :size="12" />
						</button>
					</div>
					<div v-if="u.role !== 'ADMIN'" class="user-perms">
						<label v-for="(label, perm) in permLabels" :key="perm" class="perm-check">
							<input type="checkbox" :checked="u.permissions.includes(perm)" @change="togglePerm(u, perm)" />
							{{ label }}
						</label>
						<label class="perm-check" style="margin-left: auto">
							Spieler:
							<select class="input" style="min-height: 30px; padding: 2px 6px; width: auto" :value="u.playerId ?? ''" @change="linkPlayer(u, $event)">
								<option value="">— keiner —</option>
								<option v-for="p in players" :key="p.id" :value="p.id">{{ p.name }}</option>
							</select>
						</label>
					</div>
				</div>
			</div>
		</div>
	</div>

	<AppModal v-if="showUserForm" title="Neues Konto" @close="showUserForm = false">
		<form @submit.prevent="submitUser">
			<p v-if="userError" class="form-error" role="alert">{{ userError }}</p>
			<div class="field">
				<label for="us-name">Name</label>
				<input id="us-name" v-model="userForm.name" required maxlength="100" />
			</div>
			<div class="field">
				<label for="us-email">E-Mail</label>
				<input id="us-email" v-model="userForm.email" type="email" required maxlength="120" />
			</div>
			<div class="field">
				<label for="us-pass">Passwort (min. 8 Zeichen)</label>
				<input id="us-pass" v-model="userForm.password" type="password" required minlength="8" maxlength="100" />
			</div>
			<div class="field">
				<label>Rechte</label>
				<label v-for="(label, perm) in permLabels" :key="perm" class="perm-check">
					<input v-model="userForm.permissions" type="checkbox" :value="perm" /> {{ label }}
				</label>
			</div>
			<div class="field">
				<label for="us-player">Verknüpfter Spieler (für Zu-/Absagen)</label>
				<select id="us-player" v-model="userForm.playerId">
					<option value="">— keiner —</option>
					<option v-for="p in players" :key="p.id" :value="p.id">{{ p.name }}</option>
				</select>
			</div>
			<button class="btn primary" style="width: 100%; justify-content: center">Konto anlegen</button>
		</form>
	</AppModal>
</template>

<style scoped>
.user-row { padding: 10px 14px; border-bottom: 1px solid var(--hair); }
.user-row:last-child { border-bottom: none; }
.user-top { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.user-perms { display: flex; align-items: center; gap: 14px; margin-top: 8px; flex-wrap: wrap; }
.perm-check { display: inline-flex; align-items: center; gap: 5px; font-size: 12.5px; color: var(--kreide-70); cursor: pointer; }
</style>
