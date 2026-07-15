package config

import "os"

// Config hält alle Laufzeit-Einstellungen (aus Umgebungsvariablen).
type Config struct {
	DatabaseURL string
	ListenAddr  string
}

func Load() Config {
	return Config{
		DatabaseURL: env("DATABASE_URL", "postgres://root:Ow44qsth3xHL@localhost:5432/niedduty2"),
		ListenAddr:  env("LISTEN_ADDR", ":8080"),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
