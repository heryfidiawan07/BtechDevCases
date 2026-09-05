package service

import "testing"

func TestWelcomeMessage(t *testing.T) {
	t.Parallel()

	got := WelcomeMessage("alice@demo.com")
	want := "Hello alice@demo.com, welcome back"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestNormalizeEmail(t *testing.T) {
	t.Parallel()

	got := normalizeEmail("  Alice@Demo.com ")
	if got != "alice@demo.com" {
		t.Fatalf("got %q", got)
	}
}
