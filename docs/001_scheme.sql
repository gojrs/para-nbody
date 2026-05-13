CREATE TABLE IF NOT EXISTS universes (
    id TEXT PRIMARY KEY,
    width INTEGER NOT NULL,
    height INTEGER NOT NULL,
    depth INTEGER NOT NULL,
    current_step INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS chunk_0 (
    universe_id TEXT NOT NULL,
    step INTEGER NOT NULL,
    payload BLOB NOT NULL,
    created_at TEXT NOT NULL,
    PRIMARY KEY (universe_id, step),
    FOREIGN KEY (universe_id) REFERENCES universes(id) ON DELETE CASCADE
);