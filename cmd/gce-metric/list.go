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
)

const (
	filterFlagName = "filter"
	jsonFlagName   = "json"
)

func newListCommand() (*cobra.Command, error) {
	listCmd := &cobra.Command{
		Use:     "list [--verbose] [--project ID] [--filter FILTER] [--json]",
		Short:   "List Google Cloud time-series metrics that match the filter",
		Long:    "List any Google Cloud time-series metrics that match the filter, including those reserved for Google Cloud use. The default filter will match any time-series with the prefix name 'custom.googleapis.com', which is the recommended prefix for custom metrics. Use the --json flag to include a dump of the metric descriptor.",
		Example: appName + ` list --project ID --filter 'metric.type = has_substring("my-resource")' --json`,
		RunE:    listMain,
	}
	listCmd.PersistentFlags().String(filterFlagName, "metric.type = starts_with(\"custom.googleapis.com/\")", "set the filter to use when listing metrics")
	listCmd.PersistentFlags().Bool(jsonFlagName, false, "output the descriptor for each matching metric as JSON")
	if err := viper.BindPFlag(filterFlagName, listCmd.PersistentFlags().Lookup(filterFlagName)); err != nil {
		return nil, fmt.Errorf("failed to bind '%s' pflag: %w", filterFlagName, err)
	}
	if err := viper.BindPFlag(jsonFlagName, listCmd.PersistentFlags().Lookup(jsonFlagName)); err != nil {
		return nil, fmt.Errorf("failed to bind '%s' pflag: %w", jsonFlagName, err)
	}
	return listCmd, nil
}

func listMain(_ *cobra.Command, _ []string) error {
	logger.V(0).Info("Preparing list client")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	projectID, err := effectiveProjectID(ctx)
	if err != nil {
		return err
	}
	req := monitoringpb.ListMetricDescriptorsRequest{
		Name:       "projects/" + projectID,
		Filter:     viper.GetString(filterFlagName),
		PageSize:   0,
		PageToken:  "",
		ActiveOnly: false,
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
	it := client.ListMetricDescriptors(ctx, &req)
	for {
		response, err := it.Next()
		switch {
		case errors.Is(err, iterator.Done):
			return nil
		case err != nil:
			return fmt.Errorf("failure getting list of metrics: %w", err)
		case viper.GetBool(jsonFlagName):
			fmt.Println(protojson.Format(response)) //nolint:forbidigo // The user has requested that the names of matching metrics be printed to stdout
		default:
			fmt.Println(response.Type) //nolint:forbidigo // The user has requested that the names of matching metrics be printed to stdout
		}
	}
}
