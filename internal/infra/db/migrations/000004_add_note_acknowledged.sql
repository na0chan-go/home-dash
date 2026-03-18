ALTER TABLE notes ADD COLUMN acknowledged INTEGER NOT NULL DEFAULT 0 CHECK (acknowledged IN (0, 1));
CREATE INDEX IF NOT EXISTS idx_notes_notice_order_ack ON notes(kind, pinned DESC, acknowledged ASC, created_at DESC);
