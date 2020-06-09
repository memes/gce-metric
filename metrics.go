package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/compute/metadata"
	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	timestamppb "github.com/golang/protobuf/ptypes/timestamp"
	"go.uber.org/zap"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoredrespb "google.golang.org/genproto/googleapis/api/monitoredres"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
	"github.com/google/uuid"
)

type metricConfig struct {
	projectID string
	round bool
	metricType string
	metricLabels map[string]string
	resourceType string
	resourceLabels map[string]string
}

func (m *metricConfig) Validate() error {
	if len(strings.TrimSpace(m.projectID)) < 1 {
		return fmt.Errorf("Project ID must be provided")
	}
	if len(strings.TrimSpace(m.metricType)) < 1 {
		return fmt.Errorf("MetricType must be provided")
	}
	if len(strings.TrimSpace(m.resourceType)) < 1 {
		return fmt.Errorf("ResourceType must be provided")
	}
	return nil
}

func DefaultMetricConfig() func(*metricConfig) error {
	return func(m *metricConfig) error {
		var err error
		if (metadata.OnGCE()) {
			// Get defaults from GCE metadata
			if m.projectID, err = metadata.ProjectID(); err != nil {
				return err
			}
			m.resourceType = "gce_instance"
			if m.resourceLabels["instance_id"], err = metadata.InstanceID(); err != nil {
				return err
			}
			if m.resourceLabels["zone"], err = metadata.Zone(); err != nil {
				return err
			}
		} else {
			m.resourceType = "generic_node"
			m.resourceLabels["location"] = "us-west1"
			m.resourceLabels["namespace"] = "gce-metric"
			m.resourceLabels["node_id"] = uuid.New().String()
		}
		return nil
	}
}

func ProjectID(projectID string) func(*metricConfig) error {
	return func(m *metricConfig) error {
		sanitised := strings.TrimSpace(projectID)
		if len(sanitised) > 0 {
			m.projectID = sanitised
		}
		return nil
	}
}

func IntegerMetric(round bool) func(*metricConfig) error {
	return func(m *metricConfig) error {
		m.round = round
		return nil
	}
}

func MetricType(metricType string) func(*metricConfig) error {
	return func(m* metricConfig) error {
		sanitised := strings.TrimSpace(metricType)
		if len(sanitised) > 0 {
			m.metricType = sanitised
		}
		return nil
	}
}

func MergeMetricLabels(metricLabels map[string]string) func(*metricConfig) error {
	return func(m* metricConfig) error {
		if len(metricLabels) > 0 {
			for k, v := range metricLabels {
				m.metricLabels[k] = v
			}
		}
		return nil
	}
}

func ResourceType(resourceType string) func(*metricConfig) error {
	return func(m* metricConfig) error {
		sanitised := strings.TrimSpace(resourceType)
		if len(sanitised) > 0 {
			m.resourceType = sanitised
		}
		return nil
	}
}

// Merge the key-value pairs in resourceLabels to the configuration resource labels.
func MergeResourceLabels(resourceLabels map[string]string) func(*metricConfig) error {
	return func(m* metricConfig) error {
		if len(resourceLabels) > 0 {
			for k, v := range resourceLabels {
				m.resourceLabels[k] = v
			}
		}
		return nil
	}
}

func NewMetricConfig(options...func(*metricConfig) error) (metricConfig, error) {
	metricConfig := metricConfig{
		metricLabels: make(map[string]string),
		resourceLabels: make(map[string]string),
	}
	for _, opt := range options{
		if err := opt(&metricConfig); err != nil {
			return metricConfig, err
		}
	}
	return metricConfig, nil
}

// Returns a Google Metrics Int64 typed value equivalent to rounded metric
func roundedMetricAdapter(value float64) *monitoringpb.TypedValue {
	return &monitoringpb.TypedValue{
		Value: &monitoringpb.TypedValue_Int64Value{
			Int64Value: int64(value),
		},
	}
}

func doubleMetricAdapter(value float64) *monitoringpb.TypedValue {
	return &monitoringpb.TypedValue{
		Value: &monitoringpb.TypedValue_DoubleValue{
			DoubleValue: value,
		},
	}
}

func NewCreateTimeSeriesRequestGenerator(logger *zap.SugaredLogger, metricConfig metricConfig) func(float64) *monitoringpb.CreateTimeSeriesRequest {
	requestLogger := logger.With(
		"projectID", metricConfig.projectID,
		"round", metricConfig.round,
	)
	requestLogger.Debug("Building new CreateTimeSeriesRequest generator")
	var metricValueAdapter func(float64) *monitoringpb.TypedValue
	if metricConfig.round {
		metricValueAdapter = roundedMetricAdapter
	} else {
		metricValueAdapter = doubleMetricAdapter
	}
	return func(value float64) *monitoringpb.CreateTimeSeriesRequest {
		ts := time.Now().Unix()
		return &monitoringpb.CreateTimeSeriesRequest {
			Name: "projects/" + metricConfig.projectID,
			TimeSeries: []*monitoringpb.TimeSeries{{
				Metric: &metricpb.Metric{
					Type: metricConfig.metricType,
					Labels: metricConfig.metricLabels,
				},
				Resource: &monitoredrespb.MonitoredResource{
					Type: metricConfig.resourceType,
					Labels: metricConfig.resourceLabels,
				},
				Points: []*monitoringpb.Point{{
					Interval: &monitoringpb.TimeInterval{
						StartTime: &timestamppb.Timestamp{
							Seconds: ts,
						},
						EndTime:   &timestamppb.Timestamp{
							Seconds: ts,
						},
					},
					Value: metricValueAdapter(value),
				}},
			}},
		}
	}
}

func sendMetric(ctx context.Context, request *monitoringpb.CreateTimeSeriesRequest) error {
	client, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return err
	}
	return client.CreateTimeSeries(ctx, request)
}