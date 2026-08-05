/* Web-Push im Browser: Service Worker registrieren, Abo beim Server anmelden.
   Der Server verschickt die Erinnerungen selbst (VAPID), kein Fremddienst. */

import { api } from '../services/api';

/** Kann dieses Gerät überhaupt Push? (iOS erst ab „Zum Home-Bildschirm".) */
export function pushSupported(): boolean {
	return 'serviceWorker' in navigator && 'PushManager' in window && 'Notification' in window;
}

/** Aktueller Erlaubnis-Status des Browsers. */
export function pushPermission(): NotificationPermission {
	return pushSupported() ? Notification.permission : 'denied';
}

async function registration(): Promise<ServiceWorkerRegistration> {
	const reg = await navigator.serviceWorker.register('/sw.js');
	await navigator.serviceWorker.ready;
	return reg;
}

/** Ist dieses Gerät bereits angemeldet? */
export async function pushActive(): Promise<boolean> {
	if (!pushSupported() || Notification.permission !== 'granted') return false;
	const reg = await navigator.serviceWorker.getRegistration();
	if (!reg) return false;
	return !!(await reg.pushManager.getSubscription());
}

/** Benachrichtigungen einschalten. Wirft mit deutscher Meldung, wenn es scheitert. */
export async function pushSubscribe(): Promise<void> {
	if (!pushSupported()) {
		throw new Error('Dieses Gerät unterstützt keine Benachrichtigungen. Auf dem iPhone: Seite über „Teilen → Zum Home-Bildschirm" öffnen.');
	}
	const permission = await Notification.requestPermission();
	if (permission !== 'granted') {
		throw new Error('Benachrichtigungen wurden im Browser abgelehnt.');
	}
	const { data } = await api.get<{ publicKey: string }>('/push/key');
	const reg = await registration();
	const sub =
		(await reg.pushManager.getSubscription()) ??
		(await reg.pushManager.subscribe({
			userVisibleOnly: true,
			applicationServerKey: urlBase64ToUint8Array(data.publicKey)
		}));
	await api.post('/push/subscribe', sub.toJSON());
}

/** Benachrichtigungen für dieses Gerät wieder abschalten. */
export async function pushUnsubscribe(): Promise<void> {
	const reg = await navigator.serviceWorker.getRegistration();
	const sub = await reg?.pushManager.getSubscription();
	if (!sub) return;
	await api.post('/push/unsubscribe', { endpoint: sub.endpoint }).catch(() => undefined);
	await sub.unsubscribe();
}

/** Probenachricht an die eigenen Geräte. */
export async function pushTest(): Promise<number> {
	const { data } = await api.post<{ sent: number }>('/push/test');
	return data.sent;
}

/** Der VAPID-Schlüssel kommt base64url-kodiert; die Push-API will Bytes. */
function urlBase64ToUint8Array(base64: string): Uint8Array<ArrayBuffer> {
	const padded = (base64 + '='.repeat((4 - (base64.length % 4)) % 4)).replace(/-/g, '+').replace(/_/g, '/');
	const raw = atob(padded);
	const out = new Uint8Array(new ArrayBuffer(raw.length));
	for (let i = 0; i < raw.length; i++) out[i] = raw.charCodeAt(i);
	return out;
}
