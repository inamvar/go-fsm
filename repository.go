package go_fsm

import (
	"context"
	"sync"
)

type Repository interface {
	Save(ctx context.Context, id, state string, metadata map[string]interface{}) error
	Load(ctx context.Context, id string) (state string, metadata map[string]interface{}, err error)
}

// MemoryRepository for testing
type MemoryRepository struct {
	mu     sync.RWMutex
	states map[string]struct {
		state    string
		metadata map[string]interface{}
	}
}

func NewMemoryRepository() Repository {
	return &MemoryRepository{
		states: make(map[string]struct {
			state    string
			metadata map[string]interface{}
		}),
	}
}

func (r *MemoryRepository) Save(ctx context.Context, id, state string, metadata map[string]interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	metaCopy := make(map[string]interface{})
	for k, v := range metadata {
		metaCopy[k] = v
	}

	r.states[id] = struct {
		state    string
		metadata map[string]interface{}
	}{
		state:    state,
		metadata: metaCopy,
	}
	return nil
}

func (r *MemoryRepository) Load(ctx context.Context, id string) (string, map[string]interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data, exists := r.states[id]
	if !exists {
		return "", nil, ErrNotFound
	}

	metaCopy := make(map[string]interface{})
	for k, v := range data.metadata {
		metaCopy[k] = v
	}

	return data.state, metaCopy, nil
}
