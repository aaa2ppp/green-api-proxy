package main

import (
	"cmp"
	"context"
	"crypto/tls"
	"log"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"green-api-proxy/www"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	setupLogger()

	var (
		serverAddr      = cmp.Or(os.Getenv("SERVER_ADDR"), ":8087")
		allowOrigin     = os.Getenv("ALLOW_ORIGIN")
		greenAPIBaseURL = cmp.Or(os.Getenv("GREEN_API_BASE_URL"), "https://api.green-api.com")
		readTimeout     = getTimeout("READ_TIMEOUT", 1*time.Second)
		writeTimeout    = getTimeout("WRITE_TIMEOUT", 3*time.Second)
		redirectTimeout = getTimeout("REDIRECT_TIMEOUT", 2*time.Second)
	)

	// Настраиваем прокси
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			// Перенаправляем запрос на настоящий GREEN-API
			target, _ := url.Parse(greenAPIBaseURL)
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.Host = target.Host

			slog.Debug("redirect", "toURL", maskApiToken(req.URL.String()))

			// Добавляем заголовки для CORS
			req.Header.Set("X-Forwarded-Host", req.Host)
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: false}, // проверяем TLS
		},
	}

	// Настраиваем сервер

	mux := http.NewServeMux()
	mux.Handle("/green-api/", http.StripPrefix("/green-api", http.FileServerFS(www.Static)))

	mux.Handle("/green-api-proxy/", http.StripPrefix("/green-api-proxy",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), redirectTimeout*time.Millisecond)
			defer cancel()

			// Восстанавливаем ApiToken в пути запроса из заголовка
			apiToken := r.Header.Get("X-ApiToken")
			if apiToken == "" {
				http.Error(w, "The 'X-ApiToken' header is required", http.StatusBadRequest)
				return
			}
			r = r.Clone(ctx) // чтобы не изменять исходный запрос
			r.Header.Del("X-ApiToken")
			r.URL.Path = path.Join(r.URL.Path, apiToken)

			// Разрешаем CORS только для нужных доменов
			w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "X-ApiToken, Content-Type")

			switch r.Method {
			case "OPTIONS":
				w.WriteHeader(http.StatusNoContent) // 204
				return
			case "GET", "POST":
				proxy.ServeHTTP(w, r)
				return
			}

			slog.Warn("Method not allowed", "from", r.RemoteAddr, "method", r.Method, "url", r.URL)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}),
	))

	server := &http.Server{
		Addr:         serverAddr,
		Handler:      mux,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	// Запускаем сервер
	log.Printf("Сервер запущен на %s (HTTPS)", serverAddr)
	log.Fatal(server.ListenAndServeTLS("./cert/cert.pem", "./cert/key.pem"))
}

func maskApiToken(url string) string {
	if p := strings.LastIndex(url, "/"); p != -1 {
		return url[:p+1] + "***********"
	}
	return url
}

func setupLogger() {
	level := slog.LevelInfo
	if s, ok := os.LookupEnv("LOG_LEVEL"); ok {
		var v slog.Level
		if err := v.UnmarshalText([]byte(s)); err != nil {
			slog.Warn("can't parse LOG_LEVEL", "s", s)
		} else {
			level = v
		}
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
	logger.Info("setup logger", "level", level)
	slog.SetDefault(logger)
}

func getTimeout(env string, defaultValue time.Duration) time.Duration {
	s, ok := os.LookupEnv(env)
	if !ok {
		return defaultValue
	}
	v, err := time.ParseDuration(s)
	if err != nil {
		log.Printf("ignore %s: %v", env, err)
		return defaultValue
	}
	return v
}
