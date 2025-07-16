package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	startTimeFlag = "start-time"
	endTimeFlag   = "end-time"
)

func newDataCommand() (*cobra.Command, error) {
	dataCmd := &cobra.Command{
		Use:     "data [--verbose] [--project ID] [--filter FILTER] [--start-time ISO8601] [--end-time ISO8601]",
		Short:   "Return metric data from time-series that match the filter.",
		Long:    `Returns each metric time-series that matches the supplied filter and has point-in-time data that is between the start and end times provided.`,
		Example: appName + ` data --project ID --filter 'metric.type = has_substring("my-resource")' --start-time $(date -Iseconds -v -4H)`,
		RunE:    metricData,
		Args:    cobra.NoArgs,
	}
	dataCmd.PersistentFlags().String(filterFlagName, "metric.type = starts_with(\"custom.googleapis.com/\")", "set the filter to use when listing metrics")
	dataCmd.PersistentFlags().String(startTimeFlag, "", "set the start time for filtering data, if unspecified matching time-series data points from 5 mins ago will be included")
	dataCmd.PersistentFlags().String(endTimeFlag, "", "set the end time for filtering data, if unspecified matching time-series data points up to the current time will be included")
	if err := viper.BindPFlag(filterFlagName, dataCmd.PersistentFlags().Lookup(filterFlagName)); err != nil {
		return nil, fmt.Errorf("failed to bind '%s' pflag: %w", filterFlagName, err)
	}
	if err := viper.BindPFlag(startTimeFlag, dataCmd.PersistentFlags().Lookup(startTimeFlag)); err != nil {
		return nil, fmt.Errorf("failed to bind '%s' pflag: %w", startTimeFlag, err)
	}
	if err := viper.BindPFlag(endTimeFlag, dataCmd.PersistentFlags().Lookup(endTimeFlag)); err != nil {
		return nil, fmt.Errorf("failed to bind '%s' pflag: %w", endTimeFlag, err)
	}
	return dataCmd, nil
}

func metricData(_ *cobra.Command, _ []string) error {
	logger.V(0).Info("Preparing data client")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	projectID, err := effectiveProjectID(ctx)
	if err != nil {
		return err
	}
	startTime, err := buildTimestamp(viper.GetString(startTimeFlag), time.Now().Add(-5*time.Minute))
	if err != nil {
		return err
	}
	endTime, err := buildTimestamp(viper.GetString(endTimeFlag), time.Now())
	if err != nil {
		return err
	}
	req := monitoringpb.ListTimeSeriesRequest{
		Name:   "projects/" + projectID,
		Filter: viper.GetString(filterFlagName),
		Interval: &monitoringpb.TimeInterval{
			StartTime: startTime,
			EndTime:   endTime,
		},
		PageSize:  0,
		PageToken: "",
	}
	client, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return fmt.Errorf("failure creating new metric client: %w", err)
	}
	defer func() {
		if err = client.Close(); err != nil {
			logger.Error(err, "Failed to close metric client")
		}
	}()
	it := client.ListTimeSeries(ctx, &req)
	for {
		response, err := it.Next()
		switch {
		case errors.Is(err, iterator.Done):
			return nil
		case err != nil:
			return fmt.Errorf("failure getting list of metrics: %w", err)
		default:
			fmt.Println(protojson.Format(response)) //nolint:forbidigo // The data subcommand writes to stdout deliberately
		}
	}
}

// Attempt to parse the supplied string as RFC3339, and return a Timestamp that
// is ready to use as a filter. The fallback value will be used if the string
// is empty.
func buildTimestamp(value string, fallback time.Time) (*timestamppb.Timestamp, error) {
	if value == "" {
		return timestamppb.New(fallback), nil
	}
	ts, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, fmt.Errorf("failed to parse as RFC3339: %w", err)
	}
	return timestamppb.New(ts), nil
}
