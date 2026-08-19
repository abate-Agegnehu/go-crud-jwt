// internal/config/config.go
package config

import (
    "log"
    "os"
    "time"

    "github.com/joho/godotenv"
)

type Config struct {
    DBHost     string
    DBUser     string
    DBPassword string
    DBName     string
    DBPort     string
    JWTSecret  string
    JWTExpiry  time.Duration
    Port       string
}

var AppConfig *Config

func LoadConfig() {
    err := godotenv.Load()
    if err != nil {
        log.Println("Warning: .env file not found, using environment variables")
    }

    jwtExpiry, err := time.ParseDuration(os.Getenv("JWT_EXPIRY"))
    if err != nil {
        jwtExpiry = 24 * time.Hour // default 24 hours
    }

    AppConfig = &Config{
        DBHost:     getEnv("DB_HOST", "localhost"),
        DBUser:     getEnv("DB_USER", "postgres"),
        DBPassword: getEnv("DB_PASSWORD", ""),
        DBName:     getEnv("DB_NAME", "go_crud_db"),
        DBPort:     getEnv("DB_PORT", "5432"),
        JWTSecret:  getEnv("JWT_SECRET", "default-secret-key"),
        JWTExpiry:  jwtExpiry,
        Port:       getEnv("PORT", "8080"),
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}