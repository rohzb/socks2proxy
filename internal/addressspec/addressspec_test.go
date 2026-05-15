package addressspec

import (
	"net"
	"testing"
)

func TestValidateAcceptsSupportedSpecs(t *testing.T) {
	cases := []string{
		"10.0.0.1",
		"10.0.0.0/24",
		"10.0.0.10-10.0.0.20",
		"2001:db8::1",
		"2001:db8::/32",
		"2001:db8::1-2001:db8::ff",
	}
	for _, c := range cases {
		if err := Validate(c); err != nil {
			t.Fatalf("expected valid spec %q, got error: %v", c, err)
		}
	}
}

func TestValidateRejectsInvalidSpecs(t *testing.T) {
	cases := []string{
		"not-an-ip",
		"10.0.0.20-10.0.0.10",
		"10.0.0.1-2001:db8::1",
	}
	for _, c := range cases {
		if err := Validate(c); err == nil {
			t.Fatalf("expected invalid spec %q to fail validation", c)
		}
	}
}

func TestMatchesSupportedSpecs(t *testing.T) {
	if !Matches("10.0.0.1", net.ParseIP("10.0.0.1")) {
		t.Fatalf("single IP must match")
	}
	if Matches("10.0.0.1", net.ParseIP("10.0.0.2")) {
		t.Fatalf("single IP must not match different ip")
	}
	if !Matches("10.0.0.0/24", net.ParseIP("10.0.0.42")) {
		t.Fatalf("cidr must match contained ip")
	}
	if Matches("10.0.0.0/24", net.ParseIP("10.0.1.1")) {
		t.Fatalf("cidr must not match outside ip")
	}
	if !Matches("10.0.0.10-10.0.0.20", net.ParseIP("10.0.0.15")) {
		t.Fatalf("range must match contained ip")
	}
	if Matches("10.0.0.10-10.0.0.20", net.ParseIP("10.0.0.30")) {
		t.Fatalf("range must not match outside ip")
	}
}

func TestMatchesInvalidSpecsAndBytesCompare(t *testing.T) {
	if Matches("10.0.0.1-bad", net.ParseIP("10.0.0.1")) {
		t.Fatalf("invalid range spec must not match")
	}
	if Matches("broken", net.ParseIP("10.0.0.1")) {
		t.Fatalf("invalid spec must not match")
	}
	if bytesCompare([]byte{1, 2, 3}, []byte{1, 2, 3, 4}) >= 0 {
		t.Fatalf("expected shorter slice to compare less")
	}
	if bytesCompare([]byte{1, 2, 4}, []byte{1, 2, 3}) <= 0 {
		t.Fatalf("expected greater slice to compare greater")
	}
}
