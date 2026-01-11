package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config содержит все настройки приложения
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	CORS     CORSConfig
}

type ServerConfig struct {
	Port         string
	Env          string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	URL string
}

type JWTConfig struct {
	Secret           string
	RefreshSecret    string
	AccessTokenTTL   time.Duration
	RefreshTokenTTL  time.Duration
}

type CORSConfig struct {
	AllowedOrigins []string
}

// Load загружает конфигурацию из переменных окружения
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "8080"),
			Env:          getEnv("ENV", "development"),
			ReadTimeout:  getDurationEnv("READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getDurationEnv("WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:  getDurationEnv("IDLE_TIMEOUT", 60*time.Second),
		},
		Database: DatabaseConfig{
			URL:             getEnv("DATABASE_URL", "postgres://novels:novels_dev_password@localhost:5432/novels?sslmode=disable"),
			MaxOpenConns:    getIntEnv("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getIntEnv("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getDurationEnv("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Redis: RedisConfig{
			URL: getEnv("REDIS_URL", "redis://localhost:6379"),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "dev_jwt_secret_change_in_production_32chars"),
			RefreshSecret:   getEnv("JWT_REFRESH_SECRET", "dev_jwt_refresh_secret_change_in_production_32chars"),
			AccessTokenTTL:  getDurationEnv("JWT_ACCESS_TTL", 1*time.Hour),
			RefreshTokenTTL: getDurationEnv("JWT_REFRESH_TTL", 7*24*time.Hour),
		},
		CORS: CORSConfig{
			AllowedOrigins: getSliceEnv("CORS_ORIGINS", []string{"http://localhost:3000"}),
		},
	}
}

// IsDevelopment проверяет, является ли окружение development
func (c *Config) IsDevelopment() bool {
	return c.Server.Env == "development"
}

// IsProduction проверяет, является ли окружение production
func (c *Config) IsProduction() bool {
	return c.Server.Env == "production"
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getSliceEnv(key string, defaultValue []string) []string {
	if value, exists := os.LookupEnv(key); exists {
		return strings.Split(value, ",")
	}
	return defaultValue
}
