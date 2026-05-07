package db

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/gojrs/para-nbody/types"
	_ "github.com/mattn/go-sqlite3"
)

// Internal, non-exported variable to hold the connection
var store *sql.DB

// SetDB accepts the connection from main and initializes the 3D schema
func SetDB(conn *sql.DB) error {
	store = conn

	// WHACK THE MOLE: Limit to 1 connection and add a wait timer
	store.SetMaxOpenConns(1) // Prevents "Database is Locked" by queuing writes

	// Enable WAL mode for better read/write concurrency
	_, _ = store.Exec("PRAGMA journal_mode=WAL;")
	// Wait up to 5 seconds for a lock before failing
	_, _ = store.Exec("PRAGMA busy_timeout=5000;")

	query := `
CREATE TABLE IF NOT EXISTS runs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    experiment TEXT NOT NULL,         -- THE MISSING MOLE
    label TEXT,
    config TEXT,
    result TEXT,
    parent_run_id INTEGER,           -- Ensure this matches your SaveRun too
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);`
	_, err := store.Exec(query)
	return err
}

// SaveRun catalogs the 3D simulation results
func SaveRun(experiment string, label string, cfg types.NBodyConfig, res types.NBodyResult, parentID *int64) (int64, error) {
	cfgJSON, _ := json.Marshal(cfg)
	resJSON, _ := json.Marshal(res)

	// Added 'experiment' to the columns and one extra '?' placeholder
	query := `INSERT INTO runs (experiment, label, config, result, parent_run_id) VALUES (?, ?, ?, ?, ?)`

	// Pass 'experiment' as the first argument
	result, err := store.Exec(query, experiment, label, string(cfgJSON), string(resJSON), parentID)
	if err != nil {
		fmt.Printf("SQL EXEC ERROR: %v\n", err)
		return 0, err
	}

	return result.LastInsertId()
}

// GetRunByID retrieves state for the "Time Machine"
func GetRunByID(id int64) (types.NBodyConfig, types.NBodyResult, error) {
	var cfgStr, resStr string
	query := `SELECT config, result FROM runs WHERE id = ?`
	err := store.QueryRow(query, id).Scan(&cfgStr, &resStr)
	if err != nil {
		return types.NBodyConfig{}, types.NBodyResult{}, err
	}

	var cfg types.NBodyConfig
	var res types.NBodyResult
	json.Unmarshal([]byte(cfgStr), &cfg)
	json.Unmarshal([]byte(resStr), &res)

	return cfg, res, nil
}
