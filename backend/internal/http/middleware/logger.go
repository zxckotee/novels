package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

// Logger возвращает middleware для логирования запросов
func Logger(log zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Оборачиваем ResponseWriter для получения статус-кода
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// Получаем request ID
			reqID := middleware.GetReqID(r.Context())

			defer func() {
				log.Info().
					Str("request_id", reqID).
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Str("remote_addr", r.RemoteAddr).
					Int("status", ww.Status()).
					Int("bytes", ww.BytesWritten()).
					Dur("duration", time.Since(start)).
					Msg("request completed")
			}()

			next.ServeHTTP(ww, r)
		})
	}
}
