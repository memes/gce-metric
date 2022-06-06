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

func newSawtoothCommand() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:     "sawtooth",
		Short:   "",
		Long:    "",
		PreRunE: bindViperFlags,
		RunE:    generatorMain,
		Args:    cobra.MinimumNArgs(1),
	}
	if err := addGeneratorFlags(cmd); err != nil {
		return nil, err
	}
	return cmd, nil
}

func newSineCommand() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:     "sine",
		Short:   "",
		Long:    "",
		PreRunE: bindViperFlags,
		RunE:    generatorMain,
		Args:    cobra.MinimumNArgs(1),
	}
	if err := addGeneratorFlags(cmd); err != nil {
		return nil, err
	}
	return cmd, nil
}

func newSquareCommand() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:     "square",
		Short:   "",
		Long:    "",
		PreRunE: bindViperFlags,
		RunE:    generatorMain,
		Args:    cobra.MinimumNArgs(1),
	}
	if err := addGeneratorFlags(cmd); err != nil {
		return nil, err
	}
	return cmd, nil
}

func newTriangleCommand() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:     "triangle",
		Short:   "",
		Long:    "",
		PreRunE: bindViperFlags,
		RunE:    generatorMain,
		Args:    cobra.MinimumNArgs(1),
	}
	if err := addGeneratorFlags(cmd); err != nil {
		return nil, err
	}
	return cmd, nil
}

func addGeneratorFlags(cmd *cobra.Command) error {
	cmd.PersistentFlags().Duration("sample", 60*time.Second, "sample time, specified in Go duration format. Default value is 60s")
	cmd.PersistentFlags().Duration("period", 10*time.Minute, "the period of the underlying wave function; e.g. the time for a complete cycle from floor to ceiling and back.")
	cmd.PersistentFlags().Float64("floor", 1.0, "the lowest value to send to metric; e.g. '1.0'")
	cmd.PersistentFlags().Float64("ceiling", 10.0, "the maximum value to send to metics; e.g. '10.0'")
	cmd.PersistentFlags().Bool("round", false, "Round metric value to nearest integer")
	cmd.PersistentFlags().StringToString("metriclabels", nil, "a set of metric label key=value pairs to send, separated by commas. E.g. -metriclabels=name:test,foo:bar")
	cmd.PersistentFlags().StringToString("resourcelabels", nil, "a set of resource label key=value pairs to send, separated by commas. E.g. -resourcelabels=name:test,foo:bar")
	cmd.PersistentFlags().Bool("dry-run", false, "Write metrics to stdout instead of to Google Cloud Monitoring")
	return nil
}

func bindViperFlags(cmd *cobra.Command, _ []string) error {
	if err := viper.BindPFlag("sample", cmd.PersistentFlags().Lookup("sample")); err != nil {
		return fmt.Errorf("failed to bind sample pflag: %w", err)
	}
	if err := viper.BindPFlag("period", cmd.PersistentFlags().Lookup("period")); err != nil {
		return fmt.Errorf("failed to bind period pflag: %w", err)
	}
	if err := viper.BindPFlag("floor", cmd.PersistentFlags().Lookup("floor")); err != nil {
		return fmt.Errorf("failed to bind floor pflag: %w", err)
	}
	if err := viper.BindPFlag("ceiling", cmd.PersistentFlags().Lookup("ceiling")); err != nil {
		return fmt.Errorf("failed to bind ceiling pflag: %w", err)
	}
	if err := viper.BindPFlag("round", cmd.PersistentFlags().Lookup("round")); err != nil {
		return fmt.Errorf("failed to bind round pflag: %w", err)
	}
	if err := viper.BindPFlag("metriclabels", cmd.PersistentFlags().Lookup("metriclabels")); err != nil {
		return fmt.Errorf("failed to bind metriclabels pflag: %w", err)
	}
	if err := viper.BindPFlag("resourcelabels", cmd.PersistentFlags().Lookup("resourcelabels")); err != nil {
		return fmt.Errorf("failed to bind resourcelabels pflag: %w", err)
	}
	if err := viper.BindPFlag("dry-run", cmd.PersistentFlags().Lookup("dry-run")); err != nil {
		return fmt.Errorf("failed to bind dry-run pflag: %w", err)
	}
	return nil
}

func generatorMain(cmd *cobra.Command, args []string) error {
	periodicType, err := generators.ParsePeriodicType(cmd.CalledAs())
	if err != nil {
		return fmt.Errorf("failure parsing PeriodicType: %w", err)
	}
	project := viper.GetString("project")
	sample := viper.GetDuration("sample")
	period := viper.GetDuration("period")
	floor := viper.GetFloat64("floor")
	ceiling := viper.GetFloat64("ceiling")
	dryRun := viper.GetBool("dry-run")
	rounded := viper.GetBool("round")
	logger := logger.WithValues("periodicType", periodicType.String(), "project", project, "sample", sample, "period", period, "floor", floor, "ceiling", ceiling, "dryRun", dryRun, "rounded", rounded)
	logger.V(0).Info("Building synthetic metric generator pipeline")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create the timestamped value generator
	valueGenerator, output := generators.NewPeriodicGenerator(logger,
		generators.NewPeriodicRangeCalculator(floor, ceiling, periodicType),
		period)
	// Build the pipeline from options.
	pipelineOptions := []pipeline.Option{
		pipeline.WithLogger(logger),
		pipeline.WithMetricType(args[0]),
	}
	if project != "" {
		pipelineOptions = append(pipelineOptions, pipeline.WithProjectID(project))
	}
	if rounded {
		pipelineOptions = append(pipelineOptions, pipeline.AsInteger())
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
		if err := processor(ctx, output); err != nil {
			logger.Error(err, "Pipeline processor returned an error")
			cancel()
		}
	}()
	go func() {
		logger.V(1).Info("Launching value generator")
		if err := valueGenerator(ctx, ticker.C); err != nil {
			logger.Error(err, "Generator returned an error")
			cancel()
		}
	}()
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
