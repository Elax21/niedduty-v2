// Navigations-Links. Apple Maps öffnet auf dem iPhone die vorinstallierte
// Karten-App, überall sonst ist Google Maps die sichere Wahl.
import { isIos } from './install';

export function googleMapsUrl(address: string): string {
	return `https://www.google.com/maps/dir/?api=1&destination=${encodeURIComponent(address)}`;
}

export function appleMapsUrl(address: string): string {
	return `https://maps.apple.com/?daddr=${encodeURIComponent(address)}&dirflg=d`;
}

/** Zeigt die Apple-Maps-Schaltfläche nur dort, wo sie auch etwas bewirkt. */
export function showAppleMaps(): boolean {
	return isIos();
}

/** Treffpunkt: fest 1:30 vor Anpfiff. */
export const MEET_BEFORE_MATCH_MIN = 90;

/** "13:00" minus Vorlauf → "11:30". Leere Eingabe bleibt leer. */
export function meetingTime(kickoff: string, leadMin = MEET_BEFORE_MATCH_MIN): string {
	const m = /^(\d{1,2}):(\d{2})$/.exec(kickoff.trim());
	if (!m) return '';
	const total = Number(m[1]) * 60 + Number(m[2]) - leadMin;
	if (total < 0) return '';
	return `${String(Math.floor(total / 60)).padStart(2, '0')}:${String(total % 60).padStart(2, '0')}`;
}
