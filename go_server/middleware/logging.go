package middleware

import (
	"log"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// 记录请求信息，包括查询参数
		queryParams := r.URL.Query().Encode()
		if queryParams != "" {
			log.Printf("→ [%s] %s?%s from %s", r.Method, r.URL.Path, queryParams, r.RemoteAddr)
		} else {
			log.Printf("→ [%s] %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		}
		
		wrapped := &responseWriter{
			ResponseWriter: w,
			status:        http.StatusOK,
		}
		
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("✗ PANIC [%s] %s: %v", r.Method, r.URL.Path, rec)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}()
		
		next(wrapped, r)
		
		duration := time.Since(start)
		statusIcon := "✓"
		if wrapped.status >= 400 {
			statusIcon = "✗"
		}
		
		log.Printf("%s [%s] %s - Status: %d, Size: %d bytes, Duration: %v",
			statusIcon, r.Method, r.URL.Path, wrapped.status, wrapped.size, duration)
	}
}
