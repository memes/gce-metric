package pipeline

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"

	"cloud.google.com/go/compute/metadata"
	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/memes/gce-metric/pkg/generators"
	"google.golang.org/genproto/googleapis/api/label"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

var errNotGCP = fmt.Errorf("not running on Google Cloud")

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
	asInteger                  bool
	metricLabels               map[string]string
	metricDescriptors          map[string]string
	excludeDefaultTransformers bool
	transformers               []Transformer
	emitter                    Emitter
	closer                     Closer
	client                     *monitoring.MetricClient
	descriptor                 *monitoringpb.CreateMetricDescriptorRequest
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
	req := &monitoringpb.CreateTimeSeriesRequest{}
	var err error
	for _, transformer := range p.transformers {
		req, err = transformer(req, metric)
		if err != nil {
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

func AsInteger() Option {
	return func(p *Pipeline) error {
		p.asInteger = true
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
		var createDefinition sync.Once
		p.emitter = func(ctx context.Context, req *monitoringpb.CreateTimeSeriesRequest) error {
			var err error
			createDefinition.Do(func() {
				if p.descriptor != nil {
					_, err = p.client.CreateMetricDescriptor(ctx, p.descriptor)
				}
			})
			if err != nil {
				return fmt.Errorf("failure creating metric descriptor: %w", err)
			}
			if err = p.client.CreateTimeSeries(ctx, req); err != nil {
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
		var createDefinition sync.Once
		p.emitter = func(ctx context.Context, req *monitoringpb.CreateTimeSeriesRequest) error {
			var err error
			createDefinition.Do(func() {
				if p.descriptor != nil {
					_, err = fmt.Fprintf(writer, "%+v\n", p.descriptor)
				}
			})
			if err != nil {
				return fmt.Errorf("failure writing metric descriptor: %w", err)
			}
			if _, err = fmt.Fprintf(writer, "%+v\n", req); err != nil {
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

func withOnGCE(onGCE func() bool) Option {
	return func(p *Pipeline) error {
		p.onGCE = onGCE
		return nil
	}
}

func withMetadataClient(metadataClient *metadata.Client) Option {
	return func(p *Pipeline) error {
		p.metadataClient = metadataClient
		return nil
	}
}

func NewPipeline(ctx context.Context, options ...Option) (*Pipeline, error) {
	pipeline := &Pipeline{
		logger:         logr.Discard(),
		metricType:     "custom.googleapis.com/gce_metric",
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
		baseTransformers, err := pipeline.baseTransformers(ctx)
		if err != nil {
			return nil, err
		}
		pipeline.transformers = append(baseTransformers, pipeline.transformers...)
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

func (p *Pipeline) baseTransformers(_ context.Context) ([]Transformer, error) {
	transformers := []Transformer{
		NewCreateTimeSeriesRequestTransformer(p.projectID, p.metricType, p.metricLabels),
	}
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
		transformers = append(transformers, NewGenericMonitoredResourceTransformer(p.projectID, "global", "github.com/memes/gce-metric", uuid.New().String()))
	}
	if p.asInteger {
		transformers = append(transformers, IntegerTypeValueTransformer)
	} else {
		transformers = append(transformers, DoubleTypedValueTransformer)
	}
	return transformers, nil
}

func NewCreateMetricDescriptorRequest(projectID, metricType, name, displayName string, metricLabels map[string]string, valueType metricpb.MetricDescriptor_ValueType) monitoringpb.CreateMetricDescriptorRequest {
	labels := []*label.LabelDescriptor{}
	for key, value := range metricLabels {
		labels = append(labels, &label.LabelDescriptor{
			Key:         key,
			Description: value,
			ValueType:   label.LabelDescriptor_STRING,
		})
	}
	return monitoringpb.CreateMetricDescriptorRequest{
		Name: "projects/" + projectID,
		MetricDescriptor: &metricpb.MetricDescriptor{
			Name:        name,
			Type:        metricType,
			Labels:      labels,
			MetricKind:  metricpb.MetricDescriptor_GAUGE,
			ValueType:   valueType,
			DisplayName: displayName,
		},
	}
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
