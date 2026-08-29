CREATE INDEX production_entries_received_at_id_idx
    ON production_entries (received_at DESC, id DESC);

CREATE INDEX production_color_batches_queue_idx
    ON production_color_batches (created_at ASC, entry_id ASC, position ASC, id ASC);

CREATE INDEX production_color_batches_status_queue_idx
    ON production_color_batches (status, created_at ASC, entry_id ASC, position ASC, id ASC);

---- create above / drop below ----

DROP INDEX IF EXISTS production_color_batches_status_queue_idx;
DROP INDEX IF EXISTS production_color_batches_queue_idx;
DROP INDEX IF EXISTS production_entries_received_at_id_idx;
