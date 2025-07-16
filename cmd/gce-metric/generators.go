package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/memes/gce-metric/pkg/generators"
	"github.com/memes/gce-metric/pkg/pipeline"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	sampleFlagName  = "sample"
	periodFlagName  = "period"
	floorFlagName   = "floor"
	ceilingFlagName = "ceiling"
	integerFlagName = "integer"
	dryRunFlagName  = "dry-run"
)

func newSawtoothCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sawtooth [flags] NAME",
		Short:   "Generate synthetic metrics from a sawtooth function",
		Long:    "Generate synthetic metric time-series data-points that approximate a sawtooth pattern, and send them to Google Cloud Monitoring to trigger scaling events or for other purposes.",
		Example: appName + "sawtooth --project ID custom.googleapis.com/syntheticScaler/cpu",
		PreRunE: bindViperFlags,
		RunE:    generatorMain,
		Args:    cobra.MinimumNArgs(1),
	}
	addGeneratorFlags(cmd)
	return cmd
}

func newSineCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sine [flags] NAME",
		Short:   "Generate synthetic metrics from a sine function",
		Long:    "Generate synthetic metric time-series data-points that approximate a sine pattern, and send them to Google Cloud Monitoring to trigger scaling events or for other purposes.",
		Example: appName + "sine --project ID custom.googleapis.com/syntheticScaler/cpu",
		PreRunE: bindViperFlags,
		RunE:    generatorMain,
		Args:    cobra.MinimumNArgs(1),
	}
	addGeneratorFlags(cmd)
	return cmd
}

func newSquareCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "square [flags] NAME",
		Short:   "Generate synthetic metrics from a square function",
		Long:    "Generate synthetic metric time-series data-points that approximate a square pattern, and send them to Google Cloud Monitoring to trigger scaling events or for other purposes.",
		Example: appName + "square --project ID custom.googleapis.com/syntheticScaler/cpu",
		PreRunE: bindViperFlags,
		RunE:    generatorMain,
		Args:    cobra.MinimumNArgs(1),
	}
	addGeneratorFlags(cmd)
	return cmd
}

func newTriangleCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "triangle [flags] NAME",
		Short:   "Generate synthetic metrics from a triangle function",
		Long:    "Generate synthetic metric time-series data-points that approximate a triangle pattern, and send them to Google Cloud Monitoring to trigger scaling events or for other purposes.",
		Example: appName + "triangle --project ID custom.googleapis.com/syntheticScaler/cpu",
		PreRunE: bindViperFlags,
		RunE:    generatorMain,
		Args:    cobra.MinimumNArgs(1),
	}
	addGeneratorFlags(cmd)
	return cmd
}

func addGeneratorFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().Duration(sampleFlagName, 60*time.Second, "sets the interval between sending metrics to Google Monitoring, must be valid Go duration string")
	cmd.PersistentFlags().Duration(periodFlagName, 10*time.Minute, "sets the duration for one complete cycle from floor to ceiling, must be valid Go duration string")
	cmd.PersistentFlags().Float64(floorFlagName, 1.0, "sets the minimum value for the cycles, can be an integer or floating point value")
	cmd.PersistentFlags().Float64(ceilingFlagName, 10.0, "sets the maximum value for the cycles, can be an integer of floating point value")
	cmd.PersistentFlags().Bool(integerFlagName, false, "forces the generated metrics to be integers, making them less smooth and more step-like")
	cmd.PersistentFlags().Bool(dryRunFlagName, false, "report metrics to stdout for review, without sending to Google Cloud Monitoring; for the curious!")
}

func bindViperFlags(cmd *cobra.Command, _ []string) error {
	if err := viper.BindPFlag(sampleFlagName, cmd.PersistentFlags().Lookup(sampleFlagName)); err != nil {
		return fmt.Errorf("failed to bind '%s' pflag: %w", sampleFlagName, err)
	}
	if err := viper.BindPFlag(periodFlagName, cmd.PersistentFlags().Lookup(periodFlagName)); err != nil {
		return fmt.Errorf("failed to bind '%s' pflag: %w", periodFlagName, err)
	}
	if err := viper.BindPFlag(floorFlagName, cmd.PersistentFlags().Lookup(floorFlagName)); err != nil {
		return fmt.Errorf("failed to bind '%s' pflag: %w", floorFlagName, err)
	}
	if err := viper.BindPFlag(ceilingFlagName, cmd.PersistentFlags().Lookup(ceilingFlagName)); err != nil {
		return fmt.Errorf("failed to bind '%s' pflag: %w", ceilingFlagName, err)
	}
	if err := viper.BindPFlag(integerFlagName, cmd.PersistentFlags().Lookup(integerFlagName)); err != nil {
		return fmt.Errorf("failed to bind '%s' pflag: %w", integerFlagName, err)
	}
	if err := viper.BindPFlag(dryRunFlagName, cmd.PersistentFlags().Lookup(dryRunFlagName)); err != nil {
		return fmt.Errorf("failed to bind '%s' pflag: %w", dryRunFlagName, err)
	}
	return nil
}

//nolint:funlen // Setup of options makes the function seem long
func generatorMain(cmd *cobra.Command, args []string) error {
	periodicType, err := generators.ParsePeriodicType(cmd.CalledAs())
	if err != nil {
		return fmt.Errorf("failure parsing PeriodicType: %w", err)
	}
	project := viper.GetString(projectIDFlagName)
	sample := viper.GetDuration(sampleFlagName)
	period := viper.GetDuration(periodFlagName)
	floor := viper.GetFloat64(floorFlagName)
	ceiling := viper.GetFloat64(ceilingFlagName)
	dryRun := viper.GetBool(dryRunFlagName)
	asInteger := viper.GetBool(integerFlagName)
	logger := logger.WithValues("periodicType", periodicType.String(), "project", project, "sample", sample, "period", period, floorFlagName, floor, ceilingFlagName, ceiling, "dryRun", dryRun, "asInteger", asInteger)
	logger.V(0).Info("Building synthetic metric generator pipeline")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Create the timestamped value generator
	periodicGenerator, reader, err := generators.NewPeriodicGenerator(
		generators.WithLogger(logger),
		generators.WithValueCalculator(generators.NewPeriodicRangeCalculator(floor, ceiling, periodicType)),
		generators.WithPeriod(period),
	)
	if err != nil {
		return fmt.Errorf("failure building PeriodicGenerator: %w", err)
	}
	// Build the pipeline from options.
	pipelineOptions := []pipeline.Option{
		pipeline.WithLogger(logger),
		pipeline.WithMetricType(args[0]),
	}
	if project != "" {
		pipelineOptions = append(pipelineOptions, pipeline.WithProjectID(project))
	}
	if asInteger {
		pipelineOptions = append(pipelineOptions, pipeline.WithTransformers([]pipeline.Transformer{pipeline.NewIntegerTypedValueTransformer()}))
	}
	if dryRun {
		pipelineOptions = append(pipelineOptions, pipeline.WithWriterEmitter(os.Stdout))
	}
	pipe, err := pipeline.NewPipeline(ctx, pipelineOptions...)
	if err != nil {
		return fmt.Errorf("failure creating new pipeline: %w", err)
	}
	defer func() {
		logger.V(2).Info("Closing pipeline")
		if err := pipe.Close(); err != nil {
			logger.Error(err, "Error returned while closing pipeline")
		}
	}()
	ticker := time.NewTicker(sample)
	defer ticker.Stop()
	go func() {
		logger.V(1).Info("Launching pipeline processor")
		processor := pipe.Processor()
		if err := processor(ctx, reader); err != nil {
			logger.Error(err, "Pipeline processor returned an error")
			cancel()
		}
	}()
	logger.V(1).Info("Launching periodic generator")
	go periodicGenerator(ctx, ticker.C)
	logger.V(1).Info("Goroutines launched, waiting for processing to be interrupted")
	<-ctx.Done()
	logger.V(1).Info("Context has been cancelled")
	return nil
}
