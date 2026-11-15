package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"

	"github.com/viral32111/carbon-intensity-exporter/internal/carbon"
	"github.com/viral32111/carbon-intensity-exporter/internal/config"
	"github.com/viral32111/carbon-intensity-exporter/internal/flags"
	"github.com/viral32111/carbon-intensity-exporter/internal/log"
	"github.com/viral32111/carbon-intensity-exporter/internal/metrics"
	"github.com/viral32111/carbon-intensity-exporter/internal/server"
)

// -ldflags="-X main.VERSION=X.X.X"
var VERSION = "0.0.0"

func refreshMetrics() error {
	data, err := carbon.FetchRegionalData()
	if err != nil {
		return fmt.Errorf("Failed to fetch carbon intensity data: %w", err)
	}

	for _, regionalData := range data {
		for _, region := range regionalData.Regions {
			regionLabels := prometheus.Labels{
				"region":   region.Name,
				"operator": region.Operator,
			}

			metrics.PeriodFrom.Set(float64(regionalData.From.Unix()), regionLabels)
			metrics.PeriodTo.Set(float64(regionalData.To.Unix()), regionLabels)

			metrics.IntensityForecast.With(prometheus.Labels{
				"region":   region.Name,
				"operator": region.Operator,
				"index":    region.Intensity.Index,
			}).Set(float64(region.Intensity.Forecast))

			for _, generation := range region.GenerationMix {
				metrics.GenerationMix.With(prometheus.Labels{
					"region":   region.Name,
					"operator": region.Operator,
					"fuel":     generation.Fuel,
				}).Set(generation.Percentage)
			}
		}
	}

	slog.Info("Refreshed carbon intensity metrics", "data", data)

	return nil
}

func calculateNextUpdateTime() time.Time {
	now := time.Now()

	// Truncate to the last 30 minute boundary, then 31 minutes to get the next :01 or :31
	next := now.Truncate(30 * time.Minute).Add(31 * time.Minute)

	// If that time is already in the past (we're past :31), jump ahead another 30 min
	if !next.After(now) {
		next = next.Add(30 * time.Minute)
	}

	return next
}

func main() {
	isLocalDevelopment := VERSION == "0.0.0"

	/***********************/

	flagResult := flags.Parse()

	if flagResult.Version {
		fmt.Println(VERSION)
		os.Exit(0)
	}

	if flagResult.Usage {
		flags.PrintUsage(os.Stdout)
		os.Exit(0)
	}

	/***********************/

	godotenv.Load()

	config, err := config.GetConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to load configuration: %s\n", err)
		os.Exit(1)
	}

	/***********************/

	logger := log.NewLogger(config.LogLevel, isLocalDevelopment)

	logger.Info("Hello World", "version", VERSION)

	/***********************/

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	group, ctx := errgroup.WithContext(ctx)

	/***********************/

	httpServer, err := server.New(config)
	if err != nil {
		logger.Error("Unable to create HTTP server", "error", err)
		os.Exit(1)
	}

	prometheusRegistry := metrics.NewRegistry(config.MetricsGoBuiltin)
	metrics.Register(prometheusRegistry)

	// Serve metrics on the configured path
	httpServer.HTTPMultiplexer.Handle(config.MetricsPath, metrics.Handler(prometheusRegistry))

	// Start HTTP server in the background
	group.Go(func() error {
		listenAddress := fmt.Sprintf("%s:%d", config.HTTPAddress, config.HTTPPort)
		if err := httpServer.Start(ctx, listenAddress); err != nil {
			return fmt.Errorf("HTTP server error: %w", err)
		}

		return nil
	})

	/***********************/

	group.Go(func() error {
		// Fire immediately, don't wait for the first tick
		if err := refreshMetrics(); err != nil {
			logger.Warn("Failed to refresh carbon intensity metrics", "error", err)
		}

		for {
			// Update every 30 minutes on the clock to align with the update schedule of the API
			nextUpdateTime := calculateNextUpdateTime()
			logger.Debug("Next carbon intensity data refresh scheduled", "time", nextUpdateTime)

			timer := time.NewTimer(time.Until(nextUpdateTime))

			select {
			case <-timer.C:
				if err := refreshMetrics(); err != nil {
					logger.Warn("Failed to refresh carbon intensity metrics", "error", err)
				}

			case <-ctx.Done():
				timer.Stop()
				logger.Info("Stopped refreshing carbon intensity data")
				return nil
			}
		}
	})

	/***********************/

	if err := group.Wait(); err != nil {
		logger.Error("Error in execution group", "error", err)
		os.Exit(1)
	}
}
