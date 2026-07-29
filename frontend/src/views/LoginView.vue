<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useAuthStore } from '../stores/auth';
import { apiError } from '../services/api';
import { animate, stagger } from 'animejs';
import { reducedMotion } from '../lib/motion';

const router = useRouter();
const auth = useAuthStore();

const login = ref('');
const password = ref('');
const error = ref('');
const busy = ref(false);

async function submit() {
	error.value = '';
	busy.value = true;
	try {
		await auth.login(login.value, password.value);
		router.push('/');
	} catch (e) {
		error.value = apiError(e);
	} finally {
		busy.value = false;
	}
}

onMounted(() => {
	if (reducedMotion()) return;
	animate('.login-anim', {
		opacity: [0, 1],
		translateY: [18, 0],
		delay: stagger(90, { start: 80 }),
		duration: 460,
		ease: 'outQuad'
	});
});
</script>

<template>
	<div class="login-wrap">
		<div class="login-inner">
			<div class="login-anim crest-hero">
				<img src="/logo.png" alt="Vereinswappen ASG Aramäer Ahlen" />
			</div>
			<h1 class="login-anim brand">Aramäer<br />Ahlen</h1>
			<p class="login-anim tagline">Kabine · Tabelle · Termine · Kasse</p>

			<form class="login-anim login-form" @submit.prevent="submit">
				<p v-if="error" class="form-error" role="alert">{{ error }}</p>
				<div class="field">
					<label for="login">Alias</label>
					<input id="login" v-model="login" autocapitalize="none" autocomplete="username" placeholder="dein Login-Name" required />
				</div>
				<div class="field">
					<label for="password">Passwort</label>
					<input id="password" v-model="password" type="password" autocomplete="current-password" required />
				</div>
				<button class="btn primary block" style="margin-top: 4px" :disabled="busy">
					{{ busy ? 'Anmelden …' : 'Einloggen' }}
				</button>
			</form>
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
.login-inner { width: 100%; max-width: 380px; text-align: center; }
.crest-hero {
	width: 92px; height: 92px;
	margin: 0 auto 18px;
	border-radius: 24px;
	overflow: hidden;
	background: #fff;
	box-shadow: 0 0 0 3px var(--gold-deep), 0 18px 40px rgba(0, 0, 0, 0.55);
}
.crest-hero img { width: 100%; height: 100%; object-fit: cover; }
.brand { font-size: 46px; letter-spacing: 0.01em; }
.brand :deep(br) { display: block; }
.tagline {
	color: var(--ink-3);
	font-family: var(--font-display);
	text-transform: uppercase;
	letter-spacing: 0.08em;
	font-size: 12.5px;
	margin-top: 6px;
}
.login-form { margin-top: 26px; text-align: left; }
</style>
