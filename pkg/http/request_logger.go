package http

// Taken from:
// https://blog.questionable.services/article/guide-logging-middleware-go/

import (
	"net/http"
	"runtime/debug"
	"time"

	"github.com/sirupsen/logrus"
)

// responseWriter is a minimal wrapper for http.ResponseWriter that allows the
// written HTTP status code to be captured for logging.
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}

	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true

	return
}

// LoggingMiddleware logs the incoming HTTP request & its duration.
func LoggingMiddleware(logger logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					logger.WithFields(logrus.Fields{
						"err":   err,
						"trace": debug.Stack(),
					}).Info()
				}
			}()

			start := time.Now()
			wrapped := wrapResponseWriter(w)
			next.ServeHTTP(wrapped, r)
			logger.WithFields(logrus.Fields{
				"status":   wrapped.status,
				"method":   r.Method,
				"path":     r.URL.EscapedPath(),
				"duration": time.Since(start).String(),
			}).Info()
		}

		return http.HandlerFunc(fn)
	}
}
