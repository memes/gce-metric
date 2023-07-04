//nolint:dupl // The PeriodicGenerator tests are deliberately similar but distinct
package generators_test

import (
	"errors"
	"math"
	"testing"

	"github.com/memes/gce-metric/pkg/generators"
)

const (
	generatorTolerance = 1e-6
)

func TestPeriodicTypeString(t *testing.T) {
	tests := []struct {
		name         string
		periodicType generators.PeriodicType
		expected     string
	}{
		{
			name:         "unknown",
			periodicType: 0,
			expected:     "unknown",
		},
		{
			name:         "sawtooth",
			periodicType: generators.Sawtooth,
			expected:     "sawtooth",
		},
		{
			name:         "sine",
			periodicType: generators.Sine,
			expected:     "sine",
		},
		{
			name:         "square",
			periodicType: generators.Square,
			expected:     "square",
		},
		{
			name:         "triangle",
			periodicType: generators.Triangle,
			expected:     "triangle",
		},
	}
	t.Parallel()
	for _, test := range tests {
		tst := test
		t.Run(tst.name, func(t *testing.T) {
			t.Parallel()
			result := tst.periodicType.String()
			if tst.expected != result {
				t.Errorf("Expected %v, got %v", tst.expected, result)
			}
		})
	}
}

func TestParsePeriodicType(t *testing.T) {
	tests := []struct {
		name          string
		value         string
		expected      generators.PeriodicType
		expectedError error
	}{
		{
			name:          "empty",
			expected:      generators.Invalid,
			expectedError: generators.ErrInvalidPeriodicType,
		},
		{
			name:          "invalid",
			value:         "invalid",
			expected:      generators.Invalid,
			expectedError: generators.ErrInvalidPeriodicType,
		},
		{
			name:     "sawtooth",
			value:    "sawtooth",
			expected: generators.Sawtooth,
		},
		{
			name:     "sine",
			value:    "sine",
			expected: generators.Sine,
		},
		{
			name:     "square",
			value:    "square",
			expected: generators.Square,
		},
		{
			name:     "triangle",
			value:    "triangle",
			expected: generators.Triangle,
		},
	}
	t.Parallel()
	for _, test := range tests {
		tst := test
		t.Run(tst.name, func(t *testing.T) {
			t.Parallel()
			result, err := generators.ParsePeriodicType(tst.value)
			switch {
			case tst.expectedError == nil && err != nil:
				t.Errorf("Received an unexpected error: %v", err)
			case tst.expectedError != nil && !errors.Is(err, tst.expectedError):
				t.Errorf("Expected  error %v, got %v", tst.expectedError, err)
			case tst.expected != result:
				t.Errorf("Expected %v, got %v", tst.expected, result)
			}
		})
	}
}

// Executes the supplied valueCalculator function with argument phase and verifies
// that the return value matches expectation within tolerance.
func testValueCalculator(t *testing.T, phase, expected float64, valueCalculator generators.ValueCalculator) {
	t.Helper()
	result := valueCalculator(phase)
	if ok := math.Abs(expected-result) < generatorTolerance; !ok {
		t.Errorf("Expected %f, got %f", expected, result)
	}
}

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
	calculator := generators.Sawtooth.ValueCalculator()
	for _, test := range tests {
		tst := test
		t.Run(tst.name, func(t *testing.T) {
			t.Parallel()
			testValueCalculator(t, tst.phase, tst.expected, calculator)
		})
	}
}

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
	calculator := generators.Sine.ValueCalculator()
	for _, test := range tests {
		tst := test
		t.Run(tst.name, func(t *testing.T) {
			t.Parallel()
			testValueCalculator(t, tst.phase, tst.expected, calculator)
		})
	}
}

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
	calculator := generators.Square.ValueCalculator()
	for _, test := range tests {
		tst := test
		t.Run(tst.name, func(t *testing.T) {
			t.Parallel()
			testValueCalculator(t, tst.phase, tst.expected, calculator)
		})
	}
}

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
	calculator := generators.Triangle.ValueCalculator()
	for _, test := range tests {
		tst := test
		t.Run(tst.name, func(t *testing.T) {
			t.Parallel()
			testValueCalculator(t, tst.phase, tst.expected, calculator)
		})
	}
}

//nolint:funlen // The tests table makes the function longer seem longer to linter
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
	calculator := generators.NewPeriodicRangeCalculator(low, high, generators.Sawtooth)
	for _, test := range tests {
		tst := test
		t.Run(tst.name, func(t *testing.T) {
			t.Parallel()
			testValueCalculator(t, tst.phase, tst.expected, calculator)
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
	calculator := generators.PeriodicType(-1).ValueCalculator()
	for _, test := range tests {
		tst := test
		t.Run(tst.name, func(t *testing.T) {
			t.Parallel()
			testValueCalculator(t, tst.phase, tst.expected, calculator)
		})
	}
}
