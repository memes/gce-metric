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
	SampleFlagName         = "sample"
	PeriodFlagName         = "period"
	FloorFlagName          = "floor"
	CeilingFlagName        = "ceiling"
	IntegerFlagName        = "integer"
	MetricLabelsFlagName   = "metric-labels"
	ResourceLabelsFlagName = "resource-labels"
	ResourceTypeFlagName   = "resource-type"
	DryRunFlagName         = "dry-run"
)

func newSawtoothCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sawtooth",
		Short:   "Generate synthetic metrics from a sawtooth function",
		Long:    "",
		Example: "sawtooth foo",
		PreRunE: bindViperFlags,
		RunE:    generatorMain,
		Args:    cobra.MinimumNArgs(1),
	}
	addGeneratorFlags(cmd)
	return cmd
}

func newSineCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sine",
		Short:   "Generate synthetic metrics from a sine function",
		Long:    "",
		Example: "sine foo",
		PreRunE: bindViperFlags,
		RunE:    generatorMain,
		Args:    cobra.MinimumNArgs(1),
	}
	addGeneratorFlags(cmd)
	return cmd
}

func newSquareCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "square",
		Short:   "Generate synthetic metrics from a square function",
		Long:    "",
		Example: "square foo",
		PreRunE: bindViperFlags,
		RunE:    generatorMain,
		Args:    cobra.MinimumNArgs(1),
	}
	addGeneratorFlags(cmd)
	return cmd
}

func newTriangleCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "triangle",
		Short:   "Generate synthetic metrics from a triangle function",
		Long:    "",
		Example: "triangle foo",
		PreRunE: bindViperFlags,
		RunE:    generatorMain,
		Args:    cobra.MinimumNArgs(1),
	}
	addGeneratorFlags(cmd)
	return cmd
}

func addGeneratorFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().Duration(SampleFlagName, 60*time.Second, "sample time, specified in Go duration format. Default value is 60s")
	cmd.PersistentFlags().Duration(PeriodFlagName, 10*time.Minute, "the period of the underlying wave function; e.g. the time for a complete cycle from floor to ceiling and back.")
	cmd.PersistentFlags().Float64(FloorFlagName, 1.0, "the lowest value to send to metric; e.g. '1.0'")
	cmd.PersistentFlags().Float64(CeilingFlagName, 10.0, "the maximum value to send to metics; e.g. '10.0'")
	cmd.PersistentFlags().Bool(IntegerFlagName, false, "Round metric value to nearest integer")
	cmd.PersistentFlags().StringToString(MetricLabelsFlagName, nil, "a set of metric label key=value pairs to send, separated by commas. E.g. --metric-labels=name=test,foo=bar")
	cmd.PersistentFlags().StringToString(ResourceLabelsFlagName, nil, "a set of resource label key=value pairs to send, separated by commas. E.g. --resource-labels=name=test,foo=bar")
	cmd.PersistentFlags().String(ResourceTypeFlagName, "", "TODO")
	cmd.PersistentFlags().Bool(DryRunFlagName, false, "Write metrics to stdout instead of to Google Cloud Monitoring")
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
	if err := viper.BindPFlag(MetricLabelsFlagName, cmd.PersistentFlags().Lookup(MetricLabelsFlagName)); err != nil {
		return fmt.Errorf("failed to bind '%s' pflag: %w", MetricLabelsFlagName, err)
	}
	if err := viper.BindPFlag(ResourceLabelsFlagName, cmd.PersistentFlags().Lookup(ResourceLabelsFlagName)); err != nil {
		return fmt.Errorf("failed to bind '%s' pflag: %w", ResourceLabelsFlagName, err)
	}
	if err := viper.BindPFlag(ResourceTypeFlagName, cmd.PersistentFlags().Lookup(ResourceTypeFlagName)); err != nil {
		return fmt.Errorf("failed to bind '%s' pflag: %w", ResourceTypeFlagName, err)
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
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	select {
	case <-ctx.Done():
		logger.V(1).Info("Context has been canceled, stopping ticker")
		ticker.Stop()

	case sig := <-interrupt:
		logger.V(1).Info("Interrupt received, stopping ticker and canceling context", "sig", sig)
		ticker.Stop()
		cancel()
	}
	return nil
}
