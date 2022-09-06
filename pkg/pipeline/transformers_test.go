//nolint:dupl // All transformer tests have almost identical actions and test cases
package pipeline_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/memes/gce-metric/pkg/generators"
	"github.com/memes/gce-metric/pkg/pipeline"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoredrespb "google.golang.org/genproto/googleapis/api/monitoredres"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	project       = "test-project"
	location      = "test-location"
	namespace     = "test-namespace"
	node          = "test-node"
	zone          = "test-zone"
	clusterName   = "test-cluster"
	instance      = "test-instance"
	pod           = "test-pod"
	containerName = "test-container"
)

var timestamp = time.Now()

// The NewGenericMonitoredResourceTransformer is expected to return a function
// that inserts or replaces the Resource field of every TimeSeries in the slice
// with a generic_node resource with expected field values. Any existing Metric
// or Point object should remain unchanged.
func TestNewGenericMonitoredResourceTransformer(t *testing.T) {
	transformer := pipeline.NewGenericMonitoredResourceTransformer(project, location, namespace, node)
	tests := []struct {
		name          string
		req           *monitoringpb.CreateTimeSeriesRequest
		metric        generators.Metric
		expected      *monitoringpb.CreateTimeSeriesRequest
		expectedError error
	}{
		{
			name:          "nil",
			req:           nil,
			expected:      nil,
			expectedError: pipeline.ErrNilCreateTimeSeriesRequest,
		},
		{
			name:     "default",
			req:      &monitoringpb.CreateTimeSeriesRequest{},
			expected: &monitoringpb.CreateTimeSeriesRequest{},
		},
		{
			name: "nil-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name:       "nil-series",
				TimeSeries: nil,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name:       "nil-series",
				TimeSeries: nil,
			},
		},
		{
			name: "empty-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name:       "empty-series",
				TimeSeries: []*monitoringpb.TimeSeries{},
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name:       "empty-series",
				TimeSeries: []*monitoringpb.TimeSeries{},
			},
		},
		{
			name: "insert-single-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name: "insert-single-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "insert-single-series",
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value",
									},
								},
							},
						},
					},
				},
			},
			metric: generators.Metric{
				Value:     1.1,
				Timestamp: timestamp,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name: "insert-single-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "insert-single-series",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "generic_node",
							Labels: map[string]string{
								"project_id": project,
								"location":   location,
								"namespace":  namespace,
								"node_id":    node,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "insert-multiple-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name: "insert-multiple-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "insert-multiple-series-0",
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-0",
									},
								},
							},
						},
					},
					{
						Metric: &metricpb.Metric{
							Type: "insert-multiple-series-1",
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-1",
									},
								},
							},
						},
					},
				},
			},
			metric: generators.Metric{
				Value:     2.2,
				Timestamp: timestamp,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name: "insert-multiple-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "insert-multiple-series-0",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "generic_node",
							Labels: map[string]string{
								"project_id": project,
								"location":   location,
								"namespace":  namespace,
								"node_id":    node,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-0",
									},
								},
							},
						},
					},
					{
						Metric: &metricpb.Metric{
							Type: "insert-multiple-series-1",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "generic_node",
							Labels: map[string]string{
								"project_id": project,
								"location":   location,
								"namespace":  namespace,
								"node_id":    node,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-1",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "replace-single-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name: "replace-single-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "replace-single-series",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "replace-single-series",
							Labels: map[string]string{
								"project_id": "replace-single-series",
								"location":   "replace-single-series",
								"namespace":  "replace-single-series",
								"node_id":    "replace-single-series",
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value",
									},
								},
							},
						},
					},
				},
			},
			metric: generators.Metric{
				Value:     3.3,
				Timestamp: timestamp,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name: "replace-single-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "replace-single-series",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "generic_node",
							Labels: map[string]string{
								"project_id": project,
								"location":   location,
								"namespace":  namespace,
								"node_id":    node,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "replace-multiple-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name: "replace-multiple-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "replace-multiple-series-0",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "replace-multiple-series-0",
							Labels: map[string]string{
								"project_id": "replace-multiple-series-0",
								"location":   "replace-multiple-series-0",
								"namespace":  "replace-multiple-series-0",
								"node_id":    "replace-multiple-series-0",
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-0",
									},
								},
							},
						},
					},
					{
						Metric: &metricpb.Metric{
							Type: "replace-multiple-series-1",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "replace-multiple-series-1",
							Labels: map[string]string{
								"project_id": "replace-multiple-series-1",
								"location":   "replace-multiple-series-1",
								"namespace":  "replace-multiple-series-1",
								"node_id":    "replace-multiple-series-1",
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-1",
									},
								},
							},
						},
					},
				},
			},
			metric: generators.Metric{
				Value:     4.4,
				Timestamp: timestamp,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name: "replace-multiple-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "replace-multiple-series-0",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "generic_node",
							Labels: map[string]string{
								"project_id": project,
								"location":   location,
								"namespace":  namespace,
								"node_id":    node,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-0",
									},
								},
							},
						},
					},
					{
						Metric: &metricpb.Metric{
							Type: "replace-multiple-series-1",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "generic_node",
							Labels: map[string]string{
								"project_id": project,
								"location":   location,
								"namespace":  namespace,
								"node_id":    node,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-1",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	t.Parallel()
	for _, test := range tests {
		tst := test
		t.Run(tst.name, func(t *testing.T) {
			t.Parallel()
			err := transformer(tst.req, generators.Metric{})
			switch {
			case tst.expectedError == nil && err != nil:
				t.Errorf("Transformer raised an unexpected exception: %v", err)
			case tst.expectedError != nil && !errors.Is(err, tst.expectedError):
				t.Errorf("Expected transform to raise %v, got %v", tst.expectedError, err)
			case !reflect.DeepEqual(tst.expected, tst.req):
				t.Errorf("Expected %+v, got %+v", tst.expected, tst.req)
			}
		})
	}
}

// The NewGCEMonitoredResourceTransformer is expected to return a function
// that inserts or replaces the Resource field of every TimeSeries in the slice
// with a gce_instance resource with expected field values. Any existing Metric
// or Point object should remain unchanged.
func TestNewGCEMonitoredResourceTransformer(t *testing.T) {
	transformer := pipeline.NewGCEMonitoredResourceTransformer(project, instance, zone)
	tests := []struct {
		name          string
		req           *monitoringpb.CreateTimeSeriesRequest
		metric        generators.Metric
		expected      *monitoringpb.CreateTimeSeriesRequest
		expectedError error
	}{
		{
			name:          "nil",
			req:           nil,
			expected:      nil,
			expectedError: pipeline.ErrNilCreateTimeSeriesRequest,
		},
		{
			name:     "default",
			req:      &monitoringpb.CreateTimeSeriesRequest{},
			expected: &monitoringpb.CreateTimeSeriesRequest{},
		},
		{
			name: "nil-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name:       "nil-series",
				TimeSeries: nil,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name:       "nil-series",
				TimeSeries: nil,
			},
		},
		{
			name: "empty-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name:       "empty-series",
				TimeSeries: []*monitoringpb.TimeSeries{},
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name:       "empty-series",
				TimeSeries: []*monitoringpb.TimeSeries{},
			},
		},
		{
			name: "insert-single-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name: "insert-single-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "insert-single-series",
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value",
									},
								},
							},
						},
					},
				},
			},
			metric: generators.Metric{
				Value:     1.1,
				Timestamp: timestamp,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name: "insert-single-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "insert-single-series",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "gce_instance",
							Labels: map[string]string{
								"project_id":  project,
								"instance_id": instance,
								"zone":        zone,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "insert-multiple-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name: "insert-multiple-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "insert-multiple-series-0",
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-0",
									},
								},
							},
						},
					},
					{
						Metric: &metricpb.Metric{
							Type: "insert-multiple-series-1",
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-1",
									},
								},
							},
						},
					},
				},
			},
			metric: generators.Metric{
				Value:     2.2,
				Timestamp: timestamp,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name: "insert-multiple-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "insert-multiple-series-0",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "gce_instance",
							Labels: map[string]string{
								"project_id":  project,
								"instance_id": instance,
								"zone":        zone,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-0",
									},
								},
							},
						},
					},
					{
						Metric: &metricpb.Metric{
							Type: "insert-multiple-series-1",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "gce_instance",
							Labels: map[string]string{
								"project_id":  project,
								"instance_id": instance,
								"zone":        zone,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-1",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "replace-single-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name: "replace-single-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "replace-single-series",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "replace-single-series",
							Labels: map[string]string{
								"project_id": "replace-single-series",
								"location":   "replace-single-series",
								"namespace":  "replace-single-series",
								"node_id":    "replace-single-series",
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value",
									},
								},
							},
						},
					},
				},
			},
			metric: generators.Metric{
				Value:     3.3,
				Timestamp: timestamp,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name: "replace-single-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "replace-single-series",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "gce_instance",
							Labels: map[string]string{
								"project_id":  project,
								"instance_id": instance,
								"zone":        zone,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "replace-multiple-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name: "replace-multiple-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "replace-multiple-series-0",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "replace-multiple-series-0",
							Labels: map[string]string{
								"project_id": "replace-multiple-series-0",
								"location":   "replace-multiple-series-0",
								"namespace":  "replace-multiple-series-0",
								"node_id":    "replace-multiple-series-0",
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-0",
									},
								},
							},
						},
					},
					{
						Metric: &metricpb.Metric{
							Type: "replace-multiple-series-1",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "replace-multiple-series-1",
							Labels: map[string]string{
								"project_id": "replace-multiple-series-1",
								"location":   "replace-multiple-series-1",
								"namespace":  "replace-multiple-series-1",
								"node_id":    "replace-multiple-series-1",
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-1",
									},
								},
							},
						},
					},
				},
			},
			metric: generators.Metric{
				Value:     4.4,
				Timestamp: timestamp,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name: "replace-multiple-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "replace-multiple-series-0",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "gce_instance",
							Labels: map[string]string{
								"project_id":  project,
								"instance_id": instance,
								"zone":        zone,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-0",
									},
								},
							},
						},
					},
					{
						Metric: &metricpb.Metric{
							Type: "replace-multiple-series-1",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "gce_instance",
							Labels: map[string]string{
								"project_id":  project,
								"instance_id": instance,
								"zone":        zone,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-1",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	t.Parallel()
	for _, test := range tests {
		tst := test
		t.Run(tst.name, func(t *testing.T) {
			t.Parallel()
			err := transformer(tst.req, generators.Metric{})
			switch {
			case tst.expectedError == nil && err != nil:
				t.Errorf("Transformer raised an unexpected exception: %v", err)
			case tst.expectedError != nil && !errors.Is(err, tst.expectedError):
				t.Errorf("Expected transform to raise %v, got %v", tst.expectedError, err)
			case !reflect.DeepEqual(tst.expected, tst.req):
				t.Errorf("Expected %+v, got %+v", tst.expected, tst.req)
			}
		})
	}
}

// The NewGKEMonitoredResourceTransformer is expected to return a function
// that inserts or replaces the Resource field of every TimeSeries in the slice
// with a gke_container resource with expected field values. Any existing Metric
// or Point object should remain unchanged.
func TestNewGKEMonitoredResourceTransformer(t *testing.T) {
	transformer := pipeline.NewGKEMonitoredResourceTransformer(project, clusterName, namespace, instance, pod, containerName, zone)
	tests := []struct {
		name          string
		req           *monitoringpb.CreateTimeSeriesRequest
		metric        generators.Metric
		expected      *monitoringpb.CreateTimeSeriesRequest
		expectedError error
	}{
		{
			name:          "nil",
			req:           nil,
			expected:      nil,
			expectedError: pipeline.ErrNilCreateTimeSeriesRequest,
		},
		{
			name:     "default",
			req:      &monitoringpb.CreateTimeSeriesRequest{},
			expected: &monitoringpb.CreateTimeSeriesRequest{},
		},
		{
			name: "nil-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name:       "nil-series",
				TimeSeries: nil,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name:       "nil-series",
				TimeSeries: nil,
			},
		},
		{
			name: "empty-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name:       "empty-series",
				TimeSeries: []*monitoringpb.TimeSeries{},
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name:       "empty-series",
				TimeSeries: []*monitoringpb.TimeSeries{},
			},
		},
		{
			name: "insert-single-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name: "insert-single-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "insert-single-series",
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value",
									},
								},
							},
						},
					},
				},
			},
			metric: generators.Metric{
				Value:     1.1,
				Timestamp: timestamp,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name: "insert-single-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "insert-single-series",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "gke_container",
							Labels: map[string]string{
								"project_id":     project,
								"cluster_name":   clusterName,
								"namespace_id":   namespace,
								"instance_id":    instance,
								"pod_id":         pod,
								"container_name": containerName,
								"zone":           zone,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "insert-multiple-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name: "insert-multiple-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "insert-multiple-series-0",
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-0",
									},
								},
							},
						},
					},
					{
						Metric: &metricpb.Metric{
							Type: "insert-multiple-series-1",
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-1",
									},
								},
							},
						},
					},
				},
			},
			metric: generators.Metric{
				Value:     2.2,
				Timestamp: timestamp,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name: "insert-multiple-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "insert-multiple-series-0",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "gke_container",
							Labels: map[string]string{
								"project_id":     project,
								"cluster_name":   clusterName,
								"namespace_id":   namespace,
								"instance_id":    instance,
								"pod_id":         pod,
								"container_name": containerName,
								"zone":           zone,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-0",
									},
								},
							},
						},
					},
					{
						Metric: &metricpb.Metric{
							Type: "insert-multiple-series-1",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "gke_container",
							Labels: map[string]string{
								"project_id":     project,
								"cluster_name":   clusterName,
								"namespace_id":   namespace,
								"instance_id":    instance,
								"pod_id":         pod,
								"container_name": containerName,
								"zone":           zone,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-1",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "replace-single-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name: "replace-single-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "replace-single-series",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "replace-single-series",
							Labels: map[string]string{
								"project_id": "replace-single-series",
								"location":   "replace-single-series",
								"namespace":  "replace-single-series",
								"node_id":    "replace-single-series",
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value",
									},
								},
							},
						},
					},
				},
			},
			metric: generators.Metric{
				Value:     3.3,
				Timestamp: timestamp,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name: "replace-single-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "replace-single-series",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "gke_container",
							Labels: map[string]string{
								"project_id":     project,
								"cluster_name":   clusterName,
								"namespace_id":   namespace,
								"instance_id":    instance,
								"pod_id":         pod,
								"container_name": containerName,
								"zone":           zone,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "replace-multiple-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name: "replace-multiple-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "replace-multiple-series-0",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "replace-multiple-series-0",
							Labels: map[string]string{
								"project_id": "replace-multiple-series-0",
								"location":   "replace-multiple-series-0",
								"namespace":  "replace-multiple-series-0",
								"node_id":    "replace-multiple-series-0",
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-0",
									},
								},
							},
						},
					},
					{
						Metric: &metricpb.Metric{
							Type: "replace-multiple-series-1",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "replace-multiple-series-1",
							Labels: map[string]string{
								"project_id": "replace-multiple-series-1",
								"location":   "replace-multiple-series-1",
								"namespace":  "replace-multiple-series-1",
								"node_id":    "replace-multiple-series-1",
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-1",
									},
								},
							},
						},
					},
				},
			},
			metric: generators.Metric{
				Value:     4.4,
				Timestamp: timestamp,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name: "replace-multiple-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "replace-multiple-series-0",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "gke_container",
							Labels: map[string]string{
								"project_id":     project,
								"cluster_name":   clusterName,
								"namespace_id":   namespace,
								"instance_id":    instance,
								"pod_id":         pod,
								"container_name": containerName,
								"zone":           zone,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-0",
									},
								},
							},
						},
					},
					{
						Metric: &metricpb.Metric{
							Type: "replace-multiple-series-1",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "gke_container",
							Labels: map[string]string{
								"project_id":     project,
								"cluster_name":   clusterName,
								"namespace_id":   namespace,
								"instance_id":    instance,
								"pod_id":         pod,
								"container_name": containerName,
								"zone":           zone,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-1",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	t.Parallel()
	for _, test := range tests {
		tst := test
		t.Run(tst.name, func(t *testing.T) {
			t.Parallel()
			err := transformer(tst.req, generators.Metric{})
			switch {
			case tst.expectedError == nil && err != nil:
				t.Errorf("Transformer raised an unexpected exception: %v", err)
			case tst.expectedError != nil && !errors.Is(err, tst.expectedError):
				t.Errorf("Expected transform to raise %v, got %v", tst.expectedError, err)
			case !reflect.DeepEqual(tst.expected, tst.req):
				t.Errorf("Expected %+v, got %+v", tst.expected, tst.req)
			}
		})
	}
}

// The DoubleTypedValueTransformer is expected to insert or replace the Points
// field of every TimeSeries in the slice with a DoubleValue and timestamps from
// the supplied Metric object. All other TimeSeries objects should remain unchanged.
func TestNewDoubleTypedValueTransformer(t *testing.T) {
	transformer := pipeline.NewDoubleTypedValueTransformer()
	tests := []struct {
		name          string
		req           *monitoringpb.CreateTimeSeriesRequest
		metric        generators.Metric
		expected      *monitoringpb.CreateTimeSeriesRequest
		expectedError error
	}{
		{
			name:          "nil",
			req:           nil,
			metric:        generators.Metric{},
			expected:      nil,
			expectedError: pipeline.ErrNilCreateTimeSeriesRequest,
		},
		{
			name:     "default",
			req:      &monitoringpb.CreateTimeSeriesRequest{},
			expected: &monitoringpb.CreateTimeSeriesRequest{},
		},
		{
			name: "nil-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name:       "nil-series",
				TimeSeries: nil,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name:       "nil-series",
				TimeSeries: nil,
			},
		},
		{
			name: "empty-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name:       "empty-series",
				TimeSeries: []*monitoringpb.TimeSeries{},
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name:       "empty-series",
				TimeSeries: []*monitoringpb.TimeSeries{},
			},
		},
		{
			name: "insert-single-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name: "insert-single-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "insert-single-series",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "insert-single-series",
							Labels: map[string]string{
								"project_id": "insert-single-series",
								"location":   "insert-single-series",
								"namespace":  "insert-single-series",
								"node_id":    "insert-single-series",
							},
						},
					},
				},
			},
			metric: generators.Metric{
				Value:     1.1,
				Timestamp: timestamp,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name: "insert-single-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "insert-single-series",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "insert-single-series",
							Labels: map[string]string{
								"project_id": "insert-single-series",
								"location":   "insert-single-series",
								"namespace":  "insert-single-series",
								"node_id":    "insert-single-series",
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_DoubleValue{
										DoubleValue: 1.1,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "insert-multiple-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name: "insert-multiple-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "insert-multiple-series-0",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "insert-multiple-series-0",
							Labels: map[string]string{
								"project_id": "insert-multiple-series-0",
								"location":   "insert-multiple-series-0",
								"namespace":  "insert-multiple-series-0",
								"node_id":    "insert-multiple-series-0",
							},
						},
					},
					{
						Metric: &metricpb.Metric{
							Type: "insert-multiple-series-1",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "insert-multiple-series-1",
							Labels: map[string]string{
								"project_id": "insert-multiple-series-1",
								"location":   "insert-multiple-series-1",
								"namespace":  "insert-multiple-series-1",
								"node_id":    "insert-multiple-series-1",
							},
						},
					},
				},
			},
			metric: generators.Metric{
				Value:     2.2,
				Timestamp: timestamp,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name: "insert-multiple-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "insert-multiple-series-0",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "insert-multiple-series-0",
							Labels: map[string]string{
								"project_id": "insert-multiple-series-0",
								"location":   "insert-multiple-series-0",
								"namespace":  "insert-multiple-series-0",
								"node_id":    "insert-multiple-series-0",
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_DoubleValue{
										DoubleValue: 2.2,
									},
								},
							},
						},
					},
					{
						Metric: &metricpb.Metric{
							Type: "insert-multiple-series-1",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "insert-multiple-series-1",
							Labels: map[string]string{
								"project_id": "insert-multiple-series-1",
								"location":   "insert-multiple-series-1",
								"namespace":  "insert-multiple-series-1",
								"node_id":    "insert-multiple-series-1",
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_DoubleValue{
										DoubleValue: 2.2,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "replace-single-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name: "replace-single-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "replace-single-series",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "replace-single-series",
							Labels: map[string]string{
								"project_id": "replace-single-series",
								"location":   "replace-single-series",
								"namespace":  "replace-single-series",
								"node_id":    "replace-single-series",
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value",
									},
								},
							},
						},
					},
				},
			},
			metric: generators.Metric{
				Value:     3.3,
				Timestamp: timestamp,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name: "replace-single-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "replace-single-series",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "replace-single-series",
							Labels: map[string]string{
								"project_id": "replace-single-series",
								"location":   "replace-single-series",
								"namespace":  "replace-single-series",
								"node_id":    "replace-single-series",
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_DoubleValue{
										DoubleValue: 3.3,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "replace-multiple-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name: "replace-multiple-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "replace-multiple-series-0",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "replace-multiple-series-0",
							Labels: map[string]string{
								"project_id": "replace-multiple-series-0",
								"location":   "replace-multiple-series-0",
								"namespace":  "replace-multiple-series-0",
								"node_id":    "replace-multiple-series-0",
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-0",
									},
								},
							},
						},
					},
					{
						Metric: &metricpb.Metric{
							Type: "replace-multiple-series-1",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "replace-multiple-series-1",
							Labels: map[string]string{
								"project_id": "replace-multiple-series-1",
								"location":   "replace-multiple-series-1",
								"namespace":  "replace-multiple-series-1",
								"node_id":    "replace-multiple-series-1",
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-1",
									},
								},
							},
						},
					},
				},
			},
			metric: generators.Metric{
				Value:     4.4,
				Timestamp: timestamp,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name: "replace-multiple-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "replace-multiple-series-0",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "replace-multiple-series-0",
							Labels: map[string]string{
								"project_id": "replace-multiple-series-0",
								"location":   "replace-multiple-series-0",
								"namespace":  "replace-multiple-series-0",
								"node_id":    "replace-multiple-series-0",
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_DoubleValue{
										DoubleValue: 4.4,
									},
								},
							},
						},
					},
					{
						Metric: &metricpb.Metric{
							Type: "replace-multiple-series-1",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "replace-multiple-series-1",
							Labels: map[string]string{
								"project_id": "replace-multiple-series-1",
								"location":   "replace-multiple-series-1",
								"namespace":  "replace-multiple-series-1",
								"node_id":    "replace-multiple-series-1",
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_DoubleValue{
										DoubleValue: 4.4,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	t.Parallel()
	for _, test := range tests {
		tst := test
		t.Run(tst.name, func(t *testing.T) {
			t.Parallel()
			err := transformer(tst.req, tst.metric)
			switch {
			case tst.expectedError == nil && err != nil:
				t.Errorf("Transformer raised an unexpected exception: %v", err)
			case tst.expectedError != nil && !errors.Is(err, tst.expectedError):
				t.Errorf("Expected transform to raise %v, got %v", tst.expectedError, err)
			case !reflect.DeepEqual(tst.expected, tst.req):
				t.Errorf("Expected %+v, got %+v", tst.expected, tst.req)
			}
		})
	}
}

// The IntegerTypedValueTransformer is expected to insert or replace the Points
// field of every TimeSeries in the slice with a IntegerValue and timestamps from
// the supplied Metric object. All other TimeSeries objects should remain unchanged.
func TestNewIntegerTypedValueTransformer(t *testing.T) {
	transformer := pipeline.NewIntegerTypedValueTransformer()
	tests := []struct {
		name          string
		req           *monitoringpb.CreateTimeSeriesRequest
		metric        generators.Metric
		expected      *monitoringpb.CreateTimeSeriesRequest
		expectedError error
	}{
		{
			name:          "nil",
			req:           nil,
			expected:      nil,
			expectedError: pipeline.ErrNilCreateTimeSeriesRequest,
		},
		{
			name:     "default",
			req:      &monitoringpb.CreateTimeSeriesRequest{},
			expected: &monitoringpb.CreateTimeSeriesRequest{},
		},
		{
			name: "nil-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name:       "nil-series",
				TimeSeries: nil,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name:       "nil-series",
				TimeSeries: nil,
			},
		},
		{
			name: "empty-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name:       "empty-series",
				TimeSeries: []*monitoringpb.TimeSeries{},
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name:       "empty-series",
				TimeSeries: []*monitoringpb.TimeSeries{},
			},
		},
		{
			name: "insert-single-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name: "insert-single-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "insert-single-series",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "insert-single-series",
							Labels: map[string]string{
								"project_id": "insert-single-series",
								"location":   "insert-single-series",
								"namespace":  "insert-single-series",
								"node_id":    "insert-single-series",
							},
						},
					},
				},
			},
			metric: generators.Metric{
				Value:     1.1,
				Timestamp: timestamp,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name: "insert-single-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "insert-single-series",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "insert-single-series",
							Labels: map[string]string{
								"project_id": "insert-single-series",
								"location":   "insert-single-series",
								"namespace":  "insert-single-series",
								"node_id":    "insert-single-series",
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_Int64Value{
										Int64Value: 1,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "insert-multiple-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name: "insert-multiple-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "insert-multiple-series-0",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "insert-multiple-series-0",
							Labels: map[string]string{
								"project_id": "insert-multiple-series-0",
								"location":   "insert-multiple-series-0",
								"namespace":  "insert-multiple-series-0",
								"node_id":    "insert-multiple-series-0",
							},
						},
					},
					{
						Metric: &metricpb.Metric{
							Type: "insert-multiple-series-1",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "insert-multiple-series-1",
							Labels: map[string]string{
								"project_id": "insert-multiple-series-1",
								"location":   "insert-multiple-series-1",
								"namespace":  "insert-multiple-series-1",
								"node_id":    "insert-multiple-series-1",
							},
						},
					},
				},
			},
			metric: generators.Metric{
				Value:     2.2,
				Timestamp: timestamp,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name: "insert-multiple-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "insert-multiple-series-0",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "insert-multiple-series-0",
							Labels: map[string]string{
								"project_id": "insert-multiple-series-0",
								"location":   "insert-multiple-series-0",
								"namespace":  "insert-multiple-series-0",
								"node_id":    "insert-multiple-series-0",
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_Int64Value{
										Int64Value: 2,
									},
								},
							},
						},
					},
					{
						Metric: &metricpb.Metric{
							Type: "insert-multiple-series-1",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "insert-multiple-series-1",
							Labels: map[string]string{
								"project_id": "insert-multiple-series-1",
								"location":   "insert-multiple-series-1",
								"namespace":  "insert-multiple-series-1",
								"node_id":    "insert-multiple-series-1",
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_Int64Value{
										Int64Value: 2,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "replace-single-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name: "replace-single-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "replace-single-series",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "replace-single-series",
							Labels: map[string]string{
								"project_id": "replace-single-series",
								"location":   "replace-single-series",
								"namespace":  "replace-single-series",
								"node_id":    "replace-single-series",
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value",
									},
								},
							},
						},
					},
				},
			},
			metric: generators.Metric{
				Value:     3.3,
				Timestamp: timestamp,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name: "replace-single-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "replace-single-series",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "replace-single-series",
							Labels: map[string]string{
								"project_id": "replace-single-series",
								"location":   "replace-single-series",
								"namespace":  "replace-single-series",
								"node_id":    "replace-single-series",
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_Int64Value{
										Int64Value: 3,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "replace-multiple-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name: "replace-multiple-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "replace-multiple-series-0",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "replace-multiple-series-0",
							Labels: map[string]string{
								"project_id": "replace-multiple-series-0",
								"location":   "replace-multiple-series-0",
								"namespace":  "replace-multiple-series-0",
								"node_id":    "replace-multiple-series-0",
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-0",
									},
								},
							},
						},
					},
					{
						Metric: &metricpb.Metric{
							Type: "replace-multiple-series-1",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "replace-multiple-series-1",
							Labels: map[string]string{
								"project_id": "replace-multiple-series-1",
								"location":   "replace-multiple-series-1",
								"namespace":  "replace-multiple-series-1",
								"node_id":    "replace-multiple-series-1",
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-1",
									},
								},
							},
						},
					},
				},
			},
			metric: generators.Metric{
				Value:     4.4,
				Timestamp: timestamp,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name: "replace-multiple-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "replace-multiple-series-0",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "replace-multiple-series-0",
							Labels: map[string]string{
								"project_id": "replace-multiple-series-0",
								"location":   "replace-multiple-series-0",
								"namespace":  "replace-multiple-series-0",
								"node_id":    "replace-multiple-series-0",
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_Int64Value{
										Int64Value: 4,
									},
								},
							},
						},
					},
					{
						Metric: &metricpb.Metric{
							Type: "replace-multiple-series-1",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "replace-multiple-series-1",
							Labels: map[string]string{
								"project_id": "replace-multiple-series-1",
								"location":   "replace-multiple-series-1",
								"namespace":  "replace-multiple-series-1",
								"node_id":    "replace-multiple-series-1",
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_Int64Value{
										Int64Value: 4,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	t.Parallel()
	for _, test := range tests {
		tst := test
		t.Run(tst.name, func(t *testing.T) {
			t.Parallel()
			err := transformer(tst.req, tst.metric)
			switch {
			case tst.expectedError == nil && err != nil:
				t.Errorf("Transformer raised an unexpected exception: %v", err)
			case tst.expectedError != nil && !errors.Is(err, tst.expectedError):
				t.Errorf("Expected transform to raise %v, got %v", tst.expectedError, err)
			case !reflect.DeepEqual(tst.expected, tst.req):
				t.Errorf("Expected %+v, got %+v", tst.expected, tst.req)
			}
		})
	}
}

// The NewGenericKubernetesClusterMonitoredResourceTransformer is expected to
// return a function that inserts or replaces the Resource field of every TimeSeries
// in the slice with a k8s_cluster resource with expected field values. Any existing
// Metric or Point object should remain unchanged.
func TestNewGenericKubernetesClusterMonitoredResourceTransformer(t *testing.T) {
	transformer := pipeline.NewGenericKubernetesClusterMonitoredResourceTransformer(project, location, clusterName)
	tests := []struct {
		name          string
		req           *monitoringpb.CreateTimeSeriesRequest
		metric        generators.Metric
		expected      *monitoringpb.CreateTimeSeriesRequest
		expectedError error
	}{
		{
			name:          "nil",
			req:           nil,
			expected:      nil,
			expectedError: pipeline.ErrNilCreateTimeSeriesRequest,
		},
		{
			name:     "default",
			req:      &monitoringpb.CreateTimeSeriesRequest{},
			expected: &monitoringpb.CreateTimeSeriesRequest{},
		},
		{
			name: "nil-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name:       "nil-series",
				TimeSeries: nil,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name:       "nil-series",
				TimeSeries: nil,
			},
		},
		{
			name: "empty-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name:       "empty-series",
				TimeSeries: []*monitoringpb.TimeSeries{},
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name:       "empty-series",
				TimeSeries: []*monitoringpb.TimeSeries{},
			},
		},
		{
			name: "insert-single-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name: "insert-single-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "insert-single-series",
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value",
									},
								},
							},
						},
					},
				},
			},
			metric: generators.Metric{
				Value:     1.1,
				Timestamp: timestamp,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name: "insert-single-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "insert-single-series",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "k8s_cluster",
							Labels: map[string]string{
								"project_id":   project,
								"location":     location,
								"cluster_name": clusterName,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "insert-multiple-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name: "insert-multiple-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "insert-multiple-series-0",
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-0",
									},
								},
							},
						},
					},
					{
						Metric: &metricpb.Metric{
							Type: "insert-multiple-series-1",
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-1",
									},
								},
							},
						},
					},
				},
			},
			metric: generators.Metric{
				Value:     2.2,
				Timestamp: timestamp,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name: "insert-multiple-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "insert-multiple-series-0",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "k8s_cluster",
							Labels: map[string]string{
								"project_id":   project,
								"location":     location,
								"cluster_name": clusterName,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-0",
									},
								},
							},
						},
					},
					{
						Metric: &metricpb.Metric{
							Type: "insert-multiple-series-1",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "k8s_cluster",
							Labels: map[string]string{
								"project_id":   project,
								"location":     location,
								"cluster_name": clusterName,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-1",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "replace-single-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name: "replace-single-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "replace-single-series",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "replace-single-series",
							Labels: map[string]string{
								"project_id": "replace-single-series",
								"location":   "replace-single-series",
								"namespace":  "replace-single-series",
								"node_id":    "replace-single-series",
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value",
									},
								},
							},
						},
					},
				},
			},
			metric: generators.Metric{
				Value:     3.3,
				Timestamp: timestamp,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name: "replace-single-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "replace-single-series",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "k8s_cluster",
							Labels: map[string]string{
								"project_id":   project,
								"location":     location,
								"cluster_name": clusterName,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "replace-multiple-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name: "replace-multiple-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "replace-multiple-series-0",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "replace-multiple-series-0",
							Labels: map[string]string{
								"project_id": "replace-multiple-series-0",
								"location":   "replace-multiple-series-0",
								"namespace":  "replace-multiple-series-0",
								"node_id":    "replace-multiple-series-0",
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-0",
									},
								},
							},
						},
					},
					{
						Metric: &metricpb.Metric{
							Type: "replace-multiple-series-1",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "replace-multiple-series-1",
							Labels: map[string]string{
								"project_id": "replace-multiple-series-1",
								"location":   "replace-multiple-series-1",
								"namespace":  "replace-multiple-series-1",
								"node_id":    "replace-multiple-series-1",
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-1",
									},
								},
							},
						},
					},
				},
			},
			metric: generators.Metric{
				Value:     4.4,
				Timestamp: timestamp,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name: "replace-multiple-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "replace-multiple-series-0",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "k8s_cluster",
							Labels: map[string]string{
								"project_id":   project,
								"location":     location,
								"cluster_name": clusterName,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-0",
									},
								},
							},
						},
					},
					{
						Metric: &metricpb.Metric{
							Type: "replace-multiple-series-1",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "k8s_cluster",
							Labels: map[string]string{
								"project_id":   project,
								"location":     location,
								"cluster_name": clusterName,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-1",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	t.Parallel()
	for _, test := range tests {
		tst := test
		t.Run(tst.name, func(t *testing.T) {
			t.Parallel()
			err := transformer(tst.req, generators.Metric{})
			switch {
			case tst.expectedError == nil && err != nil:
				t.Errorf("Transformer raised an unexpected exception: %v", err)
			case tst.expectedError != nil && !errors.Is(err, tst.expectedError):
				t.Errorf("Expected transform to raise %v, got %v", tst.expectedError, err)
			case !reflect.DeepEqual(tst.expected, tst.req):
				t.Errorf("Expected %+v, got %+v", tst.expected, tst.req)
			}
		})
	}
}

// The NewGenericKubernetesContainerMonitoredResourceTransformer is expected to
// return a function that inserts or replaces the Resource field of every TimeSeries
// in the slice with a k8s_container resource with expected field values. Any existing
// Metric or Point object should remain unchanged.
func TestNewGenericKubernetesContainerMonitoredResourceTransformer(t *testing.T) {
	transformer := pipeline.NewGenericKubernetesContainerMonitoredResourceTransformer(project, location, clusterName, namespace, pod, containerName)
	tests := []struct {
		name          string
		req           *monitoringpb.CreateTimeSeriesRequest
		metric        generators.Metric
		expected      *monitoringpb.CreateTimeSeriesRequest
		expectedError error
	}{
		{
			name:          "nil",
			req:           nil,
			expected:      nil,
			expectedError: pipeline.ErrNilCreateTimeSeriesRequest,
		},
		{
			name:     "default",
			req:      &monitoringpb.CreateTimeSeriesRequest{},
			expected: &monitoringpb.CreateTimeSeriesRequest{},
		},
		{
			name: "nil-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name:       "nil-series",
				TimeSeries: nil,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name:       "nil-series",
				TimeSeries: nil,
			},
		},
		{
			name: "empty-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name:       "empty-series",
				TimeSeries: []*monitoringpb.TimeSeries{},
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name:       "empty-series",
				TimeSeries: []*monitoringpb.TimeSeries{},
			},
		},
		{
			name: "insert-single-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name: "insert-single-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "insert-single-series",
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value",
									},
								},
							},
						},
					},
				},
			},
			metric: generators.Metric{
				Value:     1.1,
				Timestamp: timestamp,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name: "insert-single-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "insert-single-series",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "k8s_container",
							Labels: map[string]string{
								"project_id":     project,
								"location":       location,
								"cluster_name":   clusterName,
								"namespace_name": namespace,
								"pod_name":       pod,
								"container_name": containerName,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "insert-multiple-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name: "insert-multiple-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "insert-multiple-series-0",
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-0",
									},
								},
							},
						},
					},
					{
						Metric: &metricpb.Metric{
							Type: "insert-multiple-series-1",
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-1",
									},
								},
							},
						},
					},
				},
			},
			metric: generators.Metric{
				Value:     2.2,
				Timestamp: timestamp,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name: "insert-multiple-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "insert-multiple-series-0",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "k8s_container",
							Labels: map[string]string{
								"project_id":     project,
								"location":       location,
								"cluster_name":   clusterName,
								"namespace_name": namespace,
								"pod_name":       pod,
								"container_name": containerName,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-0",
									},
								},
							},
						},
					},
					{
						Metric: &metricpb.Metric{
							Type: "insert-multiple-series-1",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "k8s_container",
							Labels: map[string]string{
								"project_id":     project,
								"location":       location,
								"cluster_name":   clusterName,
								"namespace_name": namespace,
								"pod_name":       pod,
								"container_name": containerName,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-1",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "replace-single-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name: "replace-single-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "replace-single-series",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "replace-single-series",
							Labels: map[string]string{
								"project_id": "replace-single-series",
								"location":   "replace-single-series",
								"namespace":  "replace-single-series",
								"node_id":    "replace-single-series",
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value",
									},
								},
							},
						},
					},
				},
			},
			metric: generators.Metric{
				Value:     3.3,
				Timestamp: timestamp,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name: "replace-single-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "replace-single-series",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "k8s_container",
							Labels: map[string]string{
								"project_id":     project,
								"location":       location,
								"cluster_name":   clusterName,
								"namespace_name": namespace,
								"pod_name":       pod,
								"container_name": containerName,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "replace-multiple-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name: "replace-multiple-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "replace-multiple-series-0",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "replace-multiple-series-0",
							Labels: map[string]string{
								"project_id": "replace-multiple-series-0",
								"location":   "replace-multiple-series-0",
								"namespace":  "replace-multiple-series-0",
								"node_id":    "replace-multiple-series-0",
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-0",
									},
								},
							},
						},
					},
					{
						Metric: &metricpb.Metric{
							Type: "replace-multiple-series-1",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "replace-multiple-series-1",
							Labels: map[string]string{
								"project_id": "replace-multiple-series-1",
								"location":   "replace-multiple-series-1",
								"namespace":  "replace-multiple-series-1",
								"node_id":    "replace-multiple-series-1",
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-1",
									},
								},
							},
						},
					},
				},
			},
			metric: generators.Metric{
				Value:     4.4,
				Timestamp: timestamp,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name: "replace-multiple-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "replace-multiple-series-0",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "k8s_container",
							Labels: map[string]string{
								"project_id":     project,
								"location":       location,
								"cluster_name":   clusterName,
								"namespace_name": namespace,
								"pod_name":       pod,
								"container_name": containerName,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-0",
									},
								},
							},
						},
					},
					{
						Metric: &metricpb.Metric{
							Type: "replace-multiple-series-1",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "k8s_container",
							Labels: map[string]string{
								"project_id":     project,
								"location":       location,
								"cluster_name":   clusterName,
								"namespace_name": namespace,
								"pod_name":       pod,
								"container_name": containerName,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-1",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	t.Parallel()
	for _, test := range tests {
		tst := test
		t.Run(tst.name, func(t *testing.T) {
			t.Parallel()
			err := transformer(tst.req, generators.Metric{})
			switch {
			case tst.expectedError == nil && err != nil:
				t.Errorf("Transformer raised an unexpected exception: %v", err)
			case tst.expectedError != nil && !errors.Is(err, tst.expectedError):
				t.Errorf("Expected transform to raise %v, got %v", tst.expectedError, err)
			case !reflect.DeepEqual(tst.expected, tst.req):
				t.Errorf("Expected %+v, got %+v", tst.expected, tst.req)
			}
		})
	}
}

// The NewGenericKubernetesNodeMonitoredResourceTransformer is expected to
// return a function that inserts or replaces the Resource field of every TimeSeries
// in the slice with a k8s_node resource with expected field values. Any existing
// Metric or Point object should remain unchanged.
func TestNewGenericKubernetesNodeMonitoredResourceTransformer(t *testing.T) {
	transformer := pipeline.NewGenericKubernetesNodeMonitoredResourceTransformer(project, location, clusterName, node)
	tests := []struct {
		name          string
		req           *monitoringpb.CreateTimeSeriesRequest
		metric        generators.Metric
		expected      *monitoringpb.CreateTimeSeriesRequest
		expectedError error
	}{
		{
			name:          "nil",
			req:           nil,
			expected:      nil,
			expectedError: pipeline.ErrNilCreateTimeSeriesRequest,
		},
		{
			name:     "default",
			req:      &monitoringpb.CreateTimeSeriesRequest{},
			expected: &monitoringpb.CreateTimeSeriesRequest{},
		},
		{
			name: "nil-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name:       "nil-series",
				TimeSeries: nil,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name:       "nil-series",
				TimeSeries: nil,
			},
		},
		{
			name: "empty-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name:       "empty-series",
				TimeSeries: []*monitoringpb.TimeSeries{},
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name:       "empty-series",
				TimeSeries: []*monitoringpb.TimeSeries{},
			},
		},
		{
			name: "insert-single-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name: "insert-single-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "insert-single-series",
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value",
									},
								},
							},
						},
					},
				},
			},
			metric: generators.Metric{
				Value:     1.1,
				Timestamp: timestamp,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name: "insert-single-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "insert-single-series",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "k8s_node",
							Labels: map[string]string{
								"project_id":   project,
								"location":     location,
								"cluster_name": clusterName,
								"node_name":    node,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "insert-multiple-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name: "insert-multiple-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "insert-multiple-series-0",
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-0",
									},
								},
							},
						},
					},
					{
						Metric: &metricpb.Metric{
							Type: "insert-multiple-series-1",
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-1",
									},
								},
							},
						},
					},
				},
			},
			metric: generators.Metric{
				Value:     2.2,
				Timestamp: timestamp,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name: "insert-multiple-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "insert-multiple-series-0",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "k8s_node",
							Labels: map[string]string{
								"project_id":   project,
								"location":     location,
								"cluster_name": clusterName,
								"node_name":    node,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-0",
									},
								},
							},
						},
					},
					{
						Metric: &metricpb.Metric{
							Type: "insert-multiple-series-1",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "k8s_node",
							Labels: map[string]string{
								"project_id":   project,
								"location":     location,
								"cluster_name": clusterName,
								"node_name":    node,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-1",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "replace-single-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name: "replace-single-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "replace-single-series",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "replace-single-series",
							Labels: map[string]string{
								"project_id": "replace-single-series",
								"location":   "replace-single-series",
								"namespace":  "replace-single-series",
								"node_id":    "replace-single-series",
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value",
									},
								},
							},
						},
					},
				},
			},
			metric: generators.Metric{
				Value:     3.3,
				Timestamp: timestamp,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name: "replace-single-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "replace-single-series",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "k8s_node",
							Labels: map[string]string{
								"project_id":   project,
								"location":     location,
								"cluster_name": clusterName,
								"node_name":    node,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "replace-multiple-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name: "replace-multiple-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "replace-multiple-series-0",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "replace-multiple-series-0",
							Labels: map[string]string{
								"project_id": "replace-multiple-series-0",
								"location":   "replace-multiple-series-0",
								"namespace":  "replace-multiple-series-0",
								"node_id":    "replace-multiple-series-0",
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-0",
									},
								},
							},
						},
					},
					{
						Metric: &metricpb.Metric{
							Type: "replace-multiple-series-1",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "replace-multiple-series-1",
							Labels: map[string]string{
								"project_id": "replace-multiple-series-1",
								"location":   "replace-multiple-series-1",
								"namespace":  "replace-multiple-series-1",
								"node_id":    "replace-multiple-series-1",
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-1",
									},
								},
							},
						},
					},
				},
			},
			metric: generators.Metric{
				Value:     4.4,
				Timestamp: timestamp,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name: "replace-multiple-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "replace-multiple-series-0",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "k8s_node",
							Labels: map[string]string{
								"project_id":   project,
								"location":     location,
								"cluster_name": clusterName,
								"node_name":    node,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-0",
									},
								},
							},
						},
					},
					{
						Metric: &metricpb.Metric{
							Type: "replace-multiple-series-1",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "k8s_node",
							Labels: map[string]string{
								"project_id":   project,
								"location":     location,
								"cluster_name": clusterName,
								"node_name":    node,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-1",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	t.Parallel()
	for _, test := range tests {
		tst := test
		t.Run(tst.name, func(t *testing.T) {
			t.Parallel()
			err := transformer(tst.req, generators.Metric{})
			switch {
			case tst.expectedError == nil && err != nil:
				t.Errorf("Transformer raised an unexpected exception: %v", err)
			case tst.expectedError != nil && !errors.Is(err, tst.expectedError):
				t.Errorf("Expected transform to raise %v, got %v", tst.expectedError, err)
			case !reflect.DeepEqual(tst.expected, tst.req):
				t.Errorf("Expected %+v, got %+v", tst.expected, tst.req)
			}
		})
	}
}

// The NewGenericKubernetesPodMonitoredResourceTransformer is expected to
// return a function that inserts or replaces the Resource field of every TimeSeries
// in the slice with a k8s_pod resource with expected field values. Any existing
// Metric or Point object should remain unchanged.
func TestNewGenericKubernetesPodMonitoredResourceTransformer(t *testing.T) {
	transformer := pipeline.NewGenericKubernetesPodMonitoredResourceTransformer(project, location, clusterName, namespace, pod)
	tests := []struct {
		name          string
		req           *monitoringpb.CreateTimeSeriesRequest
		metric        generators.Metric
		expected      *monitoringpb.CreateTimeSeriesRequest
		expectedError error
	}{
		{
			name:          "nil",
			req:           nil,
			expected:      nil,
			expectedError: pipeline.ErrNilCreateTimeSeriesRequest,
		},
		{
			name:     "default",
			req:      &monitoringpb.CreateTimeSeriesRequest{},
			expected: &monitoringpb.CreateTimeSeriesRequest{},
		},
		{
			name: "nil-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name:       "nil-series",
				TimeSeries: nil,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name:       "nil-series",
				TimeSeries: nil,
			},
		},
		{
			name: "empty-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name:       "empty-series",
				TimeSeries: []*monitoringpb.TimeSeries{},
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name:       "empty-series",
				TimeSeries: []*monitoringpb.TimeSeries{},
			},
		},
		{
			name: "insert-single-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name: "insert-single-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "insert-single-series",
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value",
									},
								},
							},
						},
					},
				},
			},
			metric: generators.Metric{
				Value:     1.1,
				Timestamp: timestamp,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name: "insert-single-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "insert-single-series",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "k8s_pod",
							Labels: map[string]string{
								"project_id":     project,
								"location":       location,
								"cluster_name":   clusterName,
								"namespace_name": namespace,
								"pod_name":       pod,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "insert-multiple-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name: "insert-multiple-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "insert-multiple-series-0",
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-0",
									},
								},
							},
						},
					},
					{
						Metric: &metricpb.Metric{
							Type: "insert-multiple-series-1",
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-1",
									},
								},
							},
						},
					},
				},
			},
			metric: generators.Metric{
				Value:     2.2,
				Timestamp: timestamp,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name: "insert-multiple-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "insert-multiple-series-0",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "k8s_pod",
							Labels: map[string]string{
								"project_id":     project,
								"location":       location,
								"cluster_name":   clusterName,
								"namespace_name": namespace,
								"pod_name":       pod,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-0",
									},
								},
							},
						},
					},
					{
						Metric: &metricpb.Metric{
							Type: "insert-multiple-series-1",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "k8s_pod",
							Labels: map[string]string{
								"project_id":     project,
								"location":       location,
								"cluster_name":   clusterName,
								"namespace_name": namespace,
								"pod_name":       pod,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-1",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "replace-single-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name: "replace-single-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "replace-single-series",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "replace-single-series",
							Labels: map[string]string{
								"project_id": "replace-single-series",
								"location":   "replace-single-series",
								"namespace":  "replace-single-series",
								"node_id":    "replace-single-series",
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value",
									},
								},
							},
						},
					},
				},
			},
			metric: generators.Metric{
				Value:     3.3,
				Timestamp: timestamp,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name: "replace-single-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "replace-single-series",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "k8s_pod",
							Labels: map[string]string{
								"project_id":     project,
								"location":       location,
								"cluster_name":   clusterName,
								"namespace_name": namespace,
								"pod_name":       pod,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "replace-multiple-series",
			req: &monitoringpb.CreateTimeSeriesRequest{
				Name: "replace-multiple-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "replace-multiple-series-0",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "replace-multiple-series-0",
							Labels: map[string]string{
								"project_id": "replace-multiple-series-0",
								"location":   "replace-multiple-series-0",
								"namespace":  "replace-multiple-series-0",
								"node_id":    "replace-multiple-series-0",
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-0",
									},
								},
							},
						},
					},
					{
						Metric: &metricpb.Metric{
							Type: "replace-multiple-series-1",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "replace-multiple-series-1",
							Labels: map[string]string{
								"project_id": "replace-multiple-series-1",
								"location":   "replace-multiple-series-1",
								"namespace":  "replace-multiple-series-1",
								"node_id":    "replace-multiple-series-1",
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-1",
									},
								},
							},
						},
					},
				},
			},
			metric: generators.Metric{
				Value:     4.4,
				Timestamp: timestamp,
			},
			expected: &monitoringpb.CreateTimeSeriesRequest{
				Name: "replace-multiple-series",
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type: "replace-multiple-series-0",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "k8s_pod",
							Labels: map[string]string{
								"project_id":     project,
								"location":       location,
								"cluster_name":   clusterName,
								"namespace_name": namespace,
								"pod_name":       pod,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-0",
									},
								},
							},
						},
					},
					{
						Metric: &metricpb.Metric{
							Type: "replace-multiple-series-1",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "k8s_pod",
							Labels: map[string]string{
								"project_id":     project,
								"location":       location,
								"cluster_name":   clusterName,
								"namespace_name": namespace,
								"pod_name":       pod,
							},
						},
						Points: []*monitoringpb.Point{
							{
								Interval: &monitoringpb.TimeInterval{
									StartTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
									EndTime: &timestamppb.Timestamp{
										Seconds: timestamp.Unix(),
									},
								},
								Value: &monitoringpb.TypedValue{
									Value: &monitoringpb.TypedValue_StringValue{
										StringValue: "test-value-1",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	t.Parallel()
	for _, test := range tests {
		tst := test
		t.Run(tst.name, func(t *testing.T) {
			t.Parallel()
			err := transformer(tst.req, generators.Metric{})
			switch {
			case tst.expectedError == nil && err != nil:
				t.Errorf("Transformer raised an unexpected exception: %v", err)
			case tst.expectedError != nil && !errors.Is(err, tst.expectedError):
				t.Errorf("Expected transform to raise %v, got %v", tst.expectedError, err)
			case !reflect.DeepEqual(tst.expected, tst.req):
				t.Errorf("Expected %+v, got %+v", tst.expected, tst.req)
			}
		})
	}
}
