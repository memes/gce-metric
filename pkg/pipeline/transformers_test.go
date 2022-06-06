package pipeline_test

import (
	"time"

	"github.com/memes/gce-metric/pkg/generators"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoredrespb "google.golang.org/genproto/googleapis/api/monitoredres"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

func newMetric(value float64) generators.Metric {
	return generators.Metric{
		Value:     value,
		Timestamp: time.Now(),
	}
}

func noPointEntries() monitoringpb.CreateTimeSeriesRequest {
	return monitoringpb.CreateTimeSeriesRequest{
		Name: "projects/pkgtest",
		TimeSeries: []*monitoringpb.TimeSeries{
			{
				Metric: &metricpb.Metric{
					Type: "custom.googleapis.com",
					Labels: map[string]string{
						"blue": "green",
					},
				},
				Resource: &monitoredrespb.MonitoredResource{
					Type: "generic_node",
					Labels: map[string]string{
						"location":  "global",
						"namespace": "gce-metric",
						"node_id":   "1234567890",
					},
				},
				Points: []*monitoringpb.Point{},
			},
		},
	}
}
