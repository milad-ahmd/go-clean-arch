package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for our application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Logger   LoggerConfig
	Auth     AuthConfig
}

// ServerConfig holds all server related configuration
type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DatabaseConfig holds all database related configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// LoggerConfig holds all logger related configuration
type LoggerConfig struct {
	Level string
}

// AuthConfig holds all authentication related configuration
type AuthConfig struct {
	JWTSecret string
}

// LoadConfig loads configuration from .env file and environment variables
func LoadConfig() *Config {
	// Load .env file if it exists
	loadEnvFile()
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			ReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  getDurationEnv("SERVER_IDLE_TIMEOUT", 120*time.Second),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "clean_arch"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Logger: LoggerConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
		Auth: AuthConfig{
			JWTSecret: getEnv("JWT_SECRET", "your-secret-key"),
		},
	}
}

// Helper function to get an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// Helper function to get a duration environment variable with a default value
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return time.Duration(intValue) * time.Second
		}
	}
	return defaultValue
}

// loadEnvFile loads environment variables from .env file
func loadEnvFile() {
	// Try to find .env file in current directory and parent directories
	envFile := findEnvFile()
	if envFile != "" {
		err := godotenv.Load(envFile)
		if err != nil {
			fmt.Printf("Warning: Error loading .env file: %v\n", err)
		}
	}
}

// findEnvFile tries to find .env file in current directory and parent directories
func findEnvFile() string {
	// First try the current directory
	if _, err := os.Stat(".env"); err == nil {
		absPath, _ := filepath.Abs(".env")
		return absPath
	}

	// Try to find the project root (where go.mod is located)
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		// Check if .env exists in this directory
		envPath := filepath.Join(dir, ".env")
		if _, err := os.Stat(envPath); err == nil {
			return envPath
		}

		// Check if we're at the root directory
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return ""
}
