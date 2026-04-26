package database

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/casapps/casci/src/internal/config"
)

// Database represents the database abstraction layer
type Database struct {
	primary  *sql.DB
	replicas []*sql.DB
	cache    *sql.DB // SQLite cache for failover
	dbType   string
	mu       sync.RWMutex
}

// Driver interface for database operations
type Driver interface {
	Connect(dsn string) (*sql.DB, error)
	Migrate(ctx context.Context, db *sql.DB) error
	GetType() string
}

// New creates a new database instance
func New(ctx context.Context, cfg config.DatabaseConfig) (*Database, error) {
	var driver Driver

	// Select driver based on type
	switch cfg.Type {
	case "sqlite":
		driver = &SQLiteDriver{}
	case "postgres", "postgresql":
		driver = &PostgresDriver{}
	case "mysql", "mariadb":
		driver = &MySQLDriver{}
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	// Connect to primary database
	primary, err := driver.Connect(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to primary database: %w", err)
	}

	// Configure connection pool
	primary.SetMaxOpenConns(cfg.MaxConns)
	primary.SetMaxIdleConns(cfg.MaxIdleConn)

	// Verify connection
	if err := primary.PingContext(ctx); err != nil {
		primary.Close()
		return nil, fmt.Errorf("failed to ping primary database: %w", err)
	}

	db := &Database{
		primary: primary,
		dbType:  cfg.Type,
	}

	// Connect to replicas if configured
	for _, replicaDSN := range cfg.Replicas {
		replica, err := driver.Connect(replicaDSN)
		if err != nil {
			// Log error but don't fail
			fmt.Printf("Warning: failed to connect to replica: %v\n", err)
			continue
		}
		replica.SetMaxOpenConns(cfg.MaxConns)
		replica.SetMaxIdleConns(cfg.MaxIdleConn)
		db.replicas = append(db.replicas, replica)
	}

	// Create SQLite cache for failover (if not using SQLite as primary)
	if cfg.Type != "sqlite" {
		cacheDriver := &SQLiteDriver{}
		cache, err := cacheDriver.Connect(".casci/cache/db.sqlite")
		if err == nil {
			db.cache = cache
		}
	}

	return db, nil
}

// DB returns the underlying primary database connection
func (d *Database) DB() *sql.DB {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.primary != nil {
		return d.primary
	}
	return d.cache
}

// Migrate runs database migrations
func (d *Database) Migrate(ctx context.Context) error {
	var driver Driver

	switch d.dbType {
	case "sqlite":
		driver = &SQLiteDriver{}
	case "postgres", "postgresql":
		driver = &PostgresDriver{}
	case "mysql", "mariadb":
		driver = &MySQLDriver{}
	default:
		return fmt.Errorf("unsupported database type: %s", d.dbType)
	}

	return driver.Migrate(ctx, d.primary)
}

// Query executes a query that returns rows with failover support
func (d *Database) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// Try primary first
	if d.primary != nil {
		rows, err := d.primary.QueryContext(ctx, query, args...)
		if err == nil {
			return rows, nil
		}
		fmt.Printf("Primary query failed: %v, trying replicas\n", err)
	}

	// Try replicas
	for _, replica := range d.replicas {
		rows, err := replica.QueryContext(ctx, query, args...)
		if err == nil {
			return rows, nil
		}
	}

	// Last resort: try cache
	if d.cache != nil {
		return d.cache.QueryContext(ctx, query, args...)
	}

	return nil, fmt.Errorf("all database connections failed")
}

// QueryRow executes a query that returns a single row
func (d *Database) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.primary != nil {
		return d.primary.QueryRowContext(ctx, query, args...)
	}

	// Try replicas
	for _, replica := range d.replicas {
		return replica.QueryRowContext(ctx, query, args...)
	}

	// Try cache
	if d.cache != nil {
		return d.cache.QueryRowContext(ctx, query, args...)
	}

	return nil
}

// Exec executes a query without returning rows
func (d *Database) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.primary == nil {
		return nil, fmt.Errorf("no primary database connection")
	}

	result, err := d.primary.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	// Async replication to other databases
	go d.replicateWrite(query, args...)

	return result, nil
}

// replicateWrite asynchronously replicates writes to replicas and cache
func (d *Database) replicateWrite(query string, args ...interface{}) {
	ctx := context.Background()

	// Replicate to replicas
	for _, replica := range d.replicas {
		if _, err := replica.ExecContext(ctx, query, args...); err != nil {
			fmt.Printf("Failed to replicate to replica: %v\n", err)
		}
	}

	// Replicate to cache
	if d.cache != nil {
		if _, err := d.cache.ExecContext(ctx, query, args...); err != nil {
			fmt.Printf("Failed to replicate to cache: %v\n", err)
		}
	}
}

// BeginTx starts a transaction
func (d *Database) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.primary == nil {
		return nil, fmt.Errorf("no primary database connection")
	}

	return d.primary.BeginTx(ctx, opts)
}

// Close closes all database connections
func (d *Database) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	var errs []error

	if d.primary != nil {
		if err := d.primary.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	for _, replica := range d.replicas {
		if err := replica.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if d.cache != nil {
		if err := d.cache.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing database connections: %v", errs)
	}

	return nil
}

// Ping checks if the database is reachable
func (d *Database) Ping(ctx context.Context) error {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.primary == nil {
		return fmt.Errorf("no database connection")
	}

	return d.primary.PingContext(ctx)
}

// GetType returns the database type
func (d *Database) GetType() string {
	return d.dbType
}
