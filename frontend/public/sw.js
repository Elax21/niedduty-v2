/* Service Worker für Niedduty — zuständig allein für Push-Benachrichtigungen.
   Bewusst kein Offline-Cache: die App lebt von aktuellen Daten, ein veralteter
   Cache würde mehr schaden als nützen. */

self.addEventListener('install', () => self.skipWaiting());
self.addEventListener('activate', (e) => e.waitUntil(self.clients.claim()));

self.addEventListener('push', (event) => {
	let data = {};
	try {
		data = event.data ? event.data.json() : {};
	} catch {
		data = { title: 'Niedduty', body: event.data ? event.data.text() : '' };
	}
	const title = data.title || 'Niedduty';
	event.waitUntil(
		self.registration.showNotification(title, {
			body: data.body || '',
			icon: '/logo.png',
			badge: '/logo.png',
			tag: data.tag || 'niedduty',
			renotify: true,
			data: { url: data.url || '/' }
		})
	);
});

/* Tippen öffnet die App — vorhandenes Fenster bevorzugt. */
self.addEventListener('notificationclick', (event) => {
	event.notification.close();
	const target = (event.notification.data && event.notification.data.url) || '/';
	event.waitUntil(
		self.clients.matchAll({ type: 'window', includeUncontrolled: true }).then((list) => {
			for (const client of list) {
				if ('focus' in client) {
					client.navigate(target);
					return client.focus();
				}
			}
			return self.clients.openWindow(target);
		})
	);
});
