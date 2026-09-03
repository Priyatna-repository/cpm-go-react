package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv     string
	AppPort    string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration

	AdminName     string
	AdminEmail    string
	AdminPassword string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on environment variables")
	}

	cfg := &Config{
		AppEnv:     getEnv("APP_ENV", "local"),
		AppPort:    getEnv("APP_PORT", "8080"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "cpm"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		JWTSecret:       getEnv("JWT_SECRET", ""),
		AccessTokenTTL:  getEnvMinutes("JWT_ACCESS_TTL_MINUTES", 15),
		RefreshTokenTTL: getEnvHours("JWT_REFRESH_TTL_HOURS", 24*7),

		AdminName:     getEnv("ADMIN_NAME", "Admin"),
		AdminEmail:    getEnv("ADMIN_EMAIL", ""),
		AdminPassword: getEnv("ADMIN_PASSWORD", ""),
	}

	if len(cfg.JWTSecret) < 32 {
		log.Fatalf("JWT_SECRET must be set to a random string of at least 32 characters (got %d chars) — generate one with: openssl rand -base64 48", len(cfg.JWTSecret))
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvMinutes(key string, fallback int) time.Duration {
	return time.Duration(getEnvInt(key, fallback)) * time.Minute
}

func getEnvHours(key string, fallback int) time.Duration {
	return time.Duration(getEnvInt(key, fallback)) * time.Hour
}

func getEnvInt(key string, fallback int) int {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		log.Printf("invalid %s=%q, using default %d", key, value, fallback)
		return fallback
	}
	return n
}
