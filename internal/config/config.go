package config

import (
	"os"
	"github.com/joho/godotenv"
)

type Config struct {
	DB_SOURCE      string
	RabbitMQUrl    string
	PostgresUrl    string
	Port           string
	Environment    string

	POSTMARKServerToken string
	POSTMARKAccountToken string
	POSTMARKFromEmail    string
	POSTMARKToEmail      string	
	POSTMARKFromName     string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	
	return &Config{
		DB_SOURCE:   getEnv("DB_SOURCE", ""),
		RabbitMQUrl: getEnv("RABBITMQ_URL", ""),
		PostgresUrl: getEnv("POSTGRES_URL", ""),
		Port:        getEnv("PORT", "8000"),
		Environment: getEnv("ENVIRONMENT", "development"),
		POSTMARKServerToken: getEnv("POSTMARK_SERVER_TOKEN", ""),
		POSTMARKAccountToken: getEnv("POSTMARK_ACCOUNT_TOKEN", ""),
		POSTMARKFromEmail:    getEnv("POSTMARK_FROM_EMAIL", "info@jarabiz.com"),
		POSTMARKFromName:     getEnv("POSTMARK_FROM_NAME", "Jara-Biz"),
		POSTMARKToEmail:      getEnv("POSTMARK_TO_EMAIL", "hello@jarabiz.com"),
	}
	}

func getEnv(key, def string) string {
	val, exists := os.LookupEnv(key)
	if !exists {
		return def
	}
	return val	
}