-- 003_create_saga_state.up.sql
-- Saga execution instance
CREATE TABLE sagas (
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
);

-- Individual saga step execution records
CREATE TABLE saga_steps (
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
);

CREATE INDEX idx_sagas_status ON sagas (status);
CREATE INDEX idx_saga_steps_saga ON saga_steps (saga_id);
