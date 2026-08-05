package fussball

import "testing"

// Live-Test: prüft, dass die Spielstätte von einer echten Spielseite kommt.
// Braucht Netz — mit -short übersprungen (wie die übrigen fussball.de-Tests).
func TestFetchVenueLive(t *testing.T) {
	if testing.Short() {
		t.Skip("braucht Netz")
	}
	v, err := FetchVenue("https://www.fussball.de/spiel/aramaeer-ahlen-ahlener-sg-ii/-/spiel/031BG5CI64000000VS5489BUVUR5FS5A")
	if err != nil {
		t.Fatalf("FetchVenue: %v", err)
	}
	if v.Address == "" || v.Name == "" {
		t.Fatalf("leere Spielstätte: %+v", v)
	}
	t.Logf("Name=%q Adresse=%q", v.Name, v.Address)
}
