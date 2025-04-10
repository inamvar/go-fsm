package go_fsm

import "time"

type CallbackFunc func(transition, from, to string, args interface{}) CallbackResult

type CallbackResult struct {
	// Explicit status field makes state handling clear
	Message    string
	Status     CallbackStatus
	Metadata   map[string]interface{}
	RetryAfter time.Duration
}

type CallbackStatus int

const (
	StatusSuccess CallbackStatus = iota
	StatusFailure
	StatusUnknown // Explicit unknown state
	StatusRetry
)

type transitionConfig struct {
	maxRetries int
	backoff    func(attempt int) time.Duration
}

type TransitionOption func(*transitionConfig)

func defaultTransitionConfig() *transitionConfig {
	return &transitionConfig{
		maxRetries: 0,
		backoff:    func(int) time.Duration { return 0 },
	}
}

func WithRetry(maxRetries int, backoff func(int) time.Duration) TransitionOption {
	return func(c *transitionConfig) {
		c.maxRetries = maxRetries
		c.backoff = backoff
	}
}

func (f *FSM) RegisterBefore(condition string, cb CallbackFunc) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.callbacks.before[condition] = cb
}

func (f *FSM) RegisterAfter(condition string, cb CallbackFunc) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.callbacks.after[condition] = cb
}
