package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Port          string
	DBPath        string
	Env           string
	SessionSecret string
	HTTPSEnabled  bool
	CertFile      string
	KeyFile       string
}

// Load loads configuration from environment variables and .env file
func Load() (*Config, error) {
	// Load .env file if it exists (ignore errors for production)
	_ = godotenv.Load()

	config := &Config{
		Port:          getEnv("PORT", "8080"),
		DBPath:        getEnv("DB_PATH", "./data/app.db"),
		Env:           getEnv("ENV", "development"),
		SessionSecret: getEnv("SESSION_SECRET", "dev-secret-change-in-production"),
		HTTPSEnabled:  getEnvAsBool("HTTPS_ENABLED", false),
		CertFile:      getEnv("CERT_FILE", ""),
		KeyFile:       getEnv("KEY_FILE", ""),
	}

	return config, nil
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getEnvAsBool gets an environment variable as a boolean with a fallback value
func getEnvAsBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return fallback
}
