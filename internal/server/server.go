package server

import (
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"time"

	"green-api-proxy/internal/logger"
)

type Config struct {
	Addr            string
	TLS             bool
	AllowOrigin     string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	RedirectTimeout time.Duration
}

func (cfg *Config) Validate() error {
	if ao := cfg.AllowOrigin; ao != "" && ao != "*" {
		if _, err := url.Parse(ao); err != nil {
			return fmt.Errorf("can't parse AllowOrigin: %w", err)
		}
	}
	return nil
}

func New(cfg *Config, proxy http.Handler, fs fs.FS) (*http.Server, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	mux := http.NewServeMux()

	// Настраиваем статику
	static := http.NewServeMux()
	static.Handle("GET /green-api/", http.StripPrefix("/green-api", http.FileServerFS(fs)))
	mux.Handle("/green-api/", static)

	// Настраиваем proxy
	mux.Handle("/green-api/proxy/", http.StripPrefix("/green-api/proxy", proxyHandler(cfg, proxy)))

	return &http.Server{
		Addr:         cfg.Addr,
		Handler:      logger.HTTPLogging(slog.Default(), mux),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}, nil
}

func proxyHandler(cfg *Config, proxy http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logger.FromContext(r.Context())

		// Разрешаем CORS только для нужных доменов
		if cfg.AllowOrigin != "" {
			w.Header().Set("Access-Control-Allow-Origin", cfg.AllowOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "X-ApiToken, Content-Type")
		}

		switch r.Method {
		case "OPTIONS":
			w.WriteHeader(http.StatusNoContent) // 204
			return

		case "GET", "POST":
			logger.LogRequestDetails(log, "request details", r, logger.RequestDetailsOptions{PathContainsToken: false})

			// Извлекаем токен из заголовка
			token := r.Header.Get("X-ApiToken")
			if token == "" {
				http.Error(w, "The 'X-ApiToken' header is required", http.StatusBadRequest)
				return
			}
			escapeToken := url.QueryEscape(token)
			if escapeToken != token {
				http.Error(w, "garbage token: contains characters that need to be escaped", http.StatusBadRequest)
				return
			}
			r.Header.Del("X-ApiToken")

			// Восстанавливаем токен в пути запроса
			r.URL.Path = path.Join(r.URL.Path, token)

			proxy.ServeHTTP(w, r)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
