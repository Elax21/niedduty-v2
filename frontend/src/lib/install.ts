// „Als App aufs Handy" — sammelt alles, was der Install-Hinweis braucht.
//
// Chrome/Android feuert `beforeinstallprompt`; das Event wird abgefangen und
// aufgehoben, damit der Hinweis zum passenden Zeitpunkt kommt. iOS kennt das
// Event nicht — dort bleibt nur die Anleitung über das Teilen-Menü.
import { ref, computed } from 'vue';

type InstallEvent = Event & {
	prompt: () => Promise<void>;
	userChoice: Promise<{ outcome: 'accepted' | 'dismissed' }>;
};

// Der Hinweis kommt genau einmal von selbst (beim ersten Einloggen). Danach
// nur noch über den Menüpunkt „Auf den Startbildschirm".
const SEEN_KEY = 'ndt_install_seen';

const deferred = ref<InstallEvent | null>(null);
const hintOpen = ref(false);

/** Läuft die Seite schon als installierte App? */
export function isStandalone(): boolean {
	if (typeof window === 'undefined') return false;
	const iosStandalone = (window.navigator as Navigator & { standalone?: boolean }).standalone;
	return window.matchMedia('(display-mode: standalone)').matches || iosStandalone === true;
}

/** iOS zeigt keinen Install-Dialog — dort braucht es die Klick-Anleitung. */
export function isIos(): boolean {
	if (typeof window === 'undefined') return false;
	const ua = window.navigator.userAgent;
	return /iphone|ipad|ipod/i.test(ua) || (/Macintosh/.test(ua) && 'ontouchend' in window);
}

/** Chrome auf dem iPhone — anderes Teilen-Menü als Safari. */
export function isIosChrome(): boolean {
	if (typeof window === 'undefined') return false;
	return isIos() && /CriOS|FxiOS|EdgiOS/.test(window.navigator.userAgent);
}

/** Ein-Klick-Installation möglich (Chrome & Co.)? */
export const canPromptInstall = computed(() => deferred.value !== null);

export const installHintOpen = computed(() => hintOpen.value);

export function openInstallHint() {
	hintOpen.value = true;
}

export function closeInstallHint() {
	hintOpen.value = false;
}

function seen(): boolean {
	try {
		return localStorage.getItem(SEEN_KEY) === '1';
	} catch {
		return false;
	}
}

/** Öffnet den Hinweis einmalig und merkt sich das dauerhaft. */
function openOnce() {
	if (seen() || isStandalone()) return;
	try {
		localStorage.setItem(SEEN_KEY, '1');
	} catch {
		/* privater Modus — dann kommt er beim nächsten Mal noch einmal */
	}
	hintOpen.value = true;
}

/** Löst den echten Browser-Dialog aus. Gibt zurück, ob installiert wurde. */
export async function promptInstall(): Promise<boolean> {
	const ev = deferred.value;
	if (!ev) return false;
	await ev.prompt();
	const { outcome } = await ev.userChoice;
	deferred.value = null;
	return outcome === 'accepted';
}

/** Einmal beim App-Start aufrufen. Öffnet den Hinweis verzögert von selbst. */
export function watchInstall() {
	if (typeof window === 'undefined' || isStandalone()) return;

	window.addEventListener('beforeinstallprompt', (e) => {
		e.preventDefault();
		deferred.value = e as InstallEvent;
		window.setTimeout(openOnce, 2500);
	});

	window.addEventListener('appinstalled', () => {
		deferred.value = null;
		hintOpen.value = false;
	});

	// iOS meldet nichts — dort nach kurzer Zeit selbst anbieten.
	if (isIos()) window.setTimeout(openOnce, 3500);
}
