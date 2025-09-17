package logger

import (
	"log/slog"
	"math/rand/v2"
	"net/http"
	"runtime/debug"
	"time"
)

// HTTPLogging создает middleware для логирования HTTP-запросов. Принимает логгер
// и следующий обработчик в цепочке, возвращает новый обработчик с логированием.
func HTTPLogging(log *slog.Logger, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Генерируем уникальный ID для запроса и добавляем в логгер
		log := log.With("reqID", rand.Uint64())

		// Заменяем ResponseWriter на наш с хуком для логирования
		si := &statusInterceptor{
			ResponseWriter: w,
			log:            log,
		}

		// Добавляем логгер в контекст запроса
		origR := r
		ctx := Context(r.Context(), log)
		r = r.WithContext(ctx)

		start := time.Now()

		// Отлавливаем паники в обработчике
		defer func() {
			if p := recover(); p != nil {
				logRequest(log, "panicked request", origR, time.Since(start), si.status)
				log.Error("*** panic recovered ***",
					"panic", p,
					"stack", debug.Stack())
				http.Error(w, "internal error", 500)
			}
		}()

		// Передаем управление следующему обработчику
		h.ServeHTTP(si, r)
		logRequest(log, "request", origR, time.Since(start), si.status)
	})
}

// statusInterceptor логирует HTTP статусы и перехватывает ошибки
type statusInterceptor struct {
	http.ResponseWriter
	log    *slog.Logger
	status int // 0 = не установлен, 1xx = информационные, 2xx-5xx = основной статус
}

func (si *statusInterceptor) WriteHeader(status int) {
	switch {
	case status >= 100 && status < 200:
		si.log.Debug("informational status", "status", status)
		si.ResponseWriter.WriteHeader(status)

	case si.status == 0:
		si.status = status
		si.ResponseWriter.WriteHeader(status)

	case si.status != status:
		si.log.Warn("status code conflict", "origStatus", si.status, "newStatus", status)

	default:
		si.log.Warn("redundant WriteHeader call", "status", status)
	}
}

func (si *statusInterceptor) Write(b []byte) (int, error) {
	// NOTE: ResponseWriter гарантирует автоматический WriteHeader(200) при необходимости
	n, err := si.ResponseWriter.Write(b)
	if err != nil {
		si.log.Error("write failed", "error", err)
	}
	return n, err
}
