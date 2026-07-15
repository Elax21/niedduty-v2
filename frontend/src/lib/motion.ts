import { animate, stagger } from 'animejs';

export const reducedMotion = () =>
	window.matchMedia('(prefers-reduced-motion: reduce)').matches;

/** Listenzeilen gestaffelt einblenden (Signature-Move beim Seitenwechsel). */
export function enterRows(selector: string) {
	if (reducedMotion()) return;
	animate(selector, {
		opacity: [0, 1],
		translateY: [8, 0],
		delay: stagger(40),
		duration: 260,
		ease: 'outQuad'
	});
}

/** Anzeigetafel-Count-up: ruft cb mit dem laufenden Wert auf. */
export function countUp(to: number, cb: (v: number) => void, duration = 900) {
	if (reducedMotion() || to === 0) {
		cb(to);
		return;
	}
	const obj = { v: 0 };
	animate(obj, {
		v: to,
		duration,
		ease: 'outExpo',
		onUpdate: () => cb(Math.round(obj.v))
	});
}

/** Kreide-Quote-Balken von links aufziehen. */
export function growBars(selector: string) {
	if (reducedMotion()) return;
	animate(selector, {
		scaleX: [0, 1],
		delay: stagger(30),
		duration: 500,
		ease: 'outQuart'
	});
}
