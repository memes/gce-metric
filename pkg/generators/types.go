package generators

import (
	"errors"
	"fmt"
	"math"
)

// Defines the periodic function generators known to the package.
type PeriodicType int

// Defines a function that will return a float64 value for the given
// phase of the cycle.
type ValueCalculator func(phase float64) float64

const (
	// Represents an unrecognised periodic function that will return 0.0 on
	// all calls.
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
func (pt PeriodicType) String() string {
	if str, ok := typeStrings[pt]; ok {
		return str
	}
	return "unknown"
}

// Returns a ValueCalculator for the PeriodicType. If the periodic type does not
// match with a known implementation the Invalid function will be returned.
func (pt PeriodicType) ValueCalculator() ValueCalculator {
	if fn, ok := typeCalculators[pt]; ok {
		return fn
	}
	return typeCalculators[Invalid]
}

// Parses and returns a PeriodicType from a supplied string. If the string does
// not match an known type an error will be returned.
func ParsePeriodicType(name string) (PeriodicType, error) {
	if value, ok := stringTypes[name]; ok {
		return value, nil
	}
	return Invalid, fmt.Errorf("error parsing %q to PeriodicType: %w", name, ErrInvalidPeriodicType)
}

// Creates a new wrapped ValueCalculator from a PeriodicType that returns values
// in the range a through b.
func NewPeriodicRangeCalculator(a, b float64, periodicType PeriodicType) ValueCalculator {
	min := math.Min(a, b)
	delta := math.Abs(a - b)
	unitCalculator := periodicType.ValueCalculator()
	return func(phase float64) float64 {
		return delta*unitCalculator(phase) + min
	}
}
