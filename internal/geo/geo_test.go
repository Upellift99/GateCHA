package geo

import "testing"

// These cases are deterministic from our own filtering logic, independent of the
// embedded database contents.
func TestCountryNonPublicIsUnknown(t *testing.T) {
	tests := []struct {
		name string
		ip   string
	}{
		{"empty", ""},
		{"garbage", "not-an-ip"},
		{"loopback v4", "127.0.0.1"},
		{"loopback v6", "::1"},
		{"private 10", "10.0.0.1"},
		{"private 192", "192.168.1.1"},
		{"unspecified", "0.0.0.0"},
		{"link-local", "169.254.1.1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Country(tt.ip); got != "" {
				t.Errorf("Country(%q) = %q, want \"\"", tt.ip, got)
			}
		})
	}
}

// A well-known public address must resolve to a normalized, uppercase, 2-letter
// code. We assert the shape rather than a specific country to stay resilient to
// embedded-database updates.
func TestCountryPublicIsTwoLetterCode(t *testing.T) {
	got := Country("8.8.8.8")
	if len(got) != 2 {
		t.Fatalf("Country(8.8.8.8) = %q, want a 2-letter code", got)
	}
	for _, r := range got {
		if r < 'A' || r > 'Z' {
			t.Errorf("Country(8.8.8.8) = %q, want uppercase ASCII", got)
		}
	}
}
