package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gojrs/para-nbody/engine"
	_ "github.com/mattn/go-sqlite3"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)

	store := &SQLiteStore{
		db: db,
	}

	if err := store.init(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return store, nil
}

func (s *SQLiteStore) init() error {
	statements := []string{
		`PRAGMA journal_mode = WAL;`,
		`PRAGMA synchronous = NORMAL;`,
		`PRAGMA busy_timeout = 5000;`,
		`PRAGMA foreign_keys = ON;`,
		`CREATE TABLE IF NOT EXISTS universes (
			id TEXT PRIMARY KEY,
			width INTEGER NOT NULL,
			height INTEGER NOT NULL,
			depth INTEGER NOT NULL,
			current_step INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS chunk_0 (
			universe_id TEXT NOT NULL,
			step INTEGER NOT NULL,
			payload BLOB NOT NULL,
			created_at TEXT NOT NULL,
			PRIMARY KEY (universe_id, step),
			FOREIGN KEY (universe_id) REFERENCES universes(id) ON DELETE CASCADE
		);`,
	}

	for _, statement := range statements {
		if _, err := s.db.Exec(statement); err != nil {
			return err
		}
	}

	return nil
}

func (s *SQLiteStore) Create(id string, world *engine.World) error {
	if id == "" {
		return fmt.Errorf("id is required")
	}
	if world == nil {
		return fmt.Errorf("world is required")
	}

	payload, err := json.Marshal(world)
	if err != nil {
		return fmt.Errorf("marshal world: %w", err)
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(
		`INSERT INTO universes (id, width, height, depth, current_step, created_at, updated_at)
		 VALUES (?, ?, ?, ?, 0, ?, ?)`,
		id,
		world.Width,
		world.Height,
		world.Depth,
		now,
		now,
	); err != nil {
		return fmt.Errorf("insert universe: %w", err)
	}

	if _, err := tx.Exec(
		`INSERT INTO chunk_0 (universe_id, step, payload, created_at)
		 VALUES (?, 0, ?, ?)`,
		id,
		payload,
		now,
	); err != nil {
		return fmt.Errorf("insert chunk_0: %w", err)
	}

	return tx.Commit()
}

func (s *SQLiteStore) Get(id string) (*engine.World, bool, error) {
	if id == "" {
		return nil, false, fmt.Errorf("id is required")
	}

	var currentStep int64
	err := s.db.QueryRow(
		`SELECT current_step FROM universes WHERE id = ?`,
		id,
	).Scan(&currentStep)
	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("select universe: %w", err)
	}

	var payload []byte
	err = s.db.QueryRow(
		`SELECT payload FROM chunk_0 WHERE universe_id = ? AND step = ?`,
		id,
		currentStep,
	).Scan(&payload)
	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("select chunk_0: %w", err)
	}

	var world engine.World
	if err := json.Unmarshal(payload, &world); err != nil {
		return nil, false, fmt.Errorf("unmarshal world: %w", err)
	}

	return &world, true, nil
}

func (s *SQLiteStore) Update(id string, world *engine.World) error {
	if id == "" {
		return fmt.Errorf("id is required")
	}
	if world == nil {
		return fmt.Errorf("world is required")
	}

	payload, err := json.Marshal(world)
	if err != nil {
		return fmt.Errorf("marshal world: %w", err)
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var currentStep int64
	err = tx.QueryRow(
		`SELECT current_step FROM universes WHERE id = ?`,
		id,
	).Scan(&currentStep)
	if err == sql.ErrNoRows {
		return fmt.Errorf("universe %q not found", id)
	}
	if err != nil {
		return fmt.Errorf("select current step: %w", err)
	}

	nextStep := currentStep + 1

	if _, err := tx.Exec(
		`INSERT INTO chunk_0 (universe_id, step, payload, created_at)
		 VALUES (?, ?, ?, ?)`,
		id,
		nextStep,
		payload,
		now,
	); err != nil {
		return fmt.Errorf("insert chunk_0 step: %w", err)
	}

	if _, err := tx.Exec(
		`UPDATE universes
		 SET width = ?, height = ?, depth = ?, current_step = ?, updated_at = ?
		 WHERE id = ?`,
		world.Width,
		world.Height,
		world.Depth,
		nextStep,
		now,
		id,
	); err != nil {
		return fmt.Errorf("update universe: %w", err)
	}

	return tx.Commit()
}

func (s *SQLiteStore) Delete(id string) error {
	if id == "" {
		return fmt.Errorf("id is required")
	}

	_, err := s.db.Exec(`DELETE FROM universes WHERE id = ?`, id)
	return err
}

func (s *SQLiteStore) Close() error {
	if s == nil || s.db == nil {
		return nil
	}

	return s.db.Close()
}
