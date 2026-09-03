package api

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Upellift99/GateCHA/internal/models"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gorm.io/gorm"
)

// mcpKeyView is the shape every tool returns for an API key.
//
// It has no field for the HMAC secret, on purpose. Making the omission
// structural means no tool can leak it by forgetting to strip it: the type
// simply cannot carry one. create_key returns its own type instead.
type mcpKeyView struct {
	ID                 int64   `json:"id"`
	KeyID              string  `json:"key_id"`
	Name               string  `json:"name"`
	Domain             string  `json:"domain"`
	Enabled            bool    `json:"enabled"`
	MaxNumber          int64   `json:"max_number"`
	ExpireSeconds      int     `json:"expire_seconds"`
	Algorithm          string  `json:"algorithm"`
	RateLimitPerMin    int     `json:"rate_limit_per_min"`
	AdaptiveDifficulty bool    `json:"adaptive_difficulty"`
	HISSampling        bool    `json:"his_sampling"`
	HISEnforce         bool    `json:"his_enforce"`
	HISThreshold       float64 `json:"his_threshold"`
	CreatedAt          string  `json:"created_at"`
	UpdatedAt          string  `json:"updated_at"`
}

func mcpViewOf(key *models.APIKey) mcpKeyView {
	return mcpKeyView{
		ID:                 key.ID,
		KeyID:              key.KeyID,
		Name:               key.Name,
		Domain:             key.Domain,
		Enabled:            key.Enabled,
		MaxNumber:          key.MaxNumber,
		ExpireSeconds:      key.ExpireSeconds,
		Algorithm:          key.Algorithm,
		RateLimitPerMin:    key.RateLimitPerMin,
		AdaptiveDifficulty: key.AdaptiveDifficulty,
		HISSampling:        key.HISSampling,
		HISEnforce:         key.HISEnforce,
		HISThreshold:       key.SuspectThreshold(),
		CreatedAt:          key.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:          key.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func boolPtr(v bool) *bool { return &v }

// --- read tools -------------------------------------------------------------

type listKeysInput struct {
	Search string `json:"search,omitempty" jsonschema:"case-insensitive filter on name, domain and key ID; omit to list every key"`
}

type listKeysOutput struct {
	Keys  []mcpKeyView `json:"keys"`
	Count int          `json:"count"`
}

type getKeyInput struct {
	ID int64 `json:"id" jsonschema:"numeric ID of the key, as returned by list_keys"`
}

func registerMCPReadTools(server *mcp.Server, db *gorm.DB) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_keys",
		Description: "List the ALTCHA API keys, optionally filtered. Never returns HMAC secrets.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "List API keys",
			ReadOnlyHint:  true,
			OpenWorldHint: boolPtr(false),
		},
	}, func(_ context.Context, _ *mcp.CallToolRequest, in listKeysInput) (*mcp.CallToolResult, listKeysOutput, error) {
		keys, err := models.ListAPIKeys(db)
		if err != nil {
			return nil, listKeysOutput{}, fmt.Errorf("failed to list keys: %w", err)
		}

		query := strings.ToLower(strings.TrimSpace(in.Search))
		views := make([]mcpKeyView, 0, len(keys))
		for i := range keys {
			if query != "" && !mcpKeyMatches(&keys[i], query) {
				continue
			}
			views = append(views, mcpViewOf(&keys[i]))
		}
		return nil, listKeysOutput{Keys: views, Count: len(views)}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_key",
		Description: "Fetch a single ALTCHA API key by ID. Never returns the HMAC secret.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Get an API key",
			ReadOnlyHint:  true,
			OpenWorldHint: boolPtr(false),
		},
	}, func(_ context.Context, _ *mcp.CallToolRequest, in getKeyInput) (*mcp.CallToolResult, mcpKeyView, error) {
		key, err := mcpFetchKey(db, in.ID)
		if err != nil {
			return nil, mcpKeyView{}, err
		}
		return nil, mcpViewOf(key), nil
	})
}

func mcpKeyMatches(key *models.APIKey, query string) bool {
	return strings.Contains(strings.ToLower(key.Name), query) ||
		strings.Contains(strings.ToLower(key.Domain), query) ||
		strings.Contains(strings.ToLower(key.KeyID), query)
}

// --- write tools ------------------------------------------------------------

type createKeyInput struct {
	Name          string `json:"name" jsonschema:"human-readable name for the key, usually the site it serves"`
	Domain        string `json:"domain,omitempty" jsonschema:"domain allowed to use this key; leave empty to allow any domain"`
	MaxNumber     int64  `json:"max_number,omitempty" jsonschema:"proof-of-work difficulty; defaults to 100000 when omitted"`
	ExpireSeconds int    `json:"expire_seconds,omitempty" jsonschema:"challenge lifetime in seconds; defaults to 300 when omitted"`
	Algorithm     string `json:"algorithm,omitempty" jsonschema:"hash algorithm, SHA-256 by default"`
}

type createKeyOutput struct {
	Key mcpKeyView `json:"key"`
	// HMACSecret is returned here and nowhere else, ever.
	HMACSecret string `json:"hmac_secret" jsonschema:"the signing secret; shown only once, at creation, and unrecoverable afterwards"`
}

type updateKeyInput struct {
	ID                 int64    `json:"id" jsonschema:"numeric ID of the key to update"`
	Name               *string  `json:"name,omitempty" jsonschema:"omit to leave unchanged"`
	Domain             *string  `json:"domain,omitempty" jsonschema:"omit to leave unchanged; an empty string widens the key to any domain"`
	MaxNumber          *int64   `json:"max_number,omitempty" jsonschema:"omit to leave unchanged"`
	ExpireSeconds      *int     `json:"expire_seconds,omitempty" jsonschema:"omit to leave unchanged"`
	Algorithm          *string  `json:"algorithm,omitempty" jsonschema:"omit to leave unchanged"`
	RateLimitPerMin    *int     `json:"rate_limit_per_min,omitempty" jsonschema:"requests per minute for this key; 0 means unlimited; omit to leave unchanged"`
	AdaptiveDifficulty *bool    `json:"adaptive_difficulty,omitempty" jsonschema:"omit to leave unchanged"`
	HISSampling        *bool    `json:"his_sampling,omitempty" jsonschema:"omit to leave unchanged"`
	HISEnforce         *bool    `json:"his_enforce,omitempty" jsonschema:"reject verifications whose HIS score reaches his_threshold; off by default; omit to leave unchanged"`
	HISThreshold       *float64 `json:"his_threshold,omitempty" jsonschema:"score in (0,1] at or above which this key treats a sample as automation; higher means more bot-like; default 0.8; omit to leave unchanged"`
}

type keyIDInput struct {
	ID int64 `json:"id" jsonschema:"numeric ID of the key"`
}

func registerMCPWriteTools(server *mcp.Server, db *gorm.DB) {
	mcp.AddTool(server, &mcp.Tool{
		Name: "create_key",
		Description: "Create an ALTCHA API key. The response carries the HMAC secret, which is " +
			"shown only here and cannot be retrieved later.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Create an API key",
			DestructiveHint: boolPtr(false),
			OpenWorldHint:   boolPtr(false),
		},
	}, func(_ context.Context, _ *mcp.CallToolRequest, in createKeyInput) (*mcp.CallToolResult, createKeyOutput, error) {
		if strings.TrimSpace(in.Name) == "" {
			return nil, createKeyOutput{}, fmt.Errorf("name is required")
		}
		key, err := models.CreateAPIKey(db, strings.TrimSpace(in.Name), in.Domain, in.MaxNumber, in.ExpireSeconds, in.Algorithm)
		if err != nil {
			return nil, createKeyOutput{}, fmt.Errorf("failed to create key: %w", err)
		}
		return nil, createKeyOutput{Key: mcpViewOf(key), HMACSecret: key.HMACSecret}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "update_key",
		Description: "Change an ALTCHA API key's settings. Omitted fields keep their current value. " +
			"Cannot enable or disable a key: use enable_key or disable_key for that.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Update an API key",
			IdempotentHint:  true,
			DestructiveHint: boolPtr(false),
			OpenWorldHint:   boolPtr(false),
		},
	}, func(_ context.Context, _ *mcp.CallToolRequest, in updateKeyInput) (*mcp.CallToolResult, mcpKeyView, error) {
		// Refused, never dropped: silently ignoring a threshold the caller
		// supplied would report success while the old blocking policy stayed in
		// force, which is the one outcome an agent cannot detect.
		if in.HISThreshold != nil && !validHISThreshold(*in.HISThreshold) {
			return nil, mcpKeyView{}, fmt.Errorf("his_threshold %v is invalid: %s", *in.HISThreshold, errInvalidHISThreshold)
		}
		if _, err := mcpFetchKey(db, in.ID); err != nil {
			return nil, mcpKeyView{}, err
		}

		if err := models.UpdateAPIKeyFields(db, in.ID, mcpUpdateFields(in)); err != nil {
			return nil, mcpKeyView{}, fmt.Errorf("failed to update key %d: %w", in.ID, err)
		}
		return mcpReadBack(db, in.ID)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "enable_key",
		Description: "Enable an ALTCHA API key so it serves challenges again.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Enable an API key",
			IdempotentHint:  true,
			DestructiveHint: boolPtr(false),
			OpenWorldHint:   boolPtr(false),
		},
	}, func(_ context.Context, _ *mcp.CallToolRequest, in keyIDInput) (*mcp.CallToolResult, mcpKeyView, error) {
		return mcpSetEnabled(db, in.ID, true)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "disable_key",
		Description: "Disable an ALTCHA API key. The site using it stops being able to issue or " +
			"verify challenges, so its CAPTCHA breaks until the key is enabled again.",
		Annotations: &mcp.ToolAnnotations{
			Title:          "Disable an API key",
			IdempotentHint: true,
			// The one tool here that takes a working site down.
			DestructiveHint: boolPtr(true),
			OpenWorldHint:   boolPtr(false),
		},
	}, func(_ context.Context, _ *mcp.CallToolRequest, in keyIDInput) (*mcp.CallToolResult, mcpKeyView, error) {
		return mcpSetEnabled(db, in.ID, false)
	})
}

// mcpFetchKey looks a key up and, when it is missing, returns the one wording
// every tool uses for it. A client sees the same sentence whichever tool it
// called with a bad ID.
func mcpFetchKey(db *gorm.DB, id int64) (*models.APIKey, error) {
	key, err := models.GetAPIKeyByID(db, id)
	if err != nil {
		return nil, fmt.Errorf("no key with ID %d", id)
	}
	return key, nil
}

// mcpUpdateFields maps the fields the caller actually sent onto the columns to
// write. Nothing is read first and echoed back, so a concurrent disable_key
// cannot be undone by a rename that had already read the old value. "enabled"
// is absent from updateKeyInput, so it can never appear in this map.
func mcpUpdateFields(in updateKeyInput) map[string]any {
	fields := map[string]any{}
	if in.Name != nil {
		fields["name"] = *in.Name
	}
	if in.Domain != nil {
		fields["domain"] = *in.Domain
	}
	if in.MaxNumber != nil {
		fields["max_number"] = *in.MaxNumber
	}
	if in.ExpireSeconds != nil {
		fields["expire_seconds"] = *in.ExpireSeconds
	}
	if in.Algorithm != nil {
		fields["algorithm"] = *in.Algorithm
	}
	if in.RateLimitPerMin != nil {
		fields["rate_limit_per_min"] = *in.RateLimitPerMin
	}
	if in.AdaptiveDifficulty != nil {
		fields["adaptive_difficulty"] = *in.AdaptiveDifficulty
	}
	if in.HISSampling != nil {
		fields["his_sampling"] = *in.HISSampling
	}
	if in.HISEnforce != nil {
		fields["his_enforce"] = *in.HISEnforce
	}
	// Validated by the caller, which refuses the whole update rather than
	// letting an out-of-range value through or dropping it in silence.
	if in.HISThreshold != nil {
		fields["his_threshold"] = *in.HISThreshold
	}
	return fields
}

func mcpSetEnabled(db *gorm.DB, id int64, enabled bool) (*mcp.CallToolResult, mcpKeyView, error) {
	if _, err := mcpFetchKey(db, id); err != nil {
		return nil, mcpKeyView{}, err
	}
	if err := models.SetAPIKeyEnabled(db, id, enabled); err != nil {
		return nil, mcpKeyView{}, fmt.Errorf("failed to update key %d: %w", id, err)
	}
	return mcpReadBack(db, id)
}

// mcpReadBack returns the key as it now stands, so a tool's answer reflects the
// stored row rather than what the caller asked for.
func mcpReadBack(db *gorm.DB, id int64) (*mcp.CallToolResult, mcpKeyView, error) {
	updated, err := models.GetAPIKeyByID(db, id)
	if err != nil {
		return nil, mcpKeyView{}, fmt.Errorf("key %d disappeared after the update: %w", id, err)
	}
	return nil, mcpViewOf(updated), nil
}
