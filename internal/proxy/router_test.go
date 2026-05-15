package proxy

import (
	"testing"

	"socks2proxy/internal/config"
)

func TestParsePort(t *testing.T) {
	if p, err := ParsePort("443"); err != nil || p != 443 {
		t.Fatalf("expected parsed port 443, got p=%d err=%v", p, err)
	}
	if _, err := ParsePort("0"); err == nil {
		t.Fatalf("expected out-of-range port to fail")
	}
	if _, err := ParsePort("abc"); err == nil {
		t.Fatalf("expected invalid port string to fail")
	}
}

func TestRulesFromConfigAndDefaultRule(t *testing.T) {
	globalTLS := &config.UpstreamTLS{MinVersion: "1.2"}
	rules := RulesFromConfig(config.Routing{
		Rules: []config.RouteRule{
			{
				DstPorts:     config.PortSpecs{80},
				DstAddresses: config.AddressSpecs{"10.0.0.0/24"},
				Method:       "http",
				Upstream:     "http://proxy.example.com:3128",
			},
		},
	}, globalTLS)
	if len(rules) != 1 {
		t.Fatalf("expected one rule")
	}
	if rules[0].Method != MethodHTTP {
		t.Fatalf("expected method http")
	}
	if _, ok := rules[0].DstPorts[80]; !ok {
		t.Fatalf("expected dst port 80 in rule")
	}
	if rules[0].TLS != globalTLS {
		t.Fatalf("expected global tls defaults on rule")
	}

	def := DefaultRuleFromConfig(config.Routing{}, globalTLS)
	if def.Method != MethodReject {
		t.Fatalf("expected implicit reject default")
	}
}

func TestRuleMatchHelpers(t *testing.T) {
	r := Rule{
		DstPorts:     map[int]struct{}{443: {}},
		DstAddresses: []string{"10.0.0.0/24", "192.168.1.10-192.168.1.20"},
	}
	if !r.matchesPort(443) || r.matchesPort(80) {
		t.Fatalf("unexpected port match behavior")
	}
	if !r.matchesAddress("10.0.0.5") {
		t.Fatalf("expected cidr address match")
	}
	if !r.matchesAddress("192.168.1.15") {
		t.Fatalf("expected range address match")
	}
	if r.matchesAddress("example.com") {
		t.Fatalf("hostname should not match address specs")
	}
}
