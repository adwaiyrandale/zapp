package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	flag.Parse()

	ctx := context.Background()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://root@localhost:26257/payments?sslmode=disable"
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer pool.Close()

	command := flag.Arg(0)

	switch command {
	case "up":
		if err := runMigrations(ctx, pool); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		fmt.Println("Migrations completed successfully")
	case "down":
		fmt.Println("Down migrations not implemented")
	default:
		fmt.Println("Usage: migrate [up|down]")
	}
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS accounts (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			type VARCHAR(50) NOT NULL,
			balance BIGINT NOT NULL DEFAULT 0,
			currency CHAR(3) NOT NULL DEFAULT 'USD',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS journals (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			description TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS journal_lines (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			journal_id UUID NOT NULL REFERENCES journals(id),
			account_id UUID NOT NULL REFERENCES accounts(id),
			debit BIGINT NOT NULL DEFAULT 0,
			credit BIGINT NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS payments (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			merchant_id UUID NOT NULL,
			amount BIGINT NOT NULL CHECK (amount > 0),
			currency CHAR(3) NOT NULL,
			status VARCHAR(20) NOT NULL CHECK (status IN ('PENDING', 'AUTHORIZED', 'CAPTURED', 'CANCELLED', 'REFUNDED')),
			capture_method VARCHAR(20) NOT NULL CHECK (capture_method IN ('AUTOMATIC', 'MANUAL')),
			metadata JSONB,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS charges (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			payment_id UUID NOT NULL REFERENCES payments(id),
			kind VARCHAR(20) NOT NULL CHECK (kind IN ('AUTHORIZATION', 'CAPTURE', 'REFUND', 'VOID')),
			amount BIGINT NOT NULL,
			currency CHAR(3) NOT NULL,
			status VARCHAR(20) NOT NULL CHECK (status IN ('PENDING', 'SUCCEEDED', 'FAILED')),
			processor_ref VARCHAR(100),
			failure_code VARCHAR(50),
			failure_message TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			completed_at TIMESTAMPTZ
		)`,
		`CREATE INDEX IF NOT EXISTS idx_payments_merchant ON payments (merchant_id)`,
		`CREATE INDEX IF NOT EXISTS idx_payments_status ON payments (status)`,
		`CREATE INDEX IF NOT EXISTS idx_charges_payment ON charges (payment_id)`,
		`CREATE TABLE IF NOT EXISTS sagas (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			kind VARCHAR(50) NOT NULL,
			status VARCHAR(20) NOT NULL CHECK (status IN ('RUNNING', 'COMPLETED', 'COMPENSATING', 'COMPENSATED', 'FAILED')),
			current_step INT NOT NULL DEFAULT 0,
			input JSONB NOT NULL,
			output JSONB,
			compensation_state JSONB,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			completed_at TIMESTAMPTZ
		)`,
		`CREATE TABLE IF NOT EXISTS saga_steps (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			saga_id UUID NOT NULL REFERENCES sagas(id),
			name VARCHAR(100) NOT NULL,
			seq INT NOT NULL,
			status VARCHAR(20) NOT NULL CHECK (status IN ('PENDING', 'RUNNING', 'COMPLETED', 'COMPENSATED', 'FAILED')),
			input JSONB,
			output JSONB,
			error TEXT,
			started_at TIMESTAMPTZ,
			completed_at TIMESTAMPTZ
		)`,
		`CREATE INDEX IF NOT EXISTS idx_sagas_status ON sagas (status)`,
		`CREATE INDEX IF NOT EXISTS idx_saga_steps_saga ON saga_steps (saga_id)`,
		`CREATE TABLE IF NOT EXISTS outbox (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			topic VARCHAR(255) NOT NULL,
			key VARCHAR(255),
			value JSONB NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			published_at TIMESTAMPTZ,
			retry_count INT NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS settlements (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			merchant_id UUID NOT NULL,
			payment_id UUID,
			amount BIGINT NOT NULL CHECK (amount > 0),
			currency CHAR(3) NOT NULL,
			type VARCHAR(10) NOT NULL CHECK (type IN ('ACH', 'WIRE')),
			status VARCHAR(20) NOT NULL CHECK (status IN ('PENDING', 'PROCESSING', 'COMPLETED', 'FAILED', 'CANCELLED')),
			bank_account VARCHAR(50) NOT NULL,
			routing_number VARCHAR(20) NOT NULL,
			trace_number VARCHAR(50),
			failure_code VARCHAR(50),
			failure_message TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			completed_at TIMESTAMPTZ
		)`,
		`CREATE INDEX IF NOT EXISTS idx_settlements_merchant ON settlements (merchant_id)`,
		`CREATE INDEX IF NOT EXISTS idx_settlements_payment ON settlements (payment_id)`,
		`CREATE INDEX IF NOT EXISTS idx_settlements_status ON settlements (status)`,
	}

	for _, sql := range migrations {
		if _, err := pool.Exec(ctx, sql); err != nil {
			return fmt.Errorf("failed to run migration: %w", err)
		}
	}

	return nil
}
