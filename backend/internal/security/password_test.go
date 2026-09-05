package security

import "testing"

func TestPasswordHashAndCompare(t *testing.T) {
	t.Parallel()

	hasher := NewPasswordHasher()
	hash, err := hasher.Hash("Demo1234!")
	if err != nil {
		t.Fatal(err)
	}
	if hash == "Demo1234!" {
		t.Fatal("password stored in plaintext")
	}
	if err := hasher.Compare(hash, "Demo1234!"); err != nil {
		t.Fatalf("compare valid password: %v", err)
	}
	if err := hasher.Compare(hash, "wrong-pass"); err == nil {
		t.Fatal("expected mismatch")
	}
}
