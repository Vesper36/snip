package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Paste    PasteConfig
}

type ServerConfig struct {
	Host         string
	Port         int
	BaseURL      string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DatabaseConfig struct {
	Path string
}

type AuthConfig struct {
	AdminPassword string
	JWTSecret     string
	JWTExpiry     time.Duration
}

type PasteConfig struct {
	MaxSize        int64
	DefaultExpire  time.Duration
	AllowAnonymous bool
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Host:         envStr("SNIP_HOST", "0.0.0.0"),
			Port:         envInt("SNIP_PORT", 53524),
			BaseURL:      envStr("SNIP_BASE_URL", "http://localhost:3000"),
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
		Database: DatabaseConfig{
			Path: envStr("SNIP_DB_PATH", "./data/snip.db"),
		},
		Auth: AuthConfig{
			AdminPassword: envStr("SNIP_ADMIN_PASSWORD", ""),
			JWTSecret:     envStr("SNIP_JWT_SECRET", "snip-secret-change-me"),
			JWTExpiry:     72 * time.Hour,
		},
		Paste: PasteConfig{
			MaxSize:        int64(envInt("SNIP_MAX_SIZE", 10*1024*1024)),
			AllowAnonymous: envBool("SNIP_ALLOW_ANONYMOUS", true),
		},
	}
}

func envStr(key, fb string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fb
}
func envInt(key string, fb int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fb
}
func envBool(key string, fb bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fb
}
