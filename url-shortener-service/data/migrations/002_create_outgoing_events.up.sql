CREATE TABLE IF NOT EXISTS outgoing_events (
    id                 BIGINT PRIMARY KEY,
    payload            JSONB       NOT NULL,                      -- raw event JSON
    topic              TEXT        NOT NULL,                      -- e.g. urlshortener.metadata.requested.v1
    status             TEXT        NOT NULL DEFAULT 'PENDING',    -- PENDING | PUBLISHED | FAILED
    last_error         TEXT            NULL,
    correlation_id     TEXT        NOT NULL,
    trace_id           TEXT        NOT NULL,
    span_id            TEXT        NOT NULL,
    retry_count        INT         NOT NULL DEFAULT 0,
    created_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_outgoing_events_status ON outgoing_events(status, id);
