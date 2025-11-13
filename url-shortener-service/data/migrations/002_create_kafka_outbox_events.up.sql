CREATE TABLE IF NOT EXISTS kafka_outbox_events (
    id           BIGINT PRIMARY KEY,
    event_type   TEXT        NOT NULL,                      -- e.g. urlshortener.metadata.requested.v1
    payload      JSONB       NOT NULL,                      -- raw event JSON
    status       TEXT        NOT NULL DEFAULT 'PENDING',    -- PENDING | SENT | FAILED
    created_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_outbox_status ON kafka_outbox_events(status, id);
