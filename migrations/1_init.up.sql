CREATE TABLE IF NOT EXISTS users
(
    id        UUID PRIMARY KEY,
    email     TEXT NOT NULL UNIQUE,
    fname TEXT NOT NULL,
    lname TEXT NOT NULL,
    avatar TEXT,
    avatar_min TEXT,
    is_deleted BOOLEAN DEFAULT FALSE,
    pass_hash BYTEA NOT NULL,
    created_date  DATE DEFAULT CURRENT_DATE,
    updated_date DATE DEFAULT CURRENT_DATE
);
CREATE INDEX IF NOT EXISTS idx_email ON users (email);

CREATE TABLE IF NOT EXISTS subscribers
(
    id UUID PRIMARY KEY,
    blogger_id UUID NOT NULL,
    subscriber_id UUID NOT NULL,
    FOREIGN KEY (blogger_id) REFERENCES users(id),
    FOREIGN KEY (subscriber_id)  REFERENCES users(id),
    CONSTRAINT uniq_subscribe UNIQUE (blogger_id, subscriber_id)
);
CREATE INDEX IF NOT EXISTS inx_blogger_id ON subscribers(blogger_id);
CREATE INDEX IF NOT EXISTS inx_subscriber_id ON subscribers(subscriber_id);
