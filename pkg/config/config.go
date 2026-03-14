package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL    string
	LedgerPort     string
	PaymentPort    string
	SettlementPort string
	SagaPort       string
	DMQBroker      string
}

func Load() (*Config, error) {
	// Load .env file if exists
	_ = godotenv.Load()

	config := &Config{
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://root@localhost:26257/payments?sslmode=disable"),
		LedgerPort:     getEnv("LEDGER_PORT", "8081"),
		PaymentPort:    getEnv("PAYMENT_PORT", "8082"),
		SettlementPort: getEnv("SETTLEMENT_PORT", "8083"),
		SagaPort:       getEnv("SAGA_PORT", "8084"),
		DMQBroker:      getEnv("DMQ_BROKER", "localhost:9092"),
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (c *Config) LedgerURL() string {
	return fmt.Sprintf("http://localhost:%s", c.LedgerPort)
}

func (c *Config) PaymentURL() string {
	return fmt.Sprintf("http://localhost:%s", c.PaymentPort)
}

func (c *Config) SettlementURL() string {
	return fmt.Sprintf("http://localhost:%s", c.SettlementPort)
}

func (c *Config) SagaURL() string {
	return fmt.Sprintf("http://localhost:%s", c.SagaPort)
}
