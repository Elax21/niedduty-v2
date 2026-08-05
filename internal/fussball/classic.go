// Classic-Scraper für www.fussball.de (die „alten" ajax.team.*-Seiten).
//
// Warum zusätzlich zu den next.fussball.de-Widgets: die Widgets liefern nur
// Tabelle und eigene Spiele. Kaderstatistik (Einsätze/Einsatzminuten/Tore) und
// die Spiele *fremder* Mannschaften (Gegner-Form) gibt es nur auf den klassischen
// Seiten — und die funktionieren mit jeder beliebigen team-id.
//
// Verschleierung: dieselbe Font-Technik wie bei den Widgets, aber der Schlüssel
// steht hier pro Element in `data-obfuscation="<key>"`. Deshalb dekodieren wir
// hier schlüsselbezogen (siehe keyedDecoder).
package fussball

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

const (
	classicBase = "https://www.fussball.de"
	dateOnly    = "2006-01-02"
)

// ── Öffentliche Datentypen ──────────────────────────────────────────

// PlayerStat — eine Zeile der Kaderstatistik einer Saison.
type PlayerStat struct {
	Name       string `json:"name"`
	Matches    int    `json:"matches"`
	Minutes    int    `json:"minutes"`
	Goals      int    `json:"goals"`
	ProfileURL string `json:"profileUrl"`
}

// ClassicMatch — ein Spiel aus dem Spielplan einer beliebigen Mannschaft.
type ClassicMatch struct {
	ISODate     string `json:"isoDate"`
	Date        string `json:"date"` // "So, 26.07.26"
	Time        string `json:"time"`
	Competition string `json:"competition"`
	Home        string `json:"home"`
	Guest       string `json:"guest"`
	HomeTeamID  string `json:"homeTeamId"`
	GuestTeamID string `json:"guestTeamId"`
	HomeGoals   *int   `json:"homeGoals"`
	GuestGoals  *int   `json:"guestGoals"`
	Note        string `json:"note"` // z.B. "Absetzung", wenn kein Ergebnis
}

// ── HTML-Helfer ─────────────────────────────────────────────────────

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func hasClass(n *html.Node, class string) bool {
	for _, f := range strings.Fields(attr(n, "class")) {
		if f == class {
			return true
		}
	}
	return false
}

// findAll sammelt alle Element-Knoten, auf die pred zutrifft (Vorbestellung).
func findAll(n *html.Node, pred func(*html.Node) bool) []*html.Node {
	var out []*html.Node
	var walk func(*html.Node)
	walk = func(cur *html.Node) {
		if cur.Type == html.ElementNode && pred(cur) {
			out = append(out, cur)
		}
		for ch := cur.FirstChild; ch != nil; ch = ch.NextSibling {
			walk(ch)
		}
	}
	walk(n)
	return out
}

func firstMatch(n *html.Node, pred func(*html.Node) bool) *html.Node {
	if all := findAll(n, pred); len(all) > 0 {
		return all[0]
	}
	return nil
}

func tagged(tag string) func(*html.Node) bool {
	return func(n *html.Node) bool { return n.Data == tag }
}

func classed(tag, class string) func(*html.Node) bool {
	return func(n *html.Node) bool { return n.Data == tag && hasClass(n, class) }
}

// ── Schlüsselbezogener Decoder ──────────────────────────────────────

// keyedDecoder hält pro data-obfuscation-Schlüssel einen Font-Decoder.
// Fonts werden erst beim ersten Bedarf geladen; Fehler werden gemerkt (nil),
// damit ein kaputter Schlüssel nicht bei jedem Element neu geladen wird.
type keyedDecoder struct{ byKey map[string]*decoder }

func newKeyedDecoder() *keyedDecoder { return &keyedDecoder{byKey: map[string]*decoder{}} }

func (k *keyedDecoder) get(key string) *decoder {
	if key == "" {
		return nil
	}
	if d, ok := k.byKey[key]; ok {
		return d
	}
	d, err := newDecoder(key)
	if err != nil {
		d = nil
	}
	k.byKey[key] = d
	return d
}

// text liest den Textinhalt eines Knotens und dekodiert dabei jeden Teilbaum
// mit dem Schlüssel, der an ihm (oder einem Vorfahren) hängt.
func (k *keyedDecoder) text(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node, string)
	walk = func(cur *html.Node, key string) {
		if cur.Type == html.ElementNode {
			if v := attr(cur, "data-obfuscation"); v != "" {
				key = v
			}
		}
		if cur.Type == html.TextNode {
			if d := k.get(key); d != nil {
				b.WriteString(d.text(cur.Data))
			} else {
				b.WriteString(cur.Data)
			}
		}
		for ch := cur.FirstChild; ch != nil; ch = ch.NextSibling {
			walk(ch, key)
		}
	}
	walk(n, "")
	return strings.Join(strings.Fields(b.String()), " ")
}

func (k *keyedDecoder) intPtrOf(n *html.Node) *int {
	t := k.text(n)
	if t == "" {
		return nil
	}
	if v, err := strconv.Atoi(t); err == nil {
		return &v
	}
	return nil
}

func (k *keyedDecoder) intOf(n *html.Node) int {
	if v := k.intPtrOf(n); v != nil {
		return *v
	}
	return 0
}

// ── Kaderstatistik ──────────────────────────────────────────────────

// CurrentSeason liefert die fussball.de-Saisonkennung ("2627") zum Zeitpunkt t.
// Saisonwechsel ist der 1. Juli.
func CurrentSeason(t time.Time) string {
	y := t.Year()
	if int(t.Month()) < 7 {
		y--
	}
	return fmt.Sprintf("%02d%02d", y%100, (y+1)%100)
}

// FetchSquadStats holt die Kaderstatistik (Einsätze, Einsatzminuten, Tore)
// einer Mannschaft für eine Saison. Spielernamen sind verschleiert und werden
// über die Font dekodiert; die Zahlen liefert fussball.de im Klartext.
func FetchSquadStats(teamID, season string) ([]PlayerStat, error) {
	url := fmt.Sprintf("%s/ajax.team.squad/-/mode/PAGE/order-by/1/saison/%s/show-filter/true/team-id/%s",
		classicBase, season, teamID)
	body, err := fetch(url)
	if err != nil {
		return nil, err
	}
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	dec := newKeyedDecoder()

	table := firstMatch(doc, tagged("table"))
	if table == nil {
		return nil, fmt.Errorf("fussball.de: Kadertabelle nicht gefunden")
	}
	out := []PlayerStat{}
	for _, tr := range findAll(table, tagged("tr")) {
		cells := findAll(tr, tagged("td"))
		if len(cells) < 4 || !hasClass(cells[0], "column-player") {
			continue
		}
		name := ""
		if nameEl := firstMatch(cells[0], classed("div", "player-name")); nameEl != nil {
			name = dec.text(nameEl)
		}
		if name == "" {
			continue
		}
		profile := ""
		if a := firstMatch(cells[0], tagged("a")); a != nil {
			profile = attr(a, "href")
		}
		out = append(out, PlayerStat{
			Name:       name,
			Matches:    dec.intOf(cells[1]),
			Minutes:    dec.intOf(cells[2]),
			Goals:      dec.intOf(cells[3]),
			ProfileURL: profile,
		})
	}
	return out, nil
}

// ── Spielplan beliebiger Mannschaften ───────────────────────────────

// FetchPrevGames holt die gespielten Partien einer beliebigen Mannschaft
// (neueste zuerst) — Grundlage für die Formkurve des Gegners.
func FetchPrevGames(teamID string) ([]ClassicMatch, error) {
	return fetchClassicMatches(teamID, "prev")
}

// FetchNextGames holt die angesetzten Partien einer beliebigen Mannschaft.
func FetchNextGames(teamID string) ([]ClassicMatch, error) {
	return fetchClassicMatches(teamID, "next")
}

func fetchClassicMatches(teamID, mode string) ([]ClassicMatch, error) {
	url := fmt.Sprintf("%s/ajax.team.%s.games/-/mode/PAGE/show-token/false/team-id/%s",
		classicBase, mode, teamID)
	body, err := fetch(url)
	if err != nil {
		return nil, err
	}
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	dec := newKeyedDecoder()

	table := firstMatch(doc, tagged("table"))
	if table == nil {
		return nil, fmt.Errorf("fussball.de: Spielplan nicht gefunden")
	}

	// Der Spielplan ist eine flache Zeilenfolge: erst eine Kopfzeile mit
	// Datum/Wettbewerb, danach die Zeile mit den beiden Mannschaften.
	out := []ClassicMatch{}
	var headline string
	for _, tr := range findAll(table, tagged("tr")) {
		if hasClass(tr, "row-headline") {
			headline = dec.text(tr)
			continue
		}
		clubs := findAll(tr, classed("td", "column-club"))
		if len(clubs) < 2 {
			continue
		}
		m := ClassicMatch{
			Home:        clubName(dec, clubs[0]),
			Guest:       clubName(dec, clubs[1]),
			HomeTeamID:  clubTeamID(clubs[0]),
			GuestTeamID: clubTeamID(clubs[1]),
		}
		m.Date, m.Time, m.Competition, m.ISODate = splitHeadline(headline)
		if score := firstMatch(tr, classed("td", "column-score")); score != nil {
			left := firstMatch(score, classed("span", "score-left"))
			right := firstMatch(score, classed("span", "score-right"))
			if left != nil && right != nil {
				m.HomeGoals, m.GuestGoals = dec.intPtrOf(left), dec.intPtrOf(right)
			}
			if m.HomeGoals == nil || m.GuestGoals == nil {
				m.Note = dec.text(score) // z.B. "Absetzung"
			}
		}
		out = append(out, m)
	}
	return out, nil
}

func clubName(dec *keyedDecoder, td *html.Node) string {
	if el := firstMatch(td, classed("div", "club-name")); el != nil {
		return dec.text(el)
	}
	return dec.text(td)
}

// clubTeamID zieht die team-id aus dem Mannschafts-Link.
func clubTeamID(td *html.Node) string {
	a := firstMatch(td, tagged("a"))
	if a == nil {
		return ""
	}
	href := attr(a, "href")
	if i := strings.Index(href, "/team-id/"); i >= 0 {
		return strings.Trim(href[i+len("/team-id/"):], "/")
	}
	return ""
}

// splitHeadline zerlegt "Sonntag, 26.07.2026 - 13:00 Uhr | Kreisliga A".
func splitHeadline(s string) (date, tim, comp, iso string) {
	if i := strings.LastIndex(s, "|"); i >= 0 {
		comp = strings.TrimSpace(s[i+1:])
		s = s[:i]
	}
	iso = isoDate(s)
	if m := timeRe.FindStringSubmatch(s); m != nil {
		tim = m[1]
	}
	if m := dateRe.FindString(s); m != "" {
		date = m
	}
	return date, tim, comp, iso
}
