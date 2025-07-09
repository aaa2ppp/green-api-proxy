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

			slog.Debug("redirect", "req.URL:", req.URL)

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
			// Разрешаем CORS только для нужных доменов
			w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

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
