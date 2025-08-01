// Package pipeline creates a set of Transformers that process a Metric to make it suitable to send to GCP Monitoring.
package pipeline

import (
	"context"
	"errors"
	"fmt"
	"os"

	"cloud.google.com/go/compute/metadata"
	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/memes/gce-metric/pkg/generators"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
)

const (
	// DefaultMetricType defines the default name to give to the generated synthetic metrics timeseries.
	DefaultMetricType = "custom.googleapis.com/gce_metric"
	// DefaultLocation defines the default location added to generated synthetic metrics.
	DefaultLocation = "global"
	// DefaultNamespace defines the default namespace to give to the generated synthetic metrics timeseries.
	DefaultNamespace = "github.com/memes/gce-metric"
)

// ErrNotGCP is returned if a pipeline function requires a Google Cloud execution environment.
var ErrNotGCP = errors.New("not running on Google Cloud")

// metadataClient defines an interface that is a subset of GCP Compute Engine metadata client. This allows test cases to
// provide appropriate values that would be detected or inferred from a real GCP compute environment.
type metadataClient interface {
	ProjectID() (string, error)
	InstanceID() (string, error)
	Zone() (string, error)
	InstanceAttributeValue(string) (string, error)
}

// Emitter defines a function that will handle a generated Metrics request.
type Emitter func(context.Context, *monitoringpb.CreateTimeSeriesRequest) error

// Closer defines a function that can close and cleanup an Emitter.
type Closer func() error

// Processor defines a function that can receive a Metric from a generator and do something useful with it.
type Processor func(context.Context, <-chan generators.Metric) error

// Option defines a function type that can be used to configure a pipeline.
type Option func(*Pipeline) error

// Pipeline defines a collection of Transformers and an Emitter that can process Metric values received on a channel and
// send them to Google Cloud Metrics as a time-series.
type Pipeline struct {
	logger                     logr.Logger
	projectID                  string
	metricType                 string
	metricLabels               map[string]string
	excludeDefaultTransformers bool
	transformers               []Transformer
	emitter                    Emitter
	closer                     Closer
	client                     *monitoring.MetricClient
	// Allow unit tests to emulate a GCP environment
	onGCE          func() bool
	metadataClient metadataClient
}

// Close will execute the Closer function associated with the Pipeline, if it is defined.
func (p *Pipeline) Close() error {
	if p.closer == nil {
		return nil
	}
	return p.closer()
}

// BuildRequest creates a GCP Metrics time-series request that can be sent to GCP.
func (p *Pipeline) BuildRequest(metric generators.Metric) (*monitoringpb.CreateTimeSeriesRequest, error) {
	p.logger.V(2).Info("Building request", "metric", metric)
	req := &monitoringpb.CreateTimeSeriesRequest{
		Name: "projects/" + p.projectID,
		TimeSeries: []*monitoringpb.TimeSeries{
			{
				Metric: &metricpb.Metric{
					Type:   p.metricType,
					Labels: p.metricLabels,
				},
				MetricKind: metricpb.MetricDescriptor_GAUGE,
			},
		},
	}
	for _, transformer := range p.transformers {
		if err := transformer(req, metric); err != nil {
			return req, err
		}
	}
	return req, nil
}

// WithLogger will use the provided Logger in the pipeline.
func WithLogger(logger logr.Logger) Option {
	return func(p *Pipeline) error {
		p.logger = logger
		return nil
	}
}

// WithProjectID will force the pipeline to use the specific project identifier for the synthetic metrics in preference
// to an identifier detected from compute metadata.
func WithProjectID(projectID string) Option {
	return func(p *Pipeline) error {
		p.projectID = projectID
		return nil
	}
}

// WithMetricType will force the pipeline to use the specific metric name instead of package default.
func WithMetricType(metricType string) Option {
	return func(p *Pipeline) error {
		p.metricType = metricType
		return nil
	}
}

// WithoutDefaultTransformers will instruct the pipeline builder to exclude the default transformers.
func WithoutDefaultTransformers() Option {
	return func(p *Pipeline) error {
		p.excludeDefaultTransformers = true
		return nil
	}
}

// WithTransformers will instruct the pipeline builder to include the provided transformers in the pipeline, appending
// to any existing transformers in the pipeline.
func WithTransformers(transformers []Transformer) Option {
	return func(p *Pipeline) error {
		p.transformers = append(p.transformers, transformers...)
		return nil
	}
}

// WithEmitterAndCloser will instruct the pipeline builder to call the supplied Emitter and Closer functions instead of
// the default GCP Metrics client's writer.
func WithEmitterAndCloser(emitter Emitter, closer Closer) Option {
	return func(p *Pipeline) error {
		p.emitter = emitter
		p.closer = closer
		return nil
	}
}

// WithOnGCE will instruct the pipeline builder to use the provided function to determine if the pipeline is running in
// a Google Cloud environment instead of default Google Cloud library function.
func WithOnGCE(onGCE func() bool) Option {
	return func(p *Pipeline) error {
		p.onGCE = onGCE
		return nil
	}
}

// WithMetadataClient will instruct the pipeline builder to use the provided metadata client to determine project, zone,
// or other details of a Google Cloud environment instead of default Google Cloud runtime function.
func WithMetadataClient(client metadataClient) Option {
	return func(p *Pipeline) error {
		p.metadataClient = client
		return nil
	}
}

// NewPipeline will build a pipeline of transformers that emits value(s) to the configured target.
func NewPipeline(ctx context.Context, options ...Option) (*Pipeline, error) {
	pipeline := &Pipeline{
		logger:                     logr.Discard(),
		projectID:                  "",
		metricType:                 DefaultMetricType,
		metricLabels:               nil,
		excludeDefaultTransformers: false,
		transformers:               []Transformer{},
		emitter:                    nil,
		closer:                     nil,
		client:                     nil,
		onGCE:                      nil,
		metadataClient:             nil,
	}
	for _, option := range options {
		if err := option(pipeline); err != nil {
			return nil, err
		}
	}
	if pipeline.onGCE == nil {
		pipeline.onGCE = metadata.OnGCE
	}
	if pipeline.metadataClient == nil {
		pipeline.metadataClient = metadata.NewClient(nil)
	}
	if pipeline.projectID == "" {
		if !pipeline.onGCE() {
			return nil, ErrNotGCP
		}
		projectID, err := pipeline.metadataClient.ProjectID()
		if err != nil {
			return nil, fmt.Errorf("failure getting project identifier from metadataClient: %w", err)
		}
		pipeline.projectID = projectID
	}
	if !pipeline.excludeDefaultTransformers {
		defaultTransformers, err := pipeline.defaultTransformers(ctx)
		if err != nil {
			return nil, err
		}
		pipeline.transformers = append(defaultTransformers, pipeline.transformers...)
	}
	if pipeline.emitter == nil {
		if err := pipeline.defaultEmitter(ctx); err != nil {
			return nil, err
		}
	}
	return pipeline, nil
}

// Defines the default set of Transformer instances to use in a pipeline based on detected environment.
func (p *Pipeline) defaultTransformers(_ context.Context) ([]Transformer, error) {
	p.logger.V(1).Info("Collecting default transformers")
	transformers := []Transformer{}
	if p.onGCE() { //nolint:nestif // Determining the correct Google Cloud environment is a set of cascading tests
		p.logger.V(2).Info("Detected we're running on GCE")
		instanceID, err := p.metadataClient.InstanceID()
		if err != nil {
			return nil, fmt.Errorf("failure getting instance identifier from metadata client: %w", err)
		}
		zone, err := p.metadataClient.Zone()
		if err != nil {
			return nil, fmt.Errorf("failure getting zone from metadata client: %w", err)
		}
		p.logger.V(2).Info("Retrieved GCE metadata", "instanceID", instanceID, "zone", zone)
		if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
			// Use a transformer that add a gke_container resource type to
			// the request.
			p.logger.V(2).Info("Looks like GKE", "instanceID", instanceID, "zone", zone)
			clusterName, err := p.metadataClient.InstanceAttributeValue("cluster_name")
			if err != nil {
				return nil, fmt.Errorf("failure getting 'cluster_name' attribute from metadataClient: %w", err)
			}
			p.logger.V(2).Info("Adding GKE transformer to pipeline", "instanceID", instanceID, "zone", zone, "clusterName", clusterName)
			transformers = append(transformers, NewGKEMonitoredResourceTransformer(p.projectID, clusterName, os.Getenv("NAMESPACE"), instanceID, os.Getenv("HOSTNAME"), os.Getenv("CONTAINER_NAME"), zone))
		} else {
			p.logger.V(2).Info("Adding GCE transformer to pipeline", "instanceID", instanceID, "zone", zone)
			// Use a GCE transformer
			transformers = append(transformers, NewGCEMonitoredResourceTransformer(p.projectID, instanceID, zone))
		}
	} else {
		p.logger.V(2).Info("GCE not detected, adding generic_node transformer to pipeline")
		// Use a transformer that adds a generic_node resource type to
		// the request.
		transformers = append(transformers, NewGenericMonitoredResourceTransformer(p.projectID, DefaultLocation, DefaultNamespace, uuid.New().String()))
	}
	transformers = append(transformers, NewDoubleTypedValueTransformer())
	return transformers, nil
}

// Set the pipeline Emitter and Closer to use GCP Monitoring client, replacing any existing functions that may be set on
// the instance.
func (p *Pipeline) defaultEmitter(ctx context.Context) error {
	client, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return fmt.Errorf("failure creating new metric client: %w", err)
	}
	p.emitter = func(ctx context.Context, req *monitoringpb.CreateTimeSeriesRequest) error {
		p.logger.V(2).Info("Emitting time-series request to GCP")
		if err := client.CreateTimeSeries(ctx, req); err != nil {
			return fmt.Errorf("failure sending create time-series request: %w", err)
		}
		return nil
	}
	p.closer = func() error {
		p.logger.V(2).Info("Closing time-series emitter")
		if client == nil {
			return nil
		}
		if err := client.Close(); err != nil {
			return fmt.Errorf("failure closing metric client: %w", err)
		}
		return nil
	}
	return nil
}

// Processor returns the Processor implementation associated with the pipeline.
func (p *Pipeline) Processor() Processor {
	return func(ctx context.Context, input <-chan generators.Metric) error {
		p.logger.V(2).Info("Launching pipeline processor")
		for {
			select {
			case <-ctx.Done():
				p.logger.V(2).Info("Context has been cancelled; exiting")
				return nil
			case value, ok := <-input:
				if !ok {
					p.logger.V(2).Info("Input channel is closed; exiting")
					return nil
				}
				req, err := p.BuildRequest(value)
				if err != nil {
					return err
				}
				if err := p.emitter(ctx, req); err != nil {
					return err
				}
			}
		}
	}
}
