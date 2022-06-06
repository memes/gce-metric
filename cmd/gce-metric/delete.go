package main

import (
	"context"
	"fmt"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	"github.com/spf13/cobra"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

// Returns a new command object that performs
func newDeleteCommand() (*cobra.Command, error) {
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "delete custom metric timeseries",
		RunE:  deleteMetrics,
		Args:  cobra.MinimumNArgs(1),
	}
	if err := addGeneratorFlags(deleteCmd); err != nil {
		return nil, err
	}
	return deleteCmd, nil
}

func deleteMetrics(cmd *cobra.Command, args []string) error {
	logger.V(0).Info("Preparing delete client")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	projectID, err := effectiveProjectID(ctx)
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
