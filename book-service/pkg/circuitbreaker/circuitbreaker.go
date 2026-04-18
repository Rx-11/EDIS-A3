package circuitbreaker

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	ErrCircuitOpen = errors.New("circuit breaker is open")
)

type State int

const (
	StateClosed State = iota
	StateHalfOpen
	StateOpen
)

type CircuitState struct {
	State       State     `json:"state"`
	Failures    int       `json:"failures"`
	LastFailure time.Time `json:"last_failure"`
}

type CircuitBreaker struct {
	mu          sync.Mutex
	state       CircuitState
	maxFailures int
	timeout     time.Duration
	stateFile   string
}

func NewCircuitBreaker(maxFailures int, timeout time.Duration) *CircuitBreaker {
	cb := &CircuitBreaker{
		state:       CircuitState{State: StateClosed, Failures: 0},
		maxFailures: maxFailures,
		timeout:     timeout,
		stateFile:   "/app/state/cb_state.json",
	}
	cb.loadState()
	return cb
}

func (cb *CircuitBreaker) loadState() {
	if cb.stateFile == "" {
		return
	}
	data, err := os.ReadFile(cb.stateFile)
	if err == nil {
		json.Unmarshal(data, &cb.state)
	}
}

func (cb *CircuitBreaker) saveState() {
	if cb.stateFile == "" {
		return
	}
	dir := filepath.Dir(cb.stateFile)
	os.MkdirAll(dir, 0755)
	data, err := json.Marshal(cb.state)
	if err == nil {
		os.WriteFile(cb.stateFile, data, 0644)
	}
}

func (cb *CircuitBreaker) Execute(req func() (interface{}, error)) (interface{}, error) {
	cb.mu.Lock()
	cb.loadState()

	if cb.state.State == StateOpen {
		if time.Since(cb.state.LastFailure) > cb.timeout {
			cb.state.State = StateHalfOpen
			cb.saveState()
		} else {
			cb.mu.Unlock()
			return nil, ErrCircuitOpen
		}
	} else if cb.state.State == StateHalfOpen {
		cb.mu.Unlock()
		return nil, ErrCircuitOpen
	}
	cb.mu.Unlock()

	res, err := req()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.loadState()

	if err != nil {
		cb.state.Failures++
		cb.state.LastFailure = time.Now()
		if cb.state.Failures >= cb.maxFailures {
			cb.state.State = StateOpen
		}
		cb.saveState()
		return nil, err
	}

	cb.state.State = StateClosed
	cb.state.Failures = 0
	cb.saveState()
	return res, nil
}
