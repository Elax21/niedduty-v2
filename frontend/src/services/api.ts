import axios from 'axios';

// Alle Aufrufe laufen über den Vite-Proxy (/api → Go-Server :8080).
export const api = axios.create({ baseURL: '/api' });

export function apiError(e: unknown): string {
	if (axios.isAxiosError(e) && e.response?.data?.error) return e.response.data.error;
	return 'Unbekannter Fehler — bitte erneut versuchen';
}
