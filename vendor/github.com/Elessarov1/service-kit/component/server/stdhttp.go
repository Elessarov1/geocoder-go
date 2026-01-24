package server

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// RouteRegistrar registers service routes into mux.
type RouteRegistrar func(ctx context.Context, mux *http.ServeMux) error

// SwaggerAssets provides optional swagger UI FS and swagger YAML bytes.
type SwaggerAssets struct {
	UIFS     fs.FS
	UISubdir string
	YAML     []byte
}

// StdOptions configures standard HTTP runtime.
type StdOptions struct {
	Register RouteRegistrar

	// Optional.
	Swagger SwaggerAssets

	// Optional.
	// Logger is used by the runtime to log lifecycle events.
	Logger func(msg string, kv ...any)
}

// StdHTTPRuntime is a standard net/http server runtime.
type StdHTTPRuntime struct {
	cfg  Config
	opts StdOptions

	srv  *http.Server
	done chan error
}

func NewStdHTTPRuntime(ctx context.Context, cfg Config, opts StdOptions) (*StdHTTPRuntime, error) {
	if opts.Register == nil {
		return nil, fmt.Errorf("server.stdhttp: Register is required")
	}

	mux := http.NewServeMux()

	// Register service routes.
	if err := opts.Register(ctx, mux); err != nil {
		return nil, err
	}

	// Metrics.
	if cfg.Metrics.Enabled {
		path := cfg.Metrics.Path
		if path == "" {
			path = "/metrics"
		}
		mux.Handle(path, promhttp.Handler())
	}

	// Swagger.
	if cfg.Swagger.Enabled {
		if len(opts.Swagger.YAML) > 0 {
			yamlPath := cfg.Swagger.YAMLPath
			if yamlPath == "" {
				yamlPath = "/swagger.yaml"
			}
			mux.HandleFunc(yamlPath, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(opts.Swagger.YAML)
			})
		}

		if opts.Swagger.UIFS != nil {
			uiPath := cfg.Swagger.UIPath
			if uiPath == "" {
				uiPath = "/swagger-ui/"
			}
			sub := opts.Swagger.UISubdir
			if sub == "" {
				sub = "."
			}
			uiFS, err := fs.Sub(opts.Swagger.UIFS, sub)
			if err != nil {
				return nil, fmt.Errorf("server.stdhttp: swagger ui sub fs: %w", err)
			}
			mux.Handle(uiPath, http.StripPrefix(uiPath, http.FileServer(http.FS(uiFS))))
		}
	}

	// CORS.
	var h http.Handler = mux
	if cfg.CORS.Enabled {
		h = corsMiddleware(h)
	}

	addr := net.JoinHostPort(cfg.Host, fmt.Sprintf("%d", cfg.Port))
	rt := &StdHTTPRuntime{
		cfg:  cfg,
		opts: opts,
		srv: &http.Server{
			Addr:    addr,
			Handler: h,
		},
		done: make(chan error, 1),
	}

	if opts.Logger != nil {
		opts.Logger("HTTP server configured",
			"addr", addr,
			"cors_enabled", cfg.CORS.Enabled,
			"metrics_enabled", cfg.Metrics.Enabled,
			"metrics_path", cfg.Metrics.Path,
			"swagger_enabled", cfg.Swagger.Enabled,
			"swagger_yaml_path", cfg.Swagger.YAMLPath,
			"swagger_ui_path", cfg.Swagger.UIPath,
			"shutdown_timeout", cfg.ShutdownTimeout.String(),
		)
	}
	return rt, nil
}

func (r *StdHTTPRuntime) ListenAndServe(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		_ = r.Shutdown(context.Background())
	}()

	err := r.srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		r.done <- err
		return err
	}
	r.done <- nil
	return nil
}

// Shutdown is optional method that service-kit will call on Stop().
func (r *StdHTTPRuntime) Shutdown(ctx context.Context) error {
	timeout := r.cfg.ShutdownTimeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if r.opts.Logger != nil {
		r.opts.Logger("Shutting down HTTP server",
			"addr", r.srv.Addr,
			"timeout", timeout.String(),
		)
	}

	return r.srv.Shutdown(shutdownCtx)
}

func (r *StdHTTPRuntime) Done() <-chan error {
	return r.done
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
