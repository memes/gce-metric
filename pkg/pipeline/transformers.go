package pipeline

import (
	"math"

	"github.com/memes/gce-metric/pkg/generators"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoredrespb "google.golang.org/genproto/googleapis/api/monitoredres"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Defines a function that takes a monitoring CreateTimeSeriesRequest object and
// a moment-in-time Metric object, and returns a modified CreateTimeSeriesRequest.
type Transformer func(*monitoringpb.CreateTimeSeriesRequest, generators.Metric) (*monitoringpb.CreateTimeSeriesRequest, error)

func NewCreateTimeSeriesRequestTransformer(projectID, metricType string, metricLabels map[string]string) Transformer {
	return func(_ *monitoringpb.CreateTimeSeriesRequest, metric generators.Metric) (*monitoringpb.CreateTimeSeriesRequest, error) {
		return &monitoringpb.CreateTimeSeriesRequest{
			Name: "projects/" + projectID,
			TimeSeries: []*monitoringpb.TimeSeries{
				{
					Metric: &metricpb.Metric{
						Type:   metricType,
						Labels: metricLabels,
					},
				},
			},
		}, nil
	}
}

func NewGenericMonitoredResourceTransformer(projectID, location, namespace, nodeID string) Transformer {
	return func(req *monitoringpb.CreateTimeSeriesRequest, metric generators.Metric) (*monitoringpb.CreateTimeSeriesRequest, error) {
		for _, series := range req.TimeSeries {
			series.Resource = &monitoredrespb.MonitoredResource{
				Type: "generic_node",
				Labels: map[string]string{
					"project_id": projectID,
					"location":   location,
					"namespace":  namespace,
					"node_id":    nodeID,
				},
			}
		}
		return req, nil
	}
}

func NewGCEMonitoredResourceTransformer(projectID, instanceID, zone string) Transformer {
	return func(req *monitoringpb.CreateTimeSeriesRequest, metric generators.Metric) (*monitoringpb.CreateTimeSeriesRequest, error) {
		for _, series := range req.TimeSeries {
			series.Resource = &monitoredrespb.MonitoredResource{
				Type: "gce_instance",
				Labels: map[string]string{
					"project_id":  projectID,
					"instance_id": instanceID,
					"zone":        zone,
				},
			}
		}
		return req, nil
	}
}

// Implements a PipelineTransformer that
func NewGKEMonitoredResourceTransformer(projectID, clusterName, namespaceID, instanceID, podID, containerName, zone string) Transformer {
	return func(req *monitoringpb.CreateTimeSeriesRequest, metric generators.Metric) (*monitoringpb.CreateTimeSeriesRequest, error) {
		for _, series := range req.TimeSeries {
			series.Resource = &monitoredrespb.MonitoredResource{
				Type: "gke_container",
				Labels: map[string]string{
					"project_id":     projectID,
					"cluster_name":   clusterName,
					"namespace_id":   namespaceID,
					"instance_id":    instanceID,
					"pod_id":         podID,
					"container_name": containerName,
					"zone":           zone,
				},
			}
		}
		return req, nil
	}
}

// Implements a PipelineTransformer that replaces the time-series point-in-time
// record with the embedded value in metric.
func DoubleTypedValueTransformer(req *monitoringpb.CreateTimeSeriesRequest, metric generators.Metric) (*monitoringpb.CreateTimeSeriesRequest, error) {
	for _, series := range req.TimeSeries {
		series.Points = []*monitoringpb.Point{
			{
				Interval: &monitoringpb.TimeInterval{
					StartTime: &timestamppb.Timestamp{
						Seconds: metric.Timestamp.Unix(),
					},
					EndTime: &timestamppb.Timestamp{
						Seconds: metric.Timestamp.Unix(),
					},
				},
				Value: &monitoringpb.TypedValue{
					Value: &monitoringpb.TypedValue_DoubleValue{
						DoubleValue: metric.Value,
					},
				},
			},
		}
	}
	return req, nil
}

// Implements a PipelineTransformer that replaces the time-series point-in-time
// record with the embedded value in metric after rounding to the nearest integer.
func IntegerTypeValueTransformer(req *monitoringpb.CreateTimeSeriesRequest, metric generators.Metric) (*monitoringpb.CreateTimeSeriesRequest, error) {
	for _, series := range req.TimeSeries {
		series.Points = []*monitoringpb.Point{
			{
				Interval: &monitoringpb.TimeInterval{
					StartTime: &timestamppb.Timestamp{
						Seconds: metric.Timestamp.Unix(),
					},
					EndTime: &timestamppb.Timestamp{
						Seconds: metric.Timestamp.Unix(),
					},
				},
				Value: &monitoringpb.TypedValue{
					Value: &monitoringpb.TypedValue_Int64Value{
						Int64Value: int64(math.Round(metric.Value)),
					},
				},
			},
		}
	}
	return req, nil
}
