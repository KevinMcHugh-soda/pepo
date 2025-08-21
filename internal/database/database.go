package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"pepo/internal/db"

	"go.uber.org/zap"
)

// ConnectionConfig holds database connection configuration
type ConnectionConfig struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// DefaultConnectionConfig returns sensible defaults for database connections
func DefaultConnectionConfig() *ConnectionConfig {
	return &ConnectionConfig{
		MaxOpenConns:    25,
		MaxIdleConns:    25,
		ConnMaxLifetime: 5 * time.Minute,
	}
}

// Initialize sets up the database connection and returns the database instance and queries
func Initialize(databaseURL string, config *ConnectionConfig) (*sql.DB, *db.Queries, error) {
	if config == nil {
		config = DefaultConnectionConfig()
	}

	// Open database connection
	database, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := database.Ping(); err != nil {
		database.Close()
		return nil, nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Configure connection pool
	database.SetMaxOpenConns(config.MaxOpenConns)
	database.SetMaxIdleConns(config.MaxIdleConns)
	database.SetConnMaxLifetime(config.ConnMaxLifetime)

	// Initialize queries
	queries := db.New(database)

	zap.L().Info("database connection established",
		zap.Int("max_open_conns", config.MaxOpenConns),
		zap.Int("max_idle_conns", config.MaxIdleConns),
		zap.Duration("conn_max_lifetime", config.ConnMaxLifetime))

	return database, queries, nil
}

// Close safely closes the database connection
func Close(database *sql.DB) error {
	if database != nil {
		return database.Close()
	}
	return nil
}
