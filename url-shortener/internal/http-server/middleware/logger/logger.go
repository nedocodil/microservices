package logger

import (
	"net/http"
	"time"

	"golang.org/x/exp/slog"

	"github.com/go-chi/chi/v5/middleware"
)

func New(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		log = log.With(slog.String("component", "middleware/logger"))

		log.Info("logger middleware enabled")

		fn := func(w http.ResponseWriter, r *http.Request) {
			entry := log.With(
				slog.String("http.method", r.Method),
				slog.String("http.path", r.URL.Path),
				slog.String("http.remote_addr", r.RemoteAddr),
				slog.String("http.user_agent", r.UserAgent()),
				slog.String("http.req_id", middleware.GetReqID(r.Context())),
			)
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t1 := time.Now()
			defer func() {
				entry.Info("request completed",
					slog.Int("status", ww.Status()),
					slog.Int("bytes", ww.BytesWritten()),
					slog.String("duration", time.Since(t1).String()),
				)
			}()
		
			next.ServeHTTP(ww, r)
		}
		
		return http.HandlerFunc(fn)
	}
}

