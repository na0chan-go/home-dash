CREATE TABLE IF NOT EXISTS notes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    kind TEXT NOT NULL CHECK (kind IN ('notice', 'shopping')),
    body TEXT NOT NULL CHECK (length(body) >= 1 AND length(body) <= 200),
    pinned INTEGER NOT NULL DEFAULT 0 CHECK (pinned IN (0, 1)),
    done INTEGER NOT NULL DEFAULT 0 CHECK (done IN (0, 1)),
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    CHECK ((kind = 'notice' AND done = 0) OR (kind = 'shopping' AND pinned = 0))
);

CREATE INDEX IF NOT EXISTS idx_notes_kind_created_at ON notes(kind, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notes_notice_order ON notes(kind, pinned DESC, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notes_shopping_order ON notes(kind, done ASC, created_at DESC);
