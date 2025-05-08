CREATE TABLE short_urls
(
    short_code   TEXT PRIMARY KEY,
    original_url TEXT                     NOT NULL,
    status       TEXT                     NOT NULL,
    created_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);