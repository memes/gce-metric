package pipeline

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"cloud.google.com/go/compute/metadata"
	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/memes/gce-metric/pkg/generators"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
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

func WithDefaultEmitter() Option {
	return func(p *Pipeline) error {
		p.emitter = func(ctx context.Context, req *monitoringpb.CreateTimeSeriesRequest) error {
			if err := p.client.CreateTimeSeries(ctx, req); err != nil {
				return fmt.Errorf("failure sending create time-series request: %w", err)
			}
			return nil
		}
		p.closer = func() error {
			if p.client == nil {
				return nil
			}
			if err := p.client.Close(); err != nil {
				return fmt.Errorf("failure closing metric client: %w", err)
			}
			return nil
		}
		return nil
	}
}

func WithWriterEmitter(writer io.Writer) Option {
	return func(p *Pipeline) error {
		p.emitter = func(ctx context.Context, req *monitoringpb.CreateTimeSeriesRequest) error {
			if _, err := fmt.Fprintf(writer, "%s\n", prototext.Format(req)); err != nil {
				return fmt.Errorf("failure writing time-series request: %w", err)
			}
			return nil
		}
		p.closer = func() error {
			return nil
		}
		return nil
	}
}

func NewPipeline(ctx context.Context, options ...Option) (*Pipeline, error) {
	pipeline := &Pipeline{
		logger:         logr.Discard(),
		metricType:     DefaultMetricType,
		transformers:   []Transformer{},
		onGCE:          metadata.OnGCE,
		metadataClient: metadata.NewClient(nil),
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
		err := WithDefaultEmitter()(pipeline)
		if err != nil {
			return nil, err
		}
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

func (p *Pipeline) defaultTransformers(_ context.Context) ([]Transformer, error) {
	transformers := []Transformer{}
	if p.onGCE() {
		instanceID, err := p.metadataClient.InstanceID()
		if err != nil {
			return nil, fmt.Errorf("failure getting instance identifier from metadata client: %w", err)
		}
		zone, err := p.metadataClient.Zone()
		if err != nil {
			return nil, fmt.Errorf("failure getting zone from metadata client: %w", err)
		}
		if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
			// Use a transformer that add a gke_container resource type to
			// the request.
			clusterName, err := p.metadataClient.InstanceAttributeValue("cluster_name")
			if err != nil {
				return nil, fmt.Errorf("failure getting 'cluster_name' attribute from metadataClient: %w", err)
			}
			transformers = append(transformers, NewGKEMonitoredResourceTransformer(p.projectID, clusterName, os.Getenv("NAMESPACE"), instanceID, os.Getenv("HOSTNAME"), os.Getenv("CONTAINER_NAME"), zone))
		} else {
			//
			transformers = append(transformers, NewGCEMonitoredResourceTransformer(p.projectID, instanceID, zone))
		}
	} else {
		// Use a transformer that adds a generic_node resource type to
		// the request.
		transformers = append(transformers, NewGenericMonitoredResourceTransformer(p.projectID, DefaultLocation, DefaultNamespace, uuid.New().String()))
	}
	transformers = append(transformers, NewDoubleTypedValueTransformer())
	return transformers, nil
}

func (p *Pipeline) Processor() Processor {
	return func(ctx context.Context, input <-chan generators.Metric) error {
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
