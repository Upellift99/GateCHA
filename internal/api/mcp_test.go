package api

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Upellift99/GateCHA/internal/models"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gorm.io/gorm"
)

// startMCPServer serves the real router over HTTP so the SDK client speaks the
// actual streamable transport rather than a hand-rolled approximation of it.
func startMCPServer(t *testing.T, enabled bool) (*httptest.Server, *gorm.DB) {
	t.Helper()
	router, db := setupTestRouter(t)
	if enabled {
		if err := models.SetSetting(db, models.SettingMCPEnabled, "true"); err != nil {
			t.Fatalf("failed to enable mcp: %v", err)
		}
	}
	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)
	return srv, db
}

// bearerRoundTripper attaches a token to every request the SDK client makes.
type bearerRoundTripper struct {
	token string
	base  http.RoundTripper
}

func (b bearerRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	clone := r.Clone(r.Context())
	if b.token != "" {
		clone.Header.Set("Authorization", "Bearer "+b.token)
	}
	base := b.base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(clone)
}

func connectMCP(t *testing.T, srv *httptest.Server, token string) (*mcp.ClientSession, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil)
	transport := &mcp.StreamableClientTransport{
		Endpoint:             srv.URL + "/mcp",
		HTTPClient:           &http.Client{Transport: bearerRoundTripper{token: token}},
		MaxRetries:           -1,
		DisableStandaloneSSE: true,
	}
	session, err := client.Connect(ctx, transport, nil)
	if err == nil {
		t.Cleanup(func() { session.Close() })
	}
	return session, err
}

func TestMCPHandshakeSucceedsWithAToken(t *testing.T) {
	srv, db := startMCPServer(t, true)

	_, secret, err := models.CreateMCPToken(db, "Cursor", false)
	if err != nil {
		t.Fatalf("CreateMCPToken failed: %v", err)
	}

	session, err := connectMCP(t, srv, secret)
	if err != nil {
		t.Fatalf("expected the handshake to succeed: %v", err)
	}

	// The session initialised, which means the streamable transport negotiated a
	// protocol version end to end.
	if got := session.InitializeResult().ServerInfo.Name; got != mcpServerName {
		t.Errorf("expected server name %q, got %q", mcpServerName, got)
	}
}

func TestMCPRejectsBadCredentials(t *testing.T) {
	srv, db := startMCPServer(t, true)

	_, secret, err := models.CreateMCPToken(db, "Cursor", false)
	if err != nil {
		t.Fatalf("CreateMCPToken failed: %v", err)
	}

	cases := map[string]string{
		"no token":      "",
		"unknown token": models.MCPTokenPrefix + "00000000000000000000000000000000",
		"truncated":     secret[:len(secret)-1],
		"admin jwt":     getAdminToken(t),
	}

	for name, candidate := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := connectMCP(t, srv, candidate); err == nil {
				t.Error("expected the handshake to be refused")
			}
		})
	}
}

func TestMCPRevokedTokenStopsWorking(t *testing.T) {
	srv, db := startMCPServer(t, true)

	token, secret, err := models.CreateMCPToken(db, "Cursor", false)
	if err != nil {
		t.Fatalf("CreateMCPToken failed: %v", err)
	}
	if _, err := connectMCP(t, srv, secret); err != nil {
		t.Fatalf("expected the handshake to succeed first: %v", err)
	}

	if _, err := models.DeleteMCPToken(db, token.ID); err != nil {
		t.Fatalf("DeleteMCPToken failed: %v", err)
	}

	if _, err := connectMCP(t, srv, secret); err == nil {
		t.Error("a revoked token must stop working immediately")
	}
}

func TestMCPDisabledByDefault(t *testing.T) {
	srv, db := startMCPServer(t, false)

	_, secret, err := models.CreateMCPToken(db, "Cursor", false)
	if err != nil {
		t.Fatalf("CreateMCPToken failed: %v", err)
	}

	// A valid token is not enough while the endpoint is switched off.
	if _, err := connectMCP(t, srv, secret); err == nil {
		t.Error("expected the handshake to fail while mcp is disabled")
	}

	req, _ := http.NewRequest("POST", srv.URL+"/mcp", bytes.NewReader([]byte("{}")))
	req.Header.Set("Authorization", "Bearer "+secret)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 while disabled, got %d", resp.StatusCode)
	}
}

// The toggle is checked before the credential, so a disabled endpoint cannot be
// used as an oracle to test whether a stolen token is still valid.
func TestMCPToggleIsCheckedBeforeTheToken(t *testing.T) {
	srv, _ := startMCPServer(t, false)

	req, _ := http.NewRequest("POST", srv.URL+"/mcp", bytes.NewReader([]byte("{}")))
	req.Header.Set("Authorization", "Bearer "+models.MCPTokenPrefix+"deadbeefdeadbeefdeadbeefdeadbeef")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 rather than 401, got %d", resp.StatusCode)
	}
}

func TestMCPUnauthorizedAdvertisesBearer(t *testing.T) {
	srv, _ := startMCPServer(t, true)

	req, _ := http.NewRequest("POST", srv.URL+"/mcp", bytes.NewReader([]byte("{}")))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get("WWW-Authenticate"); got == "" {
		t.Error("a 401 should tell the client how to authenticate")
	}
}

// A token in the query string would land in proxy logs and browser history.
func TestMCPIgnoresTokenInQueryString(t *testing.T) {
	srv, db := startMCPServer(t, true)

	_, secret, err := models.CreateMCPToken(db, "Cursor", false)
	if err != nil {
		t.Fatalf("CreateMCPToken failed: %v", err)
	}

	req, _ := http.NewRequest("POST", srv.URL+"/mcp?token="+secret, bytes.NewReader([]byte("{}")))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("a query-string token must not authenticate, got %d", resp.StatusCode)
	}
}

func TestMCPRecordsTokenUsage(t *testing.T) {
	srv, db := startMCPServer(t, true)

	created, secret, err := models.CreateMCPToken(db, "Cursor", false)
	if err != nil {
		t.Fatalf("CreateMCPToken failed: %v", err)
	}
	if created.LastUsedAt != nil {
		t.Fatal("a fresh token should have no last-used timestamp")
	}

	if _, err := connectMCP(t, srv, secret); err != nil {
		t.Fatalf("handshake failed: %v", err)
	}

	tokens, err := models.ListMCPTokens(db)
	if err != nil {
		t.Fatalf("ListMCPTokens failed: %v", err)
	}
	if tokens[0].LastUsedAt == nil {
		t.Error("connecting should record the token as used")
	}
}

// The tools land in a follow-up, so there is nothing yet to assert about a
// read-only token's smaller tool set. What can be pinned down now is the
// plumbing that split depends on: the authenticated token, and its ReadOnly
// flag, must reach the code that builds the server.
func TestMCPAuthPassesTheTokenDownstream(t *testing.T) {
	router, db := setupTestRouter(t)
	if err := models.SetSetting(db, models.SettingMCPEnabled, "true"); err != nil {
		t.Fatalf("failed to enable mcp: %v", err)
	}
	_ = router

	for _, readOnly := range []bool{false, true} {
		created, secret, err := models.CreateMCPToken(db, "probe", readOnly)
		if err != nil {
			t.Fatalf("CreateMCPToken failed: %v", err)
		}

		var seen *models.MCPToken
		handler := MCPAuthMiddleware(db)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			seen = MCPTokenFromContext(r.Context())
		}))

		req := httptest.NewRequest("POST", "/mcp", nil)
		req.Header.Set("Authorization", "Bearer "+secret)
		handler.ServeHTTP(httptest.NewRecorder(), req)

		if seen == nil {
			t.Fatalf("read_only=%v: the token did not reach the handler", readOnly)
		}
		if seen.ID != created.ID {
			t.Errorf("read_only=%v: expected token %d, got %d", readOnly, created.ID, seen.ID)
		}
		if seen.ReadOnly != readOnly {
			t.Errorf("expected ReadOnly=%v to survive, got %v", readOnly, seen.ReadOnly)
		}
	}
}

func TestMCPContextIsEmptyWithoutAuth(t *testing.T) {
	if got := MCPTokenFromContext(context.Background()); got != nil {
		t.Errorf("expected no token on a bare context, got %+v", got)
	}
}
