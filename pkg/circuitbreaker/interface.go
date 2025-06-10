package circuitbreaker

import (
	"context"
	"time"
)

// CircuitBreaker interface defines the contract for circuit breaker implementations
type CircuitBreaker interface {
	// Execute runs the given function with circuit breaker protection
	Execute(func() (interface{}, error)) (interface{}, error)
	
	// State returns the current state of the circuit breaker
	State() State
	
	// Counts returns the current counts
	Counts() Counts
	
	// Name returns the circuit breaker name
	Name() string
}

// Config represents configuration for creating a circuit breaker
type Config struct {
	Name        string
	MaxRequests uint32
	Interval    time.Duration
	Timeout     time.Duration
	ReadyToTrip func(Counts) bool
}

// NewCircuitBreaker creates a new circuit breaker with the given config
func NewCircuitBreaker(config Config) CircuitBreaker {
	settings := Settings{
		Name:        config.Name,
		MaxRequests: config.MaxRequests,
		Interval:    config.Interval,
		Timeout:     config.Timeout,
		ReadyToTrip: config.ReadyToTrip,
	}
	
	return &circuitBreakerImpl{
		cb: NewCircuitBreakerInternal(settings),
	}
}

// circuitBreakerImpl implements the CircuitBreaker interface
type circuitBreakerImpl struct {
	cb *CircuitBreakerInternal
}

func (c *circuitBreakerImpl) Execute(fn func() (interface{}, error)) (interface{}, error) {
	return c.cb.Execute(context.Background(), fn)
}

func (c *circuitBreakerImpl) State() State {
	return c.cb.State()
}

func (c *circuitBreakerImpl) Counts() Counts {
	return c.cb.Counts()
}

func (c *circuitBreakerImpl) Name() string {
	return c.cb.name
}