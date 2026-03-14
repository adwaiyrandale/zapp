-- 005_create_settlements.up.sql
-- Settlement records for ACH and wire transfers
CREATE TABLE settlements (
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
);

CREATE INDEX idx_settlements_merchant ON settlements (merchant_id);
CREATE INDEX idx_settlements_payment ON settlements (payment_id);
CREATE INDEX idx_settlements_status ON settlements (status);
