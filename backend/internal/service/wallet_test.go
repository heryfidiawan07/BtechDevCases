package service

import "testing"

func TestIdempotencyKeyPattern(t *testing.T) {
	t.Parallel()

	valid := []string{"550e8400-e29b-41d4-a716-446655440000", "transfer_abc-123", "KEY.01-OK"}
	for _, key := range valid {
		if !idempotencyKeyPattern.MatchString(key) {
			t.Fatalf("expected valid key %q", key)
		}
	}

	invalid := []string{"", "abc", "has space", "bad/key", "x"}
	for _, key := range invalid {
		if idempotencyKeyPattern.MatchString(key) {
			t.Fatalf("expected invalid key %q", key)
		}
	}
}
