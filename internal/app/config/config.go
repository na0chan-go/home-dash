package config

import (
	"os"
	"strings"
)

type Config struct {
	AppAddr             string
	DBPath              string
	GarbageSchedulePath string
	CORSAllowOrigins    []string
}

func LoadFromEnv() Config {
	return Config{
		AppAddr:             getEnv("APP_ADDR", ":8080"),
		DBPath:              getEnv("DB_PATH", "/data/app.db"),
		GarbageSchedulePath: getEnv("GARBAGE_SCHEDULE_PATH", "config/garbage_schedule.json"),
		CORSAllowOrigins:    parseCSVEnv("CORS_ALLOW_ORIGINS"),
	}
}

func getEnv(key, defaultValue string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	return v
}

func parseCSVEnv(key string) []string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return []string{}
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}
