<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useAuthStore } from '../stores/auth';
import { apiError } from '../services/api';
import { animate } from 'animejs';
import { reducedMotion } from '../lib/motion';

const router = useRouter();
const auth = useAuthStore();

const email = ref('');
const password = ref('');
const error = ref('');
const busy = ref(false);

async function submit() {
	error.value = '';
	busy.value = true;
	try {
		await auth.login(email.value, password.value);
		router.push('/');
	} catch (e) {
		error.value = apiError(e);
	} finally {
		busy.value = false;
	}
}

onMounted(() => {
	if (reducedMotion()) return;
	animate('.login-card', { opacity: [0, 1], translateY: [16, 0], duration: 420, ease: 'outQuad' });
});
</script>

<template>
	<div class="login-wrap">
		<div class="login-card card">
			<div class="card-body" style="padding: 26px">
				<div class="crest-mark" style="width: 60px; height: 60px; margin-bottom: 14px"><img src="/logo.png" alt="Vereinswappen ASG Aramäer Ahlen" /></div>
				<h1 style="font-size: 28px; line-height: 1.05">Aramäer Ahlen</h1>
				<p class="login-sub">Schaltzentrale · Tabelle, Termine, Beteiligung, Kasse</p>

				<form style="margin-top: 20px" @submit.prevent="submit">
					<p v-if="error" class="form-error" role="alert">{{ error }}</p>
					<div class="field">
						<label for="email">E-Mail</label>
						<input id="email" v-model="email" type="email" autocomplete="email" required />
					</div>
					<div class="field">
						<label for="password">Passwort</label>
						<input id="password" v-model="password" type="password" autocomplete="current-password" required />
					</div>
					<button class="btn primary" style="width: 100%; justify-content: center" :disabled="busy">
						{{ busy ? 'Anmelden …' : 'Anmelden' }}
					</button>
				</form>
			</div>
		</div>
	</div>
</template>

<style scoped>
.login-wrap {
	min-height: 100vh;
	display: flex;
	align-items: center;
	justify-content: center;
	padding: 16px;
}
.login-card { width: 100%; max-width: 380px; }
.login-sub { color: var(--kreide-45); font-size: 13px; margin-top: 4px; }
</style>
