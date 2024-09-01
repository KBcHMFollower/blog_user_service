CREATE TABLE IF NOT EXISTS users
(
    id        UUID PRIMARY KEY,
    email     TEXT NOT NULL UNIQUE,
    fname TEXT NOT NULL,
    lname TEXT NOT NULL,
    avatar TEXT,
    avatar_min TEXT,
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

CREATE TABLE IF NOT EXISTS amqp_messages
(
    event_id UUID PRIMARY KEY ,
    event_type TEXT NOT NULL ,
    payload JSON NULL,
    status TEXT NOT NULL  DEFAULT 'waiting',
    retry_count INT DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_done ON amqp_messages(status);

CREATE TABLE IF NOT EXISTS request_keys
(
    id UUID PRIMARY KEY ,
    idempotency_key UUID NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_ikey ON request_keys(idempotency_key);
