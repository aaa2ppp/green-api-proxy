package proxy

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"green-api-proxy/internal/logger"
	"green-api-proxy/internal/utils"
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
			// Скрываем цепочку проксирования
			clientIP := utils.GetClientIP(r)
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
