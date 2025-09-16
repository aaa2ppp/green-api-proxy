package logger

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

const (
	requestDetailsLogLevel = slog.LevelDebug - 1
	sensitiveDataMask      = "*****"
)

type RequestDetailsOptions struct {
	PathContainsToken bool
}

func LogRequestDetails(log *slog.Logger, tag string, r *http.Request, opts RequestDetailsOptions) {
	if !slog.Default().Enabled(context.Background(), requestDetailsLogLevel) {
		return
	}

	var (
		uri  = r.URL.String()
		path = r.URL.Path
	)

	if opts.PathContainsToken {
		token := extractTokenFromPath(path)
		uri = strings.ReplaceAll(uri, url.PathEscape(token), sensitiveDataMask)
		path = strings.ReplaceAll(path, token, sensitiveDataMask)
	}

	log.Log(context.Background(), requestDetailsLogLevel, tag,
		"method", r.Method,
		"url", uri,
		"path", path,
		"query", r.URL.RawQuery,
		"host", r.Host,
		"remote_addr", r.RemoteAddr,
		"content_length", r.ContentLength,
		"headers", maskSensitiveHeaders(r.Header),
	)
}

// extractTokenFromPath (если путь содержит Token, то это последний сегмент)
// возвращает тупо все после поледнего '/', если '/' нет, то возвращает всю строку.
// Если path заканчивается на '/', последнй '/' предварительно усекаться.
func extractTokenFromPath(path string) string {
	if n := len(path); n > 0 && path[n-1] == '/' {
		path = path[:n-1]
	}

	p := strings.LastIndexByte(path, '/') + 1
	return path[p:]
}

func maskSensitiveHeaders(headers http.Header) http.Header {
	headers = headers.Clone()
	for key, values := range headers {
		for i := range values {
			if strings.Contains(strings.ToLower(key), "token") ||
				strings.Contains(strings.ToLower(key), "auth") ||
				strings.Contains(strings.ToLower(key), "cookie") {
				values[i] = sensitiveDataMask
			}
		}
	}
	return headers
}
