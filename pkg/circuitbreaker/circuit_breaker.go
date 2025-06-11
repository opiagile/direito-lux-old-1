package circuitbreaker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/opiagile/direito-lux/pkg/logger"
	"go.uber.org/zap"
)

// State represents the circuit breaker state
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// Settings configures circuit breaker behavior
type Settings struct {
	Name          string
	MaxRequests   uint32        // Max requests allowed in half-open state
	Interval      time.Duration // Interval for closed state
	Timeout       time.Duration // Timeout for open state
	ReadyToTrip   func(counts Counts) bool
	OnStateChange func(name string, from State, to State)
}

// Counts holds the numbers of requests and their successes/failures
type Counts struct {
	Requests             uint32
	TotalSuccesses       uint32
	TotalFailures        uint32
	ConsecutiveSuccesses uint32
	ConsecutiveFailures  uint32
}

// CircuitBreakerInternal implements the circuit breaker pattern
type CircuitBreakerInternal struct {
	name          string
	maxRequests   uint32
	interval      time.Duration
	timeout       time.Duration
	readyToTrip   func(counts Counts) bool
	onStateChange func(name string, from State, to State)

	mutex      sync.Mutex
	state      State
	generation uint64
	counts     Counts
	expiry     time.Time
}

// NewCircuitBreakerInternal creates a new circuit breaker
func NewCircuitBreakerInternal(st Settings) *CircuitBreakerInternal {
	cb := &CircuitBreakerInternal{
		name:          st.Name,
		maxRequests:   st.MaxRequests,
		interval:      st.Interval,
		timeout:       st.Timeout,
		readyToTrip:   st.ReadyToTrip,
		onStateChange: st.OnStateChange,
	}

	// Default settings
	if cb.maxRequests == 0 {
		cb.maxRequests = 1
	}
	if cb.interval == 0 {
		cb.interval = 0
	}
	if cb.timeout == 0 {
		cb.timeout = 60 * time.Second
	}
	if cb.readyToTrip == nil {
		cb.readyToTrip = defaultReadyToTrip
	}
	if cb.onStateChange == nil {
		cb.onStateChange = defaultOnStateChange
	}

	cb.toNewGeneration(time.Now())

	return cb
}

// defaultReadyToTrip returns true when consecutive failures >= 5
func defaultReadyToTrip(counts Counts) bool {
	return counts.ConsecutiveFailures >= 5
}

// defaultOnStateChange logs state changes
func defaultOnStateChange(name string, from, to State) {
	logger.Info("Circuit breaker state change",
		zap.String("name", name),
		zap.String("from", from.String()),
		zap.String("to", to.String()))
}

// Execute runs the given request if the circuit breaker accepts it
func (cb *CircuitBreakerInternal) Execute(ctx context.Context, req func() (interface{}, error)) (interface{}, error) {
	generation, err := cb.beforeRequest()
	if err != nil {
		return nil, err
	}

	defer func() {
		e := recover()
		if e != nil {
			cb.afterRequest(generation, false)
			panic(e)
		}
	}()

	// Execute with context
	done := make(chan struct{})
	var result interface{}
	var reqErr error

	go func() {
		defer close(done)
		result, reqErr = req()
	}()

	select {
	case <-ctx.Done():
		cb.afterRequest(generation, false)
		return nil, ctx.Err()
	case <-done:
		cb.afterRequest(generation, reqErr == nil)
		return result, reqErr
	}
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreakerInternal) State() State {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	state, _ := cb.currentState(now)
	return state
}

// Counts returns the current counts
func (cb *CircuitBreakerInternal) Counts() Counts {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	return cb.counts
}

// Reset resets the circuit breaker
func (cb *CircuitBreakerInternal) Reset() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.toNewGeneration(time.Now())
}

func (cb *CircuitBreakerInternal) beforeRequest() (uint64, error) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	state, generation := cb.currentState(now)

	if state == StateOpen {
		return generation, ErrOpenState{
			Name:      cb.name,
			State:     state,
			Remaining: cb.expiry.Sub(now),
		}
	} else if state == StateHalfOpen && cb.counts.Requests >= cb.maxRequests {
		return generation, ErrTooManyRequests{
			Name:  cb.name,
			State: state,
		}
	}

	cb.counts.Requests++
	return generation, nil
}

func (cb *CircuitBreakerInternal) afterRequest(before uint64, success bool) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	state, generation := cb.currentState(now)
	if generation != before {
		return
	}

	if success {
		cb.onSuccess(state, now)
	} else {
		cb.onFailure(state, now)
	}
}

func (cb *CircuitBreakerInternal) onSuccess(state State, now time.Time) {
	switch state {
	case StateClosed:
		cb.counts.TotalSuccesses++
		cb.counts.ConsecutiveSuccesses++
		cb.counts.ConsecutiveFailures = 0
	case StateHalfOpen:
		cb.counts.TotalSuccesses++
		cb.counts.ConsecutiveSuccesses++
		if cb.counts.ConsecutiveSuccesses >= cb.maxRequests {
			cb.toNewGeneration(now)
		}
	}
}

func (cb *CircuitBreakerInternal) onFailure(state State, now time.Time) {
	switch state {
	case StateClosed:
		cb.counts.TotalFailures++
		cb.counts.ConsecutiveFailures++
		cb.counts.ConsecutiveSuccesses = 0
		if cb.readyToTrip(cb.counts) {
			cb.setState(StateOpen, now)
		}
	case StateHalfOpen:
		cb.counts.TotalFailures++
		cb.setState(StateOpen, now)
	}
}

func (cb *CircuitBreakerInternal) currentState(now time.Time) (State, uint64) {
	switch cb.state {
	case StateClosed:
		if !cb.expiry.IsZero() && cb.expiry.Before(now) {
			cb.toNewGeneration(now)
		}
	case StateOpen:
		if cb.expiry.Before(now) {
			cb.setState(StateHalfOpen, now)
		}
	}
	return cb.state, cb.generation
}

func (cb *CircuitBreakerInternal) setState(state State, now time.Time) {
	if cb.state == state {
		return
	}

	prev := cb.state
	cb.state = state

	cb.toNewGeneration(now)

	if cb.onStateChange != nil {
		cb.onStateChange(cb.name, prev, state)
	}
}

func (cb *CircuitBreakerInternal) toNewGeneration(now time.Time) {
	cb.generation++
	cb.counts = Counts{}

	var zero time.Time
	switch cb.state {
	case StateClosed:
		if cb.interval == 0 {
			cb.expiry = zero
		} else {
			cb.expiry = now.Add(cb.interval)
		}
	case StateOpen:
		cb.expiry = now.Add(cb.timeout)
	default: // StateHalfOpen
		cb.expiry = zero
	}
}

// ErrOpenState is returned when the circuit breaker is open
type ErrOpenState struct {
	Name      string
	State     State
	Remaining time.Duration
}

func (e ErrOpenState) Error() string {
	return fmt.Sprintf("circuit breaker '%s' is %s (remaining: %v)",
		e.Name, e.State, e.Remaining.Round(time.Second))
}

// ErrTooManyRequests is returned when too many requests are made in half-open state
type ErrTooManyRequests struct {
	Name  string
	State State
}

func (e ErrTooManyRequests) Error() string {
	return fmt.Sprintf("circuit breaker '%s' is %s: too many requests", e.Name, e.State)
}

// Manager manages multiple circuit breakers
type Manager struct {
	breakers map[string]*CircuitBreakerInternal
	mutex    sync.RWMutex
}

// NewManager creates a new circuit breaker manager
func NewManager() *Manager {
	return &Manager{
		breakers: make(map[string]*CircuitBreakerInternal),
	}
}

// Get returns a circuit breaker by name, creating it if it doesn't exist
func (m *Manager) Get(name string) *CircuitBreakerInternal {
	m.mutex.RLock()
	cb, exists := m.breakers[name]
	m.mutex.RUnlock()

	if exists {
		return cb
	}

	// Create new circuit breaker
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Double-check after acquiring write lock
	if cb, exists = m.breakers[name]; exists {
		return cb
	}

	cb = NewCircuitBreakerInternal(Settings{
		Name:        name,
		MaxRequests: 3,
		Interval:    10 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
	})

	m.breakers[name] = cb
	return cb
}

// Reset resets a circuit breaker by name
func (m *Manager) Reset(name string) {
	m.mutex.RLock()
	cb, exists := m.breakers[name]
	m.mutex.RUnlock()

	if exists {
		cb.Reset()
	}
}

// ResetAll resets all circuit breakers
func (m *Manager) ResetAll() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for _, cb := range m.breakers {
		cb.Reset()
	}
}

// GetAll returns all circuit breakers and their states
func (m *Manager) GetAll() map[string]State {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	states := make(map[string]State)
	for name, cb := range m.breakers {
		states[name] = cb.State()
	}
	return states
}
