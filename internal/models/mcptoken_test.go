package models_test

import (
	"strings"
	"testing"
	"time"

	"github.com/Upellift99/GateCHA/internal/models"
	"github.com/Upellift99/GateCHA/internal/testutil"
)

func TestCreateMCPTokenReturnsSecretOnce(t *testing.T) {
	db := testutil.SetupTestDB(t)

	token, secret, err := models.CreateMCPToken(db, "Martijn laptop", false)
	if err != nil {
		t.Fatalf("CreateMCPToken failed: %v", err)
	}

	if !strings.HasPrefix(secret, models.MCPTokenPrefix) {
		t.Errorf("expected secret to start with %q, got %q", models.MCPTokenPrefix, secret)
	}
	// 16 random bytes hex-encoded, plus the prefix.
	if len(secret) != len(models.MCPTokenPrefix)+32 {
		t.Errorf("unexpected secret length %d", len(secret))
	}
	if token.Name != "Martijn laptop" || token.ReadOnly {
		t.Errorf("unexpected token: %+v", token)
	}
	if token.LastUsedAt != nil {
		t.Error("a fresh token should have no last-used timestamp")
	}
	// The project stores every date in UTC. Checked on the freshly built value:
	// asserting the location after a round-trip would test the driver instead.
	if token.CreatedAt.Location() != time.UTC {
		t.Errorf("expected CreatedAt in UTC, got %v", token.CreatedAt.Location())
	}

	// The secret must not be recoverable from storage.
	if strings.Contains(token.TokenHash, secret) || token.TokenHash == secret {
		t.Error("the stored digest must not be the secret")
	}
	if token.TokenHash != models.HashMCPToken(secret) {
		t.Error("stored digest does not match the secret digest")
	}
	if !strings.HasPrefix(secret, token.Display) {
		t.Errorf("display %q is not a prefix of the secret", token.Display)
	}
	if len(token.Display) >= len(secret) {
		t.Error("display must not reveal the whole secret")
	}
}

func TestGenerateMCPTokenIsUnique(t *testing.T) {
	seen := make(map[string]bool, 64)
	for range 64 {
		secret, err := models.GenerateMCPToken()
		if err != nil {
			t.Fatalf("GenerateMCPToken failed: %v", err)
		}
		if seen[secret] {
			t.Fatalf("generated a duplicate token: %s", secret)
		}
		seen[secret] = true
	}
}

func TestAuthenticateMCPToken(t *testing.T) {
	db := testutil.SetupTestDB(t)

	created, secret, err := models.CreateMCPToken(db, "CI", true)
	if err != nil {
		t.Fatalf("CreateMCPToken failed: %v", err)
	}

	got, err := models.AuthenticateMCPToken(db, secret)
	if err != nil {
		t.Fatalf("AuthenticateMCPToken failed: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("expected token %d, got %d", created.ID, got.ID)
	}
	if !got.ReadOnly {
		t.Error("read-only flag was not preserved")
	}
}

func TestAuthenticateMCPTokenRejects(t *testing.T) {
	db := testutil.SetupTestDB(t)

	_, secret, err := models.CreateMCPToken(db, "CI", false)
	if err != nil {
		t.Fatalf("CreateMCPToken failed: %v", err)
	}

	cases := map[string]string{
		"empty":            "",
		"missing prefix":   strings.TrimPrefix(secret, models.MCPTokenPrefix),
		"unknown token":    models.MCPTokenPrefix + "00000000000000000000000000000000",
		"truncated secret": secret[:len(secret)-1],
		"wrong prefix":     "gk_" + strings.TrimPrefix(secret, models.MCPTokenPrefix),
		"display only":     secret[:11],
	}

	for name, candidate := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := models.AuthenticateMCPToken(db, candidate); err == nil {
				t.Errorf("expected %q to be rejected", candidate)
			}
		})
	}
}

func TestAuthenticateMCPTokenToleratesWhitespace(t *testing.T) {
	db := testutil.SetupTestDB(t)

	_, secret, err := models.CreateMCPToken(db, "CI", false)
	if err != nil {
		t.Fatalf("CreateMCPToken failed: %v", err)
	}

	if _, err := models.AuthenticateMCPToken(db, "  "+secret+"\n"); err != nil {
		t.Errorf("a padded token should still authenticate: %v", err)
	}
}

func TestListMCPTokensOmitsDigest(t *testing.T) {
	db := testutil.SetupTestDB(t)

	if _, _, err := models.CreateMCPToken(db, "One", false); err != nil {
		t.Fatalf("CreateMCPToken failed: %v", err)
	}
	if _, _, err := models.CreateMCPToken(db, "Two", true); err != nil {
		t.Fatalf("CreateMCPToken failed: %v", err)
	}

	tokens, err := models.ListMCPTokens(db)
	if err != nil {
		t.Fatalf("ListMCPTokens failed: %v", err)
	}
	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(tokens))
	}
	for _, tok := range tokens {
		if tok.TokenHash != "" {
			t.Error("ListMCPTokens must not load the digest")
		}
		if tok.Name == "" || tok.Display == "" {
			t.Errorf("listing dropped a displayable field: %+v", tok)
		}
	}
}

func TestTouchMCPToken(t *testing.T) {
	db := testutil.SetupTestDB(t)

	created, _, err := models.CreateMCPToken(db, "CI", false)
	if err != nil {
		t.Fatalf("CreateMCPToken failed: %v", err)
	}

	before := time.Now().Add(-time.Second)
	if err := models.TouchMCPToken(db, created.ID); err != nil {
		t.Fatalf("TouchMCPToken failed: %v", err)
	}

	tokens, err := models.ListMCPTokens(db)
	if err != nil {
		t.Fatalf("ListMCPTokens failed: %v", err)
	}
	if tokens[0].LastUsedAt == nil {
		t.Fatal("expected last_used_at to be set")
	}
	if tokens[0].LastUsedAt.Before(before) {
		t.Errorf("last_used_at looks stale: %v", tokens[0].LastUsedAt)
	}
}

func TestDeleteMCPTokenRevokesOnlyThatToken(t *testing.T) {
	db := testutil.SetupTestDB(t)

	keep, keepSecret, err := models.CreateMCPToken(db, "Keep", false)
	if err != nil {
		t.Fatalf("CreateMCPToken failed: %v", err)
	}
	revoke, revokeSecret, err := models.CreateMCPToken(db, "Revoke", false)
	if err != nil {
		t.Fatalf("CreateMCPToken failed: %v", err)
	}

	deleted, err := models.DeleteMCPToken(db, revoke.ID)
	if err != nil {
		t.Fatalf("DeleteMCPToken failed: %v", err)
	}
	if !deleted {
		t.Fatal("expected the token to be deleted")
	}

	if _, err := models.AuthenticateMCPToken(db, revokeSecret); err == nil {
		t.Error("a revoked token must stop authenticating")
	}
	// Revoking one person's token must not disturb anyone else's.
	got, err := models.AuthenticateMCPToken(db, keepSecret)
	if err != nil {
		t.Fatalf("the surviving token stopped working: %v", err)
	}
	if got.ID != keep.ID {
		t.Errorf("expected token %d, got %d", keep.ID, got.ID)
	}

	deleted, err = models.DeleteMCPToken(db, revoke.ID)
	if err != nil {
		t.Fatalf("DeleteMCPToken failed: %v", err)
	}
	if deleted {
		t.Error("deleting an already-revoked token should report no row")
	}
}

func TestGetMCPEnabledDefaultsToFalse(t *testing.T) {
	db := testutil.SetupTestDB(t)

	enabled, err := models.GetMCPEnabled(db)
	if err != nil {
		t.Fatalf("GetMCPEnabled failed: %v", err)
	}
	if enabled {
		t.Error("the MCP endpoint must be off until an operator turns it on")
	}

	if err := models.SetSetting(db, models.SettingMCPEnabled, "true"); err != nil {
		t.Fatalf("SetSetting failed: %v", err)
	}
	enabled, err = models.GetMCPEnabled(db)
	if err != nil {
		t.Fatalf("GetMCPEnabled failed: %v", err)
	}
	if !enabled {
		t.Error("expected the setting to read back as enabled")
	}
}
