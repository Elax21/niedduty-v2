package config

import (
	"log"
	"os"
	"strings"
)

// Config hält alle Laufzeit-Einstellungen (aus Umgebungsvariablen).
type Config struct {
	DatabaseURL string
	ListenAddr  string
	// SecureCookies setzt das Secure-Flag am Session-Cookie. In Produktion
	// (HTTPS) zwingend, lokal über http:// nicht möglich.
	SecureCookies bool
	// TrustedProxies — Reverse-Proxy vor dem Server (z.B. "127.0.0.1").
	// Leer = keine Proxies vertrauen.
	TrustedProxies []string
	// Production schaltet gin in den Release-Modus.
	Production bool
}

func Load() Config {
	prod := boolEnv("PRODUCTION", false)
	cfg := Config{
		DatabaseURL:    env("DATABASE_URL", "postgres://root:Ow44qsth3xHL@localhost:5432/niedduty2"),
		ListenAddr:     env("LISTEN_ADDR", ":8080"),
		SecureCookies:  boolEnv("COOKIE_SECURE", prod),
		TrustedProxies: list(os.Getenv("TRUSTED_PROXIES")),
		Production:     prod,
	}
	if prod && os.Getenv("DATABASE_URL") == "" {
		log.Fatal("config: DATABASE_URL muss in Produktion gesetzt sein")
	}
	return cfg
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func boolEnv(key string, fallback bool) bool {
	switch strings.ToLower(os.Getenv(key)) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func list(v string) []string {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
