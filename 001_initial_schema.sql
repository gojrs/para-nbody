-- 1. THE UNIVERSAL SEED TABLE
-- This stores the "Laws of Physics" for every version of your universe.
CREATE TABLE IF NOT EXISTS universe_seeds (
    seed_id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,

    -- Referee Parameters (God-Mode Sliders)
    creation_bias REAL DEFAULT 0.85,
    matrix_throughput REAL DEFAULT 1.0,
    spleef_constant REAL DEFAULT 1.0,
    expansion_rate REAL DEFAULT 0.05,
    energy_threshold REAL DEFAULT 100.0,

    -- Metadata and Protobuf Storage
    -- We store the full Protobuf here so the Go Engine can hydrate instantly.
    raw_protobuf_blob BLOB,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 2. THE SURVIVOR LEDGER
-- This stores the results of the "Lab Runs" (The Survivor Pics).
CREATE TABLE IF NOT EXISTS survivor_reports (
    report_id INTEGER PRIMARY KEY AUTOINCREMENT,
    seed_id INTEGER NOT NULL,
    tick_count INTEGER,
    survivor_count INTEGER,

    -- The "Survivor Pic" (The Trace Lines)
    -- Stored as a BLOB so the UI can render it later.
    pic_data BLOB,

    FOREIGN KEY (seed_id) REFERENCES universe_seeds(seed_id)
);

-- 3. THE ESCROW LOG
-- Optional: If you want to audit how the "Extra E" fluctuated over time.
CREATE TABLE IF NOT EXISTS escrow_history (
    log_id INTEGER PRIMARY KEY AUTOINCREMENT,
    seed_id INTEGER,
    tick_index INTEGER,
    escrow_value REAL,
    event_type TEXT -- e.g., "LIQUIDATION" or "PAVING"
);

-- 4. INSERT THE "SANE DEFAULT" (Row #1)
-- This matches your genesis.yaml exactly.
INSERT INTO universe_seeds (
    seed_id,
    name,
    description,
    creation_bias,
    matrix_throughput,
    spleef_constant,
    expansion_rate,
    energy_threshold
) VALUES (
    1,
    'The Standard Bridge',
    'Initial bootstrap from genesis.yaml. Standard Model balance.',
    0.85,
    1.0,
    1.0,
    0.05,
    100.0
);