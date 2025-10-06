CREATE TABLE IF NOT EXISTS data (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL CHECK (type IN ('login_password', 'text', 'binary', 'bank_card')),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    data BYTEA NOT NULL,
    metadata TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_data_user_id ON data(user_id);
CREATE INDEX IF NOT EXISTS idx_data_type ON data(type);


