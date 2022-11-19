package main

import (
	"context"
	"fmt"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"github.com/spf13/cobra"
)

func newDeleteCommand() *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:   "delete [flags] metric-name ...",
		Short: "Delete time-series metric(s)",
		Long: `Delete Google Cloud time-series metrics from a GCP project. One or more fully-qualified metric names (e.g. "custom.googleapis.com/my-metric") must be provided, and each will be deleted in turn.

NOTE: This command can delete any metric given, including built-in Google Cloud metrics, provided the caller has the correct permissions.`,
		Example: "gce-metric delete --verbose --project my-google-project custom.googleapis.com/my-metric",
		RunE:    deleteMetrics,
		Args:    cobra.MinimumNArgs(1),
	}
	return deleteCmd
}

func deleteMetrics(cmd *cobra.Command, args []string) error {
	logger.V(0).Info("Preparing delete client")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	projectID, err := effectiveProjectID()
	if err != nil {
		return err
	}
	client, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return fmt.Errorf("failure creating new metric client: %w", err)
	}
	defer client.Close()
	for _, metricType := range args {
		request := &monitoringpb.DeleteMetricDescriptorRequest{
			Name: "projects/" + projectID + "/metricDescriptors/" + metricType,
		}
		if err := client.DeleteMetricDescriptor(ctx, request); err != nil {
			return fmt.Errorf("failure deleting metric descriptor: %w", err)
		}
		logger.V(0).Info("Custom metric deleted", "metricType", metricType)
	}
	return nil
}
