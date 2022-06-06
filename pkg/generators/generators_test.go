package generators_test

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/memes/gce-metric/pkg/generators"
)

const (
	generatorTolerance = 1e-6
)

// Executes the supplied valueCalculator function with argument phase and verifies
// that the return value matches expectation within tolerance.
func testValueCalculator(t *testing.T, phase, expected float64, valueCalculator generators.ValueCalculator) {
	t.Helper()
	result := valueCalculator(phase)
	if ok := math.Abs(expected-result) < generatorTolerance; !ok {
		t.Errorf("Expected %f, got %f", expected, result)
	}
}

//nolint: dupl // The periodic generator tests are deliberately similar but distinct
func TestSawtoothPeriodicGenerator(t *testing.T) {
	tests := []struct {
		name     string
		phase    float64
		expected float64
	}{
		{
			name:     "0",
			phase:    0.0,
			expected: 0.0,
		},
		{
			name:     "ϕ/4",
			phase:    0.25,
			expected: 0.25,
		},
		{
			name:     "ϕ/2",
			phase:    0.5,
			expected: 0.5,
		},
		{
			name:     "3ϕ/4",
			phase:    0.75,
			expected: 0.75,
		},
		{
			name:     "ϕ",
			phase:    1.0,
			expected: 0.0,
		},
		{
			name:     "5ϕ/4",
			phase:    1.25,
			expected: 0.25,
		},
		{
			name:     "3ϕ/2",
			phase:    1.5,
			expected: 0.5,
		},
		{
			name:     "7ϕ/4",
			phase:    1.75,
			expected: 0.75,
		},
		{
			name:     "2ϕ",
			phase:    2.0,
			expected: 0.0,
		},
	}
	t.Parallel()
	calculator := generators.Sawtooth.Calculator()
	for _, test := range tests {
		tst := test
		t.Run(tst.name, func(t *testing.T) {
			t.Parallel()
			testValueCalculator(t, tst.phase, tst.expected, calculator)
		})
	}
}

//nolint: dupl // The periodic generator tests are deliberately similar but distinct
func TestSinePeriodicGenerator(t *testing.T) {
	tests := []struct {
		name     string
		phase    float64
		expected float64
	}{
		{
			name:     "0",
			phase:    0.0,
			expected: 0.0,
		},
		{
			name:     "ϕ/4",
			phase:    0.25,
			expected: 0.5,
		},
		{
			name:     "ϕ/2",
			phase:    0.5,
			expected: 1.0,
		},
		{
			name:     "3ϕ/4",
			phase:    0.75,
			expected: 0.5,
		},
		{
			name:     "ϕ",
			phase:    1.0,
			expected: 0.0,
		},
		{
			name:     "5ϕ/4",
			phase:    1.25,
			expected: 0.5,
		},
		{
			name:     "3ϕ/2",
			phase:    1.5,
			expected: 1.0,
		},
		{
			name:     "7ϕ/4",
			phase:    1.75,
			expected: 0.5,
		},
		{
			name:     "2ϕ",
			phase:    2.0,
			expected: 0.0,
		},
	}
	t.Parallel()
	calculator := generators.Sine.Calculator()
	for _, test := range tests {
		tst := test
		t.Run(tst.name, func(t *testing.T) {
			t.Parallel()
			testValueCalculator(t, tst.phase, tst.expected, calculator)
		})
	}
}

//nolint: dupl // The periodic generator tests are deliberately similar but distinct
func TestSquarePeriodicGenerator(t *testing.T) {
	tests := []struct {
		name     string
		phase    float64
		expected float64
	}{
		{
			name:     "0",
			phase:    0.0,
			expected: 0.0,
		},
		{
			name:     "ϕ/4",
			phase:    0.25,
			expected: 0.0,
		},
		{
			name:     "ϕ/2",
			phase:    0.5,
			expected: 0.0,
		},
		{
			name:     "3ϕ/4",
			phase:    0.75,
			expected: 1.0,
		},
		{
			name:     "ϕ",
			phase:    1.0,
			expected: 1.0,
		},
		{
			name:     "5ϕ/4",
			phase:    1.25,
			expected: 0.0,
		},
		{
			name:     "3ϕ/2",
			phase:    1.5,
			expected: 0.0,
		},
		{
			name:     "7ϕ/4",
			phase:    1.75,
			expected: 1.0,
		},
		{
			name:     "2ϕ",
			phase:    2.0,
			expected: 1.0,
		},
	}
	t.Parallel()
	calculator := generators.Square.Calculator()
	for _, test := range tests {
		tst := test
		t.Run(tst.name, func(t *testing.T) {
			t.Parallel()
			testValueCalculator(t, tst.phase, tst.expected, calculator)
		})
	}
}

//nolint: dupl // The periodic generator tests are deliberately similar but distinct
func TestTrianglePeriodicGenerator(t *testing.T) {
	tests := []struct {
		name     string
		phase    float64
		expected float64
	}{
		{
			name:     "0",
			phase:    0.0,
			expected: 0.0,
		},
		{
			name:     "ϕ/4",
			phase:    0.25,
			expected: 0.5,
		},
		{
			name:     "ϕ/2",
			phase:    0.5,
			expected: 1.0,
		},
		{
			name:     "3ϕ/4",
			phase:    0.75,
			expected: 0.5,
		},
		{
			name:     "ϕ",
			phase:    1.0,
			expected: 0.0,
		},
		{
			name:     "5ϕ/4",
			phase:    1.25,
			expected: 0.5,
		},
		{
			name:     "3ϕ/2",
			phase:    1.5,
			expected: 1.0,
		},
		{
			name:     "7ϕ/4",
			phase:    1.75,
			expected: 0.5,
		},
		{
			name:     "2ϕ",
			phase:    2.0,
			expected: 0.0,
		},
	}
	t.Parallel()
	calculator := generators.Triangle.Calculator()
	for _, test := range tests {
		tst := test
		t.Run(tst.name, func(t *testing.T) {
			t.Parallel()
			testValueCalculator(t, tst.phase, tst.expected, calculator)
		})
	}
}

func TestPeriodicRangeGenerator(t *testing.T) {
	low := 10.0
	high := 20.0
	tests := []struct {
		name     string
		phase    float64
		expected float64
	}{
		{
			name:     "0",
			phase:    0.0,
			expected: low,
		},
		{
			name:     "ϕ/4",
			phase:    0.25,
			expected: low + ((high - low) / 4.0),
		},
		{
			name:     "ϕ/2",
			phase:    0.5,
			expected: low + ((high - low) / 2.0),
		},
		{
			name:     "3ϕ/4",
			phase:    0.75,
			expected: low + (3.0 * (high - low) / 4.0),
		},
		{
			name:     "ϕ",
			phase:    1.0,
			expected: low,
		},
		{
			name:     "5ϕ/4",
			phase:    1.25,
			expected: low + ((high - low) / 4.0),
		},
		{
			name:     "3ϕ/2",
			phase:    1.5,
			expected: low + ((high - low) / 2.0),
		},
		{
			name:     "7ϕ/4",
			phase:    1.75,
			expected: low + (3.0 * (high - low) / 4.0),
		},
		{
			name:     "2ϕ",
			phase:    2.0,
			expected: low,
		},
	}
	t.Parallel()
	periodRangeFunction := generators.NewPeriodicRangeCalculator(low, high, generators.Sawtooth)
	for _, test := range tests {
		tst := test
		t.Run(tst.name, func(t *testing.T) {
			t.Parallel()
			testValueCalculator(t, tst.phase, tst.expected, periodRangeFunction)
		})
	}
}

func TestInvalidPeriodicGenerator(t *testing.T) {
	tests := []struct {
		name     string
		phase    float64
		expected float64
	}{
		{
			name:     "0",
			phase:    0.0,
			expected: 0.0,
		},
		{
			name:     "ϕ/4",
			phase:    0.25,
			expected: 0.0,
		},
		{
			name:     "ϕ/2",
			phase:    0.5,
			expected: 0.0,
		},
		{
			name:     "3ϕ/4",
			phase:    0.75,
			expected: 0.0,
		},
		{
			name:     "ϕ",
			phase:    1.0,
			expected: 0.0,
		},
		{
			name:     "5ϕ/4",
			phase:    1.25,
			expected: 0.0,
		},
		{
			name:     "3ϕ/2",
			phase:    1.5,
			expected: 0.0,
		},
		{
			name:     "7ϕ/4",
			phase:    1.75,
			expected: 0.0,
		},
		{
			name:     "2ϕ",
			phase:    2.0,
			expected: 0.0,
		},
	}
	t.Parallel()
	calculator := generators.PeriodicType(-1).Calculator()
	for _, test := range tests {
		tst := test
		t.Run(tst.name, func(t *testing.T) {
			t.Parallel()
			testValueCalculator(t, tst.phase, tst.expected, calculator)
		})
	}
}

// Helper function to listen to input and increment counter when a Metric is
// received.
func metricCounter(counter *int, input <-chan generators.Metric) {
	for {
		_, ok := <-input
		if !ok {
			return
		}
		*counter++
	}
}

// Verify that the periodic generator function will exit when context is cancelled,
// without emitting further values.
func TestPeriodicGeneratorCancel(t *testing.T) {
	t.Parallel()
	valueCount := 0
	generator, output := generators.NewPeriodicGenerator(logr.Discard(), generators.Sawtooth.Calculator(), 10*time.Second)
	ticker := time.NewTimer(100 * time.Millisecond)
	defer ticker.Stop()
	go metricCounter(&valueCount, output)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		if err := generator(ctx, ticker.C); err != nil {
			t.Errorf("Unexpected error in periodic generator: %v", err)
		}
	}()
	time.Sleep(1 * time.Second)
	cancel()
	_, ok := <-output
	if ok {
		t.Errorf("Expected output channel to be closed")
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
	generator, output := generators.NewPeriodicGenerator(logr.Discard(), generators.Sawtooth.Calculator(), 10*time.Second)
	ticker := time.NewTimer(100 * time.Millisecond)
	defer ticker.Stop()
	go metricCounter(&valueCount, output)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	go func() {
		if err := generator(ctx, ticker.C); err != nil {
			t.Errorf("Unexpected error in periodic generator: %v", err)
		}
	}()
	time.Sleep(1 * time.Second)
	_, ok := <-output
	if ok {
		t.Errorf("Expected output channel to be closed")
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
