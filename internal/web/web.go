// Package web bettet das gebaute Frontend ins Server-Binary ein, damit in
// Produktion ein einziges Artefakt ausgeliefert wird (kein separater Webserver).
//
// `cd frontend && npm run build` schreibt nach internal/web/dist.
// Fehlt der Build, läuft der Server trotzdem — dann gibt es nur die API
// (so bleibt `go run ./cmd/server` neben dem Vite-Dev-Server nutzbar).
package web

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed all:dist
var embedded embed.FS

// Assets liefert das eingebettete Frontend oder nil, wenn nicht gebaut wurde.
func Assets() fs.FS {
	sub, err := fs.Sub(embedded, "dist")
	if err != nil {
		return nil
	}
	if _, err := fs.Stat(sub, "index.html"); err != nil {
		return nil
	}
	return sub
}

// Version — Fingerabdruck des ausgelieferten Frontends.
//
// Als installierte App wird die Seite nach dem Start nie wieder geladen; ein
// neues Binary käme also nie an. Die App fragt diesen Wert regelmäßig ab und
// lädt sich neu, sobald er sich ändert. index.html genügt als Grundlage: Vite
// hängt an jeden Build neue Dateinamen mit Hash hinein.
func Version() string {
	assets := Assets()
	if assets == nil {
		return "dev"
	}
	data, err := fs.ReadFile(assets, "index.html")
	if err != nil {
		return "dev"
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:6])
}

// Mount hängt das Frontend in den Router: echte Dateien direkt, alles andere
// auf index.html (History-Routing von vue-router). /api bleibt unberührt.
func Mount(r *gin.Engine) bool {
	assets := Assets()
	if assets == nil {
		return false
	}
	fileServer := http.FileServer(http.FS(assets))

	r.NoRoute(func(c *gin.Context) {
		p := strings.TrimPrefix(path.Clean(c.Request.URL.Path), "/")
		if strings.HasPrefix(p, "api/") || p == "api" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Unbekannter Endpunkt"})
			return
		}
		if p != "" {
			if f, err := assets.Open(p); err == nil {
				f.Close()
				// Gehashte Build-Dateien dürfen lange im Cache bleiben.
				if strings.HasPrefix(p, "assets/") {
					c.Header("Cache-Control", "public, max-age=31536000, immutable")
				}
				fileServer.ServeHTTP(c.Writer, c.Request)
				return
			}
		}
		// Unbekannter Pfad → App-Shell, der Router entscheidet.
		c.Header("Cache-Control", "no-cache")
		c.FileFromFS("/", http.FS(assets))
	})
	return true
}
