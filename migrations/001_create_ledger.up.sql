-- 001_create_ledger.up.sql
-- Chart of accounts
CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('ASSET', 'LIABILITY', 'EQUITY', 'REVENUE', 'EXPENSE')),
    currency CHAR(3) NOT NULL,
    normal_balance VARCHAR(10) NOT NULL CHECK (normal_balance IN ('DEBIT', 'CREDIT')),
    parent_id UUID REFERENCES accounts(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Journal entry header
CREATE TABLE journals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reference_type VARCHAR(50) NOT NULL,
    reference_id UUID NOT NULL,
    description TEXT,
    posted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Journal lines (individual debit/credit entries)
CREATE TABLE journal_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    journal_id UUID NOT NULL REFERENCES journals(id),
    account_id UUID NOT NULL REFERENCES accounts(id),
    amount BIGINT NOT NULL CHECK (amount > 0),
    direction VARCHAR(10) NOT NULL CHECK (direction IN ('DEBIT', 'CREDIT')),
    currency CHAR(3) NOT NULL,
    memo TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for faster lookups
CREATE INDEX idx_journals_reference ON journals (reference_type, reference_id);
CREATE INDEX idx_journal_lines_journal ON journal_lines (journal_id);
CREATE INDEX idx_journal_lines_account ON journal_lines (account_id);

-- Enforce double-entry: all lines in a journal must balance
CREATE OR REPLACE FUNCTION check_double_entry()
RETURNS TRIGGER AS $$
DECLARE
    debit_total BIGINT;
    credit_total BIGINT;
BEGIN
    SELECT 
        COALESCE(SUM(CASE WHEN direction = 'DEBIT' THEN amount ELSE 0 END), 0),
        COALESCE(SUM(CASE WHEN direction = 'CREDIT' THEN amount ELSE 0 END), 0)
    INTO debit_total, credit_total
    FROM journal_lines
    WHERE journal_id = NEW.journal_id;
    
    IF debit_total != credit_total THEN
        RAISE EXCEPTION 'Double-entry violation: debits (%) != credits (%)', 
            debit_total, credit_total;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER double_entry_trigger
    AFTER INSERT ON journal_lines
    FOR EACH ROW EXECUTE FUNCTION check_double_entry();

-- Insert default accounts
INSERT INTO accounts (code, name, type, currency, normal_balance) VALUES
    ('1000', 'Cash', 'ASSET', 'USD', 'DEBIT'),
    ('1100', 'Accounts Receivable', 'ASSET', 'USD', 'DEBIT'),
    ('2000', 'Revenue', 'REVENUE', 'USD', 'CREDIT'),
    ('2100', 'Refunds Payable', 'LIABILITY', 'USD', 'CREDIT'),
    ('3000', 'Payment Processing Fees', 'EXPENSE', 'USD', 'DEBIT');
