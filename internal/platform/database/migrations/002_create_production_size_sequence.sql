CREATE TABLE production_size_sequence (
    singleton_id SMALLINT PRIMARY KEY CHECK (singleton_id = 1),
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE production_size_sequence_items (
    sequence_id SMALLINT NOT NULL REFERENCES production_size_sequence(singleton_id) ON DELETE CASCADE,
    name TEXT NOT NULL CHECK (length(trim(name)) > 0),
    position INTEGER NOT NULL CHECK (position > 0),
    PRIMARY KEY (sequence_id, position)
);

CREATE UNIQUE INDEX production_size_sequence_items_name_unique
    ON production_size_sequence_items (sequence_id, lower(trim(name)));

---- create above / drop below ----

DROP TABLE IF EXISTS production_size_sequence_items;
DROP TABLE IF EXISTS production_size_sequence;
