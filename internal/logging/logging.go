package logging

import (
	"log/slog"
	"os"
	"time"

	"echo-server/internal/config"

	clog "github.com/charmbracelet/log"
)

func SetupLogging(env config.Env) {
	defer slog.Info("The App is running in Env", "ENV", string(env))

	if env == config.PROD {
		logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
		slog.SetDefault(logger)
		return
	}

	handler := clog.NewWithOptions(os.Stderr, clog.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
		TimeFormat:      time.Kitchen,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
