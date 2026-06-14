-- Tabla de Cuentas
CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_number VARCHAR(20) UNIQUE NOT NULL,
    client_id VARCHAR(50) NOT NULL,
    type VARCHAR(20) NOT NULL, -- 'checking' o 'savings'
    currency VARCHAR(3) DEFAULT 'USD' NOT NULL,
    balance NUMERIC(15, 2) NOT NULL DEFAULT 0.00,
    status VARCHAR(20) DEFAULT 'ACTIVE' NOT NULL, -- 'active' | 'blocked' | 'closed'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,

    -- RESTRICCIÓN DE NEGOCIO: El saldo nunca puede ser negativo
    CONSTRAINT chk_positive_balance CHECK (balance >= 0.00)
);

-- Tabla de Transacciones
CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    reference_unique VARCHAR(20) UNIQUE NOT NULL, -- Idempotencia y búsqueda
    ammount NUMERIC(15, 2) NOT NULL,
    type VARCHAR(20) NOT NULL, -- 'deposit' | 'withdrawal' | 'transfer_out' | 'transfer_in'
    status VARCHAR(20) DEFAULT 'SUCCESSFUL' NOT NULL, -- 'successful' | 'failed' | 'reverted'
    transfer_details JSONB, -- Detalles complementarios de transferencias
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
);

-- Indices para optimizar consultas
CREATE UNIQUE INDEX IF NOT EXISTS idx_accounts_number ON accounts(account_number);
CREATE INDEX IF NOT EXISTS idx_transactions_account_date ON transactions(account_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_accounts_client ON accounts(client_id);
CREATE INDEX IF NOT EXISTS idx_transactions_type_date ON transactions(account_id, type, timestamp DESC);