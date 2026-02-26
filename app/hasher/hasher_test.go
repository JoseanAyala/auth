package hasher

import (
	"testing"
)

func TestErrQueueFull(t *testing.T) {
	// Create dispatcher but do NOT start workers — jobs will pile up.
	d := NewDispatcher()

	// Fill the buffer.
	for range cap(d.jobs) {
		err := d.Submit(HashJob{Password: "x", Result: make(chan HashResult, 1)})
		if err != nil {
			t.Fatalf("unexpected error filling buffer: %v", err)
		}
	}

	// Next submit must fail.
	err := d.Submit(HashJob{Password: "overflow", Result: make(chan HashResult, 1)})
	if err != ErrQueueFull {
		t.Fatalf("expected ErrQueueFull, got: %v", err)
	}
}
