package socks5

import "testing"

func TestNormalizeTargetHost(t *testing.T) {
	if got := normalizeTargetHost("1.2.3.4", 443); got != "1.2.3.4" {
		t.Fatalf("unexpected ipv4 normalize: %s", got)
	}
	if got := normalizeTargetHost("[2001:db8::1]:443", 443); got != "2001:db8::1" {
		t.Fatalf("unexpected bracket ipv6 normalize: %s", got)
	}
	if got := normalizeTargetHost("example.com:443", 443); got != "example.com" {
		t.Fatalf("unexpected host:port normalize: %s", got)
	}
	if got := normalizeTargetHost("example.com:444", 443); got != "example.com:444" {
		t.Fatalf("expected unmatched port to remain unchanged: %s", got)
	}
}
