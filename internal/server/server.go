package server

import (
	"Geocoder"
	"Geocoder/internal/common/logger"
	"Geocoder/internal/config"
	"Geocoder/internal/geoip"
	"Geocoder/internal/server/middleware"
	"Geocoder/internal/server/oas"
	"context"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/oschwald/maxminddb-golang"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

const shutdownTimeout = 10 * time.Second

type GeoCoderServer struct {
	oas.UnimplementedHandler

	startTime time.Time
	server    *http.Server
	lg        *zap.Logger

	geo  *geoip.Store
	mmdb *maxminddb.Reader
}

var _ oas.Handler = (*GeoCoderServer)(nil) // static check

func NewServer(ctx context.Context, cfg *config.ServerConfig, geo *geoip.Store, mmdbPath string) (*GeoCoderServer, error) {
	lg := logger.FromContext(ctx).Named("server")

	db, err := maxminddb.Open(mmdbPath)
	if err != nil {
		return nil, fmt.Errorf("open mmdb for ip lookup: %w", err)
	}

	s := &GeoCoderServer{
		lg:        lg,
		startTime: time.Now(),
		geo:       geo,
		mmdb:      db,
	}

	// OpenAPI handler (ogen-generated)
	oasServer, err := oas.NewServer(s)
	if err != nil {
		return nil, fmt.Errorf("failed to create oas server: %w", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", oasServer)
	mux.Handle("/metrics", promhttp.Handler())

	// Swagger UI + yaml
	if cfg.Swagger {
		swaggerUI, err := fs.Sub(Geocoder.SwaggerUI, "_openapi/swaggerui")
		if err != nil {
			return nil, fmt.Errorf("failed to create swagger ui sub fs: %w", err)
		}
		mux.HandleFunc("/swagger.yaml", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write(Geocoder.Swagger)
		})
		mux.Handle("/swagger-ui/", http.StripPrefix("/swagger-ui/", http.FileServer(http.FS(swaggerUI))))
	}

	if cfg.CorsEnabled {
		lg.Info("CORS enabled")
	} else {
		lg.Info("CORS disabled")
	}

	handler := middleware.Wrap(mux, middleware.LoggerMiddleware(lg, cfg.CorsEnabled))

	s.server = &http.Server{
		Handler: handler,
		Addr:    net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
	}

	return s, nil
}

func (s *GeoCoderServer) ListenAndServe(ctx context.Context) error {
	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		s.lg.Info("Shutting down HTTP server", zap.Duration("timeout", shutdownTimeout))
		if err := s.server.Shutdown(shutdownCtx); err != nil {
			s.lg.Error("failed to shutdown server", zap.Error(err))
			return
		}
		s.lg.Info("HTTP server shutdown complete")
	}()

	s.lg.Info("Starting HTTP server", zap.String("addr", s.server.Addr))
	err := s.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *GeoCoderServer) Close() {
	_ = s.mmdb.Close()
	_ = s.server.Close()
	s.lg.Info("HTTP server closed")
}

func (s *GeoCoderServer) NewError(ctx context.Context, err error) *oas.DefaultErrorStatusCode {
	s.lg.Error("API request error", zap.Error(err))
	return ErrResponse(http.StatusInternalServerError, "internal.error", err.Error())
}
