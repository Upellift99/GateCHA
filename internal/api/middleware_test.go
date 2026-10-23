package api

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Upellift99/GateCHA/internal/auth"
	"github.com/Upellift99/GateCHA/internal/models"
	"github.com/Upellift99/GateCHA/internal/testutil"
)

func TestAuthenticateAPIKey_QueryParam(t *testing.T) {
	db := testutil.SetupTestDB(t)
	key, _ := models.CreateAPIKey(db, "Test", "", 0, 0, "")

	req := httptest.NewRequest("GET", "/api/v1/challenge?apiKey="+key.KeyID, nil)
	w := httptest.NewRecorder()

	result, ok := authenticateAPIKey(db, w, req)
	if !ok {
		t.Fatal("expected authentication to succeed")
	}

	ctxKey := GetAPIKeyFromContext(result)
	if ctxKey == nil {
		t.Fatal("expected API key in context")
	}
	if ctxKey.ID != key.ID {
		t.Errorf("expected key ID %d, got %d", key.ID, ctxKey.ID)
	}
}

func TestAuthenticateAPIKey_BearerHeader(t *testing.T) {
	db := testutil.SetupTestDB(t)
	key, _ := models.CreateAPIKey(db, "Test", "", 0, 0, "")

	req := httptest.NewRequest("GET", "/api/v1/challenge", nil)
	req.Header.Set("Authorization", "Bearer "+key.KeyID)
	w := httptest.NewRecorder()

	_, ok := authenticateAPIKey(db, w, req)
	if !ok {
		t.Fatal("expected authentication to succeed via Bearer header")
	}
}

func TestAuthenticateAPIKey_Missing(t *testing.T) {
	db := testutil.SetupTestDB(t)

	req := httptest.NewRequest("GET", "/api/v1/challenge", nil)
	w := httptest.NewRecorder()

	_, ok := authenticateAPIKey(db, w, req)
	if ok {
		t.Fatal("expected authentication to fail for missing key")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthenticateAPIKey_InvalidPrefix(t *testing.T) {
	db := testutil.SetupTestDB(t)

	req := httptest.NewRequest("GET", "/api/v1/challenge?apiKey=invalid_prefix", nil)
	w := httptest.NewRecorder()

	_, ok := authenticateAPIKey(db, w, req)
	if ok {
		t.Fatal("expected authentication to fail for invalid prefix")
	}
}

func TestAuthenticateAPIKey_NonexistentKey(t *testing.T) {
	db := testutil.SetupTestDB(t)

	req := httptest.NewRequest("GET", "/api/v1/challenge?apiKey=gk_nonexistent000000000000", nil)
	w := httptest.NewRecorder()

	_, ok := authenticateAPIKey(db, w, req)
	if ok {
		t.Fatal("expected authentication to fail for nonexistent key")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthenticateAPIKey_Disabled(t *testing.T) {
	db := testutil.SetupTestDB(t)
	key, _ := models.CreateAPIKey(db, "Test", "", 0, 0, "")
	models.UpdateAPIKey(db, key.ID, models.UpdateAPIKeyParams{
		Name: key.Name, Domain: key.Domain, MaxNumber: key.MaxNumber,
		ExpireSeconds: key.ExpireSeconds, Algorithm: key.Algorithm, Enabled: false,
	})

	req := httptest.NewRequest("GET", "/api/v1/challenge?apiKey="+key.KeyID, nil)
	w := httptest.NewRecorder()

	_, ok := authenticateAPIKey(db, w, req)
	if ok {
		t.Fatal("expected authentication to fail for disabled key")
	}
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestAuthenticateAPIKey_DomainAllowed(t *testing.T) {
	db := testutil.SetupTestDB(t)
	key, _ := models.CreateAPIKey(db, "Test", "allowed.com", 0, 0, "")

	req := httptest.NewRequest("GET", "/api/v1/challenge?apiKey="+key.KeyID, nil)
	req.Header.Set("Origin", "https://allowed.com")
	w := httptest.NewRecorder()

	_, ok := authenticateAPIKey(db, w, req)
	if !ok {
		t.Fatal("expected authentication to succeed for matching domain")
	}
}

func TestAuthenticateAPIKey_DomainBlocked(t *testing.T) {
	db := testutil.SetupTestDB(t)
	key, _ := models.CreateAPIKey(db, "Test", "allowed.com", 0, 0, "")

	req := httptest.NewRequest("GET", "/api/v1/challenge?apiKey="+key.KeyID, nil)
	req.Header.Set("Origin", "https://evil.com")
	w := httptest.NewRecorder()

	_, ok := authenticateAPIKey(db, w, req)
	if ok {
		t.Fatal("expected authentication to fail for non-matching domain")
	}
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestAuthenticateAPIKey_NoDomainNoOrigin(t *testing.T) {
	db := testutil.SetupTestDB(t)
	key, _ := models.CreateAPIKey(db, "Test", "restricted.com", 0, 0, "")

	// No Origin header = no domain check
	req := httptest.NewRequest("GET", "/api/v1/challenge?apiKey="+key.KeyID, nil)
	w := httptest.NewRecorder()

	_, ok := authenticateAPIKey(db, w, req)
	if !ok {
		t.Fatal("expected authentication to succeed when no Origin header")
	}
}

func TestAuthenticateAdmin_Valid(t *testing.T) {
	secret := "test-secret"
	token, _, _ := auth.GenerateJWT("admin", secret)

	req := httptest.NewRequest("GET", "/api/admin/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	ok := authenticateAdmin(secret, w, req)
	if !ok {
		t.Fatal("expected admin authentication to succeed")
	}
}

func TestAuthenticateAdmin_NoHeader(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/admin/me", nil)
	w := httptest.NewRecorder()

	ok := authenticateAdmin("secret", w, req)
	if ok {
		t.Fatal("expected admin authentication to fail without header")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthenticateAdmin_InvalidToken(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/admin/me", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	ok := authenticateAdmin("secret", w, req)
	if ok {
		t.Fatal("expected admin authentication to fail with invalid token")
	}
}

func TestCORSMiddleware_AllowAll(t *testing.T) {
	handler := CORSMiddleware(true)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("expected CORS origin *")
	}
	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("expected CORS methods header")
	}
}

func TestCORSMiddleware_SpecificOrigin(t *testing.T) {
	handler := CORSMiddleware(false)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Errorf("expected origin echo, got %q", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSMiddleware_Preflight(t *testing.T) {
	handler := CORSMiddleware(true)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("OPTIONS", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204 for preflight, got %d", w.Code)
	}
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	handler := SecurityHeadersMiddleware(false)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	checks := map[string]string{
		"X-Frame-Options":        "DENY",
		"X-Content-Type-Options": "nosniff",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
	}
	for header, want := range checks {
		if got := w.Header().Get(header); got != want {
			t.Errorf("%s = %q, want %q", header, got, want)
		}
	}
	csp := w.Header().Get("Content-Security-Policy")
	if csp == "" {
		t.Error("expected a Content-Security-Policy header")
	}
	// The ALTCHA widget runs its proof-of-work in a blob: Worker; without this
	// directive the login captcha would be blocked by the CSP.
	if !strings.Contains(csp, "worker-src 'self' blob:") {
		t.Errorf("CSP must allow blob: workers for the ALTCHA widget, got %q", csp)
	}
	if w.Header().Get("Strict-Transport-Security") != "" {
		t.Error("HSTS should be absent when disabled")
	}
}

func TestSecurityHeadersMiddleware_HSTS(t *testing.T) {
	handler := SecurityHeadersMiddleware(true)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Header().Get("Strict-Transport-Security") == "" {
		t.Error("expected HSTS header when enabled")
	}
}

func TestMaxBodyBytesMiddleware_Rejects(t *testing.T) {
	// Handler reads the whole body; the limiter should make oversized reads fail.
	handler := MaxBodyBytesMiddleware(10)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.ReadAll(r.Body); err != nil {
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/", strings.NewReader("this body is definitely longer than ten bytes"))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected oversized body to be rejected, got %d", w.Code)
	}
}

func TestMaxBodyBytesMiddleware_AllowsSmall(t *testing.T) {
	handler := MaxBodyBytesMiddleware(1024)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.ReadAll(r.Body); err != nil {
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/", strings.NewReader("small"))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected small body to pass, got %d", w.Code)
	}
}

func TestMatchDomain(t *testing.T) {
	tests := []struct {
		url, domain string
		expected    bool
	}{
		{"https://example.com/path", "example.com", true},
		{"http://example.com:8080/path", "example.com", true},
		{"https://other.com", "example.com", false},
		{"https://EXAMPLE.COM", "example.com", true},
		{"example.com/path", "example.com", true},
		// Multiple domains (one per line, commas/whitespace tolerated).
		{"https://app.com", "example.com\napp.com", true},
		{"https://other.com", "example.com\napp.com", false},
		{"https://app.com", "example.com, app.com", true},
		// Wildcard: matches subdomains and the bare domain.
		{"https://api.example.com", "*.example.com", true},
		{"https://deep.api.example.com", "*.example.com", true},
		{"https://example.com", "*.example.com", true},
		{"https://EXAMPLE.COM", "*.example.com", true},
		{"https://notexample.com", "*.example.com", false},
		{"https://example.com.evil.com", "*.example.com", false},
		// Empty config never matches a present origin (caller gates on this).
		{"https://example.com", "", false},
	}

	for _, tt := range tests {
		result := matchDomain(tt.url, tt.domain)
		if result != tt.expected {
			t.Errorf("matchDomain(%q, %q) = %v, want %v", tt.url, tt.domain, result, tt.expected)
		}
	}
}

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSON(w, http.StatusOK, map[string]string{"hello": "world"})

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected application/json, got %s", w.Header().Get("Content-Type"))
	}

	var body map[string]string
	json.NewDecoder(w.Body).Decode(&body)
	if body["hello"] != "world" {
		t.Errorf("expected hello=world, got %v", body)
	}
}

func requestLogLine(t *testing.T, logClientIP bool) string {
	t.Helper()
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	h := RequestLogMiddleware(logger, logClientIP)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("ok"))
	}))
	req := httptest.NewRequest("GET", "/api/v1/challenge?apiKey=gk_secretish", nil)
	req.RemoteAddr = "203.0.113.7:51234"
	h.ServeHTTP(httptest.NewRecorder(), req)
	return buf.String()
}

func TestRequestLog_OmitsClientIPAndQuery(t *testing.T) {
	line := requestLogLine(t, false)
	for _, want := range []string{"method=GET", "path=/api/v1/challenge", "status=418", "bytes=2", "duration="} {
		if !strings.Contains(line, want) {
			t.Errorf("log line %q is missing %q", line, want)
		}
	}
	// The visitor's IP and the ?apiKey= query must not reach the log.
	for _, leak := range []string{"203.0.113.7", "client_ip", "apiKey", "gk_secretish"} {
		if strings.Contains(line, leak) {
			t.Errorf("log line %q leaks %q", line, leak)
		}
	}
}

func TestRequestLog_IncludesClientIPWhenEnabled(t *testing.T) {
	line := requestLogLine(t, true)
	if !strings.Contains(line, "client_ip=203.0.113.7") {
		t.Errorf("expected client_ip in %q", line)
	}
	if strings.Contains(line, "gk_secretish") {
		t.Errorf("the query string must stay out of the log even with client IPs: %q", line)
	}
}

func TestRequestLog_DefaultsStatusTo200(t *testing.T) {
	var buf bytes.Buffer
	h := RequestLogMiddleware(slog.New(slog.NewTextHandler(&buf, nil)), false)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/healthz", nil))
	if !strings.Contains(buf.String(), "status=200") {
		t.Errorf("a handler that writes nothing is a 200, got %q", buf.String())
	}
}

func TestRequestLog_StripsNewlinesFromPath(t *testing.T) {
	var buf bytes.Buffer
	h := RequestLogMiddleware(slog.New(slog.NewTextHandler(&buf, nil)), false)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	req := httptest.NewRequest("GET", "/x", nil)
	req.URL.Path = "/x\nlevel=ERROR msg=forged"
	h.ServeHTTP(httptest.NewRecorder(), req)
	if lines := strings.Count(strings.TrimRight(buf.String(), "\n"), "\n"); lines != 0 {
		t.Fatalf("expected one log line, got %d extra in %q", lines, buf.String())
	}
	if strings.Contains(buf.String(), "\\n") || !strings.Contains(buf.String(), "/xlevel=ERROR msg=forged") {
		t.Errorf("expected the newline stripped from the path, got %q", buf.String())
	}
}
