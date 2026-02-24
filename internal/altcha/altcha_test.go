package altcha

import (
	"testing"
)

func TestGenerateChallenge(t *testing.T) {
	secret := "test-hmac-secret-0123456789abcdef"

	challenge, err := GenerateChallenge(secret, 100000, "SHA-256", 300)
	if err != nil {
		t.Fatalf("GenerateChallenge failed: %v", err)
	}

	if challenge.Challenge == "" {
		t.Error("expected non-empty challenge")
	}
	if challenge.Salt == "" {
		t.Error("expected non-empty salt")
	}
	if challenge.Algorithm != "SHA-256" {
		t.Errorf("expected SHA-256, got %s", challenge.Algorithm)
	}
	if challenge.MaxNumber != 100000 {
		t.Errorf("expected maxnumber 100000, got %d", challenge.MaxNumber)
	}
}

func TestGenerateChallenge_SHA512(t *testing.T) {
	challenge, err := GenerateChallenge("secret", 50000, "SHA-512", 60)
	if err != nil {
		t.Fatalf("GenerateChallenge SHA-512 failed: %v", err)
	}
	if challenge.Algorithm != "SHA-512" {
		t.Errorf("expected SHA-512, got %s", challenge.Algorithm)
	}
}

func TestVerifyPayload_Invalid(t *testing.T) {
	ok, err := VerifyPayload("some-secret", "not-a-valid-payload")
	if err == nil && ok {
		t.Error("expected failure for invalid payload")
	}
}

func TestVerifyPayload_EmptyPayload(t *testing.T) {
	ok, err := VerifyPayload("some-secret", "")
	if err == nil && ok {
		t.Error("expected failure for empty payload")
	}
}
