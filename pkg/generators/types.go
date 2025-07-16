package generators

import (
	"errors"
	"fmt"
	"math"
)

// PeriodicType defines the periodic function generators known to the package.
type PeriodicType int

// ValueCalculator defines a function that will return a float64 value for the given phase of the cycle depending on the
// PeriodicType used.
type ValueCalculator func(phase float64) float64

const (
	// Invalid represents an unrecognised periodic function that will return 0.0 on all calls.
	Invalid PeriodicType = iota
	// Sawtooth represents a periodic function that generates a sawtooth wave, rising linearly from 0.0 to 1.0 over
	// one cycle, before falling back to 0.0 and repeating.
	Sawtooth
	// Sine represents a periodic function that generates a sine wave, resized to return a value between 0.0 and 1.0
	// inclusive, phase-shifted so that values calculated when phase is close to an integer will be 0.0.
	Sine
	// Square represents a periodic function that generates 0.0 or 1.0 for the first half or second half of each
	// cycle, respectively.
	Square
	// Triangle represents a periodic function that generates a triangle wave, rising linearly from 0.0 to 1.0 over
	// first half cycle, then falling linearly to 0.0 for second half of cycle.
	Triangle
)

// ErrInvalidPeriodicType is returned when an unknown PeriodicType is encountered.
var ErrInvalidPeriodicType = errors.New("invalid PeriodicType name")

// String returns a string identifier for the PeriodicType, or "unknown" if it is an unrecognised type.
func (pt PeriodicType) String() string {
	switch pt {
	case Invalid:
		return "unknown"
	case Sawtooth:
		return "sawtooth"
	case Sine:
		return "sine"
	case Square:
		return "square"
	case Triangle:
		return "triangle"
	default:
		return "unknown"
	}
}

// ValueCalculator returns a concrete ValueCalculator for the PeriodicType. If the periodic type does not match with a
// known implementation the Invalid function will be returned.
func (pt PeriodicType) ValueCalculator() ValueCalculator {
	switch pt {
	case Invalid:
		return func(_ float64) float64 {
			return 0.0
		}
	case Sawtooth:
		return func(phase float64) float64 {
			return phase - math.Floor(phase)
		}
	case Sine:
		return func(phase float64) float64 {
			// Shift phase by pi/2 to get a value that starts at zero
			// instead of 0.5.
			return 0.5 + math.Sin(math.Pi*2.0*(phase-0.25))/2.0
		}
	case Square:
		return func(phase float64) float64 {
			if math.Signbit(math.Sin(math.Pi * 2.0 * phase)) {
				return 1.0
			}
			return 0.0
		}
	case Triangle:
		return func(phase float64) float64 {
			return math.Abs(2.0 * (phase - math.Floor(0.5+(phase))))
		}
	default:
		return func(_ float64) float64 {
			return 0.0
		}
	}
}

// ParsePeriodicType returns a valid PeriodicType from a supplied string. If the string does not match an known type an
// error will be returned.
func ParsePeriodicType(name string) (PeriodicType, error) {
	switch name {
	case "sawtooth":
		return Sawtooth, nil
	case "sine":
		return Sine, nil
	case "square":
		return Square, nil
	case "triangle":
		return Triangle, nil
	default:
		return Invalid, fmt.Errorf("error parsing %q to PeriodicType: %w", name, ErrInvalidPeriodicType)
	}
}

// NewPeriodicRangeCalculator returns new wrapped ValueCalculator from a PeriodicType that returns values in the range a
// through b.
func NewPeriodicRangeCalculator(a, b float64, periodicType PeriodicType) ValueCalculator {
	minimumValue := math.Min(a, b)
	delta := math.Abs(a - b)
	unitCalculator := periodicType.ValueCalculator()
	return func(phase float64) float64 {
		return delta*unitCalculator(phase) + minimumValue
	}
}
