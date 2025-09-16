package server

import (
	"context"
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

func New(cfg *Config, proxy http.Handler, fs fs.FS) (*http.Server, error) {
	if ao := cfg.AllowOrigin; ao != "" && ao != "*" {
		if _, err := url.Parse(ao); err != nil {
			return nil, fmt.Errorf("can't parse AllowOrigin: %w", err)
		}
	}

	// Настраиваем статику
	static := http.NewServeMux()
	static.Handle("GET /green-api/", http.StripPrefix("/green-api", http.FileServerFS(fs)))

	mux := http.NewServeMux()
	mux.Handle("/green-api/", static)

	// Настраиваем proxy
	mux.Handle("/green-api/proxy/", http.StripPrefix("/green-api/proxy",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
				logger.LogRequestDetails(log, "new request details", r, logger.RequestDetailsOptions{PathContainsToken: false})

				ctx, cancel := context.WithTimeout(r.Context(), cfg.RedirectTimeout)
				defer cancel()

				r = r.Clone(ctx) // чтобы не изменять исходный запрос

				// Получаем токен из заголовка
				token := r.Header.Get("X-ApiToken")
				if token == "" {
					http.Error(w, "The 'X-ApiToken' header is required", http.StatusBadRequest)
					return
				}
				r.Header.Del("X-ApiToken")

				// Восстанавливаем токен в пути запроса.
				escapeToken := url.QueryEscape(token)
				if escapeToken != token {
					http.Error(w, "garbage token: contains characters that need to be escaped", http.StatusBadRequest)
					return
				}
				r.URL.Path = path.Join(r.URL.Path, token)

				proxy.ServeHTTP(w, r)
				return
			}

			log.Warn("Method not allowed", "from", r.RemoteAddr, "method", r.Method, "url", r.URL)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}),
	))

	return &http.Server{
		Addr:         cfg.Addr,
		Handler:      logger.HTTPLogging(slog.Default(), mux),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}, nil
}
