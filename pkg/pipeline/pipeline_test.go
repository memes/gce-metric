package pipeline

import (
	"context"
	"errors"
	"log"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/go-logr/stdr"
	"github.com/google/uuid"
	"github.com/memes/gce-metric/pkg/generators"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoredrespb "google.golang.org/genproto/googleapis/api/monitoredres"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	testProjectID     = "test-project"
	testInstanceID    = "test-instance"
	testZone          = "test-zone1-a"
	testNamespace     = "test-namespace"
	testClusterName   = "test-cluster"
	testPodName       = "test-pod"
	testContainerName = "test-container"
	testHost          = "test-host"
)

// Define an object to override GCP metadata client for testing.
type testClient struct {
	projectID  string
	instanceID string
	zone       string
	attributes map[string]string
}

// Implements the metadataClient interface requirement for ProjectID.
func (t *testClient) ProjectID() (string, error) {
	return t.projectID, nil
}

// Implements the metadataClient interface requirement for InstanceID.
func (t *testClient) InstanceID() (string, error) {
	return t.instanceID, nil
}

// Implements the metadataClient interface requirement for Zone.
func (t *testClient) Zone() (string, error) {
	return t.zone, nil
}

// Implements the metadataClient interface requirement for InstanceAttributeValue.
func (t *testClient) InstanceAttributeValue(name string) (string, error) {
	return t.attributes[name], nil
}

// Implement an Option that allows changing the OnGCE function used by Pipeline
// to determine if it is executing in a Google Cloud environment.
func withOnGCE(onGCE bool) Option {
	return func(p *Pipeline) error {
		p.onGCE = func() bool { return onGCE }
		return nil
	}
}

// Implement an Option that allows changing the metadata client used by Pipeline
// to query Google Cloud environment.
func withMetadataClient(client *testClient) Option {
	return func(p *Pipeline) error {
		p.metadataClient = client
		return nil
	}
}

// Helper function to create a new Pipeline object that will appear to be running
// outside of GCP.
func newNonGCPTestPipeline(t *testing.T, options ...Option) (*Pipeline, error) {
	t.Helper()
	client := &testClient{}
	return NewPipeline(context.Background(), append(options, withOnGCE(false), withMetadataClient(client))...)
}

func TestNonGCPDefault(t *testing.T) {
	t.Parallel()
	_, err := newNonGCPTestPipeline(t)
	switch {
	case err != nil && !errors.Is(err, errNotGCP):
		t.Errorf("Expected NewPipeline to raise %v, got %v", errNotGCP, err)
	case err == nil:
		t.Errorf("Expected NewPipeline to raise %v, but it didn't", errNotGCP)
	}
}

func TestNonGCPExplicitProjectID(t *testing.T) {
	t.Parallel()
	pipeline, err := newNonGCPTestPipeline(t, WithProjectID(testProjectID))
	if err != nil {
		t.Fatalf("Unexpected error returned from NewPipeline: %v", err)
	}
	defer pipeline.Close()
	metric := generators.Metric{
		Value:     1.1,
		Timestamp: time.Now(),
	}
	expectedWithoutNodeID := &monitoringpb.CreateTimeSeriesRequest{
		Name: "projects/" + testProjectID,
		TimeSeries: []*monitoringpb.TimeSeries{
			{
				Metric: &metricpb.Metric{
					Type: DefaultMetricType,
				},
				MetricKind: metricpb.MetricDescriptor_GAUGE,
				Resource: &monitoredrespb.MonitoredResource{
					Type: "generic_node",
					Labels: map[string]string{
						"project_id": testProjectID,
						"location":   DefaultLocation,
						"namespace":  DefaultNamespace,
					},
				},
				Points: []*monitoringpb.Point{
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
				},
			},
		},
	}
	req, err := pipeline.BuildRequest(metric)
	if err != nil {
		t.Fatalf("Unexpected error from BuildRequest: %v", err)
	}
	_, err = uuid.Parse(req.TimeSeries[0].Resource.Labels["node_id"])
	if err != nil {
		t.Errorf("Error parsing UUID from resource label 'node_id': %v", err)
	}
	delete(req.TimeSeries[0].Resource.Labels, "node_id")
	if !reflect.DeepEqual(req, expectedWithoutNodeID) {
		t.Errorf("Expected %+v, got %+v", expectedWithoutNodeID, req)
	}
}

// Helper function to create a new Pipeline object that will appear to be running
// in a Compute Engine VM.
func newGCETestPipeline(t *testing.T, options ...Option) (*Pipeline, error) {
	t.Helper()
	client := &testClient{
		projectID:  testProjectID,
		instanceID: testInstanceID,
		zone:       testZone,
	}
	return NewPipeline(context.Background(), append(options, withOnGCE(true), withMetadataClient(client))...)
}

func TestGCEPipelineDefault(t *testing.T) {
	t.Parallel()
	pipeline, err := newGCETestPipeline(t)
	if err != nil {
		t.Fatalf("Unexpected error returned from NewPipeline: %v", err)
	}
	pipeline.Close()
	metric := generators.Metric{
		Value:     1.1,
		Timestamp: time.Now(),
	}
	expected := &monitoringpb.CreateTimeSeriesRequest{
		Name: "projects/" + testProjectID,
		TimeSeries: []*monitoringpb.TimeSeries{
			{
				Metric: &metricpb.Metric{
					Type: DefaultMetricType,
				},
				MetricKind: metricpb.MetricDescriptor_GAUGE,
				Resource: &monitoredrespb.MonitoredResource{
					Type: "gce_instance",
					Labels: map[string]string{
						"project_id":  testProjectID,
						"instance_id": testInstanceID,
						"zone":        testZone,
					},
				},
				Points: []*monitoringpb.Point{
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
				},
			},
		},
	}
	req, err := pipeline.BuildRequest(metric)
	if err != nil {
		t.Fatalf("Unexpected error from BuildRequest: %v", err)
	}
	if !reflect.DeepEqual(req, expected) {
		t.Errorf("Expected %+v, got %+v", expected, req)
	}
}

// Helper function to create a new Pipeline object that will appear to be running
// in a GKE container.
func newGKETestPipeline(t *testing.T, options ...Option) (*Pipeline, error) {
	t.Helper()
	t.Setenv("KUBERNETES_SERVICE_HOST", testHost)
	t.Setenv("NAMESPACE", testNamespace)
	t.Setenv("HOSTNAME", testPodName)
	t.Setenv("CONTAINER_NAME", testContainerName)
	client := &testClient{
		projectID:  testProjectID,
		instanceID: testInstanceID,
		zone:       testZone,
		attributes: map[string]string{
			"cluster_name": testClusterName,
		},
	}
	return NewPipeline(context.Background(), append(options, withOnGCE(true), withMetadataClient(client))...)
}

func TestGKEPipelineDefault(t *testing.T) { //nolint: paralleltest // simulating GKE requires t.SetEnv() - incompatible with t.Parallel()
	pipeline, err := newGKETestPipeline(t)
	if err != nil {
		t.Fatalf("Unexpected error returned from NewPipeline: %v", err)
	}
	pipeline.Close()
	metric := generators.Metric{
		Value:     1.1,
		Timestamp: time.Now(),
	}
	expected := &monitoringpb.CreateTimeSeriesRequest{
		Name: "projects/" + testProjectID,
		TimeSeries: []*monitoringpb.TimeSeries{
			{
				Metric: &metricpb.Metric{
					Type: DefaultMetricType,
				},
				MetricKind: metricpb.MetricDescriptor_GAUGE,
				Resource: &monitoredrespb.MonitoredResource{
					Type: "gke_container",
					Labels: map[string]string{
						"project_id":     testProjectID,
						"cluster_name":   testClusterName,
						"namespace_id":   testNamespace,
						"instance_id":    testInstanceID,
						"pod_id":         testPodName,
						"container_name": testContainerName,
						"zone":           testZone,
					},
				},
				Points: []*monitoringpb.Point{
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
				},
			},
		},
	}
	req, err := pipeline.BuildRequest(metric)
	if err != nil {
		t.Fatalf("Unexpected error from BuildRequest: %v", err)
	}
	if !reflect.DeepEqual(req, expected) {
		t.Errorf("Expected %+v, got %+v", expected, req)
	}
}

func Example() {
	// Use Go's standard logger as the logr implementation
	logger := stdr.NewWithOptions(log.New(os.Stderr, "", log.Lshortfile), stdr.Options{LogCaller: stdr.All})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	metrics := make(chan generators.Metric, 2)
	pipeline, err := NewPipeline(ctx,
		WithLogger(logger),
		WithProjectID("my-google-project-id"),
		WithMetricType("custom.googleapis.com/my-synthetic-metric"),
		WithWriterEmitter(os.Stdout),
		WithTransformers([]Transformer{
			func(req *monitoringpb.CreateTimeSeriesRequest, _ generators.Metric) error {
				for _, series := range req.TimeSeries {
					series.Resource.Labels["node_id"] = "example"
				}
				return nil
			},
		}),
	)
	if err != nil {
		logger.Error(err, "NewPipeline returned an error")
		return
	}
	defer pipeline.Close()

	// Launch a pipeline processor that will emit each value received from
	// metrics channel.
	go func() {
		if err := pipeline.Processor()(ctx, metrics); err != nil {
			logger.Error(err, "Pipeline processor returned an error")
			cancel()
		}
	}()

	metrics <- generators.Metric{
		Value:     1.0,
		Timestamp: time.Unix(1, 0),
	}

	metrics <- generators.Metric{
		Value:     2.0,
		Timestamp: time.Unix(2, 0),
	}
	<-ctx.Done()
	// Output:
	// name: "projects/my-google-project-id"
	// time_series: {
	//   metric: {
	//     type: "custom.googleapis.com/my-synthetic-metric"
	//   }
	//   resource: {
	//     type: "generic_node"
	//     labels: {
	//       key: "location"
	//       value: "global"
	//     }
	//     labels: {
	//       key: "namespace"
	//       value: "github.com/memes/gce-metric"
	//     }
	//     labels: {
	//       key: "node_id"
	//       value: "example"
	//     }
	//     labels: {
	//       key: "project_id"
	//       value: "my-google-project-id"
	//     }
	//   }
	//   metric_kind: GAUGE
	//   points: {
	//     interval: {
	//       end_time: {
	//         seconds: 1
	//       }
	//       start_time: {
	//         seconds: 1
	//       }
	//     }
	//     value: {
	//       double_value: 1
	//     }
	//   }
	// }
	//
	// name: "projects/my-google-project-id"
	// time_series: {
	//   metric: {
	//     type: "custom.googleapis.com/my-synthetic-metric"
	//   }
	//   resource: {
	//     type: "generic_node"
	//     labels: {
	//       key: "location"
	//       value: "global"
	//     }
	//     labels: {
	//       key: "namespace"
	//       value: "github.com/memes/gce-metric"
	//     }
	//     labels: {
	//       key: "node_id"
	//       value: "example"
	//     }
	//     labels: {
	//       key: "project_id"
	//       value: "my-google-project-id"
	//     }
	//   }
	//   metric_kind: GAUGE
	//   points: {
	//     interval: {
	//       end_time: {
	//         seconds: 2
	//       }
	//       start_time: {
	//         seconds: 2
	//       }
	//     }
	//     value: {
	//       double_value: 2
	//     }
	//   }
	// }
}
