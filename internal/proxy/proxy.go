package proxy

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"green-api-proxy/internal/logger"
)

const hideToken = true

type Config struct {
	GreenApiBaseURL string
}

func New(cfg *Config) (*httputil.ReverseProxy, error) {

	target, err := url.Parse(cfg.GreenApiBaseURL)

	if err != nil {
		return nil, fmt.Errorf("can't parse GreenApiBaseURL: %w", err)
	}

	return &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			// Получаем IP колиента (или что сможем)
			var clientIP string
			if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
				// Берём первый IP в цепочке (до первой запятой)
				p := strings.IndexByte(xff, ',')
				if p == -1 {
					p = len(xff)
				}
				clientIP = strings.TrimSpace(xff[:p])
			} else {
				// Если пустой - используем RemoteAddr
				clientIP = r.RemoteAddr
				// Вырезаем только IP часть
				if host, _, err := net.SplitHostPort(clientIP); err == nil {
					clientIP = host
				}
			}

			// Скрываем цепочку проксирования
			r.Header.Set("X-Real-IP", clientIP)
			r.Header.Set("X-Forwarded-For", clientIP)

			if r.Header.Get("X-Forwarded-Host") == "" {
				r.Header.Set("X-Forwarded-Host", r.Host)
			}

			if r.Header.Get("X-Forwarded-Proto") == "" {
				if r.TLS != nil {
					r.Header.Set("X-Forwarded-Proto", "https")
				} else {
					r.Header.Set("X-Forwarded-Proto", "http")
				}
			}

			// Перенаправляем запрос на настоящий GREEN-API
			r.URL.Scheme = target.Scheme
			r.URL.Host = target.Host
			r.Host = target.Host

			log := logger.FromContext(r.Context())
			logger.LogRequestDetails(log, "redirect details", r, logger.RequestDetailsOptions{PathContainsToken: hideToken})
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: false}, // проверяем TLS
		},
	}, nil
}
