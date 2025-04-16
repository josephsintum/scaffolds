package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"echo-server/internal/config"
	"echo-server/internal/database"
)

type Server struct {
	port int
	host string

	db database.Service
}

func NewServer(cfg *config.Config) *http.Server {
	db, err := database.New(cfg)
	if err != nil {
		slog.Error(
			"failed to connect to database",
			"error", err,
		)
		os.Exit(1)
	}

	NewServer := &Server{
		port: cfg.Env.Port,
		host: "localhost",

		db: db,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", NewServer.host, NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
