package go_fsm

import "errors"

var (
	ErrInvalidTransition  = errors.New("invalid transition")
	ErrMaxRetriesExceeded = errors.New("max retries exceeded")
	ErrNotFound           = errors.New("not found")
	ErrUnknownState       = errors.New("unknown state detected")
)
