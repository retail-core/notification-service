package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	RabbitMQUrl    string
	PostgresUrl    string
	Port           string
	Environment    string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	
	return &Config{
		RabbitMQUrl: getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		PostgresUrl: getEnv("POSTGRES_URL", "postgres://user:password@localhost:5432/dbname?sslmode=disable"),
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
	}
}

func getEnv(key, def string) string {
	val, exists := os.LookupEnv(key)
	if !exists {
		return def
	}
	return val	
}