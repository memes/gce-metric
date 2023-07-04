package generators_test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/memes/gce-metric/pkg/generators"
)

// Helper function to listen to input and increment counter when a Metric is
// received.
func metricCounter(ctx context.Context, counter *int, reader <-chan generators.Metric) {
	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-reader:
			if !ok {
				return
			}
			*counter++
		}
	}
}

// Verify that the periodic generator function will exit when context is cancelled,
// without emitting further values.
func TestPeriodicGeneratorCancel(t *testing.T) {
	t.Parallel()
	valueCount := 0
	periodicGenerator, reader, err := generators.NewPeriodicGenerator(
		generators.WithLogger(logr.Discard()),
		generators.WithValueCalculator(generators.Sawtooth.ValueCalculator()),
		generators.WithPeriod(1*time.Minute),
	)
	if err != nil {
		t.Errorf("NewPeriodicGenerator raised an error: %v", err)
	}
	ticker := time.NewTimer(100 * time.Millisecond)
	defer ticker.Stop()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go metricCounter(ctx, &valueCount, reader)
	go periodicGenerator(ctx, ticker.C)
	time.Sleep(1 * time.Second)
	cancel()
	_, ok := <-reader
	if ok {
		t.Errorf("Expected reader channel to be closed")
	}
	if valueCount < 1 {
		t.Error("Expected valueCount to be >0")
	}
	oldValueCount := valueCount
	time.Sleep(1 * time.Second)
	if oldValueCount != valueCount {
		t.Errorf("Expected valueCount to be %d, got %d", oldValueCount, valueCount)
	}
}

// Verify that the periodic generator function will exit when context reaches a
// deadline, without emitting further values.
func TestPeriodicGeneratorTimeout(t *testing.T) {
	t.Parallel()
	valueCount := 0
	periodicGenerator, reader, err := generators.NewPeriodicGenerator(
		generators.WithLogger(logr.Discard()),
		generators.WithValueCalculator(generators.Sawtooth.ValueCalculator()),
		generators.WithPeriod(1*time.Minute),
	)
	if err != nil {
		t.Errorf("NewPeriodicGenerator raised an error: %v", err)
	}
	ticker := time.NewTimer(100 * time.Millisecond)
	defer ticker.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	go metricCounter(ctx, &valueCount, reader)
	go periodicGenerator(ctx, ticker.C)
	time.Sleep(1 * time.Second)
	_, ok := <-reader
	if ok {
		t.Errorf("Expected reader channel to be closed")
	}
	if valueCount < 1 {
		t.Error("Expected valueCount to be >0")
	}
	oldValueCount := valueCount
	time.Sleep(1 * time.Second)
	if oldValueCount != valueCount {
		t.Errorf("Expected valueCount to be %d, got %d", oldValueCount, valueCount)
	}
}

func Example() { //nolint:testableexamples // The output would include a timestamp
	// Create the timestamped value generator
	periodicGenerator, reader, err := generators.NewPeriodicGenerator(
		generators.WithValueCalculator(generators.NewPeriodicRangeCalculator(0.0, 100.0, generators.Square)),
		generators.WithPeriod(10*time.Second),
	)
	if err != nil {
		log.Printf("Error creating NewPeriodicGenerator: %v", err)
		return
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	go periodicGenerator(ctx, ticker.C)
loop:
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			break loop

		case metric := <-reader:
			fmt.Printf("Metric with value %f for timestamp %s\n", metric.Value, metric.Timestamp.Format("15:04:05"))
		}
	}
}
