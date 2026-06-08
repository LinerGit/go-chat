package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port       string
	DBDSN      string
	JWTSecret  string
	MaxMsgLen  int
	HistoryLim int
	PingPeriod time.Duration
	PongWait   time.Duration
	WriteWait  time.Duration
}

func MustLoad() *Config {
	cfg := &Config{
		Port:       getEnv("PORT", "8080"),
		DBDSN:      mustGetEnv("DB_DSN"),
		JWTSecret:  mustGetEnv("JWT_SECRET"),
		MaxMsgLen:  getEnvInt("MAX_MSG_LEN", 4096),
		HistoryLim: getEnvInt("HISTORY_LIMIT", 50),
		PongWait:   60 * time.Second,
		WriteWait:  10 * time.Second,
	}
	cfg.PingPeriod = (cfg.PongWait * 9) / 10
	return cfg
}

func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic("required env variable not set: " + key)
	}
	return v
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
