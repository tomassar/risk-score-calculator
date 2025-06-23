package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// DB wraps the SQL database connection
type DB struct {
	*sql.DB
}

// NewSQLiteDB creates a new SQLite database connection
func NewSQLiteDB(dbPath string) (*DB, error) {
	// Ensure the directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	wrapper := &DB{DB: db}

	// Run migrations
	if err := wrapper.migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return wrapper, nil
}

// migrate runs database migrations
func (db *DB) migrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS borrowers (
			id TEXT PRIMARY KEY,
			full_name TEXT NOT NULL,
			email TEXT NOT NULL,
			age INTEGER NOT NULL,
			monthly_income REAL NOT NULL,
			employment_status TEXT NOT NULL,
			dependents INTEGER NOT NULL,
			existing_loans REAL NOT NULL,
			savings_balance REAL NOT NULL,
			risk_score INTEGER DEFAULT 100,
			score_explanation TEXT DEFAULT '',
			decision TEXT DEFAULT 'REVIEW',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS rules (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT NOT NULL,
			field TEXT NOT NULL,
			operator TEXT NOT NULL,
			value TEXT NOT NULL,
			weight INTEGER NOT NULL,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_borrowers_risk_score ON borrowers(risk_score)`,
		`CREATE INDEX IF NOT EXISTS idx_borrowers_decision ON borrowers(decision)`,
		`CREATE INDEX IF NOT EXISTS idx_rules_active ON rules(is_active)`,
		`CREATE INDEX IF NOT EXISTS idx_rules_field ON rules(field)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("failed to execute migration: %w", err)
		}
	}

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}
