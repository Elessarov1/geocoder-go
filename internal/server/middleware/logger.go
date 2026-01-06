package middleware

import (
	"Geocoder/internal/common/logger"
	"net"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
	cors   bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.cors {
		rw.ResponseWriter.Header().Add("Access-Control-Allow-Credentials", "true")
		rw.ResponseWriter.Header().Add("Access-Control-Allow-Headers", "*")
		rw.ResponseWriter.Header().Add("Access-Control-Allow-Origin", "*")
	}

	rw.ResponseWriter.WriteHeader(code)
	rw.status = code
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.size += n
	return n, err
}

func LoggerMiddleware(lg *zap.Logger, cors bool) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			ctx := r.Context()
			ctx = logger.WithLogger(ctx, lg)

			start := time.Now()
			ip := getIP(r)

			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK, cors: cors}
			next.ServeHTTP(rw, r.WithContext(ctx))

			path := r.URL.Path
			if path == "/v1/health" || path == "/metrics" {
				return
			}

			lg.Debug("Request",
				zap.String("method", r.Method),
				zap.Int("status", rw.status),
				zap.String("path", path),
				zap.Int("response_size", rw.size),
				zap.Duration("duration", time.Since(start)),
				zap.String("client_ip", ip),
			)
		})
	}
}

func getIP(r *http.Request) string {
	var ip string

	// X-Forwarded-For
	ip = r.Header.Get("X-Forwarded-For")
	ci := strings.Index(ip, ",")
	if ci == -1 {
		ci = len(ip)
	}
	ip = strings.TrimSpace(ip[:ci])
	if ip != "" && net.ParseIP(ip) != nil {
		return ip
	}

	// X-Real-IP
	ip = r.Header.Get("X-Real-IP")
	if ip != "" && net.ParseIP(ip) != nil {
		return ip
	}

	// True-Client-IP
	ip = r.Header.Get("True-Client-IP")
	if ip != "" && net.ParseIP(ip) != nil {
		return ip
	}

	// Socket
	ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	return ip
}
