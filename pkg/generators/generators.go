// Package generators provides functionality to generatehas types that can generate periodic, timestamped, values
// that can be used as a source of synthetic metrics.
package generators

import (
	"context"
	"sync"
	"time"

	"github.com/go-logr/logr"
)

// Metric represents a point-in-time generated value which will be written
// to the output channel of the PeriodicGenerator function.
type Metric struct {
	// The value of the generated metric.
	Value float64
	// The timestamp of the generated metric.
	Timestamp time.Time
}

// Defines a function that will block until the context is cancelled,
// emitting a Metric value to the output channel on each tick received
// on the ticker channel.
type PeriodicGenerator func(context.Context, <-chan time.Time)

// Accumulates the fluent configuration options that will be used to create the
// PeriodicGenerator function and Metic channel.
type config struct {
	logger     logr.Logger
	calculator ValueCalculator
	period     time.Duration
	bufferSize int
}

// Defines a generator configuration option function.
type Option func(*config) error

// Use the supplied Logger instance for the generator function.
func WithLogger(logger logr.Logger) Option {
	return func(c *config) error {
		c.logger = logger
		return nil
	}
}

// Use the supplied ValueCalculator as the point-in-time generator function.
func WithValueCalculator(calculator ValueCalculator) Option {
	return func(c *config) error {
		c.calculator = calculator
		return nil
	}
}

// Sets the duration of a single waveform cycle. For example, if period is 60s
// then the periodic generator will complete a full cycle of values every minute.
func WithPeriod(period time.Duration) Option {
	return func(c *config) error {
		c.period = period
		return nil
	}
}

// Returns a PeriodicGenerator function that will generate a Metric value on each
// tick, and a read-only channel that will receive the generated value.
// The default generator is a sawtooth waveform in the range 0 <= value <= 100
// with a period of 20 minutes, and a buffered channel with single Metric capacity.
// The various Option functions can be used to change this.
func NewPeriodicGenerator(options ...Option) (PeriodicGenerator, <-chan Metric, error) {
	config := &config{
		logger:     logr.Discard(),
		calculator: NewPeriodicRangeCalculator(0.0, 100.0, Sawtooth),
		period:     20 * time.Minute,
		bufferSize: 1,
	}
	for _, option := range options {
		if err := option(config); err != nil {
			return nil, nil, err
		}
	}
	config.logger.V(2).Info("Building PeriodicGenerator and channel")
	ch := make(chan Metric, config.bufferSize)
	return func(ctx context.Context, ticker <-chan time.Time) {
		defer close(ch)
		var firstTick sync.Once
		var tZero time.Time
		for {
			select {
			case <-ctx.Done():
				config.logger.V(2).Info("Context has been cancelled; exiting")
				return
			// NOTE: ticker channel is never closed; context must reach
			// a deadline or be cancelled to prevent deadlock.
			case tick := <-ticker:
				// Set tZero to the timestamp of the first received tick
				firstTick.Do(func() { tZero = tick })
				metric := Metric{
					Value:     config.calculator(tick.Sub(tZero).Seconds() / config.period.Seconds()),
					Timestamp: tick,
				}
				select {
				case ch <- metric:
					config.logger.V(2).Info("Wrote new value to output channel", "metric", metric)
				default:
					config.logger.V(2).Info("Can't write to output channel; dropping value", "metric", metric)
				}
			}
		}
	}, ch, nil
}
