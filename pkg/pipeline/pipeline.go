package pipeline

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"cloud.google.com/go/compute/metadata"
	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/memes/gce-metric/pkg/generators"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	"google.golang.org/protobuf/encoding/prototext"
)

const (
	DefaultMetricType = "custom.googleapis.com/gce_metric"
	DefaultLocation   = "global"
	DefaultNamespace  = "github.com/memes/gce-metric"
)

// This error will be returned if a pipeline function requires a Google Cloud
// execution environment.
var errNotGCP = errors.New("not running on Google Cloud")

type metadataClient interface {
	ProjectID() (string, error)
	InstanceID() (string, error)
	Zone() (string, error)
	InstanceAttributeValue(string) (string, error)
}

type Emitter func(context.Context, *monitoringpb.CreateTimeSeriesRequest) error

type Closer func() error

type Processor func(context.Context, <-chan generators.Metric) error

type Option func(*Pipeline) error

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

func (p *Pipeline) Close() error {
	if p.closer == nil {
		return nil
	}
	return p.closer()
}

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

func WithLogger(logger logr.Logger) Option {
	return func(p *Pipeline) error {
		p.logger = logger
		return nil
	}
}

// Use a specific project identifier for the synthetic metrics in preference to
// detecting from metadata.
func WithProjectID(projectID string) Option {
	return func(p *Pipeline) error {
		p.projectID = projectID
		return nil
	}
}

func WithMetricType(metricType string) Option {
	return func(p *Pipeline) error {
		p.metricType = metricType
		return nil
	}
}

func WithoutDefaultTransformers() Option {
	return func(p *Pipeline) error {
		p.excludeDefaultTransformers = true
		return nil
	}
}

func WithTransformers(transformers []Transformer) Option {
	return func(p *Pipeline) error {
		p.transformers = append(p.transformers, transformers...)
		return nil
	}
}

func WithWriterEmitter(writer io.Writer) Option {
	return func(p *Pipeline) error {
		p.emitter = func(_ context.Context, req *monitoringpb.CreateTimeSeriesRequest) error {
			p.logger.V(2).Info("Emitting time-series request to writer")
			if _, err := fmt.Fprintf(writer, "%s\n", prototext.Format(req)); err != nil {
				return fmt.Errorf("failure writing time-series request: %w", err)
			}
			return nil
		}
		p.closer = func() error {
			p.logger.V(2).Info("Closing time-series writer emitter")
			return nil
		}
		return nil
	}
}

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
		onGCE:                      metadata.OnGCE,
		metadataClient:             metadata.NewClient(nil),
	}
	for _, option := range options {
		if err := option(pipeline); err != nil {
			return nil, err
		}
	}
	if pipeline.projectID == "" {
		if !pipeline.onGCE() {
			return nil, errNotGCP
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
		pipeline.emitter = pipeline.defaultEmitter
	}
	if pipeline.closer == nil {
		pipeline.closer = pipeline.defaultCloser
	}
	if pipeline.client == nil {
		client, err := monitoring.NewMetricClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("failure creating new metric client: %w", err)
		}
		pipeline.client = client
	}
	return pipeline, nil
}

func (p *Pipeline) defaultEmitter(ctx context.Context, req *monitoringpb.CreateTimeSeriesRequest) error {
	p.logger.V(2).Info("Emitting time-series request to GCP")
	if err := p.client.CreateTimeSeries(ctx, req); err != nil {
		return fmt.Errorf("failure sending create time-series request: %w", err)
	}
	return nil
}

func (p *Pipeline) defaultCloser() error {
	p.logger.V(2).Info("Closing time-series emitter")
	if p.client == nil {
		return nil
	}
	if err := p.client.Close(); err != nil {
		return fmt.Errorf("failure closing metric client: %w", err)
	}
	return nil
}

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
