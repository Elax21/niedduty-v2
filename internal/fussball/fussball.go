// Package fussball holt Tabellen- und Spieldaten von next.fussball.de und
// dekodiert die per Custom-Font verschleierten Texte (Anti-Scraping).
//
// Funktionsweise: Die Widget-Seiten (https://next.fussball.de/widget/<type>/<id>)
// sind Next.js-Seiten mit den Daten im <script id="__NEXT_DATA__">-JSON. Zahlen
// und Namen sind mit Private-Use-Codepoints (U+E000–U+F8FF) verschleiert; die
// zugehörige Font (Schlüssel in pageProps.obfuscatedFont) mappt jeden Codepoint
// auf eine Glyphe mit echtem Namen (z.B. U+F04E → "zero"). Wir lesen die
// Glyphennamen aus der Font und übersetzen sie zurück in echte Zeichen.
package fussball

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/image/font/sfnt"
)

const (
	widgetBase = "https://next.fussball.de/widget"
	fontBase   = "https://www.fussball.de/export.fontface/-/format/ttf/id/%s/type/font"
	referer    = "https://niedduty.com/"
	userAgent  = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36"
)

var httpClient = &http.Client{Timeout: 20 * time.Second}

var nextDataRe = regexp.MustCompile(`(?s)id="__NEXT_DATA__"[^>]*>(.*?)</script>`)

// ── Öffentliche Datentypen ──────────────────────────────────────────

type TableEntry struct {
	Position     int    `json:"position"`
	TeamName     string `json:"teamName"`
	Played       int    `json:"played"`
	Won          int    `json:"won"`
	Drawn        int    `json:"drawn"`
	Lost         int    `json:"lost"`
	GoalsFor     int    `json:"goalsFor"`
	GoalsAgainst int    `json:"goalsAgainst"`
	Points       int    `json:"points"`
	IsOwn        bool   `json:"isOwn"`
	LogoURL      string `json:"logoUrl"`
}

type MatchTeam struct {
	Name    string `json:"name"`
	LogoURL string `json:"logoUrl"`
	IsOwn   bool   `json:"isOwn"`
}

type Match struct {
	ID         string    `json:"id"`
	Date       string    `json:"date"`    // z.B. "Sonntag, 16.08.2026"
	ISODate    string    `json:"isoDate"` // "2026-08-16" (für Termin-Einsortierung)
	Time       string    `json:"time"`    // "13:00"
	URL        string    `json:"url"`     // fussball.de-Spielseite
	Home       MatchTeam `json:"home"`
	Guest      MatchTeam `json:"guest"`
	HomeGoals  *int      `json:"homeGoals"`
	GuestGoals *int      `json:"guestGoals"`
	Played     bool      `json:"played"`
}

type Matches struct {
	Previous []Match `json:"previous"`
	Next     []Match `json:"next"`
}

// ── Fetch + Extraktion ──────────────────────────────────────────────

func fetch(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Referer", referer)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fussball.de: HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func nextData(widgetType, id string) (json.RawMessage, error) {
	body, err := fetch(fmt.Sprintf("%s/%s/%s", widgetBase, widgetType, id))
	if err != nil {
		return nil, err
	}
	m := nextDataRe.FindSubmatch(body)
	if m == nil {
		return nil, fmt.Errorf("fussball.de: __NEXT_DATA__ nicht gefunden")
	}
	return m[1], nil
}

// ── Decoder (Font-Glyphennamen → echte Zeichen) ─────────────────────

type decoder struct {
	font  *sfnt.Font
	buf   sfnt.Buffer
	cache map[rune]rune
}

func newDecoder(fontKey string) (*decoder, error) {
	body, err := fetch(fmt.Sprintf(fontBase, fontKey))
	if err != nil {
		return nil, err
	}
	f, err := sfnt.Parse(body)
	if err != nil {
		return nil, fmt.Errorf("font parse: %w", err)
	}
	return &decoder{font: f, cache: map[rune]rune{}}, nil
}

func (d *decoder) rune(r rune) rune {
	if r < 0x2000 { // normale Zeichen unverändert
		return r
	}
	if c, ok := d.cache[r]; ok {
		return c
	}
	c := r
	if idx, err := d.font.GlyphIndex(&d.buf, r); err == nil && idx != 0 {
		if name, err := d.font.GlyphName(&d.buf, idx); err == nil && name != "" {
			c = glyphToRune(name)
		}
	}
	d.cache[r] = c
	return c
}

func (d *decoder) text(s string) string {
	if s == "" {
		return s
	}
	var b strings.Builder
	for _, r := range s {
		b.WriteRune(d.rune(r))
	}
	return strings.TrimSpace(b.String())
}

func (d *decoder) intVal(s string) int {
	t := strings.TrimSpace(d.text(s))
	if t == "" {
		return 0
	}
	n, _ := strconv.Atoi(t)
	return n
}

func (d *decoder) intPtr(s string) *int {
	t := strings.TrimSpace(d.text(s))
	if t == "" {
		return nil
	}
	if n, err := strconv.Atoi(t); err == nil {
		return &n
	}
	return nil
}

// glyphToRune übersetzt einen Adobe-Glyphennamen in sein Zeichen.
func glyphToRune(name string) rune {
	if r, ok := glyphMap[name]; ok {
		return r
	}
	if rs := []rune(name); len(rs) == 1 { // z.B. "b", "E", "O"
		return rs[0]
	}
	if strings.HasPrefix(name, "uni") && len(name) == 7 {
		if v, err := strconv.ParseInt(name[3:], 16, 32); err == nil {
			return rune(v)
		}
	}
	return '?'
}

var glyphMap = map[string]rune{
	"zero": '0', "one": '1', "two": '2', "three": '3', "four": '4',
	"five": '5', "six": '6', "seven": '7', "eight": '8', "nine": '9',
	"space": ' ', "colon": ':', "period": '.', "comma": ',', "hyphen": '-',
	"slash": '/', "backslash": '\\', "ampersand": '&', "plus": '+', "minus": '-',
	"parenleft": '(', "parenright": ')', "bracketleft": '[', "bracketright": ']',
	"quoteright": '\'', "quoteleft": '\'', "quotesingle": '\'', "quotedbl": '"',
	"endash": '-', "emdash": '-', "underscore": '_', "numbersign": '#',
	"asterisk": '*', "percent": '%', "exclam": '!', "question": '?',
	"periodcentered": '·', "bullet": '•', "at": '@', "degree": '°',
	// Deutsch
	"adieresis": 'ä', "odieresis": 'ö', "udieresis": 'ü',
	"Adieresis": 'Ä', "Odieresis": 'Ö', "Udieresis": 'Ü', "germandbls": 'ß',
	// häufige Akzente (Vereinsnamen)
	"aacute": 'á', "eacute": 'é', "iacute": 'í', "oacute": 'ó', "uacute": 'ú',
	"agrave": 'à', "egrave": 'è', "igrave": 'ì', "ograve": 'ò', "ugrave": 'ù',
	"acircumflex": 'â', "ecircumflex": 'ê', "icircumflex": 'î', "ocircumflex": 'ô', "ucircumflex": 'û',
	"Aacute": 'Á', "Eacute": 'É', "Iacute": 'Í', "Oacute": 'Ó', "Uacute": 'Ú',
	"Egrave": 'È', "Agrave": 'À', "ccedilla": 'ç', "Ccedilla": 'Ç',
	"ntilde": 'ñ', "Ntilde": 'Ñ', "atilde": 'ã', "otilde": 'õ',
	"scaron": 'š', "zcaron": 'ž', "cacute": 'ć', "lslash": 'ł',
	"oslash": 'ø', "aring": 'å', "ae": 'æ',
}

// ── Öffentliche Fetch-Funktionen ────────────────────────────────────

// FetchTable holt die Ligatabelle. ownName markiert das eigene Team.
func FetchTable(tableID, ownName string) ([]TableEntry, error) {
	raw, err := nextData("table", tableID)
	if err != nil {
		return nil, err
	}
	var pd struct {
		Props struct {
			PageProps struct {
				ObfuscatedFont string `json:"obfuscatedFont"`
				Table          struct {
					Entries []struct {
						TeamName     string `json:"teamName"`
						Position     string `json:"position"`
						Matches      string `json:"matches"`
						MatchesWon   string `json:"matchesWon"`
						MatchesDrawn string `json:"matchesDrawn"`
						MatchesLost  string `json:"matchesLost"`
						GoalsPlus    string `json:"goalsPlus"`
						GoalsMinus   string `json:"goalsMinus"`
						Points       string `json:"points"`
						ClubLogoURL  string `json:"clubLogoURL"`
					} `json:"entries"`
				} `json:"table"`
			} `json:"pageProps"`
		} `json:"props"`
	}
	if err := json.Unmarshal(raw, &pd); err != nil {
		return nil, err
	}
	dec, err := newDecoder(pd.Props.PageProps.ObfuscatedFont)
	if err != nil {
		return nil, err
	}
	own := normalize(ownName)
	var out []TableEntry
	for _, e := range pd.Props.PageProps.Table.Entries {
		name := dec.text(e.TeamName)
		if name == "" {
			continue
		}
		out = append(out, TableEntry{
			Position:     dec.intVal(e.Position),
			TeamName:     name,
			Played:       dec.intVal(e.Matches),
			Won:          dec.intVal(e.MatchesWon),
			Drawn:        dec.intVal(e.MatchesDrawn),
			Lost:         dec.intVal(e.MatchesLost),
			GoalsFor:     dec.intVal(e.GoalsPlus),
			GoalsAgainst: dec.intVal(e.GoalsMinus),
			Points:       dec.intVal(e.Points),
			IsOwn:        own != "" && strings.Contains(normalize(name), own),
			LogoURL:      e.ClubLogoURL,
		})
	}
	return out, nil
}

// FetchMatches holt letzte + kommende Spiele der Mannschaft.
func FetchMatches(matchesID, ownName string) (*Matches, error) {
	raw, err := nextData("team-matches", matchesID)
	if err != nil {
		return nil, err
	}
	var pd struct {
		Props struct {
			PageProps struct {
				ObfuscatedFont  string     `json:"obfuscatedFont"`
				PreviousMatches []rawMatch `json:"previousMatches"`
				NextMatches     []rawMatch `json:"nextMatches"`
			} `json:"pageProps"`
		} `json:"props"`
	}
	if err := json.Unmarshal(raw, &pd); err != nil {
		return nil, err
	}
	dec, err := newDecoder(pd.Props.PageProps.ObfuscatedFont)
	if err != nil {
		return nil, err
	}
	own := normalize(ownName)
	return &Matches{
		Previous: decodeMatches(dec, own, pd.Props.PageProps.PreviousMatches),
		Next:     decodeMatches(dec, own, pd.Props.PageProps.NextMatches),
	}, nil
}

type rawMatch struct {
	ID      string `json:"id"`
	Kickoff struct {
		DateWithWeekday string `json:"dateWithWeekday"`
		Time            string `json:"time"`
	} `json:"kickoff"`
	HomeTeam  rawTeam `json:"homeTeam"`
	GuestTeam rawTeam `json:"guestTeam"`
	Result    struct {
		HomeResult  string `json:"homeResult"`
		GuestResult string `json:"guestResult"`
	} `json:"result"`
}

type rawTeam struct {
	Name        string `json:"name"`
	ClubLogoURL string `json:"clubLogoURL"`
}

func decodeMatches(dec *decoder, own string, raw []rawMatch) []Match {
	out := make([]Match, 0, len(raw))
	for _, m := range raw {
		home := dec.text(m.HomeTeam.Name)
		guest := dec.text(m.GuestTeam.Name)
		if home == "" && guest == "" {
			continue
		}
		hg := dec.intPtr(m.Result.HomeResult)
		gg := dec.intPtr(m.Result.GuestResult)
		date := dec.text(m.Kickoff.DateWithWeekday)
		url := ""
		if m.ID != "" {
			url = matchURLBase + m.ID
		}
		out = append(out, Match{
			ID:         m.ID,
			Date:       date,
			ISODate:    isoDate(date),
			Time:       dec.text(m.Kickoff.Time),
			URL:        url,
			Home:       MatchTeam{Name: home, LogoURL: m.HomeTeam.ClubLogoURL, IsOwn: own != "" && strings.Contains(normalize(home), own)},
			Guest:      MatchTeam{Name: guest, LogoURL: m.GuestTeam.ClubLogoURL, IsOwn: own != "" && strings.Contains(normalize(guest), own)},
			HomeGoals:  hg,
			GuestGoals: gg,
			Played:     hg != nil && gg != nil,
		})
	}
	return out
}

func normalize(s string) string {
	return strings.TrimSpace(strings.ToLower(s))
}

const matchURLBase = "https://www.fussball.de/spiel/-/spiel/"

var dateRe = regexp.MustCompile(`(\d{2})\.(\d{2})\.(\d{4})`)

// isoDate zieht aus "Sonntag, 16.08.2026" das "2026-08-16".
func isoDate(s string) string {
	m := dateRe.FindStringSubmatch(s)
	if m == nil {
		return ""
	}
	return m[3] + "-" + m[2] + "-" + m[1]
}
