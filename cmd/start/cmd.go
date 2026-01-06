package start

import (
	"context"
	"fmt"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/Elessarov1/geocoder-go/cmd"
	"github.com/Elessarov1/geocoder-go/internal/common/logger"
	"github.com/Elessarov1/geocoder-go/internal/config"
	"github.com/Elessarov1/geocoder-go/internal/geocoder_api"
	"github.com/Elessarov1/geocoder-go/internal/geoip"
	grpc_server "github.com/Elessarov1/geocoder-go/internal/gprc_server"
	"github.com/Elessarov1/geocoder-go/internal/server"

	"github.com/go-faster/errors"
	"github.com/oschwald/maxminddb-golang"
	"github.com/urfave/cli/v3"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type App struct {
	cfg        config.Config
	geoip      *geoip.Store
	httpServer *server.GeoCoderServer
}

func CmdStart() *cli.Command {
	app := &App{}
	return &cli.Command{
		Name:   "start",
		Usage:  "Start geocoder",
		Before: app.before,
		Action: app.action,
	}
}

func (app *App) before(ctx context.Context, _ *cli.Command) (context.Context, error) {
	var appCtx, cfg, err = cmd.ReadConfig(ctx)
	if err != nil {
		return ctx, err
	}
	app.cfg = cfg

	return appCtx, nil
}

func (app *App) action(ctx context.Context, _ *cli.Command) error {
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	cfg := app.cfg
	log := logger.FromContext(ctx)

	log.Info("Starting geocoder",
		zap.String("path", cfg.GeoCoder.GeoIPDbPath),
		zap.Bool("debug", cfg.GeoCoder.Debug),
	)

	logMem(log, "mem_before_load")

	store, err := app.loadGeoIP(ctx, cfg)
	if err != nil {
		return err
	}

	logMem(log, "mem_after_load")
	app.geoip = store

	st := store.Stats()
	log.Info("GeoIP database loaded",
		zap.Int("total_networks", st.TotalNetworks),
		zap.Int("unique_countries", st.UniqueCountries),
		zap.Int("ipv4_networks", st.V4Networks),
		zap.Int("ipv6_networks", st.V6Networks),
	)

	//runtime.GC()
	//logMem(log, "mem_after_load_after_gc")

	mmdb, err := maxminddb.Open(cfg.GeoCoder.GeoIPDbPath)
	if err != nil {
		return fmt.Errorf("open mmdb: %w", err)
	}
	defer mmdb.Close()

	api := geocoder_api.NewService(store, mmdb, time.Now())

	srv, err := server.NewServer(ctx, &cfg.Server, api)
	if err != nil {
		return fmt.Errorf("failed to create geocoder http server: %w", err)
	}
	defer srv.Close()

	grpcSrv, err := grpc_server.New(ctx, &cfg.GRPC, api)
	if err != nil {
		return fmt.Errorf("failed to create geocoder grpc server: %w", err)
	}
	defer grpcSrv.Close()

	app.httpServer = srv

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return srv.ListenAndServe(ctx) })
	g.Go(func() error { return grpcSrv.Serve(ctx) })

	g.Go(func() error {
		<-ctx.Done()
		log.Info("Stop signal received, gracefully shutting down", zap.Error(ctx.Err()))
		return ctx.Err()
	})

	if err := g.Wait(); err != nil {
		if !errors.Is(err, context.Canceled) {
			return err
		}
	}

	log.Info("Shutdown complete")
	return nil
}

func (app *App) loadGeoIP(ctx context.Context, cfg config.Config) (*geoip.Store, error) {
	opt := geoip.DefaultOptions()
	return geoip.Load(ctx, cfg.GeoCoder.GeoIPDbPath, opt)
}

func logMem(log *zap.Logger, prefix string) {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	log.Debug(prefix,
		zap.Uint64("heap_alloc_bytes", ms.HeapAlloc),
		zap.Float64("heap_alloc_mb", float64(ms.HeapAlloc)/1024/1024),

		zap.Uint64("heap_inuse_bytes", ms.HeapInuse),
		zap.Float64("heap_inuse_mb", float64(ms.HeapInuse)/1024/1024),

		zap.Uint64("sys_bytes", ms.Sys),
		zap.Float64("sys_mb", float64(ms.Sys)/1024/1024),

		zap.Uint32("gc_cycles", ms.NumGC),
	)
}
