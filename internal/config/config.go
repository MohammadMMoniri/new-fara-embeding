// internal/config/config.go
package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	MinIO    MinIOConfig
	OpenAI   OpenAIConfig
	LogLevel string
}

type ServerConfig struct {
	Port int
}

type DatabaseConfig struct {
	URL string
}

type MinIOConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	BucketName      string
}

type OpenAIConfig struct {
	APIKey   string
	BaseURL  string
	Model    string
	MaxRetries int
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnvAsInt("PORT", 8080),
		},
		Database: DatabaseConfig{
			URL: getEnv("DATABASE_URL", "postgres://localhost/embeddings?sslmode=disable"),
		},
		MinIO: MinIOConfig{
			Endpoint:        getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKeyID:     getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			SecretAccessKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
			UseSSL:          getEnvAsBool("MINIO_USE_SSL", false),
			BucketName:      getEnv("MINIO_BUCKET", "documents"),
		},
		OpenAI: OpenAIConfig{
			APIKey:     getEnv("OPENAI_API_KEY", ""),
			BaseURL:    getEnv("OPENAI_BASE_URL", "https://api.openai.com/v1"),
			Model:      getEnv("OPENAI_MODEL", "text-embedding-3-small"),
			MaxRetries: getEnvAsInt("OPENAI_MAX_RETRIES", 3),
		},
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

