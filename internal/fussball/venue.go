package fussball

import (
	"fmt"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

// Spielstätte. Die Widgets liefern `venue: null`, die klassische Spielseite
// dagegen einen fertigen Ortslink:
//
//	<a href="https://www.google.de/maps?q=Im+Elsken+60%2C+59227+Ahlen" class="location">
//	  Kunstrasenplatz, Sportpark Nord Kunstrasen, Im Elsken 60, 59227 Ahlen
//	</a>
//
// Daraus holen wir Anzeigename und reine Anschrift — letztere taugt für
// Google- und Apple-Maps-Links gleichermaßen.

// Venue — Spielstätte eines Spiels.
type Venue struct {
	// Name — vollständige Beschreibung, z.B. "Kunstrasenplatz, Sportpark Nord …".
	Name string `json:"name"`
	// Address — nur die Anschrift, z.B. "Im Elsken 60, 59227 Ahlen".
	Address string `json:"address"`
}

// FetchVenue liest die Spielstätte von einer fussball.de-Spielseite.
// Fehlt die Angabe (unbekannter Platz), kommt eine leere Venue ohne Fehler
// zurück — ein Spiel ohne Adresse ist kein Grund, den Abruf scheitern zu lassen.
func FetchVenue(matchURL string) (Venue, error) {
	if !strings.HasPrefix(matchURL, classicBase+"/spiel/") {
		return Venue{}, fmt.Errorf("fussball.de: keine Spielseite")
	}
	body, err := fetch(matchURL)
	if err != nil {
		return Venue{}, err
	}
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return Venue{}, err
	}

	link := firstMatch(doc, classed("a", "location"))
	if link == nil {
		return Venue{}, nil
	}

	v := Venue{Name: strings.TrimSpace(textOf(link))}
	if q := mapsQuery(attr(link, "href")); q != "" {
		v.Address = q
	} else {
		v.Address = v.Name
	}
	return v, nil
}

// mapsQuery zieht die Anschrift aus dem `maps?q=…`-Link.
func mapsQuery(href string) string {
	u, err := url.Parse(href)
	if err != nil {
		return ""
	}
	return u.Query().Get("q")
}

// textOf sammelt den sichtbaren Text eines Knotens (hier unverschleiert).
func textOf(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(x *html.Node) {
		if x.Type == html.TextNode {
			b.WriteString(x.Data)
		}
		for c := x.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return strings.Join(strings.Fields(b.String()), " ")
}
