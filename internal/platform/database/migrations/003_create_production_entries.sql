CREATE TABLE production_entries (
    id TEXT PRIMARY KEY,
    received_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE production_color_batches (
    id TEXT PRIMARY KEY,
    entry_id TEXT NOT NULL REFERENCES production_entries(id) ON DELETE CASCADE,
    color_name TEXT NOT NULL CHECK (length(trim(color_name)) > 0),
    color_key TEXT NOT NULL CHECK (length(trim(color_key)) > 0),
    position INTEGER NOT NULL CHECK (position > 0),
    status TEXT NOT NULL CHECK (length(trim(status)) > 0),
    created_at TIMESTAMPTZ NOT NULL,
    started_at TIMESTAMPTZ NULL,
    completed_at TIMESTAMPTZ NULL,
    UNIQUE (entry_id, color_key),
    UNIQUE (entry_id, position)
);

CREATE TABLE production_size_runs (
    id TEXT PRIMARY KEY,
    color_batch_id TEXT NOT NULL REFERENCES production_color_batches(id) ON DELETE CASCADE,
    size_name TEXT NOT NULL CHECK (length(trim(size_name)) > 0),
    position INTEGER NOT NULL CHECK (position > 0),
    status TEXT NOT NULL CHECK (length(trim(status)) > 0),
    quantity INTEGER NULL CHECK (quantity IS NULL OR quantity >= 0),
    started_at TIMESTAMPTZ NULL,
    completed_at TIMESTAMPTZ NULL,
    UNIQUE (color_batch_id, position)
);

---- create above / drop below ----

DROP TABLE IF EXISTS production_size_runs;
DROP TABLE IF EXISTS production_color_batches;
DROP TABLE IF EXISTS production_entries;
