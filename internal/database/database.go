package database

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"echo-server/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Service interface {
	Health() map[string]string
	Close() error
	Ping() error
}

type DB struct {
	Client *pgxpool.Pool
}

var dbInstance *DB

func New(cfg *config.Config) (*DB, error) {
	ctx := context.Background()
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance, nil
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.Env.DBUsername,
		cfg.Env.DBPassword,
		cfg.Env.DBHost,
		strconv.Itoa(cfg.Env.DBPort),
		cfg.Env.DBDatabase,
	)

	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to parse pool config: %w", err)
	}

	// Set pool configuration
	poolConfig.MaxConns = 50                       // Maximum number of connections in the pool
	poolConfig.MinConns = 5                        // Minimum number of connections in the pool
	poolConfig.MaxConnLifetime = 1 * time.Hour     // Maximum lifetime of a connection
	poolConfig.MaxConnIdleTime = 30 * time.Minute  // Maximum idle time for a connection
	poolConfig.HealthCheckPeriod = 1 * time.Minute // How often to check connection health

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create database pool: %w", err)
	}

	dbInstance = &DB{
		Client: pool,
	}

	return dbInstance, nil

}

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (db *DB) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	// Ping the database
	err := db.Client.Ping(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		slog.Error("db down", "error", err)
		return stats
	}

	// Database is up, add more statistics
	stats["status"] = "up"
	stats["message"] = "It's healthy"

	// Get detailed pool statistics
	poolStats := db.Client.Stat()
	stats["total_connections"] = strconv.Itoa(int(poolStats.TotalConns()))
	stats["acquired_connections"] = strconv.Itoa(int(poolStats.AcquiredConns()))
	stats["idle_connections"] = strconv.Itoa(int(poolStats.IdleConns()))
	stats["max_connections"] = strconv.Itoa(int(poolStats.MaxConns()))
	stats["canceled_acquires"] = strconv.FormatInt(poolStats.CanceledAcquireCount(), 10)
	stats["acquisition_duration"] = poolStats.AcquireDuration().String()
	stats["construction_wait_count"] = strconv.FormatInt(poolStats.EmptyAcquireCount(), 10)

	// Evaluate stats to provide a health message
	if poolStats.TotalConns() > 40 { // Assuming 50 is the max for this example
		stats["message"] = "The database is experiencing heavy load."
	}

	if poolStats.EmptyAcquireCount() > 1000 {
		stats["message"] = "High number of acquisitions that had to wait for connection construction, indicating potential bottlenecks."
	}

	if int(poolStats.MaxIdleDestroyCount()) > int(poolStats.TotalConns())/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if int(poolStats.MaxLifetimeDestroyCount()) > int(poolStats.TotalConns())/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return stats
}

func (db *DB) Close() error {
	slog.Info("Closing database connection pool")
	db.Client.Close()
	return nil
}

func (db *DB) Ping() error {
	ctx := context.Background()
	err := db.Client.Ping(ctx)
	if err != nil {
		slog.Error("Failed to ping database", "error", err)
		return err
	}
	return nil
}
