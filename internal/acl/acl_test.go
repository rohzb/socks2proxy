package acl

import (
	"net"
	"testing"
)

func TestNewAddressAllowlistAndContains(t *testing.T) {
	list, err := NewAddressAllowlist([]string{"10.0.0.0/24", "192.168.1.10-192.168.1.20"})
	if err != nil {
		t.Fatalf("unexpected error creating allowlist: %v", err)
	}
	if !list.Contains(net.ParseIP("10.0.0.5")) {
		t.Fatalf("expected ip to be allowed")
	}
	if !list.Contains(net.ParseIP("192.168.1.15")) {
		t.Fatalf("expected ip in range to be allowed")
	}
	if list.Contains(net.ParseIP("172.16.0.1")) {
		t.Fatalf("unexpected allowed ip")
	}
}

func TestNewAddressAllowlistRejectsInvalidSpec(t *testing.T) {
	if _, err := NewAddressAllowlist([]string{"broken"}); err == nil {
		t.Fatalf("expected invalid spec to fail")
	}
}

func TestPortAllowlist(t *testing.T) {
	ports := NewPortAllowlist([]int{80, 443})
	if !ports.Allows(80) || !ports.Allows(443) {
		t.Fatalf("expected configured ports to be allowed")
	}
	if ports.Allows(22) {
		t.Fatalf("unexpected allowed port")
	}
	if (&PortAllowlist{}).Allows(22) {
		t.Fatalf("empty allowlist should not allow")
	}
	all := NewAllowAllPortAllowlist()
	if !all.Allows(22) || !all.Allows(65535) {
		t.Fatalf("allow-all should allow any port")
	}
}
