package api

import (
	"context"
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/Upellift99/GateCHA/internal/models"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gorm.io/gorm"
	"net/http/httptest"
)

// connectWithTools brings up the endpoint and returns a live session for a
// token with the requested access level.
func connectWithTools(t *testing.T, readOnly bool) (*mcp.ClientSession, *gorm.DB, *httptest.Server) {
	t.Helper()
	srv, db := startMCPServer(t, true)
	_, secret, err := models.CreateMCPToken(db, "tools", readOnly)
	if err != nil {
		t.Fatalf("CreateMCPToken failed: %v", err)
	}
	session, err := connectMCP(t, srv, secret)
	if err != nil {
		t.Fatalf("handshake failed: %v", err)
	}
	return session, db, srv
}

func toolNames(t *testing.T, session *mcp.ClientSession) []string {
	t.Helper()
	res, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}
	names := make([]string, 0, len(res.Tools))
	for _, tool := range res.Tools {
		names = append(names, tool.Name)
	}
	slices.Sort(names)
	return names
}

// callTool runs a tool and decodes its structured result.
func callTool(t *testing.T, session *mcp.ClientSession, name string, args map[string]any, out any) *mcp.CallToolResult {
	t.Helper()
	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("CallTool(%s) transport error: %v", name, err)
	}
	if res.IsError {
		return res
	}
	if out != nil {
		raw, err := json.Marshal(res.StructuredContent)
		if err != nil {
			t.Fatalf("failed to re-marshal %s output: %v", name, err)
		}
		if err := json.Unmarshal(raw, out); err != nil {
			t.Fatalf("failed to decode %s output: %v", name, err)
		}
	}
	return res
}

func mustCall(t *testing.T, session *mcp.ClientSession, name string, args map[string]any, out any) {
	t.Helper()
	if res := callTool(t, session, name, args, out); res.IsError {
		t.Fatalf("%s returned an error: %s", name, resultText(res))
	}
}

func resultText(res *mcp.CallToolResult) string {
	var b strings.Builder
	for _, c := range res.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			b.WriteString(tc.Text)
		}
	}
	return b.String()
}

func TestMCPToolSurface(t *testing.T) {
	session, _, _ := connectWithTools(t, false)

	want := []string{"create_key", "disable_key", "enable_key", "get_key", "list_keys", "update_key"}
	got := toolNames(t, session)
	if !slices.Equal(got, want) {
		t.Errorf("unexpected tool set:\n got %v\nwant %v", got, want)
	}

	// Deleting a key breaks a live site with no undo, so it is not offered.
	if slices.Contains(got, "delete_key") {
		t.Error("delete_key must not be exposed over MCP")
	}
}

func TestMCPReadOnlyTokenOnlyGetsReadTools(t *testing.T) {
	session, _, _ := connectWithTools(t, true)

	got := toolNames(t, session)
	want := []string{"get_key", "list_keys"}
	if !slices.Equal(got, want) {
		t.Errorf("a read-only token should see only the read tools:\n got %v\nwant %v", got, want)
	}
}

// Least privilege has to hold at call time, not just in the advertised list: a
// client that names a withheld tool anyway must be refused.
func TestMCPReadOnlyTokenCannotCallWriteTools(t *testing.T) {
	session, db, _ := connectWithTools(t, true)

	key, err := models.CreateAPIKey(db, "Site", "site.com", 0, 0, "")
	if err != nil {
		t.Fatalf("CreateAPIKey failed: %v", err)
	}

	for _, tc := range []struct {
		name string
		args map[string]any
	}{
		{"create_key", map[string]any{"name": "Sneaky"}},
		{"update_key", map[string]any{"id": key.ID, "name": "Renamed"}},
		{"enable_key", map[string]any{"id": key.ID}},
		{"disable_key", map[string]any{"id": key.ID}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: tc.name, Arguments: tc.args})
			if err == nil && !res.IsError {
				t.Fatalf("%s must be refused for a read-only token", tc.name)
			}
		})
	}

	// Nothing was written despite the attempts.
	after, err := models.GetAPIKeyByID(db, key.ID)
	if err != nil {
		t.Fatalf("GetAPIKeyByID failed: %v", err)
	}
	if after.Name != "Site" || !after.Enabled {
		t.Errorf("a read-only token changed a key: %+v", after)
	}
	keys, err := models.ListAPIKeys(db)
	if err != nil {
		t.Fatalf("ListAPIKeys failed: %v", err)
	}
	if len(keys) != 1 {
		t.Errorf("expected the key count to be unchanged, got %d", len(keys))
	}
}

func TestMCPCreateKeyReturnsTheSecretOnce(t *testing.T) {
	session, db, _ := connectWithTools(t, false)

	var created createKeyOutput
	mustCall(t, session, "create_key", map[string]any{"name": "Blog", "domain": "blog.example.com"}, &created)

	if created.HMACSecret == "" {
		t.Fatal("create_key must return the signing secret")
	}
	if created.Key.KeyID == "" || !strings.HasPrefix(created.Key.KeyID, "gk_") {
		t.Errorf("unexpected key_id %q", created.Key.KeyID)
	}
	if created.Key.Name != "Blog" || created.Key.Domain != "blog.example.com" {
		t.Errorf("unexpected key: %+v", created.Key)
	}
	// CreateAPIKey owns the defaults; the tool must not invent its own.
	if created.Key.MaxNumber != 100000 || created.Key.ExpireSeconds != 300 || created.Key.Algorithm != "SHA-256" {
		t.Errorf("defaults were not applied: %+v", created.Key)
	}
	if !created.Key.Enabled {
		t.Error("a new key should be enabled")
	}

	// The secret is unreachable from every other tool.
	var fetched mcpKeyView
	mustCall(t, session, "get_key", map[string]any{"id": created.Key.ID}, &fetched)

	res, _ := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "get_key", Arguments: map[string]any{"id": created.Key.ID},
	})
	if body := resultText(res); strings.Contains(body, created.HMACSecret) || strings.Contains(body, "hmac_secret") {
		t.Error("get_key leaked the HMAC secret")
	}

	listRes, _ := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "list_keys"})
	if body := resultText(listRes); strings.Contains(body, created.HMACSecret) || strings.Contains(body, "hmac_secret") {
		t.Error("list_keys leaked the HMAC secret")
	}

	// The secret really is the stored one.
	stored, err := models.GetAPIKeyByID(db, created.Key.ID)
	if err != nil {
		t.Fatalf("GetAPIKeyByID failed: %v", err)
	}
	if stored.HMACSecret != created.HMACSecret {
		t.Error("the returned secret does not match the stored one")
	}
}

func TestMCPCreateKeyRequiresAName(t *testing.T) {
	session, db, _ := connectWithTools(t, false)

	for _, name := range []string{"", "   "} {
		res := callTool(t, session, "create_key", map[string]any{"name": name}, nil)
		if !res.IsError {
			t.Errorf("name %q should be refused", name)
		}
	}

	keys, err := models.ListAPIKeys(db)
	if err != nil {
		t.Fatalf("ListAPIKeys failed: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("a refused create must not persist a key, got %d", len(keys))
	}
}

func TestMCPListKeysFilters(t *testing.T) {
	session, db, _ := connectWithTools(t, false)

	alpha, _ := models.CreateAPIKey(db, "Alpha", "alpha.com", 0, 0, "")
	models.CreateAPIKey(db, "Bravo", "bravo.com", 0, 0, "")
	models.CreateAPIKey(db, "Charlie", "", 0, 0, "")

	var all listKeysOutput
	mustCall(t, session, "list_keys", nil, &all)
	if all.Count != 3 || len(all.Keys) != 3 {
		t.Fatalf("expected 3 keys, got %d", all.Count)
	}

	cases := map[string]int{
		"alpha":       1,
		"ALPHA":       1,
		"  alpha  ":   1,
		"bravo.com":   1,
		"o":           2, // Bravo by name, Alpha by domain (".com"); Charlie has neither
		"nothinghere": 0,
	}
	for query, want := range cases {
		var out listKeysOutput
		mustCall(t, session, "list_keys", map[string]any{"search": query}, &out)
		if out.Count != want {
			t.Errorf("search %q: expected %d, got %d", query, want, out.Count)
		}
	}

	// Filtering also matches on the key ID, which is what an operator has in hand
	// when debugging a live site.
	var byID listKeysOutput
	mustCall(t, session, "list_keys", map[string]any{"search": alpha.KeyID}, &byID)
	if byID.Count != 1 || byID.Keys[0].ID != alpha.ID {
		t.Errorf("searching by key ID failed: %+v", byID)
	}
}

func TestMCPGetKeyRejectsUnknownID(t *testing.T) {
	session, _, _ := connectWithTools(t, false)

	res := callTool(t, session, "get_key", map[string]any{"id": 99999}, nil)
	if !res.IsError {
		t.Error("expected an error for an unknown key")
	}
}

func TestMCPUpdateKeyLeavesOmittedFieldsAlone(t *testing.T) {
	session, db, _ := connectWithTools(t, false)

	key, err := models.CreateAPIKey(db, "Original", "original.com", 20000, 120, "SHA-512")
	if err != nil {
		t.Fatalf("CreateAPIKey failed: %v", err)
	}

	var updated mcpKeyView
	mustCall(t, session, "update_key", map[string]any{"id": key.ID, "name": "Renamed"}, &updated)

	if updated.Name != "Renamed" {
		t.Errorf("expected the name to change, got %q", updated.Name)
	}
	// This is the trap: UpdateAPIKey writes every column, so an omitted field
	// must be read back from the existing row rather than sent as a zero value.
	if updated.Domain != "original.com" {
		t.Errorf("omitting domain blanked it: %q", updated.Domain)
	}
	if updated.MaxNumber != 20000 {
		t.Errorf("omitting max_number reset it: %d", updated.MaxNumber)
	}
	if updated.ExpireSeconds != 120 {
		t.Errorf("omitting expire_seconds reset it: %d", updated.ExpireSeconds)
	}
	if updated.Algorithm != "SHA-512" {
		t.Errorf("omitting algorithm reset it: %q", updated.Algorithm)
	}
	// Enabled is the dangerous one: silently writing false here would take a live
	// site's CAPTCHA down on an unrelated rename.
	if !updated.Enabled {
		t.Error("a partial update disabled the key")
	}

	// And the other direction: updating a disabled key must not revive it.
	if _, _, err := func() (*mcp.CallToolResult, mcpKeyView, error) {
		return mcpSetEnabled(db, key.ID, false)
	}(); err != nil {
		t.Fatalf("failed to disable the key: %v", err)
	}

	var again mcpKeyView
	mustCall(t, session, "update_key", map[string]any{"id": key.ID, "name": "Renamed twice"}, &again)
	if again.Enabled {
		t.Error("a partial update re-enabled a disabled key")
	}
}

// 0 disables the per-key limit, so it has to be distinguishable from "omitted".
func TestMCPUpdateKeyAcceptsExplicitZero(t *testing.T) {
	session, db, _ := connectWithTools(t, false)

	key, _ := models.CreateAPIKey(db, "Limited", "", 0, 0, "")
	if err := models.UpdateAPIKey(db, key.ID, models.UpdateAPIKeyParams{
		Name: key.Name, Domain: key.Domain, MaxNumber: key.MaxNumber,
		ExpireSeconds: key.ExpireSeconds, Algorithm: key.Algorithm,
		RateLimitPerMin: 60, Enabled: true,
	}); err != nil {
		t.Fatalf("UpdateAPIKey failed: %v", err)
	}

	var updated mcpKeyView
	mustCall(t, session, "update_key", map[string]any{"id": key.ID, "rate_limit_per_min": 0}, &updated)
	if updated.RateLimitPerMin != 0 {
		t.Errorf("an explicit 0 should clear the limit, got %d", updated.RateLimitPerMin)
	}
}

// update_key must not be the thing that takes a site's CAPTCHA down. Asserted
// on the published schema rather than on behaviour alone: the field has to be
// absent, not merely ignored, so a client never offers it in the first place.
func TestMCPUpdateKeyHasNoEnabledField(t *testing.T) {
	session, db, _ := connectWithTools(t, false)

	res, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}
	for _, tool := range res.Tools {
		if tool.Name != "update_key" {
			continue
		}
		raw, err := json.Marshal(tool.InputSchema)
		if err != nil {
			t.Fatalf("failed to marshal the schema: %v", err)
		}
		var schema struct {
			Properties map[string]any `json:"properties"`
		}
		if err := json.Unmarshal(raw, &schema); err != nil {
			t.Fatalf("failed to decode the schema: %v", err)
		}
		if _, found := schema.Properties["enabled"]; found {
			t.Error("update_key must not accept an enabled field")
		}
		if _, found := schema.Properties["name"]; !found {
			t.Error("sanity check failed: update_key should accept a name")
		}
	}

	// Smuggling the field in is refused by schema validation before the handler
	// runs, so this asserts the schema, not the handler. Preserving Enabled on a
	// legitimate partial update is covered by TestMCPUpdateKeyLeavesOmittedFieldsAlone.
	key, _ := models.CreateAPIKey(db, "Live", "live.com", 0, 0, "")
	_, _ = session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "update_key",
		Arguments: map[string]any{"id": key.ID, "enabled": false},
	})

	after, err := models.GetAPIKeyByID(db, key.ID)
	if err != nil {
		t.Fatalf("GetAPIKeyByID failed: %v", err)
	}
	if !after.Enabled {
		t.Error("update_key disabled a key; only disable_key may do that")
	}
}

func TestMCPEnableAndDisableKey(t *testing.T) {
	session, db, _ := connectWithTools(t, false)

	key, _ := models.CreateAPIKey(db, "Site", "site.com", 0, 0, "")

	var disabled mcpKeyView
	mustCall(t, session, "disable_key", map[string]any{"id": key.ID}, &disabled)
	if disabled.Enabled {
		t.Error("disable_key did not disable the key")
	}
	stored, _ := models.GetAPIKeyByID(db, key.ID)
	if stored.Enabled {
		t.Error("the key is still enabled in the database")
	}

	// Disabling must not disturb anything else about the key.
	if stored.Name != "Site" || stored.Domain != "site.com" {
		t.Errorf("disable_key changed more than the flag: %+v", stored)
	}

	var enabled mcpKeyView
	mustCall(t, session, "enable_key", map[string]any{"id": key.ID}, &enabled)
	if !enabled.Enabled {
		t.Error("enable_key did not enable the key")
	}

	// Both are idempotent, as their annotations claim.
	mustCall(t, session, "enable_key", map[string]any{"id": key.ID}, &enabled)
	if !enabled.Enabled {
		t.Error("enabling twice should be a no-op, not a reversal")
	}
}

func TestMCPEnableDisableRejectUnknownID(t *testing.T) {
	session, _, _ := connectWithTools(t, false)

	for _, name := range []string{"enable_key", "disable_key"} {
		if res := callTool(t, session, name, map[string]any{"id": 99999}, nil); !res.IsError {
			t.Errorf("%s should refuse an unknown key", name)
		}
	}
}

// The annotations drive the consent prompt a client shows, so they have to say
// the truth about which tool takes a site down.
func TestMCPToolAnnotations(t *testing.T) {
	session, _, _ := connectWithTools(t, false)

	res, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}

	for _, tool := range res.Tools {
		if tool.Annotations == nil {
			t.Errorf("%s has no annotations", tool.Name)
			continue
		}
		switch tool.Name {
		case "list_keys", "get_key":
			if !tool.Annotations.ReadOnlyHint {
				t.Errorf("%s should be marked read-only", tool.Name)
			}
		case "disable_key":
			if tool.Annotations.DestructiveHint == nil || !*tool.Annotations.DestructiveHint {
				t.Error("disable_key should be marked destructive")
			}
		default:
			if tool.Annotations.DestructiveHint == nil || *tool.Annotations.DestructiveHint {
				t.Errorf("%s should not be marked destructive", tool.Name)
			}
		}
	}
}
