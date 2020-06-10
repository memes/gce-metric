// Copyright 2020 Matthew Emes. All rights reserved.
// Use of this source code is governed by the MIT license included in the source.
package main

import (
	"context"
	"flag"
	"math"
	"os"
	"os/signal"
	"time"

	"github.com/peterbourgon/ff/v3/ffcli"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	var (
		verbose     bool
		sample      time.Duration
		period      time.Duration
		floor       float64
		ceiling     float64
		round       bool
		metricType  string
		projectID   string
		ticker      *time.Ticker
		output      chan float64
		rootFlagSet *flag.FlagSet
	)
	rootFlagSet = flag.NewFlagSet("gce-metric", flag.ExitOnError)
	rootFlagSet.BoolVar(&verbose, "verbose", false, "enable DEBUG log level")
	rootFlagSet.DurationVar(&sample, "sample", 60*time.Second, "sample time, specified in Go duration format. Default value is 60s")
	rootFlagSet.DurationVar(&period, "period", 10*time.Minute, "the period of the underlying wave function; e.g. the time for a complete cycle from floor to ceiling and back.")
	rootFlagSet.Float64Var(&floor, "floor", 1.0, "the lowest value to send to metric; e.g. '1.0'")
	rootFlagSet.Float64Var(&ceiling, "ceiling", 10.0, "the maximum value to send to metics; e.g. '10.0'")
	rootFlagSet.BoolVar(&round, "round", false, "Round metric value to nearest integer")
	rootFlagSet.StringVar(&metricType, "type", "custom.googleapis.com/gce_metric", "The custom metric type to use; e.g. custom.googleapis.com/gce_metric")
	rootFlagSet.StringVar(&projectID, "project", "", "the GCP project id to use; specify if not running on GCE or to override project id")
	logger, level := initLogger()
	defer func() {
		_ = logger.Sync()
	}()
	logger.Debug("Starting application")
	sawtooth := &ffcli.Command{
		Name:       "sawtooth",
		ShortUsage: "gce-metric [flags] sawtooth",
		ShortHelp:  "generate metrics that match a sawtooth wave function",
		Exec: func(ctx context.Context, _ []string) error {
			generator := NewSawtoothGenerator(logger, period)
			return generator(ctx, ticker, output)
		},
	}
	sine := &ffcli.Command{
		Name:       "sine",
		ShortUsage: "gce-metric [flags] sine",
		ShortHelp:  "generate metrics that match a sine wave function",
		Exec: func(ctx context.Context, _ []string) error {
			generator := NewSineGenerator(logger, period)
			return generator(ctx, ticker, output)
		},
	}
	square := &ffcli.Command{
		Name:       "square",
		ShortUsage: "gce-metric [flags] square",
		ShortHelp:  "generate metrics that match a square wave function",
		Exec: func(ctx context.Context, _ []string) error {
			generator := NewSquareGenerator(logger, period)
			return generator(ctx, ticker, output)
		},
	}
	triangle := &ffcli.Command{
		Name:       "triangle",
		ShortUsage: "gce-metric [flags] triangle",
		ShortHelp:  "generate metrics that match a triangle wave function",
		Exec: func(ctx context.Context, _ []string) error {
			generator := NewTriangleGenerator(logger, period)
			return generator(ctx, ticker, output)
		},
	}
	root := &ffcli.Command{
		ShortUsage:  "gce-metric [flags] <subcommand>",
		Subcommands: []*ffcli.Command{sawtooth, sine, square, triangle},
		FlagSet:     rootFlagSet,
		Exec: func(context.Context, []string) error {
			return flag.ErrHelp
		},
	}
	if err := root.Parse(os.Args[1:]); err != nil {
		logger.Fatalw("Error parsing command line",
			"error", err,
		)
		os.Exit(1)
	}
	if verbose {
		level.SetLevel(zapcore.DebugLevel)
	}
	ticker = time.NewTicker(sample)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		logger.Debug("Signal handler started, waiting for signal")
		sig := <-interrupt
		ticker.Stop()
		logger.Infow("Interrupt received, cancelling context",
			"sig", sig,
		)
		cancel()
		logger.Debug("Signal handler exit")
	}()
	output = make(chan float64, 1)
	go func(ctx context.Context) {
		logger.Debug("Output handler started")
		delta := ceiling - floor
		metricConfig, err := NewMetricConfig(DefaultMetricConfig(), Round(round), MetricType(metricType), ProjectID(projectID))
		if err != nil {
			logger.Errorw("Error configuring metrics client",
				"err", err,
			)
			// Bail out, but make sure all other processes get terminated
			interrupt <- os.Interrupt
			return
		}
		if err := metricConfig.Validate(); err != nil {
			logger.Errorw("Error validating metrics client configuration",
				"err", err,
			)
			// Bail out, but make sure all other processes get terminated
			interrupt <- os.Interrupt
			return
		}
		requestGenerator := NewCreateTimeSeriesRequestGenerator(logger, metricConfig)
		for {
			select {
			case <-ctx.Done():
				logger.Infow("Output has been cancelled")
				return
			case factor := <-output:
				value := delta*factor + floor
				logger.Infow("Value updated",
					"factor", factor,
					"value", value,
				)
				if err = sendMetric(ctx, requestGenerator(value)); err != nil {
					logger.Errorw("Error sending metric to GCP",
						"err", err,
					)
				}
			}
		}
	}(ctx)
	if err := root.Run(ctx); err != nil {
		logger.Fatalw("Error running command",
			"error", err,
		)
		os.Exit(1)
	}
	logger.Debug("Exit main")
}

// initLogger returns a sugared Zap logger and a level that can be adjusted at run-time
func initLogger() (*zap.SugaredLogger, zap.AtomicLevel) {
	config := zap.NewProductionEncoderConfig()
	encoder := zapcore.NewJSONEncoder(config)
	level := zap.NewAtomicLevel()
	logger := zap.New(zapcore.NewCore(encoder, zapcore.Lock(os.Stdout), level))
	return logger.Sugar(), level
}

// NewSawtoothGenerator returns a cancellable function that will generate a value
// between 0.0 and 1.0 using a sawtooth wave function.
func NewSawtoothGenerator(logger *zap.SugaredLogger, period time.Duration) func(context.Context, *time.Ticker, chan float64) error {
	tzero := time.Now()
	periodSecs := period.Seconds()
	sawtoothLogger := logger.With(
		"generator", "sawtooth",
		"period", period,
	)
	return func(ctx context.Context, ticker *time.Ticker, output chan float64) error {
		for {
			select {
			case <-ctx.Done():
				sawtoothLogger.Infow("Function has been cancelled")
				return nil
			case tick := <-ticker.C:
				t := tick.Sub(tzero).Seconds()
				// Note: because output should start at floor, and increase linearly
				// to limit(ceiling), swatooth wave is phase shifted
				value := ((t / periodSecs) - math.Floor(t/periodSecs))
				sawtoothLogger.Debugw("Tick", "t", t, "value", value)
				output <- value
			}
		}
	}
}

// NewSineGenerator returns a cancellable function that will generate a value
// between 0.0 and 1.0 using a sine wave function.
func NewSineGenerator(logger *zap.SugaredLogger, period time.Duration) func(context.Context, *time.Ticker, chan float64) error {
	// Want to start the wave function close to floor, so offset tzero by half-phase
	tzero := time.Now().Add(period / 2.0)
	periodSecs := period.Seconds()
	sineLogger := logger.With(
		"generator", "sine",
		"period", period,
	)
	return func(ctx context.Context, ticker *time.Ticker, output chan float64) error {
		for {
			select {
			case <-ctx.Done():
				sineLogger.Infow("Function has been cancelled")
				return nil
			case tick := <-ticker.C:
				t := tick.Sub(tzero).Seconds()
				value := 0.5 + math.Sin(math.Pi*t/periodSecs)/2.0
				sineLogger.Debugw("Tick", "t", t, "value", value)
				output <- value
			}
		}
	}
}

// NewSquareGenerator returns a cancellable function that will generate a value
// of 0.0 or 1.0 using a square wave function.
func NewSquareGenerator(logger *zap.SugaredLogger, period time.Duration) func(context.Context, *time.Ticker, chan float64) error {
	// Want to start the wave function close to floor, so offset tzero by half-phase
	tzero := time.Now().Add(period / 2.0)
	periodSecs := period.Seconds()
	squareLogger := logger.With(
		"generator", "square",
		"period", period,
	)
	return func(ctx context.Context, ticker *time.Ticker, output chan float64) error {
		var value float64
		for {
			select {
			case <-ctx.Done():
				squareLogger.Infow("Function has been cancelled")
				return nil
			case tick := <-ticker.C:
				t := tick.Sub(tzero).Seconds()
				// Value is simply the sign of the sine wave at t
				switch value = math.Sin(math.Pi * t / periodSecs); {
				case value < 0:
					value = 0.0
				default:
					value = 1.0
				}
				squareLogger.Debugw("Tick", "t", t, "value", value)
				output <- value
			}
		}
	}
}

// NewTriangleGenerator returns a cancellable function that will generate a value
// between 0.0 and 1.0 using a triangle wave function.
func NewTriangleGenerator(logger *zap.SugaredLogger, period time.Duration) func(context.Context, *time.Ticker, chan float64) error {
	tzero := time.Now()
	periodSecs := period.Seconds()
	triangleLogger := logger.With(
		"generator", "triangle",
		"period", period,
	)
	return func(ctx context.Context, ticker *time.Ticker, output chan float64) error {
		for {
			select {
			case <-ctx.Done():
				triangleLogger.Infow("Function has been cancelled")
				return nil
			case tick := <-ticker.C:
				t := tick.Sub(tzero).Seconds()
				value := math.Abs(2.0 * ((t / periodSecs) - math.Floor(0.5+(t/periodSecs))))
				triangleLogger.Debugw("Tick", "t", t, "value", value)
				output <- value
			}
		}
	}
}
