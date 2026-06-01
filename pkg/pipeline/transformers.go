package pipeline

import (
	"errors"
	"math"

	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"github.com/memes/gce-metric/pkg/generators"
	monitoredrespb "google.golang.org/genproto/googleapis/api/monitoredres"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	containerNameKey = "container_name"
	clusterNameKey   = "cluster_name"
	instanceIDKey    = "instance_id"
	locationKey      = "location"
	namespaceIDKey   = "namespace_id"
	namespaceNameKey = "namespace_name"
	namespaceKey     = "namespace"
	nodeIDKey        = "node_id"
	nodeNameKey      = "node_name"
	podIDKey         = "pod_id"
	podNameKey       = "pod_name"
	projectIDKey     = "project_id"
	zoneKey          = "zone"
)

// ErrNilCreateTimeSeriesRequest is returned when a Transformer receives a nil CreateTimeSeriesRequest.
var ErrNilCreateTimeSeriesRequest = errors.New("transformer received nil as CreateTimeSeriesRequest")

// Transformer defines a function that mutates a monitoring CreateTimeSeriesRequest object using the supplied
// moment-in-time Metric object.
type Transformer func(*monitoringpb.CreateTimeSeriesRequest, generators.Metric) error

// NewGenericMonitoredResourceTransformer returns a Transformer that will insert a generic_node resource into each
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
					projectIDKey: projectID,
					locationKey:  location,
					namespaceKey: namespace,
					nodeIDKey:    nodeID,
				},
			}
		}
		return nil
	}
}

// NewGCEMonitoredResourceTransformer returns a Transformer that will insert a gce_instance resource into each
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
					projectIDKey:  projectID,
					instanceIDKey: instanceID,
					zoneKey:       zone,
				},
			}
		}
		return nil
	}
}

// NewGKEMonitoredResourceTransformer returns a Transformer that will insert a gke_container resource into each
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
					projectIDKey:     projectID,
					clusterNameKey:   clusterName,
					namespaceIDKey:   namespaceID,
					instanceIDKey:    instanceID,
					podIDKey:         podID,
					containerNameKey: containerName,
					zoneKey:          zone,
				},
			}
		}
		return nil
	}
}

// NewDoubleTypedValueTransformer returns a Transformer that replaces the time-series point-in-time record with the
// embedded value in metric.
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

// NewIntegerTypedValueTransformer returns a Transformer that replaces the time-series point-in-time record with the
// embedded value in metric after rounding to the nearest integer.
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

// NewGenericKubernetesClusterMonitoredResourceTransformer returns a Transformer that will insert a k8s_cluster resource
// into each time-series value.
func NewGenericKubernetesClusterMonitoredResourceTransformer(projectID, location, clusterName string) Transformer {
	return func(req *monitoringpb.CreateTimeSeriesRequest, _ generators.Metric) error {
		if req == nil {
			return ErrNilCreateTimeSeriesRequest
		}
		for _, series := range req.TimeSeries {
			series.Resource = &monitoredrespb.MonitoredResource{
				Type: "k8s_cluster",
				Labels: map[string]string{
					projectIDKey:   projectID,
					locationKey:    location,
					clusterNameKey: clusterName,
				},
			}
		}
		return nil
	}
}

// NewGenericKubernetesContainerMonitoredResourceTransformer returns a Transformer that will insert a k8s_container
// resource into each time-series value.
func NewGenericKubernetesContainerMonitoredResourceTransformer(projectID, location, clusterName, namespaceID, podID, containerName string) Transformer {
	return func(req *monitoringpb.CreateTimeSeriesRequest, _ generators.Metric) error {
		if req == nil {
			return ErrNilCreateTimeSeriesRequest
		}
		for _, series := range req.TimeSeries {
			series.Resource = &monitoredrespb.MonitoredResource{
				Type: "k8s_container",
				Labels: map[string]string{
					projectIDKey:     projectID,
					locationKey:      location,
					clusterNameKey:   clusterName,
					namespaceNameKey: namespaceID,
					podNameKey:       podID,
					containerNameKey: containerName,
				},
			}
		}
		return nil
	}
}

// NewGenericKubernetesNodeMonitoredResourceTransformer returns a Transformer that will insert a k8s_node resource into
// each time-series value.
func NewGenericKubernetesNodeMonitoredResourceTransformer(projectID, location, clusterName, nodeName string) Transformer {
	return func(req *monitoringpb.CreateTimeSeriesRequest, _ generators.Metric) error {
		if req == nil {
			return ErrNilCreateTimeSeriesRequest
		}
		for _, series := range req.TimeSeries {
			series.Resource = &monitoredrespb.MonitoredResource{
				Type: "k8s_node",
				Labels: map[string]string{
					projectIDKey:   projectID,
					locationKey:    location,
					clusterNameKey: clusterName,
					nodeNameKey:    nodeName,
				},
			}
		}
		return nil
	}
}

// NewGenericKubernetesPodMonitoredResourceTransformer returns a Transformer that will insert a k8s_pod resource into
// each time-series value.
func NewGenericKubernetesPodMonitoredResourceTransformer(projectID, location, clusterName, namespaceID, podID string) Transformer {
	return func(req *monitoringpb.CreateTimeSeriesRequest, _ generators.Metric) error {
		if req == nil {
			return ErrNilCreateTimeSeriesRequest
		}
		for _, series := range req.TimeSeries {
			series.Resource = &monitoredrespb.MonitoredResource{
				Type: "k8s_pod",
				Labels: map[string]string{
					projectIDKey:     projectID,
					locationKey:      location,
					clusterNameKey:   clusterName,
					namespaceNameKey: namespaceID,
					podNameKey:       podID,
				},
			}
		}
		return nil
	}
}
