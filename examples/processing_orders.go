package main

import (
	"context"
	"fmt"
	"time"

	fsm "github.com/inamvar/go-fsm"
)

func main() {
	// Unknown state handler moves to "error" state
	unknownHandler := func(currentState, condition string, args interface{}) string {
		fmt.Printf("Unknown state detected! Current: %s, Condition: %s, args: %+v\n",
			currentState, condition, args)
		return "error"
	}

	repo := fsm.NewMemoryRepository()
	orderFSM := fsm.New("order-123", "created", repo, unknownHandler)

	// Setup transitions
	orderFSM.AddTransition("created", "processing", "process")
	orderFSM.AddTransition("processing", "done", "final")
	orderFSM.AddTransition("error", "created", "reset")

	// Register callback that might return unknown state
	orderFSM.RegisterBefore("process", func(_, from, to string, _ interface{}) fsm.CallbackResult {
		// Simulate external service failure
		if time.Now().Unix()%2 == 0 { // Random failure
			return fsm.CallbackResult{
				Status:  fsm.StatusUnknown,
				Message: "connection timeout",
				Metadata: map[string]interface{}{
					"service": "payment_gateway",
					"error":   "connection timeout",
				},
			}
		}
		return fsm.CallbackResult{Status: fsm.StatusSuccess}
	})

	// Execute transition
	err := orderFSM.Transition(context.Background(), "process", map[string]any{
		"order_id": "order-123",
		"amount":   234.65,
		"tax":      12,
		"customer": "john doe",
	})

	if err != nil {
		fmt.Println("Transition error:", err)
	}

	fmt.Println("Final state:", orderFSM.Current())
}
