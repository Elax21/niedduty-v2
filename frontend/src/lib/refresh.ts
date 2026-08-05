// Als installierte App gibt es keine Browserleiste — also einen eigenen
// Auffrischen-Knopf. Jede Seite meldet hier ihre Ladefunktion an; der Knopf in
// der Top-Bar ruft alle angemeldeten auf.
import { ref, onMounted, onUnmounted } from 'vue';
import { api } from '../services/api';

type Loader = () => unknown | Promise<unknown>;

const loaders = new Set<Loader>();
export const refreshing = ref(false);

// ── Neue Version erkennen ───────────────────────────────────────
// Als installierte App wird die Seite nie neu geladen — ein neues Binary käme
// sonst erst an, wenn jemand die App vom Startbildschirm wirft. Deshalb merken
// wir uns die Build-Kennung vom Start und vergleichen sie.
let loadedVersion = '';

async function serverVersion(): Promise<string> {
	try {
		const { data } = await api.get<{ version: string }>('/version');
		return data.version;
	} catch {
		return '';
	}
}

/** Lädt die Seite neu, wenn auf dem Server ein neuer Build läuft. */
export async function reloadIfOutdated(): Promise<boolean> {
	const current = await serverVersion();
	if (!current) return false;
	if (!loadedVersion) {
		loadedVersion = current;
		return false;
	}
	if (current !== loadedVersion) {
		window.location.reload();
		return true;
	}
	return false;
}

/** Einmal beim App-Start: Kennung merken und regelmäßig nachsehen. */
export function watchVersion(everyMs = 15 * 60 * 1000) {
	void reloadIfOutdated();
	window.setInterval(() => void reloadIfOutdated(), everyMs);
	// Zurück aus dem Hintergrund ist der häufigste Moment für ein Update.
	document.addEventListener('visibilitychange', () => {
		if (document.visibilityState === 'visible') void reloadIfOutdated();
	});
}

/** In einer View aufrufen: die eigene load()-Funktion anmelden. */
export function useRefresh(load: Loader) {
	onMounted(() => loaders.add(load));
	onUnmounted(() => loaders.delete(load));
}

/** Lädt die Daten der offenen Seite neu. */
export async function refreshAll() {
	if (refreshing.value) return;
	refreshing.value = true;
	try {
		// Erst nachsehen, ob es eine neue Version gibt — dann lädt die Seite
		// ohnehin neu und die Loader wären umsonst.
		if (await reloadIfOutdated()) return;
		await Promise.all([...loaders].map((l) => Promise.resolve(l()).catch(() => null)));
	} finally {
		// Kurz stehen lassen, damit die Drehung sichtbar ist.
		setTimeout(() => (refreshing.value = false), 350);
	}
}
