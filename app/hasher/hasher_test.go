package hasher

import (
	"strings"
	"testing"
)

func TestHashPassword(t *testing.T) {
	d := NewDispatcher()
	d.Start()
	defer d.Stop()

	result := make(chan HashResult, 1)
	if err := d.Submit(HashJob{Password: "s3cret", Result: result}); err != nil {
		t.Fatalf("submit: %v", err)
	}

	r := <-result
	if r.Err != nil {
		t.Fatalf("hash error: %v", r.Err)
	}
	if !strings.HasPrefix(r.Hash, "$argon2id$v=19$") {
		t.Fatalf("unexpected hash format: %s", r.Hash)
	}
}

func TestVerifyPassword(t *testing.T) {
	d := NewDispatcher()
	d.Start()
	defer d.Stop()

	// Hash a password first.
	hashResult := make(chan HashResult, 1)
	if err := d.Submit(HashJob{Password: "correct-horse", Result: hashResult}); err != nil {
		t.Fatalf("submit hash: %v", err)
	}
	hr := <-hashResult
	if hr.Err != nil {
		t.Fatalf("hash error: %v", hr.Err)
	}

	// Verify with the same password.
	verifyResult := make(chan VerifyResult, 1)
	if err := d.Submit(VerifyJob{Password: "correct-horse", StoredHash: hr.Hash, Result: verifyResult}); err != nil {
		t.Fatalf("submit verify: %v", err)
	}
	vr := <-verifyResult
	if vr.Err != nil {
		t.Fatalf("verify error: %v", vr.Err)
	}
	if !vr.Match {
		t.Fatal("expected password to match")
	}
}

func TestVerifyWrongPassword(t *testing.T) {
	d := NewDispatcher()
	d.Start()
	defer d.Stop()

	// Hash one password.
	hashResult := make(chan HashResult, 1)
	if err := d.Submit(HashJob{Password: "right-password", Result: hashResult}); err != nil {
		t.Fatalf("submit hash: %v", err)
	}
	hr := <-hashResult
	if hr.Err != nil {
		t.Fatalf("hash error: %v", hr.Err)
	}

	// Verify with a different password.
	verifyResult := make(chan VerifyResult, 1)
	if err := d.Submit(VerifyJob{Password: "wrong-password", StoredHash: hr.Hash, Result: verifyResult}); err != nil {
		t.Fatalf("submit verify: %v", err)
	}
	vr := <-verifyResult
	if vr.Err != nil {
		t.Fatalf("verify error: %v", vr.Err)
	}
	if vr.Match {
		t.Fatal("expected password NOT to match")
	}
}

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
