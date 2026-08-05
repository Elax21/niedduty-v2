<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { api, apiError } from '../services/api';
import { useAuthStore } from '../stores/auth';
import { animate, stagger } from 'animejs';
import { reducedMotion } from '../lib/motion';

const route = useRoute();
const router = useRouter();
const auth = useAuthStore();

const token = String(route.params.token || '');
const state = ref<'checking' | 'ok' | 'invalid'>('checking');
const clubName = ref('Aramäer Ahlen');

const firstName = ref('');
const lastName = ref('');
const alias = ref('');
const password = ref('');
const birthday = ref('');
const error = ref('');
const busy = ref(false);

onMounted(async () => {
	try {
		const { data } = await api.get<{ valid: boolean; clubName?: string }>(`/invite/${token}`);
		if (data.valid) {
			state.value = 'ok';
			if (data.clubName) clubName.value = data.clubName;
			if (!reducedMotion()) {
				requestAnimationFrame(() =>
					animate('.reg-anim', { opacity: [0, 1], translateY: [16, 0], delay: stagger(80, { start: 60 }), duration: 420, ease: 'outQuad' })
				);
			}
		} else {
			state.value = 'invalid';
		}
	} catch {
		state.value = 'invalid';
	}
});

// Alias-Vorschlag aus Vorname
function suggestAlias() {
	if (!alias.value && firstName.value) {
		alias.value = firstName.value.trim().toLowerCase().replace(/[^a-z0-9._-]/g, '');
	}
}

async function submit() {
	error.value = '';
	busy.value = true;
	try {
		await auth.register({
			token,
			firstName: firstName.value.trim(),
			lastName: lastName.value.trim(),
			alias: alias.value.trim().toLowerCase(),
			password: password.value,
			birthday: birthday.value
		});
		router.push('/');
	} catch (e) {
		error.value = apiError(e);
	} finally {
		busy.value = false;
	}
}
</script>

<template>
	<div class="login-wrap">
		<div class="login-inner">
			<div class="reg-anim crest-hero"><img src="/logo.png" alt="Vereinswappen" /></div>

			<template v-if="state === 'ok'">
				<h1 class="reg-anim brand">Willkommen<br />im Team</h1>
				<p class="reg-anim tagline">{{ clubName }} · Konto anlegen</p>

				<form class="reg-anim reg-form" @submit.prevent="submit">
					<p v-if="error" class="form-error" role="alert">{{ error }}</p>
					<div class="row2">
						<div class="field">
							<label for="fn">Vorname</label>
							<input id="fn" v-model="firstName" autocomplete="given-name" required maxlength="60" @blur="suggestAlias" />
						</div>
						<div class="field">
							<label for="ln">Nachname</label>
							<input id="ln" v-model="lastName" autocomplete="family-name" required maxlength="60" />
						</div>
					</div>
					<div class="field">
						<label for="al">Alias (Login-Name)</label>
						<input id="al" v-model="alias" autocapitalize="none" autocomplete="username" placeholder="z.B. tuma9" required />
						<span class="hint">3–24 Zeichen: a–z, 0–9, . _ -</span>
					</div>
					<div class="field">
						<label for="bd">Geburtstag (für die Glückwünsche)</label>
						<input id="bd" v-model="birthday" type="date" autocomplete="bday" />
					</div>
					<div class="field">
						<label for="pw">Passwort</label>
						<input id="pw" v-model="password" type="password" autocomplete="new-password" minlength="8" required placeholder="min. 8 Zeichen" />
					</div>
					<button class="btn primary block" style="margin-top: 4px" :disabled="busy">
						{{ busy ? 'Anlegen …' : 'Konto anlegen & los' }}
					</button>
				</form>
				<p class="reg-anim foot">Schon dabei? <RouterLink to="/login" style="color: var(--gold)">Einloggen</RouterLink></p>
			</template>

			<template v-else-if="state === 'invalid'">
				<h1 class="brand" style="font-size: 34px">Link ungültig</h1>
				<p class="tagline" style="text-transform: none; letter-spacing: 0; font-size: 14px; margin-top: 12px; color: var(--ink-2)">
					Dieser Einladungslink ist abgelaufen oder wurde zurückgezogen. Frag Alessandro nach einem neuen.
				</p>
				<RouterLink to="/login" class="btn ghost block" style="margin-top: 22px">Zum Login</RouterLink>
			</template>

			<div v-else class="tagline" style="margin-top: 30px">Einladung wird geprüft …</div>
		</div>
	</div>
</template>

<style scoped>
.login-wrap {
	min-height: 100vh;
	display: flex;
	align-items: center;
	justify-content: center;
	padding: 24px 20px calc(24px + env(safe-area-inset-bottom, 0px));
	background:
		radial-gradient(120% 55% at 50% -8%, rgba(244, 177, 37, 0.16), transparent 60%),
		radial-gradient(90% 40% at 82% 6%, rgba(214, 58, 53, 0.12), transparent 60%),
		var(--bg);
	background-attachment: fixed;
}
.login-inner { width: 100%; max-width: 400px; text-align: center; }
.crest-hero {
	width: 82px; height: 82px;
	margin: 0 auto 16px;
	border-radius: 22px;
	overflow: hidden;
	background: #fff;
	box-shadow: 0 0 0 3px var(--gold-deep), 0 18px 40px rgba(0, 0, 0, 0.55);
}
.crest-hero img { width: 100%; height: 100%; object-fit: cover; }
.brand { font-size: 40px; letter-spacing: 0.01em; }
.tagline {
	color: var(--ink-3);
	font-family: var(--font-display);
	text-transform: uppercase;
	letter-spacing: 0.06em;
	font-size: 12.5px;
	margin-top: 6px;
}
.reg-form { margin-top: 24px; text-align: left; }
.hint { font-size: 11.5px; color: var(--ink-3); }
.foot { margin-top: 16px; font-size: 13px; color: var(--ink-3); }
</style>
