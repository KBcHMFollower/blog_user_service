CREATE TABLE IF NOT EXISTS users
(
    id        UUID PRIMARY KEY,
    email     TEXT NOT NULL UNIQUE,
    pass_hash BYTEA NOT NULL,
    created_date  DATE,
    updated_date DATE
);
CREATE INDEX IF NOT EXISTS idx_email ON users (email);
