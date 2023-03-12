package main

import (
	"context"

	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"prometheus-sabnzbd-exporter/internal/config"
	"prometheus-sabnzbd-exporter/internal/exporter"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	// go build -ldflags="-X \"main.version=${VERSION}\"" ./cmd/awair-exporter/awair-exporter.go
	appName = "sabnzbd-exporter"
	version = "x.x.x"
)

var (
	infoMetricOpts = prometheus.GaugeOpts{
		Namespace: "exporter",
		Name:      "info",
		Help:      "Info about this sabnzbd-exporter",
	}
)

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

func newHealthCheckHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, "OK")
	})
}

func main() {
	cfg, err := config.LoadConfig(appName, os.Args[1:])
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}
	err = cfg.Validate()
	if err != nil {
		log.Fatal().Err(err).Msg("Invalid config")
	}
	logLevel, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		// Sanity Check, should be unreachable due to validation.
		log.Fatal().Err(err).Msg("Invalid log level")
	}
	zerolog.SetGlobalLevel(logLevel)

	var srv http.Server

	idleConnsClosed := make(chan struct{})
	go func() {
		sigchan := make(chan os.Signal, 1)

		signal.Notify(sigchan, os.Interrupt)
		signal.Notify(sigchan, syscall.SIGTERM)
		sig := <-sigchan
		log.Info().
			Str("signal", sig.String()).
			Msg("Stopping in response to signal")
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatal().Err(err).Msg("Failed to gracefully close http server")
		}
		close(idleConnsClosed)
	}()

	log.Info().
		Str("app_name", appName).
		Str("version", version).
		Str("listen_port", cfg.ListenPort).
		Str("base_url", cfg.BaseURL).
		Msg("Exporter Started.")

	ex, err := exporter.NewSabnzbdExporter(cfg.BaseURL, cfg.ApiKey)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed to build SabnzbD Collector.")
	}

	infoMetricOpts.ConstLabels = prometheus.Labels{
		"app_name": appName,
		"version":  version,
		"base_url": cfg.BaseURL,
	}

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(
		prometheus.NewGaugeFunc(
			infoMetricOpts,
			func() float64 { return 1 },
		),
		ex,
	)
	if cfg.GoCollector {
		reg.MustRegister(collectors.NewGoCollector())
	}
	if cfg.ProcessCollector {
		reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	}
	router := http.NewServeMux()
	router.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	router.Handle("/healthz", newHealthCheckHandler())
	srv.Addr = fmt.Sprintf(":%s", cfg.ListenPort)
	srv.Handler = router
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("Failed to start HTTP Server")
	}
	<-idleConnsClosed
}
