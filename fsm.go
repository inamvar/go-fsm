package go_fsm

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type FSM struct {
	mu                  sync.RWMutex
	id                  string
	current             string
	transitions         map[string]map[string]string // from -> condition -> to
	metadata            map[string]interface{}
	repo                Repository
	callbacks           callbacks
	unknownStateHandler func(currentState, condition string, args interface{}) string
}

type callbacks struct {
	before map[string]CallbackFunc
	after  map[string]CallbackFunc
}

func New(id, initialState string, repo Repository, unknownHandler func(string, string, interface{}) string) *FSM {
	return &FSM{
		id:          id,
		current:     initialState,
		transitions: make(map[string]map[string]string),
		metadata:    make(map[string]interface{}),
		repo:        repo,
		callbacks: callbacks{
			before: make(map[string]CallbackFunc),
			after:  make(map[string]CallbackFunc),
		},
		unknownStateHandler: unknownHandler,
	}
}

func (f *FSM) AddTransition(from, to, condition string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.transitions[from]; !exists {
		f.transitions[from] = make(map[string]string)
	}
	f.transitions[from][condition] = to
}

func (f *FSM) Transition(ctx context.Context, condition string, args interface{}, opts ...TransitionOption) error {
	config := defaultTransitionConfig()
	for _, opt := range opts {
		opt(config)
	}

	for attempt := 0; attempt <= config.maxRetries; attempt++ {
		err := f.executeTransition(ctx, condition, args)
		if err == nil {
			return nil
		}

		if errors.Is(err, ErrUnknownState) {
			return fmt.Errorf("operation aborted due to unknown state: %w", err)
		}

		if attempt < config.maxRetries {
			select {
			case <-time.After(config.backoff(attempt)):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	return ErrMaxRetriesExceeded
}

func (f *FSM) executeTransition(ctx context.Context, condition string, args interface{}) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	from := f.current
	to, valid := f.transitions[from][condition]
	if !valid {
		return ErrInvalidTransition
	}

	// Execute before callback
	if cb, exists := f.callbacks.before[condition]; exists {
		result := cb(condition, from, to, args)
		if result.Status == StatusFailure {
			return f.handleCallbackError(condition, from, to, result.Message)
		}
		if result.Status == StatusUnknown {
			return f.handleUnknownState(condition, from, args)
		}
		f.mergeMetadata(result.Metadata)
	}

	// Change state
	f.current = to
	if err := f.persist(ctx); err != nil {
		f.current = from // Rollback
		return fmt.Errorf("persistence failed: %w", err)
	}

	// Execute after callback
	if cb, exists := f.callbacks.after[condition]; exists {
		result := cb(condition, from, to, args)
		if result.Status == StatusFailure {
			return f.handleCallbackError(condition, from, to, result.Message)
		}
		if result.Status == StatusUnknown {
			return f.handleUnknownState(condition, from, args)
		}
		f.mergeMetadata(result.Metadata)
	}

	return nil
}

func (f *FSM) handleUnknownState(condition, from string, args interface{}) error {
	if f.unknownStateHandler != nil {
		f.current = f.unknownStateHandler(from, condition, args)
		f.persist(context.Background()) // Best effort save
	}

	f.metadata["unknown_state"] = map[string]interface{}{
		"condition": condition,
		"from":      from,
		"args":      args,
		"timestamp": time.Now(),
	}
	return ErrUnknownState
}

func (f *FSM) handleCallbackError(condition, from, to string, message string) error {
	f.metadata["callback_error"] = map[string]interface{}{
		"condition": condition,
		"from":      from,
		"to":        to,
		"error":     message,
		"timestamp": time.Now(),
	}
	return errors.New(message)
}

func (f *FSM) mergeMetadata(updates map[string]interface{}) {
	if updates == nil {
		return
	}
	for k, v := range updates {
		f.metadata[k] = v
	}
}

func (f *FSM) persist(ctx context.Context) error {
	if f.repo == nil {
		return nil
	}
	return f.repo.Save(ctx, f.id, f.current, f.metadata)
}

// Additional helper methods
func (f *FSM) Current() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.current
}

func (f *FSM) Metadata() map[string]interface{} {
	f.mu.RLock()
	defer f.mu.RUnlock()
	metaCopy := make(map[string]interface{})
	for k, v := range f.metadata {
		metaCopy[k] = v
	}
	return metaCopy
}
