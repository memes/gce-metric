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
	FilterFlagName = "filter"
	JSONFlagName   = "json"
)

func newListCommand() (*cobra.Command, error) {
	listCmd := &cobra.Command{
		Use:     "list [--verbose] [--project ID] [--filter FILTER] [--json]",
		Short:   "List Google Cloud time-series metrics that match the filter",
		Long:    "List any Google Cloud time-series metrics that match the filter, including those reserved for Google Cloud use. The default filter will match any time-series with the prefix name 'custom.googleapis.com', which is the recommended prefix for custom metrics. Use the --json flag to include a dump of the metric descriptor.",
		Example: AppName + ` list --project ID --filter 'metric.type = has_substring("my-resource")' --json`,
		RunE:    listMain,
	}
	listCmd.PersistentFlags().String(FilterFlagName, "metric.type = starts_with(\"custom.googleapis.com/\")", "set the filter to use when listing metrics")
	listCmd.PersistentFlags().Bool(JSONFlagName, false, "output the descriptor for each matching metric as JSON")
	if err := viper.BindPFlag(FilterFlagName, listCmd.PersistentFlags().Lookup(FilterFlagName)); err != nil {
		return nil, fmt.Errorf("failed to bind '%s' pflag: %w", FilterFlagName, err)
	}
	if err := viper.BindPFlag(JSONFlagName, listCmd.PersistentFlags().Lookup(JSONFlagName)); err != nil {
		return nil, fmt.Errorf("failed to bind '%s' pflag: %w", JSONFlagName, err)
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
		case viper.GetBool(JSONFlagName):
			fmt.Println(protojson.Format(response)) //nolint:forbidigo // The user has requested that the names of matching metrics be printed to stdout
		default:
			fmt.Println(response.Type) //nolint:forbidigo // The user has requested that the names of matching metrics be printed to stdout
		}
	}
}
