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
	SampleFlagName  = "sample"
	PeriodFlagName  = "period"
	FloorFlagName   = "floor"
	CeilingFlagName = "ceiling"
	IntegerFlagName = "integer"
	DryRunFlagName  = "dry-run"
)

func newSawtoothCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sawtooth [flags] NAME",
		Short:   "Generate synthetic metrics from a sawtooth function",
		Long:    "Generate synthetic metric time-series data-points that approximate a sawtooth pattern, and send them to Google Cloud Monitoring to trigger scaling events or for other purposes.",
		Example: AppName + "sawtooth --project ID custom.googleapis.com/syntheticScaler/cpu",
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
		Example: AppName + "sine --project ID custom.googleapis.com/syntheticScaler/cpu",
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
		Example: AppName + "square --project ID custom.googleapis.com/syntheticScaler/cpu",
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
		Example: AppName + "triangle --project ID custom.googleapis.com/syntheticScaler/cpu",
		PreRunE: bindViperFlags,
		RunE:    generatorMain,
		Args:    cobra.MinimumNArgs(1),
	}
	addGeneratorFlags(cmd)
	return cmd
}

func addGeneratorFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().Duration(SampleFlagName, 60*time.Second, "sets the interval between sending metrics to Google Monitoring, must be valid Go duration string")
	cmd.PersistentFlags().Duration(PeriodFlagName, 10*time.Minute, "sets the duration for one complete cycle from floor to ceiling, must be valid Go duration string")
	cmd.PersistentFlags().Float64(FloorFlagName, 1.0, "sets the minimum value for the cycles, can be an integer or floating point value")
	cmd.PersistentFlags().Float64(CeilingFlagName, 10.0, "sets the maximum value for the cycles, can be an integer of floating point value")
	cmd.PersistentFlags().Bool(IntegerFlagName, false, "forces the generated metrics to be integers, making them less smooth and more step-like")
	cmd.PersistentFlags().Bool(DryRunFlagName, false, "report metrics to stdout for review, without sending to Google Cloud Monitoring; for the curious!")
}

func bindViperFlags(cmd *cobra.Command, _ []string) error {
	if err := viper.BindPFlag(SampleFlagName, cmd.PersistentFlags().Lookup(SampleFlagName)); err != nil {
		return fmt.Errorf("failed to bind '%s' pflag: %w", SampleFlagName, err)
	}
	if err := viper.BindPFlag(PeriodFlagName, cmd.PersistentFlags().Lookup(PeriodFlagName)); err != nil {
		return fmt.Errorf("failed to bind '%s' pflag: %w", PeriodFlagName, err)
	}
	if err := viper.BindPFlag(FloorFlagName, cmd.PersistentFlags().Lookup(FloorFlagName)); err != nil {
		return fmt.Errorf("failed to bind '%s' pflag: %w", FloorFlagName, err)
	}
	if err := viper.BindPFlag(CeilingFlagName, cmd.PersistentFlags().Lookup(CeilingFlagName)); err != nil {
		return fmt.Errorf("failed to bind '%s' pflag: %w", CeilingFlagName, err)
	}
	if err := viper.BindPFlag(IntegerFlagName, cmd.PersistentFlags().Lookup(IntegerFlagName)); err != nil {
		return fmt.Errorf("failed to bind '%s' pflag: %w", IntegerFlagName, err)
	}
	if err := viper.BindPFlag(DryRunFlagName, cmd.PersistentFlags().Lookup(DryRunFlagName)); err != nil {
		return fmt.Errorf("failed to bind '%s' pflag: %w", DryRunFlagName, err)
	}
	return nil
}

func generatorMain(cmd *cobra.Command, args []string) error {
	periodicType, err := generators.ParsePeriodicType(cmd.CalledAs())
	if err != nil {
		return fmt.Errorf("failure parsing PeriodicType: %w", err)
	}
	project := viper.GetString(ProjectIDFlagName)
	sample := viper.GetDuration(SampleFlagName)
	period := viper.GetDuration(PeriodFlagName)
	floor := viper.GetFloat64(FloorFlagName)
	ceiling := viper.GetFloat64(CeilingFlagName)
	dryRun := viper.GetBool(DryRunFlagName)
	asInteger := viper.GetBool(IntegerFlagName)
	logger := logger.WithValues("periodicType", periodicType.String(), "project", project, "sample", sample, "period", period, FloorFlagName, floor, CeilingFlagName, ceiling, "dryRun", dryRun, "asInteger", asInteger)
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
