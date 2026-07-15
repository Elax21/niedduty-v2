// WhatsApp-Status-Bild der Mannschaftskasse: 1080×1920 Canvas in
// Vereinsfarben, geteilt via navigator.share (mobil) oder Download.

export interface ShareRow {
	name: string;
	open: number; // Cent
	count: number;
}

export interface ShareData {
	clubName: string;
	rows: ShareRow[];
	totalOpen: number;
	iban?: string;
	inhaber?: string;
}

const GOLD = '#f0a81c';
const ROT = '#c0272d';
const KREIDE = '#f2efe4';
const BG = '#16110c';
const BG_CARD = '#1d1710';

function euro(cents: number) {
	return (cents / 100).toLocaleString('de-DE', { style: 'currency', currency: 'EUR' });
}

async function loadLogo(): Promise<HTMLImageElement | null> {
	return new Promise((resolve) => {
		const img = new Image();
		img.onload = () => resolve(img);
		img.onerror = () => resolve(null);
		img.src = '/logo.png';
	});
}

export async function buildKasseImage(data: ShareData): Promise<Blob | null> {
	const W = 1080;
	const H = 1920;
	const canvas = document.createElement('canvas');
	canvas.width = W;
	canvas.height = H;
	const ctx = canvas.getContext('2d');
	if (!ctx) return null;

	// Hintergrund + Flutlicht-Glow
	ctx.fillStyle = BG;
	ctx.fillRect(0, 0, W, H);
	const glow = ctx.createRadialGradient(W / 2, -200, 100, W / 2, -200, 1100);
	glow.addColorStop(0, 'rgba(240, 168, 28, 0.22)');
	glow.addColorStop(1, 'rgba(240, 168, 28, 0)');
	ctx.fillStyle = glow;
	ctx.fillRect(0, 0, W, H);

	// Kopf: Logo + Vereinsname
	const logo = await loadLogo();
	if (logo) {
		const s = 150;
		ctx.save();
		ctx.beginPath();
		ctx.roundRect(W / 2 - s / 2, 90, s, s, 24);
		ctx.clip();
		ctx.fillStyle = '#fff';
		ctx.fillRect(W / 2 - s / 2, 90, s, s);
		ctx.drawImage(logo, W / 2 - s / 2, 90, s, s);
		ctx.restore();
	}
	ctx.textAlign = 'center';
	ctx.fillStyle = KREIDE;
	ctx.font = '700 64px "Saira Condensed", sans-serif';
	ctx.fillText(data.clubName.toUpperCase(), W / 2, 330);
	ctx.fillStyle = GOLD;
	ctx.font = '600 44px "Saira Condensed", sans-serif';
	ctx.fillText('MANNSCHAFTSKASSE', W / 2, 395);
	ctx.fillStyle = 'rgba(242, 239, 228, 0.5)';
	ctx.font = '400 30px Saira, sans-serif';
	ctx.fillText(
		new Date().toLocaleDateString('de-DE', { day: '2-digit', month: 'long', year: 'numeric' }),
		W / 2, 445
	);

	// Rote Linie
	ctx.fillStyle = ROT;
	ctx.fillRect(W / 2 - 60, 480, 120, 6);

	// Spieler-Zeilen (max. 14, sortiert kommt vom Aufrufer)
	const rows = data.rows.slice(0, 14);
	const startY = 560;
	const rowH = 78;
	ctx.font = '400 36px Saira, sans-serif';
	rows.forEach((r, i) => {
		const y = startY + i * rowH;
		// Zeilen-Karte
		ctx.fillStyle = i % 2 === 0 ? BG_CARD : 'rgba(240, 168, 28, 0.05)';
		ctx.beginPath();
		ctx.roundRect(70, y, W - 140, rowH - 12, 10);
		ctx.fill();

		ctx.textAlign = 'left';
		ctx.fillStyle = KREIDE;
		ctx.font = '500 38px Saira, sans-serif';
		ctx.fillText(r.name, 100, y + 45);

		ctx.textAlign = 'right';
		ctx.fillStyle = GOLD;
		ctx.font = '600 40px "Chivo Mono", monospace';
		ctx.fillText(euro(r.open), W - 100, y + 46);
	});
	if (data.rows.length > 14) {
		ctx.textAlign = 'center';
		ctx.fillStyle = 'rgba(242, 239, 228, 0.5)';
		ctx.font = '400 30px Saira, sans-serif';
		ctx.fillText(`… und ${data.rows.length - 14} weitere`, W / 2, startY + 14 * rowH + 20);
	}

	// Summe
	const sumY = Math.min(startY + rows.length * rowH + 90, H - 340);
	ctx.textAlign = 'center';
	ctx.fillStyle = 'rgba(242, 239, 228, 0.6)';
	ctx.font = '600 34px "Saira Condensed", sans-serif';
	ctx.fillText('OFFEN GESAMT', W / 2, sumY);
	ctx.fillStyle = ROT;
	ctx.font = '700 96px "Chivo Mono", monospace';
	ctx.fillText(euro(data.totalOpen), W / 2, sumY + 105);

	// Fußzeile: IBAN
	if (data.iban) {
		ctx.fillStyle = 'rgba(242, 239, 228, 0.55)';
		ctx.font = '400 30px Saira, sans-serif';
		ctx.fillText('Überweisung an:', W / 2, H - 190);
		ctx.fillStyle = KREIDE;
		ctx.font = '500 34px "Chivo Mono", monospace';
		ctx.fillText(data.iban, W / 2, H - 140);
		if (data.inhaber) {
			ctx.fillStyle = 'rgba(242, 239, 228, 0.55)';
			ctx.font = '400 30px Saira, sans-serif';
			ctx.fillText(data.inhaber, W / 2, H - 95);
		}
	}
	ctx.fillStyle = 'rgba(240, 168, 28, 0.4)';
	ctx.font = '600 26px "Saira Condensed", sans-serif';
	ctx.fillText('NIEDDUTY', W / 2, H - 40);

	return new Promise((resolve) => canvas.toBlob((b) => resolve(b), 'image/png'));
}

// Teilen (mobil: WhatsApp-Status via Share-Sheet) mit Download-Fallback.
export async function shareKasseImage(data: ShareData): Promise<'shared' | 'downloaded' | 'failed'> {
	const blob = await buildKasseImage(data);
	if (!blob) return 'failed';
	const file = new File([blob], 'mannschaftskasse.png', { type: 'image/png' });
	if (navigator.canShare?.({ files: [file] })) {
		try {
			await navigator.share({ files: [file], title: 'Mannschaftskasse' });
			return 'shared';
		} catch {
			// abgebrochen → Fallback Download
		}
	}
	const url = URL.createObjectURL(blob);
	const a = document.createElement('a');
	a.href = url;
	a.download = 'mannschaftskasse.png';
	a.click();
	URL.revokeObjectURL(url);
	return 'downloaded';
}
