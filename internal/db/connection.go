package db

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // postgres driver
)

// DB holds the database connection
type DB struct {
	*sqlx.DB
}

// Connect establishes a connection to the database
func Connect(databaseURL string) (*DB, error) {
	sqlxDB, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		return nil, err
	}

	// Test the connection
	if err := sqlxDB.Ping(); err != nil {
		sqlxDB.Close()
		return nil, err
	}

	return &DB{sqlxDB}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}

// Health checks if the database connection is healthy
func (db *DB) Health() error {
	return db.Ping()
}
