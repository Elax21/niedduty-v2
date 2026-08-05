<script setup lang="ts">
/* Steckbrief zum nächsten Pflichtspiel: wer, wo, wie stark.
   Alle Daten kommen aus /fussball/scouting — kein Umweg über fussball.de. */
import { computed } from 'vue';
import { ExternalLink } from 'lucide-vue-next';
import type { Scouting } from '../types';

const props = defineProps<{ scouting: Scouting }>();

const match = computed(() => props.scouting.match);
const opp = computed(() => props.scouting.opponent);
const ownForm = computed(() => props.scouting.ownForm ?? []);

const platzText = computed(() => {
	const o = opp.value;
	if (!o?.inTable) return 'Noch kein Tabellenplatz';
	return `${o.position}. Platz · ${o.played} Spiele`;
});
const tordiff = computed(() => {
	const o = opp.value;
	if (!o) return '0';
	const d = o.goalsFor - o.goalsAgainst;
	return d > 0 ? `+${d}` : String(d);
});
</script>

<template>
	<section v-if="match && opp" class="card">
		<div class="card-head">
			<h2>{{ scouting.atHome ? 'Heimspiel gegen' : 'Auswärts bei' }}</h2>
			<a v-if="match.url" :href="match.url" target="_blank" rel="noopener" class="meta">
				fussball.de <ExternalLink :size="12" style="vertical-align: -1px" />
			</a>
		</div>
		<div class="card-body oppcard">
			<div class="opphead">
				<img v-if="opp.logoUrl" class="logo" :src="opp.logoUrl" :alt="'Wappen ' + opp.name" loading="lazy" />
				<div style="min-width: 0">
					<div class="nm">{{ opp.name }}</div>
					<div class="pl">{{ platzText }}</div>
				</div>
			</div>

			<p class="oppsum">{{ opp.summary }}</p>

			<div v-if="opp.inTable" class="oppgrid">
				<div class="c">
					<div class="v" style="color: var(--gold)">{{ opp.points }}</div>
					<div class="k">Punkte</div>
				</div>
				<div class="c">
					<div class="v">{{ opp.goalsFor }}:{{ opp.goalsAgainst }}</div>
					<div class="k">Tore</div>
				</div>
				<div class="c">
					<div class="v" :style="{ color: opp.goalsFor >= opp.goalsAgainst ? 'var(--gruen)' : 'var(--bad)' }">{{ tordiff }}</div>
					<div class="k">Diff</div>
				</div>
			</div>

			<div v-if="opp.form.length" class="formrow" style="margin-top: 14px">
				<span class="overline" style="margin-right: 4px">Gegner</span>
				<span v-for="(f, i) in opp.form" :key="i" class="formchip" :class="f.result" :title="`${f.score} gegen ${f.opponent}`">
					{{ f.result }}
				</span>
			</div>
			<div v-if="ownForm.length" class="formrow" style="margin-top: 8px">
				<span class="overline" style="margin-right: 4px">Wir</span>
				<span v-for="(f, i) in ownForm" :key="i" class="formchip" :class="f.result" :title="`${f.score} gegen ${f.opponent}`">
					{{ f.result }}
				</span>
			</div>

			<div v-if="opp.meetings.length" class="meetings">
				<div class="overline">Zuletzt gegeneinander</div>
				<div v-for="(m, i) in opp.meetings" :key="i" class="meet">
					<span class="mono dt">{{ m.date.split('-').reverse().join('.') }}</span>
					<span class="grow">{{ m.home ? 'Heim' : 'Auswärts' }}</span>
					<span v-if="m.score" class="mono sc" :class="m.result">{{ m.score }}</span>
					<span v-else class="sc dim">{{ m.note || '–' }}</span>
				</div>
			</div>
		</div>
	</section>
</template>

<style scoped>
.meetings { margin-top: 16px; padding-top: 13px; border-top: 1px solid var(--line); }
.meet {
	display: flex;
	align-items: center;
	gap: 10px;
	font-size: 13px;
	padding: 7px 0;
}
.meet .dt { font-size: 12px; color: var(--ink-3); }
.meet .grow { flex: 1; color: var(--ink-2); }
.meet .sc { font-weight: 700; font-size: 13.5px; }
.meet .sc.S { color: var(--gruen); }
.meet .sc.N { color: var(--bad); }
.meet .sc.U { color: var(--warn); }
.meet .sc.dim { color: var(--ink-3); font-weight: 400; }
</style>
