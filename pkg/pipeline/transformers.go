package pipeline

import (
	"errors"
	"math"

	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"github.com/memes/gce-metric/pkg/generators"
	monitoredrespb "google.golang.org/genproto/googleapis/api/monitoredres"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var ErrNilCreateTimeSeriesRequest = errors.New("transformer received nil as CreateTimeSeriesRequest")

// Defines a function that mutates a monitoring CreateTimeSeriesRequest object
// using the supplied moment-in-time Metric object.
type Transformer func(*monitoringpb.CreateTimeSeriesRequest, generators.Metric) error

// Returns a Transformer that will insert a generic_node resource into each
// time-series value.
func NewGenericMonitoredResourceTransformer(projectID, location, namespace, nodeID string) Transformer {
	return func(req *monitoringpb.CreateTimeSeriesRequest, _ generators.Metric) error {
		if req == nil {
			return ErrNilCreateTimeSeriesRequest
		}
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
		return nil
	}
}

// Returns a Transformer that will insert a gce_instance resource into each
// time-series value.
func NewGCEMonitoredResourceTransformer(projectID, instanceID, zone string) Transformer {
	return func(req *monitoringpb.CreateTimeSeriesRequest, _ generators.Metric) error {
		if req == nil {
			return ErrNilCreateTimeSeriesRequest
		}
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
		return nil
	}
}

// Returns a Transformer that will insert a gke_container resource into each
// time-series value.
func NewGKEMonitoredResourceTransformer(projectID, clusterName, namespaceID, instanceID, podID, containerName, zone string) Transformer {
	return func(req *monitoringpb.CreateTimeSeriesRequest, _ generators.Metric) error {
		if req == nil {
			return ErrNilCreateTimeSeriesRequest
		}
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
		return nil
	}
}

// Returns a Transformer that replaces the time-series point-in-time record with
// the embedded value in metric.
func NewDoubleTypedValueTransformer() Transformer {
	return func(req *monitoringpb.CreateTimeSeriesRequest, metric generators.Metric) error {
		if req == nil {
			return ErrNilCreateTimeSeriesRequest
		}
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
		return nil
	}
}

// Returns a Transformer that replaces the time-series point-in-time record with
// the embedded value in metric after rounding to the nearest integer.
func NewIntegerTypedValueTransformer() Transformer {
	return func(req *monitoringpb.CreateTimeSeriesRequest, metric generators.Metric) error {
		if req == nil {
			return ErrNilCreateTimeSeriesRequest
		}
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
		return nil
	}
}

// Returns a Transformer that will insert a k8s_cluster resource into each
// time-series value.
func NewGenericKubernetesClusterMonitoredResourceTransformer(projectID, location, clusterName string) Transformer {
	return func(req *monitoringpb.CreateTimeSeriesRequest, _ generators.Metric) error {
		if req == nil {
			return ErrNilCreateTimeSeriesRequest
		}
		for _, series := range req.TimeSeries {
			series.Resource = &monitoredrespb.MonitoredResource{
				Type: "k8s_cluster",
				Labels: map[string]string{
					"project_id":   projectID,
					"location":     location,
					"cluster_name": clusterName,
				},
			}
		}
		return nil
	}
}

// Returns a Transformer that will insert a k8s_container resource into each
// time-series value.
func NewGenericKubernetesContainerMonitoredResourceTransformer(projectID, location, clusterName, namespaceID, podID, containerName string) Transformer {
	return func(req *monitoringpb.CreateTimeSeriesRequest, _ generators.Metric) error {
		if req == nil {
			return ErrNilCreateTimeSeriesRequest
		}
		for _, series := range req.TimeSeries {
			series.Resource = &monitoredrespb.MonitoredResource{
				Type: "k8s_container",
				Labels: map[string]string{
					"project_id":     projectID,
					"location":       location,
					"cluster_name":   clusterName,
					"namespace_name": namespaceID,
					"pod_name":       podID,
					"container_name": containerName,
				},
			}
		}
		return nil
	}
}

// Returns a Transformer that will insert a k8s_node resource into each
// time-series value.
func NewGenericKubernetesNodeMonitoredResourceTransformer(projectID, location, clusterName, nodeName string) Transformer {
	return func(req *monitoringpb.CreateTimeSeriesRequest, _ generators.Metric) error {
		if req == nil {
			return ErrNilCreateTimeSeriesRequest
		}
		for _, series := range req.TimeSeries {
			series.Resource = &monitoredrespb.MonitoredResource{
				Type: "k8s_node",
				Labels: map[string]string{
					"project_id":   projectID,
					"location":     location,
					"cluster_name": clusterName,
					"node_name":    nodeName,
				},
			}
		}
		return nil
	}
}

// Returns a Transformer that will insert a k8s_pod resource into each time-series
// value.
func NewGenericKubernetesPodMonitoredResourceTransformer(projectID, location, clusterName, namespaceID, podID string) Transformer {
	return func(req *monitoringpb.CreateTimeSeriesRequest, _ generators.Metric) error {
		if req == nil {
			return ErrNilCreateTimeSeriesRequest
		}
		for _, series := range req.TimeSeries {
			series.Resource = &monitoredrespb.MonitoredResource{
				Type: "k8s_pod",
				Labels: map[string]string{
					"project_id":     projectID,
					"location":       location,
					"cluster_name":   clusterName,
					"namespace_name": namespaceID,
					"pod_name":       podID,
				},
			}
		}
		return nil
	}
}
