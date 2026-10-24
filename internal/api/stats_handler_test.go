package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Upellift99/GateCHA/internal/models"
	"gorm.io/gorm"
)

func issueStatsToken(t *testing.T, router http.Handler, keyID int64) string {
	t.Helper()
	req := httptest.NewRequest("POST", "/api/admin/keys/"+strconv.FormatInt(keyID, 10)+"/stats-token", nil)
	req.Header.Set("Authorization", bearerPrefix+getAdminToken(t))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("issue stats token: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if !strings.HasPrefix(resp["stats_token"], models.StatsTokenPrefix) {
		t.Fatalf("expected a gs_ token, got %q", resp["stats_token"])
	}
	return resp["stats_token"]
}

func getStats(router http.Handler, auth, query string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", "/api/v1/stats"+query, nil)
	if auth != "" {
		req.Header.Set("Authorization", bearerPrefix+auth)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func seedTodayStats(t *testing.T, db *gorm.DB, keyID int64, ok, fail int) {
	t.Helper()
	for range ok {
		if err := models.IncrementVerificationsOK(db, keyID); err != nil {
			t.Fatal(err)
		}
		if err := models.IncrementCountryVerification(db, keyID, "FR", true); err != nil {
			t.Fatal(err)
		}
	}
	for range fail {
		if err := models.IncrementVerificationsFail(db, keyID); err != nil {
			t.Fatal(err)
		}
	}
}

type statsResp struct {
	KeyID   string `json:"key_id"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
	models.KeyStatsReport
}

func decodeStats(t *testing.T, w *httptest.ResponseRecorder) statsResp {
	t.Helper()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp statsResp
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	return resp
}

func TestKeyStats_ReadsOnlyItsOwnKey(t *testing.T) {
	router, db := setupTestRouter(t)
	mine, _ := models.CreateAPIKey(db, "Mine", "", 0, 0, "")
	other, _ := models.CreateAPIKey(db, "Other", "", 0, 0, "")
	seedTodayStats(t, db, mine.ID, 3, 1)
	seedTodayStats(t, db, other.ID, 50, 50)

	token := issueStatsToken(t, router, mine.ID)
	resp := decodeStats(t, getStats(router, token, ""))

	if resp.KeyID != mine.KeyID || resp.Name != "Mine" || !resp.Enabled {
		t.Errorf("wrong key in response: %+v", resp)
	}
	if resp.Totals.VerificationsOK != 3 || resp.Totals.VerificationsFail != 1 {
		t.Errorf("totals leak or miss: %+v", resp.Totals)
	}
	if len(resp.Countries) != 1 || resp.Countries[0].Country != "FR" || resp.Countries[0].VerificationsOK != 3 {
		t.Errorf("countries: %+v", resp.Countries)
	}
	if resp.Days != 30 || len(resp.Daily) != 30 {
		t.Fatalf("expected 30 zero-filled days, got days=%d len=%d", resp.Days, len(resp.Daily))
	}
	today := time.Now().UTC().Format("2006-01-02")
	if last := resp.Daily[len(resp.Daily)-1]; last.Date != today || last.VerificationsOK != 3 {
		t.Errorf("last day should be today with the counts, got %+v", last)
	}
	if strings.Contains(getStats(router, token, "").Body.String(), mine.HMACSecret) {
		t.Error("stats response must not carry the HMAC secret")
	}
}

func TestKeyStats_Days(t *testing.T) {
	router, db := setupTestRouter(t)
	key, _ := models.CreateAPIKey(db, "k", "", 0, 0, "")
	token := issueStatsToken(t, router, key.ID)

	if resp := decodeStats(t, getStats(router, token, "?days=7")); len(resp.Daily) != 7 {
		t.Errorf("days=7: got %d days", len(resp.Daily))
	}
	for _, q := range []string{"?days=0", "?days=91", "?days=abc"} {
		if w := getStats(router, token, q); w.Code != http.StatusBadRequest {
			t.Errorf("%s: expected 400, got %d", q, w.Code)
		}
	}
}

func TestKeyStats_RejectsBadCredentials(t *testing.T) {
	router, db := setupTestRouter(t)
	key, _ := models.CreateAPIKey(db, "k", "", 0, 0, "")
	issueStatsToken(t, router, key.ID)

	cases := map[string]string{
		"no header":         "",
		"public key ID":     key.KeyID,
		"HMAC secret":       key.HMACSecret,
		"unknown gs_ token": models.StatsTokenPrefix + strings.Repeat("0", 32),
	}
	for name, auth := range cases {
		if w := getStats(router, auth, ""); w.Code != http.StatusUnauthorized {
			t.Errorf("%s: expected 401, got %d", name, w.Code)
		}
	}
	// The key ID in the query string, the way /challenge takes it, opens nothing.
	if w := getStats(router, "", "?apiKey="+key.KeyID); w.Code != http.StatusUnauthorized {
		t.Errorf("apiKey query: expected 401, got %d", w.Code)
	}
}

func TestKeyStats_ReissueAndRevoke(t *testing.T) {
	router, db := setupTestRouter(t)
	key, _ := models.CreateAPIKey(db, "k", "", 0, 0, "")
	first := issueStatsToken(t, router, key.ID)
	second := issueStatsToken(t, router, key.ID)

	if w := getStats(router, first, ""); w.Code != http.StatusUnauthorized {
		t.Errorf("replaced token: expected 401, got %d", w.Code)
	}
	if w := getStats(router, second, ""); w.Code != http.StatusOK {
		t.Errorf("current token: expected 200, got %d", w.Code)
	}

	req := httptest.NewRequest("DELETE", "/api/admin/keys/"+strconv.FormatInt(key.ID, 10)+"/stats-token", nil)
	req.Header.Set("Authorization", bearerPrefix+getAdminToken(t))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("revoke: expected 204, got %d", w.Code)
	}
	if w := getStats(router, second, ""); w.Code != http.StatusUnauthorized {
		t.Errorf("revoked token: expected 401, got %d", w.Code)
	}
	got, _ := models.GetAPIKeyByID(db, key.ID)
	if got.StatsTokenHash != nil || got.StatsTokenDisplay != "" {
		t.Errorf("revoke left %v / %q", got.StatsTokenHash, got.StatsTokenDisplay)
	}

	// Revoking twice is not an error: the key exists, it just has no token.
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Errorf("second revoke: expected 204, got %d", w.Code)
	}
}

func TestKeyStats_DisabledKeyStillReadable(t *testing.T) {
	router, db := setupTestRouter(t)
	key, _ := models.CreateAPIKey(db, "k", "", 0, 0, "")
	token := issueStatsToken(t, router, key.ID)
	models.SetAPIKeyEnabled(db, key.ID, false)

	if resp := decodeStats(t, getStats(router, token, "")); resp.Enabled {
		t.Error("expected enabled=false for a disabled key")
	}
}

func TestStatsToken_AdminErrors(t *testing.T) {
	router, _ := setupTestRouter(t)
	for _, method := range []string{"POST", "DELETE"} {
		for path, want := range map[string]int{
			"/api/admin/keys/999/stats-token": http.StatusNotFound,
			"/api/admin/keys/abc/stats-token": http.StatusBadRequest,
		} {
			req := httptest.NewRequest(method, path, nil)
			req.Header.Set("Authorization", bearerPrefix+getAdminToken(t))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != want {
				t.Errorf("%s %s: expected %d, got %d", method, path, want, w.Code)
			}
		}
		// Without admin auth the route is closed.
		req := httptest.NewRequest(method, "/api/admin/keys/1/stats-token", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s without auth: expected 401, got %d", method, w.Code)
		}
	}
}

func TestStatsToken_ShownInKeyButNeverItsHash(t *testing.T) {
	router, db := setupTestRouter(t)
	key, _ := models.CreateAPIKey(db, "k", "", 0, 0, "")
	token := issueStatsToken(t, router, key.ID)
	hash := models.HashMCPToken(token)

	for _, path := range []string{"/api/admin/keys", "/api/admin/keys/" + strconv.FormatInt(key.ID, 10)} {
		req := httptest.NewRequest("GET", path, nil)
		req.Header.Set("Authorization", bearerPrefix+getAdminToken(t))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		body := w.Body.String()
		if strings.Contains(body, hash) || strings.Contains(body, token) {
			t.Errorf("%s leaks the stats token or its hash", path)
		}
		if !strings.Contains(body, token[:11]) {
			t.Errorf("%s should show the token's display prefix %q", path, token[:11])
		}
	}
}
