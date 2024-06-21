CREATE TABLE IF NOT EXISTS users
(
    id        UUID PRIMARY KEY,
    email     TEXT NOT NULL UNIQUE,
    fname TEXT NOT NULL,
    lname TEXT NOT NULL,
    pass_hash BYTEA NOT NULL,
    created_date  DATE,
    updated_date DATE
);
CREATE INDEX IF NOT EXISTS idx_email ON users (email);

CREATE TABLE IF NOT EXISTS subscribers
(
    id UUID PRIMARY KEY,
    bloger_id UUID NOT NULL,
    subscriber_id UUID NOT NULL,
    FOREIGN KEY (bloger_id) REFERENCES users(id),
    FOREIGN KEY (subscriber_id)  REFERENCES users(id)
);
