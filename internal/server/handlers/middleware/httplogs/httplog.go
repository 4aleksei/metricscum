package httplogs

import (
	"net/http"
	"time"

	"github.com/4aleksei/metricscum/internal/common/logger"
	"go.uber.org/zap"
)

type (
	responseData struct {
		status int
		size   int
	}
	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func WithLogging(next http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		next.ServeHTTP(&lw, r)
		duration := time.Since(start)
		logger.Log.Info("got incoming HTTP request",
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.String("AcceptEnc", r.Header.Get("Accept-Encoding")),
			zap.String("ContentEnc", r.Header.Get("Content-Encoding")),
			zap.String("Accept", r.Header.Get("Accept")),
			zap.String("ContentType", r.Header.Get("Content-Type")),
			zap.Duration("duration", duration),
			zap.Int("resp_status", responseData.status),
			zap.Int("resp_size", responseData.size))
	}
	return http.HandlerFunc(logFn)
}
