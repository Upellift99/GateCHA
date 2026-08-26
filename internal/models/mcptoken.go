package models

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

// MCPTokenPrefix marks a token as an MCP admin credential, keeping it visually
// distinct from the gk_ site API keys.
const MCPTokenPrefix = "gm_"

// mcpTokenDisplayLen is how much of the token is stored in clear for display.
// It covers the prefix plus 8 hex characters, enough to tell tokens apart in a
// list without narrowing a brute-force search in any meaningful way.
const mcpTokenDisplayLen = len(MCPTokenPrefix) + 8

// ErrMCPTokenNotFound is returned when no enabled token matches a secret.
var ErrMCPTokenNotFound = errors.New("mcp token not found")

// MCPToken is a long-lived credential for the MCP endpoint, issued per person
// rather than shared, so it can be revoked without disturbing anyone else.
//
// The secret is stored as a SHA-256 digest, deliberately not bcrypt. Bcrypt is
// slow by design to defend low-entropy human-chosen passwords; these tokens are
// 128 bits of CSPRNG output, so there is nothing to slow down, and bcrypt would
// add its work factor to every single MCP request for no gain.
type MCPToken struct {
	ID   int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name string `gorm:"not null;default:''" json:"name"`
	// TokenHash is the hex-encoded SHA-256 of the secret. Never serialised.
	TokenHash string `gorm:"not null;uniqueIndex;size:64" json:"-"`
	// Display holds the token's leading characters so the dashboard can show
	// which token a row refers to after the secret itself is gone.
	Display string `gorm:"not null;default:''" json:"display"`
	// ReadOnly restricts the token to the read-only MCP tools.
	ReadOnly bool `gorm:"not null;default:false" json:"read_only"`
	// LastUsedAt surfaces dormant tokens that should be revoked. Nil until the
	// token authenticates for the first time.
	LastUsedAt *time.Time `json:"last_used_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

// HashMCPToken returns the storage digest for a token secret.
func HashMCPToken(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

// GenerateMCPToken returns a new token secret with 128 bits of entropy.
func GenerateMCPToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return MCPTokenPrefix + hex.EncodeToString(b), nil
}

// CreateMCPToken issues a token and returns it alongside the stored record.
// The secret is returned only here; it is unrecoverable afterwards.
func CreateMCPToken(db *gorm.DB, name string, readOnly bool) (*MCPToken, string, error) {
	secret, err := GenerateMCPToken()
	if err != nil {
		return nil, "", err
	}

	token := &MCPToken{
		Name:      name,
		TokenHash: HashMCPToken(secret),
		Display:   secret[:mcpTokenDisplayLen],
		ReadOnly:  readOnly,
		CreatedAt: time.Now(),
	}
	if err := db.Create(token).Error; err != nil {
		return nil, "", err
	}
	return token, secret, nil
}

// ListMCPTokens returns every token, newest first. Digests are never loaded.
func ListMCPTokens(db *gorm.DB) ([]MCPToken, error) {
	var tokens []MCPToken
	return tokens, db.Omit("token_hash").Order("created_at desc").Find(&tokens).Error
}

// DeleteMCPToken revokes a token. It reports whether a row was removed, so a
// repeated revoke can be told apart from one that never existed.
func DeleteMCPToken(db *gorm.DB, id int64) (bool, error) {
	res := db.Delete(&MCPToken{}, id)
	return res.RowsAffected > 0, res.Error
}

// AuthenticateMCPToken resolves a presented secret to its token record.
//
// The lookup is by digest, so the database never sees the secret and an
// attacker cannot probe for a prefix. The constant-time comparison that follows
// guards the row we did find; the indexed lookup itself is not a useful timing
// oracle because the attacker cannot steer SHA-256 output.
func AuthenticateMCPToken(db *gorm.DB, secret string) (*MCPToken, error) {
	secret = strings.TrimSpace(secret)
	// Early-out, not a security boundary: rejection is already guaranteed by the
	// digest lookup below, since a malformed secret cannot hash to a stored row.
	// This only saves a database round-trip on the obviously-wrong input an
	// internet-facing endpoint gets probed with.
	if !strings.HasPrefix(secret, MCPTokenPrefix) {
		return nil, ErrMCPTokenNotFound
	}

	digest := HashMCPToken(secret)

	var token MCPToken
	err := db.Where("token_hash = ?", digest).First(&token).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrMCPTokenNotFound
	}
	if err != nil {
		return nil, err
	}

	if subtle.ConstantTimeCompare([]byte(token.TokenHash), []byte(digest)) != 1 {
		return nil, ErrMCPTokenNotFound
	}
	return &token, nil
}

// TouchMCPToken records that a token was just used.
func TouchMCPToken(db *gorm.DB, id int64) error {
	now := time.Now()
	return db.Model(&MCPToken{}).Where("id = ?", id).Update("last_used_at", now).Error
}
