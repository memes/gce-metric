package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/api/iterator"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

func newListCommand() (*cobra.Command, error) {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "list custom metric timeseries",
		RunE:  listMain,
	}
	listCmd.PersistentFlags().String("filter", "metric.type = starts_with(\"custom.googleapis.com/\")", "set the filter to use when listing metrics")
	if err := viper.BindPFlag("filter", listCmd.PersistentFlags().Lookup("filter")); err != nil {
		return nil, fmt.Errorf("failed to bind filter pflag: %w", err)
	}
	return listCmd, nil
}

func listMain(cmd *cobra.Command, _ []string) error {
	logger.V(0).Info("Preparing list client")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	projectID, err := effectiveProjectID(ctx)
	if err != nil {
		return err
	}
	req := monitoringpb.ListMetricDescriptorsRequest{
		Name: "projects/" + projectID,
	}
	if filter := viper.GetString("filter"); filter != "" {
		req.Filter = filter
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
			fmt.Println(response.Type)
		}
	}
}
