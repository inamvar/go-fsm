# Finite State Machine (FSM) in Go

A production-ready finite state machine implementation with:
- Thread-safe operations
- State persistence
- Retry mechanisms
- Callback hooks
- Metadata tracking

## Features

- **Simple API**: Easy to integrate
- **Persistence**: Save/load state
- **Error Handling**: Built-in retry logic
- **Extensible**: Custom storage backends

## Usage

```go
// Initialize
repo := fsm.NewMemoryRepository()
fsm := fsm.New("order-123", "created", repo)

// Add transitions
fsm.AddTransition("created", "processing", "process")

// Register callbacks
fsm.RegisterBefore("process", func(_, from, to string, _ interface{}) fsm.CallbackResult {
    fmt.Printf("Transitioning from %s to %s\n", from, to)
    return fsm.CallbackResult{}
})

// Execute transition
err := fsm.Transition(ctx, "process", nil, fsm.WithRetry(3, backoff))
```

## Options

- `WithRetry(maxRetries int, backoff func(int) time.Duration)` - Enable retries