package config

import "testing"

func TestAllowedPortsAndHasNonRejectDefault(t *testing.T) {
	cfg := Default()
	cfg.Routing.Rules = []RouteRule{
		{DstPorts: PortSpecs{80, 443}},
		{DstPort: PortSpecs{22}},
	}
	ports := cfg.AllowedPorts()
	if len(ports) != 3 {
		t.Fatalf("expected 3 allowed ports, got %d", len(ports))
	}

	if cfg.HasNonRejectDefault() {
		t.Fatalf("expected false when default is not set")
	}
	cfg.Routing.Default = &DefaultRule{Method: "reject"}
	if cfg.HasNonRejectDefault() {
		t.Fatalf("expected false for reject default")
	}
	cfg.Routing.Default = &DefaultRule{Method: "direct"}
	if !cfg.HasNonRejectDefault() {
		t.Fatalf("expected true for non-reject default")
	}
}

func TestParseUpstreamEndpoint(t *testing.T) {
	ok := []string{
		"http://proxy.example.com:3128",
		"https://proxy.example.com:8443",
	}
	for _, raw := range ok {
		ep, err := ParseUpstreamEndpoint(raw)
		if err != nil {
			t.Fatalf("expected valid upstream %q, got error: %v", raw, err)
		}
		if ep.Scheme == "" || ep.Address == "" || ep.ServerName == "" {
			t.Fatalf("expected parsed endpoint fields for %q", raw)
		}
	}

	bad := []string{
		"",
		"proxy.example.com:3128",
		"socks5://proxy.example.com:1080",
		"http://proxy.example.com:3128/path",
		"http://proxy.example.com:3128?x=1",
		"http://user:pass@proxy.example.com:3128",
		"http://proxy.example.com",
	}
	for _, raw := range bad {
		if _, err := ParseUpstreamEndpoint(raw); err == nil {
			t.Fatalf("expected invalid upstream %q to fail", raw)
		}
	}
}
