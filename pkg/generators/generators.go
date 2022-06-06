// Package generators
package generators

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/go-logr/logr"
)

type (
	Metric struct {
		Value     float64
		Timestamp time.Time
	}
	// Defines a function that will block until the context is cancelled,
	// emitting a Metric value to the output channel on each tick received
	// by the ticker channel.
	PeriodicGenerator func(ctx context.Context, ticker <-chan time.Time) error
	// Defines a function that will return a float64 value for the given
	// phase value.
	ValueCalculator func(phase float64) float64
	// Defines the periodic function generators known to the package.
	PeriodicType int
)

const (
	Invalid PeriodicType = iota
	// Represents a periodic function that generates a sawtooth wave, rising
	// linearly from 0.0 to 1.0 over one cycle, before falling back to 0.0
	// and repeating.
	Sawtooth
	// Represents a periodic function that generates a sine wave, resized to
	// return a value between 0.0 and 1.0 inclusive, phase-shifted so that
	// values calculated when phase is close to an integer will be 0.0.
	Sine
	// Represents a periodic function that generates 0.0 or 1.0 for the first
	// half or second half of each cycle, respectively.
	Square
	// Represents a periodic function that generates a triangle wave, rising
	// linearly from 0.0 to 1.0 over first half cycle, then falling linearly
	// to 0.0 for second half of cycle.
	Triangle
)

var (
	ErrInvalidPeriodicType = errors.New("invalid PeriodicType name")
	typeStrings            = map[PeriodicType]string{
		Sawtooth: "sawtooth",
		Sine:     "sine",
		Square:   "square",
		Triangle: "triangle",
	}
	stringTypes = map[string]PeriodicType{
		"sawtooth": Sawtooth,
		"sine":     Sine,
		"square":   Square,
		"triangle": Triangle,
	}
	typeCalculators = map[PeriodicType]ValueCalculator{
		Invalid: func(_ float64) float64 {
			return 0.0
		},
		Sawtooth: func(phase float64) float64 {
			return phase - math.Floor(phase)
		},
		Sine: func(phase float64) float64 {
			// Shift phase by pi/2 to get a value that starts at zero
			// instead of 0.5.
			return 0.5 + math.Sin(math.Pi*2.0*(phase-0.25))/2.0
		},
		Square: func(phase float64) float64 {
			if math.Signbit(math.Sin(math.Pi * 2.0 * phase)) {
				return 1.0
			}
			return 0.0
		},
		Triangle: func(phase float64) float64 {
			return math.Abs(2.0 * (phase - math.Floor(0.5+(phase))))
		},
	}
)

// Returns a string identifier for the PeriodicType, or "unknown" if it is an
// unrecognised type.
func (v PeriodicType) String() string {
	if str, ok := typeStrings[v]; ok {
		return str
	}
	return "unknown"
}

// Returns a ValueCalculator for the PeriodicType. If the periodic type does not
// match with a known implementation the Invalid function will be returned.
func (v PeriodicType) Calculator() ValueCalculator {
	if fn, ok := typeCalculators[v]; ok {
		return fn
	}
	return typeCalculators[Invalid]
}

func ParsePeriodicType(name string) (PeriodicType, error) {
	if value, ok := stringTypes[name]; ok {
		return value, nil
	}
	return Invalid, fmt.Errorf("i%q: %w", name, ErrInvalidPeriodicType)
}

// Creates a new wrapped ValueCalculator from a PeriodicType that returns values
// in the range a through b.
func NewPeriodicRangeCalculator(a, b float64, periodicType PeriodicType) ValueCalculator {
	min := math.Min(a, b)
	delta := math.Abs(a - b)
	unitCalculator := periodicType.Calculator()
	return func(phase float64) float64 {
		return delta*unitCalculator(phase) + min
	}
}

// Returns a generator function that when called will generate a Metric value
// using the calculator function provided on each received ticker, writing the
// value to the returned output channel.
func NewPeriodicGenerator(logger logr.Logger, periodicCalculator ValueCalculator, period time.Duration) (PeriodicGenerator, <-chan Metric) {
	logger = logger.WithValues("period", period)
	logger.V(1).Info("Building periodic value generator")
	output := make(chan Metric, 1)
	tZero := time.Now()
	return func(ctx context.Context, ticker <-chan time.Time) error {
		defer close(output)
		for {
			select {
			case <-ctx.Done():
				logger.V(2).Info("Context has been cancelled; exiting")
				return nil
			// NOTE: ticker channel is never closed; context must reach
			// a deadline or be cancelled to prevent deadlock.
			case tick := <-ticker:
				metric := Metric{
					Value:     periodicCalculator(tick.Sub(tZero).Seconds() / period.Seconds()),
					Timestamp: tick,
				}
				select {
				case output <- metric:
					logger.V(2).Info("Wrote new value to output channel", "metric", metric)
				default:
					logger.V(2).Info("Can't write to output channel; dropping value", "metric", metric)
				}
			}
		}
	}, output
}
