package config

import "os"

type Config struct {
	AppAddr             string
	DBPath              string
	GarbageSchedulePath string
}

func LoadFromEnv() Config {
	return Config{
		AppAddr:             getEnv("APP_ADDR", ":8080"),
		DBPath:              getEnv("DB_PATH", "/data/app.db"),
		GarbageSchedulePath: getEnv("GARBAGE_SCHEDULE_PATH", "config/garbage_schedule.json"),
	}
}

func getEnv(key, defaultValue string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	return v
}
