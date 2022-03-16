// Copyright 2020 Matthew Emes. All rights reserved.
// Use of this source code is governed by the MIT license included in the source.
package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"cloud.google.com/go/compute/metadata"
	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	timestamppb "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/google/uuid"
	"github.com/peterbourgon/ff/v3/ffcli"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/api/iterator"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoredrespb "google.golang.org/genproto/googleapis/api/monitoredres"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

const (
	// Google metrics require a minimum interval between samples of 10s
	GCPMinimumSampleDuration = 10 * time.Second
)

type metricValueGenerator func(*zap.SugaredLogger, float64) *monitoringpb.TypedValue
type metricRequestGenerator func(float64) *monitoringpb.CreateTimeSeriesRequest

func main() {
	var (
		err                error
		verbose            bool
		sample             time.Duration
		period             time.Duration
		floor              float64
		ceiling            float64
		round              bool
		projectID          string
		metricLabels       map[string]string
		resourceType       string
		resourceLabels     map[string]string
		ticker             *time.Ticker
		overrideProjectID  string
		metricLabelsArgs   string
		resourceLabelsArgs string
		filter             string
		quiet              bool
	)
	logger, level := initLogger()
	defer func() {
		_ = logger.Sync()
	}()
	logger.Debug("Starting application")

	rootFlagSet := flag.NewFlagSet("gce-metric", flag.ExitOnError)
	rootFlagSet.BoolVar(&verbose, "verbose", false, "enable DEBUG log level")
	rootFlagSet.BoolVar(&quiet, "quiet", false, "disable all but ERROR and FATAL logging")
	generatorFlagSet := flag.NewFlagSet("generator", flag.ExitOnError)
	generatorFlagSet.BoolVar(&verbose, "verbose", false, "enable DEBUG log level")
	generatorFlagSet.BoolVar(&quiet, "quiet", false, "disable all but ERROR and FATAL logging")
	generatorFlagSet.DurationVar(&sample, "sample", 60*time.Second, "sample time, specified in Go duration format. Default value is 60s")
	generatorFlagSet.DurationVar(&period, "period", 10*time.Minute, "the period of the underlying wave function; e.g. the time for a complete cycle from floor to ceiling and back.")
	generatorFlagSet.Float64Var(&floor, "floor", 1.0, "the lowest value to send to metric; e.g. '1.0'")
	generatorFlagSet.Float64Var(&ceiling, "ceiling", 10.0, "the maximum value to send to metics; e.g. '10.0'")
	generatorFlagSet.BoolVar(&round, "round", false, "Round metric value to nearest integer")
	generatorFlagSet.StringVar(&overrideProjectID, "project", "", "the GCP project id to use; specify if not running on GCE or to override project id")
	generatorFlagSet.StringVar(&metricLabelsArgs, "metriclabels", "", "a set of metric label key:value pairs to send, separated by commas. E.g. -metriclabels=name:test,foo:bar")
	generatorFlagSet.StringVar(&resourceLabelsArgs, "resourcelabels", "", "a set of resource label key:value pairs to send, separated by commas. E.g. -resourcelabels=name:test,foo:bar")
	deleteFlagSet := flag.NewFlagSet("delete", flag.ExitOnError)
	deleteFlagSet.BoolVar(&verbose, "verbose", false, "enable DEBUG log level")
	deleteFlagSet.BoolVar(&quiet, "quiet", false, "disable all but ERROR and FATAL logging")
	deleteFlagSet.StringVar(&overrideProjectID, "project", "", "the GCP project id to use; specify if not running on GCE or to override project id")
	listFlagSet := flag.NewFlagSet("delete", flag.ExitOnError)
	listFlagSet.BoolVar(&verbose, "verbose", false, "enable DEBUG log level")
	listFlagSet.BoolVar(&quiet, "quiet", false, "disable all but ERROR and FATAL logging")
	listFlagSet.StringVar(&overrideProjectID, "project", "", "the GCP project id to use; specify if not running on GCE or to override project id")
	listFlagSet.StringVar(&filter, "filter", "metric.type = starts_with(\"custom.googleapis.com/\")", "set the filter to use when listing metrics")
	sawtooth := &ffcli.Command{
		Name:       "sawtooth",
		ShortUsage: "gce-metric sawtooth [flags] type",
		ShortHelp:  "generate metrics that match a sawtooth wave function",
		FlagSet:    generatorFlagSet,
		Exec: func(ctx context.Context, metricTypes []string) error {
			if err := validateGenerateParameters(metricTypes, projectID, resourceType, resourceLabels, sample); err != nil {
				return err
			}
			requestGenerator := newCreateTimeSeriesRequestGenerator(logger, round, metricTypes[0], projectID, metricLabels, resourceType, resourceLabels)
			valueGenerator := newSawtoothGenerator(logger, period)
			return valueGenerator(ctx, ticker, outputChannel(ctx, logger, floor, ceiling, requestGenerator))
		},
	}
	sine := &ffcli.Command{
		Name:       "sine",
		ShortUsage: "gce-metric sine [flags] type",
		ShortHelp:  "generate metrics that match a sine wave function",
		FlagSet:    generatorFlagSet,
		Exec: func(ctx context.Context, metricTypes []string) error {
			if err := validateGenerateParameters(metricTypes, projectID, resourceType, resourceLabels, sample); err != nil {
				return err
			}
			requestGenerator := newCreateTimeSeriesRequestGenerator(logger, round, metricTypes[0], projectID, metricLabels, resourceType, resourceLabels)
			valueGenerator := newSineGenerator(logger, period)
			return valueGenerator(ctx, ticker, outputChannel(ctx, logger, floor, ceiling, requestGenerator))
		},
	}
	square := &ffcli.Command{
		Name:       "square",
		ShortUsage: "gce-metric square [flags] type",
		ShortHelp:  "generate metrics that match a square wave function",
		FlagSet:    generatorFlagSet,
		Exec: func(ctx context.Context, metricTypes []string) error {
			if err := validateGenerateParameters(metricTypes, projectID, resourceType, resourceLabels, sample); err != nil {
				return err
			}
			requestGenerator := newCreateTimeSeriesRequestGenerator(logger, round, metricTypes[0], projectID, metricLabels, resourceType, resourceLabels)
			valueGenerator := newSquareGenerator(logger, period)
			return valueGenerator(ctx, ticker, outputChannel(ctx, logger, floor, ceiling, requestGenerator))
		},
	}
	triangle := &ffcli.Command{
		Name:       "triangle",
		ShortUsage: "gce-metric triangle [flags] type",
		ShortHelp:  "generate metrics that match a triangle wave function",
		FlagSet:    generatorFlagSet,
		Exec: func(ctx context.Context, metricTypes []string) error {
			if err := validateGenerateParameters(metricTypes, projectID, resourceType, resourceLabels, sample); err != nil {
				return err
			}
			requestGenerator := newCreateTimeSeriesRequestGenerator(logger, round, metricTypes[0], projectID, metricLabels, resourceType, resourceLabels)
			valueGenerator := newTriangleGenerator(logger, period)
			return valueGenerator(ctx, ticker, outputChannel(ctx, logger, floor, ceiling, requestGenerator))
		},
	}
	delete := &ffcli.Command{
		Name:       "delete",
		ShortUsage: "gce-metric delete [flags] type...",
		ShortHelp:  "delete a custom metric timeseries",
		FlagSet:    deleteFlagSet,
		Exec: func(ctx context.Context, metricTypes []string) error {
			if err := validateDeleteParameters(metricTypes, projectID); err != nil {
				return err
			}
			return deleteMetrics(ctx, logger, metricTypes, projectID)
		},
	}
	list := &ffcli.Command{
		Name:       "list",
		ShortUsage: "gce-metric list [flags]",
		ShortHelp:  "list custom metric timeseries",
		FlagSet:    listFlagSet,
		Exec: func(ctx context.Context, _ []string) error {
			if err := validateListParameters(projectID, filter); err != nil {
				return err
			}
			return listMetrics(ctx, logger, projectID, filter)
		},
	}
	root := &ffcli.Command{
		ShortUsage:  "gce-metric <subcommand> [flags]",
		Subcommands: []*ffcli.Command{sawtooth, sine, square, triangle, delete, list},
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
	if quiet {
		level.SetLevel(zapcore.ErrorLevel)
	}

	// Get default values from execution environment
	if projectID, metricLabels, resourceType, resourceLabels, err = newEnvironmentParameters(logger); err != nil {
		logger.Fatalw("Error inspecting running environment",
			"projectID", projectID,
			"metricLabels", metricLabels,
			"resourceType", resourceType,
			"resourceLabels", resourceLabels,
			"error", err,
		)
		os.Exit(1)
	}

	// Allow a project ID flag to override anything picked up from GCE metadata
	if len(strings.TrimSpace(overrideProjectID)) > 0 {
		projectID = overrideProjectID
	}
	// Parse and update metricLabels map from command line argument
	if err = mergeMapArgs(metricLabels, metricLabelsArgs); err != nil {
		logger.Fatalw("Error parsing metriclabels argument",
			"error", err,
		)
		os.Exit(1)
	}

	// Parse and update resourceLabels map from command line argument
	if err = mergeMapArgs(resourceLabels, resourceLabelsArgs); err != nil {
		logger.Fatalw("Error parsing resourceLabels argument",
			"error", err,
		)
		os.Exit(1)
	}

	// Start a ticker; not needed for delete/list operations, but safer to do here
	ticker = time.NewTicker(sample)

	// Create a cancellable context and launch a goroutine to handle OS signals
	ctx := cancellableContext(context.Background(), logger, ticker)

	// Run the command
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

// Returns appropriate values for projectID, resourceType, and labels required for sending metrics to GCP.
// Where possible, these values will be taken from GCP metadata
func newEnvironmentParameters(logger *zap.SugaredLogger) (projectID string, metricLabels map[string]string, resourceType string, resourceLabels map[string]string, err error) {
	metricLabels = make(map[string]string)
	resourceLabels = make(map[string]string)
	logger.Debug("Examining environment for default parameter values")
	if metadata.OnGCE() {
		logger.Debug("Running on GCE")
		resourceType = "gce_instance"
		// Get defaults from GCE metadata
		if projectID, err = metadata.ProjectID(); err != nil {
			return
		}
		if resourceLabels["instance_id"], err = metadata.InstanceID(); err != nil {
			return
		}
		if resourceLabels["zone"], err = metadata.Zone(); err != nil {
			return
		}
	} else {
		logger.Debug("Not running on GCE")
		resourceType = "generic_node"
		resourceLabels["location"] = "global"
		resourceLabels["namespace"] = "gce-metric"
		resourceLabels["node_id"] = uuid.New().String()
	}
	return
}

// Merge a string containing key:value,... pairs into an existing map
func mergeMapArgs(existing map[string]string, updates string) error {
	if len(strings.TrimSpace(updates)) == 0 {
		return nil
	}
	pairs := strings.Split(updates, ",")
	if len(pairs) == 0 {
		return nil
	}
	for _, pair := range pairs {
		args := strings.SplitN(strings.TrimSpace(pair), ":", 2)
		if len(args) != 2 {
			return fmt.Errorf("unbalanced arguments: must be key:value pairs")
		}
		k := strings.TrimSpace(args[0])
		v := strings.TrimSpace(args[1])
		if len(k) == 0 || len(v) == 0 {
			return fmt.Errorf("unbalanced arguments: must be key:value pairs '%s:%s'", k, v)
		}
		existing[k] = v
	}
	return nil
}

// Return an error if the parameters required to generate and send a series of metrics are incomplete
func validateGenerateParameters(metricTypes []string, projectID string, resourceType string, resourceLabels map[string]string, sample time.Duration) error {
	if len(metricTypes) != 1 {
		return fmt.Errorf("exactly one metric type must be provided: %d supplied", len(metricTypes))
	}
	if len(strings.TrimSpace(metricTypes[0])) < 1 {
		return fmt.Errorf("metric type must be a non-empty string")
	}
	if len(strings.TrimSpace(projectID)) < 1 {
		return fmt.Errorf("project ID must be provided")
	}
	if len(strings.TrimSpace(resourceType)) < 1 {
		return fmt.Errorf("resourceType must be provided")
	}
	if sample < GCPMinimumSampleDuration {
		return fmt.Errorf("sample value must be %v or greater: %v", GCPMinimumSampleDuration, sample)
	}
	return nil
}

// Return an error if the parameters required to delete a custom metric are incomplete
func validateDeleteParameters(metricTypes []string, projectID string) error {
	if len(metricTypes) < 1 {
		return fmt.Errorf("at least one metic type must be provided to delete")
	}
	for _, metricType := range metricTypes {
		if len(strings.TrimSpace(metricType)) < 1 {
			return fmt.Errorf("metric type must be provided")
		}
	}
	if len(strings.TrimSpace(projectID)) < 1 {
		return fmt.Errorf("project ID must be provided")
	}
	return nil
}

// Return an error if the parameters required to list custom metrics are incomplete
func validateListParameters(projectID string, filter string) error {
	if len(strings.TrimSpace(projectID)) < 1 {
		return fmt.Errorf("project ID must be provided")
	}
	return nil
}

// Returns a cancellable context, and a handler that will intercept OS signals to shutdown
// the ticker (if active), and propagate cancel through the context
func cancellableContext(ctx context.Context, logger *zap.SugaredLogger, ticker *time.Ticker) context.Context {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	newCtx, cancel := context.WithCancel(ctx)
	go func() {
		logger.Debug("Signal handler started, waiting for signal")
		sig := <-interrupt
		if ticker != nil {
			logger.Debug("Stopping timer")
			ticker.Stop()
		}
		logger.Infow("Interrupt received, cancelling context",
			"sig", sig,
		)
		cancel()
		logger.Debug("Signal handler exit")
	}()
	return newCtx
}

// Creates a channel that will post the values added to it to Google Cloud Metrics
func outputChannel(ctx context.Context, logger *zap.SugaredLogger, floor float64, ceiling float64, requestGenerator metricRequestGenerator) chan float64 {
	output := make(chan float64, 1)
	go func(ctx context.Context) {
		logger.Debug("Output handler started")
		delta := ceiling - floor
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
				// Note: the routine does not cancel or fail if the push to GCP metrics fails
				if err := sendMetric(ctx, requestGenerator(value)); err != nil {
					logger.Errorw("Error sending metric to GCP",
						"err", err,
					)
				}
			}
		}
	}(ctx)
	return output
}

// Returns a cancellable function that when called will generate a value between 0.0 and 1.0 using a sawtooth wave function
// and send it to a channel.
func newSawtoothGenerator(logger *zap.SugaredLogger, period time.Duration) func(context.Context, *time.Ticker, chan float64) error {
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

// Returns a cancellable function that when called will generate a value between 0.0 and 1.0 using a sine wave function
// and send it to a channel.
func newSineGenerator(logger *zap.SugaredLogger, period time.Duration) func(context.Context, *time.Ticker, chan float64) error {
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

// Returns a cancellable function that when called will generate a value between 0.0 and 1.0 using a square wave function
// and send it to a channel.
func newSquareGenerator(logger *zap.SugaredLogger, period time.Duration) func(context.Context, *time.Ticker, chan float64) error {
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

// Returns a cancellable function that when called will generate a value between 0.0 and 1.0 using a triangle wave function
// and send it to a channel.
func newTriangleGenerator(logger *zap.SugaredLogger, period time.Duration) func(context.Context, *time.Ticker, chan float64) error {
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

// Returns a Google Metrics Int64 typed value equivalent to rounded metric
func roundedMetricAdapter(logger *zap.SugaredLogger, value float64) *monitoringpb.TypedValue {
	typedValue := &monitoringpb.TypedValue{
		Value: &monitoringpb.TypedValue_Int64Value{
			Int64Value: int64(value),
		},
	}
	logger.Debugw("Generated rounded metric",
		"value", value,
		"typedValue", typedValue,
	)
	return typedValue
}

// Returns a Google Metrics DoubleValue typed value for the provided metric
func doubleMetricAdapter(logger *zap.SugaredLogger, value float64) *monitoringpb.TypedValue {
	typedValue := &monitoringpb.TypedValue{
		Value: &monitoringpb.TypedValue_DoubleValue{
			DoubleValue: value,
		},
	}
	logger.Debugw("Generated double metric",
		"value", value,
		"typedValue", typedValue,
	)
	return typedValue
}

// Returns a function that accepts a float64 value and returns a CreateTimeSeriesRequest that can be sent to GCP metrics
func newCreateTimeSeriesRequestGenerator(logger *zap.SugaredLogger, round bool, metricType string, projectID string, metricLabels map[string]string, resourceType string, resourceLabels map[string]string) metricRequestGenerator {
	requestLogger := logger.With(
		"round", round,
		"metricType", metricType,
		"projectID", projectID,
		"metricLabels", metricLabels,
		"resourceType", resourceType,
		"resourceLabels", resourceLabels,
	)
	requestLogger.Debug("Building new CreateTimeSeriesRequest generator")
	var metricValueAdapter metricValueGenerator
	if round {
		metricValueAdapter = roundedMetricAdapter
	} else {
		metricValueAdapter = doubleMetricAdapter
	}
	return func(value float64) *monitoringpb.CreateTimeSeriesRequest {
		ts := time.Now().Unix()
		req := &monitoringpb.CreateTimeSeriesRequest{
			Name: "projects/" + projectID,
			TimeSeries: []*monitoringpb.TimeSeries{{
				Metric: &metricpb.Metric{
					Type:   metricType,
					Labels: metricLabels,
				},
				Resource: &monitoredrespb.MonitoredResource{
					Type:   resourceType,
					Labels: resourceLabels,
				},
				Points: []*monitoringpb.Point{{
					Interval: &monitoringpb.TimeInterval{
						StartTime: &timestamppb.Timestamp{
							Seconds: ts,
						},
						EndTime: &timestamppb.Timestamp{
							Seconds: ts,
						},
					},
					Value: metricValueAdapter(requestLogger, value),
				}},
			}},
		}
		requestLogger.Debugw("new CreateTimeSeriesRequest",
			"value", value,
			"req", req,
		)
		return req
	}
}

// Send a metric request to GCP
func sendMetric(ctx context.Context, req *monitoringpb.CreateTimeSeriesRequest) error {
	client, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return err
	}
	return client.CreateTimeSeries(ctx, req)
}

// Delete custom GCP metric series
func deleteMetrics(ctx context.Context, logger *zap.SugaredLogger, metricTypes []string, projectID string) error {
	client, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return err
	}
	for _, metricType := range metricTypes {
		request := &monitoringpb.DeleteMetricDescriptorRequest{
			Name: "projects/" + projectID + "/metricDescriptors/" + metricType,
		}
		if err := client.DeleteMetricDescriptor(ctx, request); err != nil {
			return err
		}
		logger.Infow("Custom metric deleted",
			"metricType", metricType,
		)
	}
	return nil
}

// List custom metrics
func listMetrics(ctx context.Context, logger *zap.SugaredLogger, projectID string, filter string) error {
	client, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return err
	}
	request := &monitoringpb.ListMetricDescriptorsRequest{
		Name: "projects/" + projectID,
	}
	f := strings.TrimSpace(filter)
	if len(f) > 0 {
		request.Filter = f
	}
	it := client.ListMetricDescriptors(ctx, request)
	for {
		response, err := it.Next()
		switch {
		case err == iterator.Done:
			return nil
		case err != nil:
			return err
		default:
			logger.Debugw("Metric descriptor response",
				"response", response,
			)
			fmt.Println(response.Type)
		}
	}
}
