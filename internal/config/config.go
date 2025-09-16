package config

import (
	"log/slog"
	"time"

	"green-api-proxy/internal/logger"
	"green-api-proxy/internal/proxy"
	"green-api-proxy/internal/server"
)

type (
	Logger = logger.Config
	Server = server.Config
	Proxy  = proxy.Config
)

type Config struct {
	Logger *Logger
	Server *Server
	Proxy  *Proxy
}

func Load() (Config, error) {
	var ge getenv
	const required = true
	cfg := Config{
		Logger: &Logger{
			Level:     ge.LogLevel("LOG_LEVEL", !required, slog.LevelInfo),
			Plaintext: ge.Bool("LOG_PAINTEXT", !required, false),
		},
		Server: &Server{
			Addr:            ge.String("SERVER_ADDR", !required, ":8087"),
			TLS:             ge.Bool("SERVER_TLS", !required, false),
			AllowOrigin:     ge.String("ALLOW_ORIGIN", !required, ""),
			ReadTimeout:     ge.Duration("READ_TIMEOUT", !required, 1*time.Second),
			WriteTimeout:    ge.Duration("WRITE_TIMEOUT", !required, 4*time.Second),
			RedirectTimeout: ge.Duration("REDIRECT_TIMEOUT", !required, 4*time.Second),
		},
		Proxy: &Proxy{
			GreenApiBaseURL: ge.String("GREEN_API_BASE_URL", !required, "https://api.green-api.com"),
		},
	}
	return cfg, ge.Err()
}
