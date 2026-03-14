package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

var ErrCircuitOpen = errors.New("circuit breaker is open")

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

type Config struct {
	FailureThreshold int
	SuccessThreshold int
	Timeout          time.Duration
}

type CircuitBreaker struct {
	mu              sync.RWMutex
	state           State
	failures        int
	successes       int
	lastFailureTime time.Time
	config          Config
}

func New(config Config) *CircuitBreaker {
	return &CircuitBreaker{
		state:  StateClosed,
		config: config,
	}
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
	if !cb.allowRequest() {
		return ErrCircuitOpen
	}

	err := fn()

	cb.recordResult(err)
	return err
}

func (cb *CircuitBreaker) allowRequest() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(cb.lastFailureTime) > cb.config.Timeout {
			cb.state = StateHalfOpen
			cb.successes = 0
			return true
		}
		return false
	case StateHalfOpen:
		return true
	}
	return false
}

func (cb *CircuitBreaker) recordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		cb.lastFailureTime = time.Now()

		if cb.state == StateHalfOpen || cb.failures >= cb.config.FailureThreshold {
			cb.state = StateOpen
		}
	} else {
		cb.successes++
		cb.failures = 0

		if cb.state == StateHalfOpen {
			if cb.successes >= cb.config.SuccessThreshold {
				cb.state = StateClosed
			}
		}
	}
}

func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

func (cb *CircuitBreaker) Failures() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.failures
}

func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = StateClosed
	cb.failures = 0
	cb.successes = 0
}

type Option func(*CircuitBreaker)

func WithFailureThreshold(n int) Option {
	return func(cb *CircuitBreaker) {
		cb.config.FailureThreshold = n
	}
}

func WithSuccessThreshold(n int) Option {
	return func(cb *CircuitBreaker) {
		cb.config.SuccessThreshold = n
	}
}

func WithTimeout(d time.Duration) Option {
	return func(cb *CircuitBreaker) {
		cb.config.Timeout = d
	}
}

func NewWithOptions(opts ...Option) *CircuitBreaker {
	cb := &CircuitBreaker{
		state: StateClosed,
		config: Config{
			FailureThreshold: 5,
			SuccessThreshold: 3,
			Timeout:          30 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(cb)
	}
	return cb
}

type Service struct {
	breakers map[string]*CircuitBreaker
	mu       sync.RWMutex
}

func NewService() *Service {
	return &Service{
		breakers: make(map[string]*CircuitBreaker),
	}
}

func (s *Service) GetBreaker(name string) *CircuitBreaker {
	s.mu.RLock()
	cb, exists := s.breakers[name]
	s.mu.RUnlock()

	if exists {
		return cb
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if cb, exists = s.breakers[name]; exists {
		return cb
	}

	cb = New(Config{
		FailureThreshold: 5,
		SuccessThreshold: 3,
		Timeout:          30 * time.Second,
	})
	s.breakers[name] = cb
	return cb
}

func (s *Service) Execute(name string, fn func() error) error {
	return s.GetBreaker(name).Execute(fn)
}
