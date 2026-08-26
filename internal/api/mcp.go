package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Upellift99/GateCHA/internal/models"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gorm.io/gorm"
)

// mcpServerName identifies this server to MCP clients.
const mcpServerName = "gatecha"

type mcpTokenCtxKey struct{}

// MCPTokenFromContext returns the token that authenticated the current request.
func MCPTokenFromContext(ctx context.Context) *models.MCPToken {
	token, _ := ctx.Value(mcpTokenCtxKey{}).(*models.MCPToken)
	return token
}

// MCPEnabledMiddleware refuses every request while the endpoint is switched off.
//
// It answers 404 because that is what the path is: nothing is served here. This
// does not hide the endpoint's existence, and is not meant to. The SPA fallback
// answers 200 for unknown paths, so a 404 is distinguishable either way, and the
// dashboard already announces which product is running at this domain.
func MCPEnabledMiddleware(db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			enabled, err := models.GetMCPEnabled(db)
			if err != nil {
				slog.Error("failed to read the mcp toggle", "error", err)
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
				return
			}
			if !enabled {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "mcp endpoint is disabled"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// MCPAuthMiddleware authenticates a bearer MCP token.
//
// The token is deliberately not accepted from the query string: URLs end up in
// proxy logs and browser history, and this credential carries admin capability.
func MCPAuthMiddleware(db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			secret := bearerToken(r)
			if secret == "" {
				unauthorizedMCP(w)
				return
			}

			token, err := models.AuthenticateMCPToken(db, secret)
			if err != nil {
				if !errors.Is(err, models.ErrMCPTokenNotFound) {
					slog.Error("failed to authenticate an mcp token", "error", err)
				}
				unauthorizedMCP(w)
				return
			}

			// Best effort: a failed bookkeeping write must not deny a valid caller.
			if err := models.TouchMCPToken(db, token.ID); err != nil {
				slog.Error("failed to record mcp token usage", "error", err, "token_id", token.ID)
			}

			ctx := context.WithValue(r.Context(), mcpTokenCtxKey{}, token)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func bearerToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if len(header) < 7 || !strings.EqualFold(header[:7], "bearer ") {
		return ""
	}
	return strings.TrimSpace(header[7:])
}

func unauthorizedMCP(w http.ResponseWriter) {
	// RFC 6750: tell the client how to authenticate rather than failing blankly.
	w.Header().Set("WWW-Authenticate", `Bearer realm="gatecha-mcp"`)
	writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid or missing mcp token"})
}

// newMCPServer builds the MCP server exposed to one authenticated token.
//
// A server is built per request rather than shared, so a read-only token can be
// given a strictly smaller set of tools. Withholding the tool is stronger than
// checking a flag inside it: a tool that was never registered cannot be called
// and is never advertised in tools/list.
func newMCPServer(db *gorm.DB, version string, token *models.MCPToken) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    mcpServerName,
		Title:   "GateCHA",
		Version: version,
	}, nil)

	registerMCPReadTools(server, db)
	if token != nil && !token.ReadOnly {
		registerMCPWriteTools(server, db)
	}
	return server
}

// MCPHandler returns the streamable HTTP handler for the MCP endpoint.
//
// Stateless mode is required to serve protocol version 2026-07-28, and suits
// this endpoint anyway: there is no server-to-client request to make, so there
// is nothing for a session to carry between requests.
func MCPHandler(db *gorm.DB, version string) http.Handler {
	return mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server {
			return newMCPServer(db, version, MCPTokenFromContext(r.Context()))
		},
		&mcp.StreamableHTTPOptions{Stateless: true},
	)
}
