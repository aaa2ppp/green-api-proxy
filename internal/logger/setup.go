package logger

import (
	"log/slog"
	"os"
)

type Config struct {
	Level     slog.Level
	Plaintext bool
}

func SetupDefault(cfg *Config) {
	if cfg.Plaintext {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.Level})))
	} else {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.Level})))
	}
}
