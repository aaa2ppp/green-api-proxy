package main

import (
	"cmp"
	"crypto/tls"
	"log"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strings"

	"green-api-proxy/www"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	serverAddr := cmp.Or(os.Getenv("SERVER_ADDR"), ":8087")
	allowOrigin := os.Getenv("ALLOW_ORIGIN")
	greenAPIBaseURL := cmp.Or(os.Getenv("GREEN_API_BASE_URL"), "https://api.green-api.com")

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
	http.Handle("/green-api/", http.StripPrefix("/green-api", http.FileServerFS(www.Static)))

	http.Handle("/green-api-proxy/", http.StripPrefix("/green-api-proxy",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Восстанавливаем ApiToken в пути запроса из заголовка
			apiToken := r.Header.Get("X-ApiToken")
			if apiToken == "" {
				http.Error(w, "The 'X-ApiToken' header is required", http.StatusBadRequest)
				return
			}
			r = r.Clone(r.Context()) // чтобы не изменять исходный запрос
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

	// Запускаем сервер
	log.Println("Сервер запущен на :8087 (HTTPS)")
	log.Fatal(http.ListenAndServeTLS(serverAddr, "./cert/cert.pem", "./cert/key.pem", nil))
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
