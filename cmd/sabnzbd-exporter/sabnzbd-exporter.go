package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"prometheus-sabnzbd-exporter/internal/exporter"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	// go build -ldflags="-X \"main.version=${VERSION}\"" ./cmd/awair-exporter/awair-exporter.go
	app_name = "sabnzbd-exporter"
	version  = "x.x.x"
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
	debug := flag.Bool("debug", false, "sets log level to debug")
	goCollector := flag.Bool("gocollector", false, "enables go stats exporter")
	processCollector := flag.Bool("processcollector", false, "enables process stats exporter")
	flag.Parse()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	err := godotenv.Load(".env")
	if err != nil {
		// Typical use will be via direct env in kubernetes,
		// don't fail here.
		log.Warn().Err(err).Msg("No .env file loaded")
	}

	// TODO: validation
	base_url := os.Getenv("SABNZBD_BASE_URL")
	if base_url == "" {
		log.Fatal().
			Msg("SABNZBD_BASE_URL must be set to the base url of your sabnzbd instance.")
	}

	api_key := os.Getenv("SABNZBD_API_KEY")
	if api_key == "" {
		log.Fatal().
			Msg("SABNZBD_API_KEY must be set to the api key of your sabnzbd instance.")
	}

	listen_port := os.Getenv("SABNZBD_EXPORTER_PORT")
	if listen_port == "" {
		listen_port = "8080"
	}

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
		Str("app_name", app_name).
		Str("version", version).
		Str("listen_port", listen_port).
		Msg("Exporter Started.")

	ex, err := exporter.NewSabnzbdExporter(base_url, api_key)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed to connect to Awair device.")
	}

	infoMetricOpts.ConstLabels = prometheus.Labels{
		"app_name": app_name,
		"version":  version,
		"base_url": base_url,
	}

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(
		prometheus.NewGaugeFunc(
			infoMetricOpts,
			func() float64 { return 1 },
		),
		ex,
	)
	if *goCollector {
		reg.MustRegister(collectors.NewGoCollector())
	}
	if *processCollector {
		reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	}
	router := http.NewServeMux()
	router.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	router.Handle("/healthz", newHealthCheckHandler())
	srv.Addr = fmt.Sprintf(":%s", listen_port)
	srv.Handler = router
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("Failed to start HTTP Server")
	}
	<-idleConnsClosed
}
