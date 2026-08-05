package fussball

import (
	"testing"
	"time"
)

// Echte Mannschafts-ID der ASG Aramäer Ahlen (Herren).
const testTeamID = "011MIC2EF8000000VTVG0001VTR8C1K7"

func TestCurrentSeason(t *testing.T) {
	cases := map[string]string{
		"2026-08-01": "2627",
		"2026-06-30": "2526",
		"2027-01-15": "2627",
	}
	for in, want := range cases {
		d, _ := time.Parse(dateOnly, in)
		if got := CurrentSeason(d); got != want {
			t.Errorf("CurrentSeason(%s) = %s, erwartet %s", in, got, want)
		}
	}
}

func TestSplitHeadline(t *testing.T) {
	date, tim, comp, iso := splitHeadline("Sonntag, 26.07.2026 - 13:00 Uhr | Kreisliga A")
	if date != "26.07.2026" || tim != "13:00" || comp != "Kreisliga A" || iso != "2026-07-26" {
		t.Fatalf("unerwartet: %q %q %q %q", date, tim, comp, iso)
	}
}

// Netz-Tests: nur mit `go test -tags=live` bzw. ohne -short sinnvoll.
func TestFetchSquadStatsLive(t *testing.T) {
	if testing.Short() {
		t.Skip("braucht Netzzugang zu fussball.de")
	}
	stats, err := FetchSquadStats(testTeamID, "2526")
	if err != nil {
		t.Fatal(err)
	}
	if len(stats) == 0 {
		t.Fatal("keine Spieler geliefert")
	}
	top := stats[0]
	if top.Name == "" || top.Matches == 0 {
		t.Fatalf("Kaderzeile nicht dekodiert: %+v", top)
	}
}

func TestFetchPrevGamesLive(t *testing.T) {
	if testing.Short() {
		t.Skip("braucht Netzzugang zu fussball.de")
	}
	games, err := FetchPrevGames(testTeamID)
	if err != nil {
		t.Fatal(err)
	}
	if len(games) == 0 {
		t.Fatal("keine Spiele geliefert")
	}
	var withResult int
	for _, g := range games {
		if g.HomeGoals != nil && g.GuestGoals != nil {
			withResult++
		}
		if g.Home == "" || g.Guest == "" {
			t.Fatalf("Mannschaftsname fehlt: %+v", g)
		}
	}
	if withResult == 0 {
		t.Fatal("kein einziges Ergebnis dekodiert")
	}
}
