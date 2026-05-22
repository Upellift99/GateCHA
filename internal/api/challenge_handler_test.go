package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Upellift99/GateCHA/internal/models"
	"github.com/Upellift99/GateCHA/internal/testutil"
)

// challengeRequestWithKey builds a request carrying key in the context, the way
// APIKeyMiddleware would, so the handler can be exercised in isolation.
func challengeRequestWithKey(key *models.APIKey) *http.Request {
	req := httptest.NewRequest("GET", "/api/v1/challenge", nil)
	if key == nil {
		return req
	}
	ctx := context.WithValue(req.Context(), apiKeyContextKey, key)
	return req.WithContext(ctx)
}

// Defensive branch: no API key in context (middleware normally guards this, so
// the full-router path never reaches it) -> 401.
func TestChallengeHandler_NilKey(t *testing.T) {
	db := testutil.SetupTestDB(t)
	h := &ChallengeHandler{DB: db}

	w := httptest.NewRecorder()
	h.ServeHTTP(w, challengeRequestWithKey(nil))

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// Challenge generation fails for an unsupported algorithm -> 500.
func TestChallengeHandler_GenerateError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	h := &ChallengeHandler{DB: db}

	key := &models.APIKey{
		HMACSecret:    "secret",
		MaxNumber:     100,
		Algorithm:     "NOT-A-REAL-ALGO",
		ExpireSeconds: 60,
	}

	w := httptest.NewRecorder()
	h.ServeHTTP(w, challengeRequestWithKey(key))

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 for unsupported algorithm, got %d: %s", w.Code, w.Body.String())
	}
}

// Incrementing the counter fails (daily_stats table missing) but the challenge
// is still served: the error is logged and the handler returns 200.
func TestChallengeHandler_IncrementError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	key, err := models.CreateAPIKey(db, "Test", "", 100, 60, "SHA-256")
	if err != nil {
		t.Fatalf("CreateAPIKey failed: %v", err)
	}

	// Drop daily_stats so IncrementChallengesIssued errors.
	if err := db.Migrator().DropTable(&models.DailyStat{}); err != nil {
		t.Fatalf("failed to drop daily_stats: %v", err)
	}

	h := &ChallengeHandler{DB: db}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, challengeRequestWithKey(key))

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 (increment error is non-fatal), got %d", w.Code)
	}
}
