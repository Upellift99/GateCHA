// Package geo resolves a client IP to an ISO 3166-1 alpha-2 country code using
// a pure-Go, embedded database (no external service, no file to mount). It is
// privacy-preserving by construction: callers resolve the country at request
// time and immediately discard the IP — GateCHA never stores or logs the raw
// address, only the aggregated country.
package geo

import (
	"net"
	"strings"

	"github.com/phuslu/iploc"
)

// Country returns the uppercase ISO 3166-1 alpha-2 code for ip, or "" when the
// address is empty, unparseable, private/loopback/link-local, or unknown to the
// embedded database. An empty result means "unknown" and should render without
// a flag.
func Country(ip string) string {
	if ip == "" {
		return ""
	}
	parsed := net.ParseIP(ip)
	if parsed == nil ||
		parsed.IsLoopback() ||
		parsed.IsPrivate() ||
		parsed.IsUnspecified() ||
		parsed.IsLinkLocalUnicast() {
		return ""
	}
	code := strings.ToUpper(iploc.Country(parsed))
	// iploc returns "ZZ" for addresses it cannot place; normalize to unknown.
	if len(code) != 2 || code == "ZZ" {
		return ""
	}
	return code
}
