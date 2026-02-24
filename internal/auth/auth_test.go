package auth

import (
	"testing"

	"github.com/Upellift99/GateCHA/internal/testutil"
)

func TestEnsureAdminUser(t *testing.T) {
	db := testutil.SetupTestDB(t)

	if err := EnsureAdminUser(db, "admin", "password123"); err != nil {
		t.Fatalf("EnsureAdminUser failed: %v", err)
	}

	// Second call should be no-op
	if err := EnsureAdminUser(db, "admin", "different"); err != nil {
		t.Fatalf("EnsureAdminUser (2nd call) failed: %v", err)
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM admin_users").Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 admin user, got %d", count)
	}
}

func TestValidateCredentials(t *testing.T) {
	db := testutil.SetupTestDB(t)
	EnsureAdminUser(db, "admin", "password123")

	ok, err := ValidateCredentials(db, "admin", "password123")
	if err != nil {
		t.Fatalf("ValidateCredentials failed: %v", err)
	}
	if !ok {
		t.Error("expected valid credentials")
	}

	ok, err = ValidateCredentials(db, "admin", "wrong")
	if err != nil {
		t.Fatalf("ValidateCredentials (wrong pw) failed: %v", err)
	}
	if ok {
		t.Error("expected invalid credentials for wrong password")
	}

	ok, err = ValidateCredentials(db, "nonexistent", "password123")
	if err != nil {
		t.Fatalf("ValidateCredentials (bad user) failed: %v", err)
	}
	if ok {
		t.Error("expected invalid for nonexistent user")
	}
}

func TestGenerateAndValidateJWT(t *testing.T) {
	secret := "test-secret-key"

	token, expiresAt, err := GenerateJWT("admin", secret)
	if err != nil {
		t.Fatalf("GenerateJWT failed: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
	if expiresAt.IsZero() {
		t.Error("expected non-zero expiry")
	}

	claims, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("ValidateJWT failed: %v", err)
	}
	if claims["sub"] != "admin" {
		t.Errorf("expected sub=admin, got %v", claims["sub"])
	}
}

func TestValidateJWT_WrongSecret(t *testing.T) {
	token, _, _ := GenerateJWT("admin", "correct-secret")

	_, err := ValidateJWT(token, "wrong-secret")
	if err == nil {
		t.Error("expected error for wrong secret")
	}
}

func TestValidateJWT_InvalidToken(t *testing.T) {
	_, err := ValidateJWT("invalid.token.here", "secret")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestChangePassword(t *testing.T) {
	db := testutil.SetupTestDB(t)
	EnsureAdminUser(db, "admin", "old-password")

	if err := ChangePassword(db, "admin", "new-password"); err != nil {
		t.Fatalf("ChangePassword failed: %v", err)
	}

	ok, _ := ValidateCredentials(db, "admin", "old-password")
	if ok {
		t.Error("old password should be invalid after change")
	}

	ok, _ = ValidateCredentials(db, "admin", "new-password")
	if !ok {
		t.Error("new password should be valid after change")
	}
}
