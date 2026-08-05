<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { api, apiError } from '../services/api';
import { useAuthStore } from '../stores/auth';
import { Plus, Trash2, Copy, Share2, RefreshCw, Link2 } from 'lucide-vue-next';
import AppModal from '../components/AppModal.vue';
import type { Club, Invite, Player, User } from '../types';

const auth = useAuthStore();
const users = ref<User[]>([]);
const players = ref<Player[]>([]);
const clubForm = ref<Club | null>(null);
const clubMsg = ref('');
const error = ref('');

// ── Einladungslink ──
const invite = ref<Invite | null>(null);
const inviteMsg = ref('');
const inviteLink = computed(() => (invite.value ? `${location.origin}/register/${invite.value.token}` : ''));

async function loadInvite() {
	const { data } = await api.get<Invite | null>('/invite');
	invite.value = data;
}
async function newInvite() {
	const { data } = await api.post<Invite>('/invite');
	invite.value = data;
	inviteMsg.value = 'Neuer Link erzeugt – der alte ist jetzt ungültig.';
	setTimeout(() => (inviteMsg.value = ''), 3500);
}
async function stopInvite() {
	if (!window.confirm('Aktuellen Einladungslink ungültig machen?')) return;
	await api.delete('/invite');
	invite.value = null;
}
async function copyLink() {
	try {
		await navigator.clipboard.writeText(inviteLink.value);
		inviteMsg.value = 'Link kopiert!';
		setTimeout(() => (inviteMsg.value = ''), 2000);
	} catch {
		inviteMsg.value = 'Kopieren nicht möglich – Link markieren.';
	}
}
async function shareLink() {
	const nav = navigator as Navigator & { share?: (d: ShareData) => Promise<void> };
	if (nav.share) {
		try {
			await nav.share({ title: `${auth.club?.name ?? 'Team'} – Einladung`, text: 'Tritt unserer Team-App bei:', url: inviteLink.value });
		} catch { /* abgebrochen */ }
	} else {
		copyLink();
	}
}

const permLabels: Record<string, string> = {
	strafen: 'Strafen aufschreiben',
	termine: 'Termine verwalten',
	beteiligung: 'Trainingsbeteiligung sehen'
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
const userForm = ref({ alias: '', email: '', name: '', password: '', permissions: [] as string[], playerId: '', noPlayer: false });

function openUserCreate() {
	userForm.value = { alias: '', email: '', name: '', password: '', permissions: [], playerId: '', noPlayer: false };
	userError.value = '';
	showUserForm.value = true;
}
async function submitUser() {
	userError.value = '';
	try {
		await api.post('/users', { ...userForm.value, playerId: userForm.value.playerId || null });
		showUserForm.value = false;
		await load();
	} catch (e) {
		userError.value = apiError(e);
	}
}
async function togglePerm(u: User, perm: string) {
	const perms = u.permissions.includes(perm) ? u.permissions.filter((p) => p !== perm) : [...u.permissions, perm];
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

onMounted(() => { load(); loadInvite(); });
</script>

<template>
	<div class="page-head">
		<h1>Verwaltung</h1>
		<span class="sub">Nur Admin</span>
	</div>

	<p v-if="error" class="form-error" role="alert">{{ error }}</p>

	<div class="stack">
		<!-- Einladungslink -->
		<div class="card">
			<div class="card-head">
				<h2><Link2 :size="16" style="vertical-align: -3px; color: var(--gold)" /> Einladungslink</h2>
			</div>
			<div class="card-body">
				<p class="hint" style="margin-bottom: 12px">
					Teile diesen Link mit dem Team. Jeder legt sich damit selbst ein Konto an (Vorname, Nachname, Alias, Passwort) und landet automatisch im Kader.
				</p>
				<p v-if="inviteMsg" class="form-ok" role="status">{{ inviteMsg }}</p>

				<template v-if="invite">
					<div class="invite-link">{{ inviteLink }}</div>
					<div class="invite-actions">
						<button class="btn sm gold" @click="shareLink"><Share2 :size="14" /> Teilen</button>
						<button class="btn sm" @click="copyLink"><Copy :size="14" /> Kopieren</button>
						<button class="btn sm ghost" @click="newInvite"><RefreshCw :size="14" /> Neu</button>
						<button class="btn sm danger" style="margin-left: auto" @click="stopInvite">Deaktivieren</button>
					</div>
					<p class="hint" style="margin-top: 10px">Bisher genutzt: {{ invite.useCount }}×{{ invite.maxUses ? ' / ' + invite.maxUses : ' (unbegrenzt)' }}</p>
				</template>
				<template v-else>
					<p class="empty" style="padding: 12px">Noch kein aktiver Link.</p>
					<button class="btn primary block" @click="newInvite"><Plus :size="15" /> Einladungslink erstellen</button>
				</template>
			</div>
		</div>

		<!-- Verein -->
		<div class="card">
			<div class="card-head"><h2>Verein</h2><span v-if="clubMsg" class="meta" style="color: var(--gruen)">{{ clubMsg }}</span></div>
			<div v-if="clubForm" class="card-body">
				<form @submit.prevent="saveClub">
					<div class="row2">
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

					<div class="chalk-divider" />
					<p class="hint">fussball.de-Widget-IDs (aus <code>next.fussball.de/widgets</code> → Widget öffnen → ID aus der Adresszeile). Nur die ID, z.B. <code>aab8a3a1-…</code></p>
					<div class="field">
						<label for="cl-w-table">Tabelle (table)</label>
						<input id="cl-w-table" v-model="clubForm.fussballTableId" maxlength="64" placeholder="85bc8155-…" />
					</div>
					<div class="field">
						<label for="cl-w-matches">Spiele – letzte & nächste (team-matches)</label>
						<input id="cl-w-matches" v-model="clubForm.fussballMatchesId" maxlength="64" placeholder="82539853-…" />
					</div>
					<div class="field">
						<label for="cl-w-next">Nächstes Spiel (next-match)</label>
						<input id="cl-w-next" v-model="clubForm.fussballNextMatchId" maxlength="64" placeholder="aab8a3a1-…" />
					</div>

					<div class="field">
						<label for="cl-team">Mannschafts-ID fussball.de (wird automatisch erkannt)</label>
						<input id="cl-team" v-model="clubForm.fussballTeamId" maxlength="64" placeholder="011MIC2EF8…" />
					</div>

					<div class="chalk-divider" />
					<div class="field">
						<label for="cl-gcal">Google Team-Kalender (optional)</label>
						<input id="cl-gcal" v-model="clubForm.googleCalendarUrl" maxlength="400" placeholder="https://calendar.google.com/…" />
					</div>
					<div class="field">
						<label for="cl-insta">Instagram (optional)</label>
						<input id="cl-insta" v-model="clubForm.instagramUrl" maxlength="200" placeholder="https://www.instagram.com/…" />
					</div>

					<div class="chalk-divider" />
					<div class="row2">
						<div class="field">
							<label for="cl-iban">Kasse IBAN</label>
							<input id="cl-iban" v-model="clubForm.kasseIban" maxlength="40" />
						</div>
						<div class="field">
							<label for="cl-inhaber">Inhaber</label>
							<input id="cl-inhaber" v-model="clubForm.kasseInhaber" maxlength="100" />
						</div>
					</div>
					<button class="btn gold block">Speichern</button>
				</form>
			</div>
		</div>

		<!-- Konten -->
		<div class="card">
			<div class="card-head">
				<h2>Konten &amp; Rechte</h2>
				<button class="btn sm" @click="openUserCreate"><Plus :size="13" /> Konto</button>
			</div>
			<div class="card-body flush">
				<div v-for="u in users" :key="u.id" class="user-row">
					<div class="user-top">
						<strong>{{ u.name }}</strong>
						<span class="chip" :class="{ admin: u.role === 'ADMIN' }">{{ u.role === 'ADMIN' ? 'Admin' : 'Mitglied' }}</span>
						<button v-if="u.role !== 'ADMIN'" class="btn sm icon danger" style="margin-left: auto" aria-label="Konto löschen" @click="removeUser(u)"><Trash2 :size="13" /></button>
					</div>
					<div class="user-mail"><span class="mono" style="color: var(--gold)">@{{ u.alias }}</span><template v-if="u.email"> · {{ u.email }}</template></div>
					<div v-if="u.role !== 'ADMIN'" class="user-perms">
						<label v-for="(label, perm) in permLabels" :key="perm" class="perm-check">
							<input type="checkbox" :checked="u.permissions.includes(perm)" @change="togglePerm(u, perm)" /> {{ label }}
						</label>
						<label class="perm-check">
							Spieler:
							<select class="input" style="min-height: 34px; padding: 4px 8px; width: auto; flex: 1" :value="u.playerId ?? ''" @change="linkPlayer(u, $event)">
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
			<div class="row2">
				<div class="field">
					<label for="us-alias">Alias (Login)</label>
					<input id="us-alias" v-model="userForm.alias" autocapitalize="none" required placeholder="z.B. tuma9" />
				</div>
				<div class="field">
					<label for="us-email">E-Mail (optional)</label>
					<input id="us-email" v-model="userForm.email" type="email" maxlength="120" />
				</div>
			</div>
			<div class="field">
				<label for="us-pass">Passwort (min. 8 Zeichen)</label>
				<input id="us-pass" v-model="userForm.password" type="password" required minlength="8" maxlength="100" />
			</div>
			<div class="field">
				<label>Rechte</label>
				<label v-for="(label, perm) in permLabels" :key="perm" class="perm-check" style="margin-top: 4px">
					<input v-model="userForm.permissions" type="checkbox" :value="perm" /> {{ label }}
				</label>
			</div>
			<div class="field">
				<label for="us-player">Verknüpfter Spieler (für Zu-/Absagen)</label>
				<select id="us-player" v-model="userForm.playerId">
					<option value="">— neuen Kader-Eintrag anlegen —</option>
					<option v-for="p in players" :key="p.id" :value="p.id">{{ p.name }}</option>
				</select>
				<label class="perm-check" style="margin-top: 8px">
					<input v-model="userForm.noPlayer" type="checkbox" />
					Kein Kader-Eintrag (Trainer, Betreuer)
				</label>
			</div>
			<button class="btn primary block">Konto anlegen</button>
		</form>
	</AppModal>
</template>

<style scoped>
.hint { font-size: 12.5px; color: var(--ink-3); margin-bottom: 10px; }
.invite-link {
	font-family: var(--font-mono);
	font-size: 12.5px;
	word-break: break-all;
	color: var(--ink-2);
	background: var(--bg);
	border: 1px solid var(--line-2);
	border-radius: 11px;
	padding: 11px 12px;
}
.invite-actions { display: flex; align-items: center; gap: 7px; flex-wrap: wrap; margin-top: 10px; }
.hint code { font-family: var(--font-mono); color: var(--ink-2); font-size: 11.5px; }
.user-row { padding: 13px 15px; border-bottom: 1px solid var(--line); }
.user-row:last-child { border-bottom: none; }
.user-top { display: flex; align-items: center; gap: 8px; }
.user-mail { font-size: 12.5px; color: var(--ink-3); margin-top: 3px; }
.user-perms { display: flex; flex-direction: column; gap: 9px; margin-top: 11px; }
.perm-check { display: inline-flex; align-items: center; gap: 8px; font-size: 14px; color: var(--ink-2); cursor: pointer; }
.perm-check input { width: 20px; height: 20px; accent-color: var(--gold); flex-shrink: 0; }
</style>
