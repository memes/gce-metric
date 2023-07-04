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
)

const FilterFlagName = "filter"

func newListCommand() (*cobra.Command, error) {
	listCmd := &cobra.Command{
		Use:     "list [--verbose] [--project ID] [--filter FILTER]",
		Short:   "List Google Cloud time-series metrics that match the filter",
		Long:    "List any Google Cloud time-series metrics that match the filter, including those reserved for Google Cloud use. The default filter will match any time-series with the prefix name 'custom.googleapis.com', which is the recommended prefix for custom metrics.",
		Example: AppName + `list --project ID --filter 'metric.type = has_substring("my-resource")'`,
		RunE:    listMain,
	}
	listCmd.PersistentFlags().String(FilterFlagName, "metric.type = starts_with(\"custom.googleapis.com/\")", "set the filter to use when listing metrics")
	if err := viper.BindPFlag(FilterFlagName, listCmd.PersistentFlags().Lookup(FilterFlagName)); err != nil {
		return nil, fmt.Errorf("failed to bind '%s' pflag: %w", FilterFlagName, err)
	}
	return listCmd, nil
}

func listMain(_ *cobra.Command, _ []string) error {
	logger.V(0).Info("Preparing list client")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	projectID, err := effectiveProjectID()
	if err != nil {
		return err
	}
	req := monitoringpb.ListMetricDescriptorsRequest{
		Name:      "projects/" + projectID,
		Filter:    viper.GetString(FilterFlagName),
		PageSize:  0,
		PageToken: "",
	}
	client, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return fmt.Errorf("failure creating new metric client: %w", err)
	}
	defer client.Close()
	it := client.ListMetricDescriptors(ctx, &req)
	for {
		response, err := it.Next()
		switch {
		case errors.Is(err, iterator.Done):
			return nil
		case err != nil:
			return fmt.Errorf("failure getting list of metrics: %w", err)

		default:
			fmt.Println(response.Type) //nolint:forbidigo // The user has requested that the names of matching metrics be printed to stdout
		}
	}
}
