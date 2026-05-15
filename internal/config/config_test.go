package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestValidateSupportsAllRoutingMethods(t *testing.T) {
	cfg := Default()
	cfg.Routing.Rules = []RouteRule{
		{DstPorts: PortSpecs{80}, DstAddresses: AddressSpecs{"0.0.0.0/0"}, Method: "http", Upstream: "http://proxy.example.com:3128"},
		{DstPorts: PortSpecs{443}, DstAddresses: AddressSpecs{"0.0.0.0/0"}, Method: "connect", Upstream: "http://proxy.example.com:3128"},
		{DstPorts: PortSpecs{22}, DstAddresses: AddressSpecs{"0.0.0.0/0"}, Method: "direct"},
		{DstPorts: PortSpecs{25}, DstAddresses: AddressSpecs{"0.0.0.0/0"}, Method: "reject"},
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid config, got error: %v", err)
	}
}

func TestValidateSupportsRoutingDefault(t *testing.T) {
	cfg := Default()
	cfg.Routing.Rules = []RouteRule{
		{DstPorts: PortSpecs{443}, DstAddresses: AddressSpecs{"0.0.0.0/0"}, Method: "connect", Upstream: "http://proxy.example.com:3128"},
	}
	cfg.Routing.Default = &DefaultRule{Method: "direct"}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid config with routing.default, got error: %v", err)
	}
}

func TestValidateRejectsInvalidRoutingDefault(t *testing.T) {
	cfg := Default()
	cfg.Routing.Default = &DefaultRule{Method: "connect"}
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected validation error for default connect without upstream")
	}
}

func TestValidateDirectMustNotHaveUpstream(t *testing.T) {
	cfg := Default()
	cfg.Routing.Rules = []RouteRule{
		{DstPorts: PortSpecs{22}, DstAddresses: AddressSpecs{"0.0.0.0/0"}, Method: "direct", Upstream: "http://proxy.example.com:3128"},
	}

	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected validation error for direct+upstream")
	}
}

func TestValidateHTTPAndConnectRequireUpstream(t *testing.T) {
	tests := []RouteRule{
		{DstPorts: PortSpecs{80}, DstAddresses: AddressSpecs{"0.0.0.0/0"}, Method: "http"},
		{DstPorts: PortSpecs{443}, DstAddresses: AddressSpecs{"0.0.0.0/0"}, Method: "connect"},
	}

	for _, rule := range tests {
		cfg := Default()
		cfg.Routing.Rules = []RouteRule{rule}
		if err := cfg.Validate(); err == nil {
			t.Fatalf("expected validation error for method %q without upstream", rule.Method)
		}
	}
}

func TestValidateRejectsUnknownLogLevel(t *testing.T) {
	cfg := Default()
	cfg.Routing.Rules = []RouteRule{
		{DstPorts: PortSpecs{80}, DstAddresses: AddressSpecs{"0.0.0.0/0"}, Method: "http", Upstream: "http://proxy.example.com:3128"},
	}
	cfg.Logging.Level = "verbose"
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected validation error for unknown logging.level")
	}
}

func TestValidateRejectsInvalidUpstreamPort(t *testing.T) {
	cfg := Default()
	cfg.Routing.Rules = []RouteRule{
		{DstPorts: PortSpecs{80}, DstAddresses: AddressSpecs{"0.0.0.0/0"}, Method: "http", Upstream: "proxy.example:70000"},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected validation error for invalid upstream port")
	}
}

func TestValidateRejectsInvalidListenPort(t *testing.T) {
	cfg := Default()
	cfg.Listen = ":0"
	cfg.Routing.Rules = []RouteRule{
		{DstPorts: PortSpecs{80}, DstAddresses: AddressSpecs{"0.0.0.0/0"}, Method: "http", Upstream: "http://proxy.example.com:3128"},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected validation error for invalid listen port")
	}
}

func TestValidateRejectsInvalidClientCIDR(t *testing.T) {
	cfg := Default()
	cfg.AllowedClientAddresses = []string{"172.31.11.0/99"}
	cfg.Routing.Rules = []RouteRule{
		{DstPorts: PortSpecs{80}, DstAddresses: AddressSpecs{"0.0.0.0/0"}, Method: "http", Upstream: "http://proxy.example.com:3128"},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected validation error for invalid allowed_client_addresses entry")
	}
}

func TestValidateSupportsHTTPAndHTTPSUpstreamURLs(t *testing.T) {
	cfg := Default()
	cfg.Routing.Rules = []RouteRule{
		{DstPorts: PortSpecs{80}, DstAddresses: AddressSpecs{"0.0.0.0/0"}, Method: "http", Upstream: "http://proxy.example.com:3128"},
		{DstPorts: PortSpecs{443}, DstAddresses: AddressSpecs{"0.0.0.0/0"}, Method: "connect", Upstream: "https://proxy.example.com:8443"},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid config for http/https upstream URLs, got: %v", err)
	}
}

func TestValidateRejectsUnsupportedUpstreamScheme(t *testing.T) {
	cfg := Default()
	cfg.Routing.Rules = []RouteRule{
		{DstPorts: PortSpecs{80}, DstAddresses: AddressSpecs{"0.0.0.0/0"}, Method: "http", Upstream: "socks5://proxy.example.com:1080"},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected validation error for unsupported upstream URL scheme")
	}
}

func TestValidateRejectsUpstreamWithoutScheme(t *testing.T) {
	cfg := Default()
	cfg.Routing.Rules = []RouteRule{
		{DstPorts: PortSpecs{80}, DstAddresses: AddressSpecs{"0.0.0.0/0"}, Method: "http", Upstream: "proxy.example.com:3128"},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected validation error for upstream without scheme")
	}
}

func TestValidateSupportsDstAddressesSingleCIDRAndRange(t *testing.T) {
	cfg := Default()
	cfg.Routing.Rules = []RouteRule{
		{
			DstPorts:     PortSpecs{443},
			DstAddresses: AddressSpecs{"10.0.0.10", "172.16.0.0/16", "192.168.1.10-192.168.1.200"},
			Method:       "connect",
			Upstream:     "https://proxy.example.com:8443",
		},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid config for dst_addresses formats, got: %v", err)
	}
}

func TestValidateRejectsInvalidDstAddressRange(t *testing.T) {
	cfg := Default()
	cfg.Routing.Rules = []RouteRule{
		{
			DstPorts:     PortSpecs{443},
			DstAddresses: AddressSpecs{"192.168.1.200-192.168.1.10"},
			Method:       "connect",
			Upstream:     "https://proxy.example.com:8443",
		},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected validation error for invalid dst_addresses IP range")
	}
}

func TestValidateSupportsInsecureSkipVerifyForHTTPSUpstream(t *testing.T) {
	cfg := Default()
	cfg.Routing.Rules = []RouteRule{
		{
			DstPorts:     PortSpecs{80},
			DstAddresses: AddressSpecs{"0.0.0.0/0"},
			Method:       "http",
			Upstream:     "https://proxy.example.com:8443",
			TLS: &UpstreamTLS{
				InsecureSkipVerify: true,
				MinVersion:         "1.2",
			},
		},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid config for https upstream with insecure_skip_verify, got: %v", err)
	}
}

func TestValidateRejectsMissingTLSCACertFile(t *testing.T) {
	cfg := Default()
	cfg.Routing.Rules = []RouteRule{
		{
			DstPorts:     PortSpecs{80},
			DstAddresses: AddressSpecs{"0.0.0.0/0"},
			Method:       "http",
			Upstream:     "https://proxy.example.com:8443",
			TLS: &UpstreamTLS{
				CACertFile: "/path/does/not/exist.pem",
			},
		},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected validation error for missing tls.ca_cert_file")
	}
}

func TestValidateRejectsTLSOptionsForHTTPUpstream(t *testing.T) {
	cfg := Default()
	cfg.Routing.Rules = []RouteRule{
		{
			DstPorts:     PortSpecs{80},
			DstAddresses: AddressSpecs{"0.0.0.0/0"},
			Method:       "http",
			Upstream:     "http://proxy.example.com:3128",
			TLS: &UpstreamTLS{
				InsecureSkipVerify: true,
				MinVersion:         "1.3",
			},
		},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected validation error for tls options with non-https upstream")
	}
}

func TestValidateRejectsUnsupportedTLSMinVersion(t *testing.T) {
	cfg := Default()
	cfg.Routing.Rules = []RouteRule{
		{
			DstPorts:     PortSpecs{443},
			DstAddresses: AddressSpecs{"0.0.0.0/0"},
			Method:       "connect",
			Upstream:     "https://proxy.example.com:8443",
			TLS: &UpstreamTLS{
				MinVersion: "1.1",
			},
		},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected validation error for unsupported tls.min_version")
	}
}

func TestValidateSupportsGlobalTLSDefaultsWithHTTPSRules(t *testing.T) {
	cfg := Default()
	cfg.TLS = &UpstreamTLS{
		MinVersion: "1.3",
	}
	cfg.Routing.Rules = []RouteRule{
		{
			DstPorts:     PortSpecs{443},
			DstAddresses: AddressSpecs{"0.0.0.0/0"},
			Method:       "connect",
			Upstream:     "https://proxy.example.com:8443",
		},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid config with global tls defaults, got: %v", err)
	}
}

func TestValidateAllowsGlobalTLSDefaultsWithHTTPUpstream(t *testing.T) {
	cfg := Default()
	cfg.TLS = &UpstreamTLS{
		MinVersion:         "1.3",
		InsecureSkipVerify: true,
	}
	cfg.Routing.Rules = []RouteRule{
		{
			DstPorts:     PortSpecs{80},
			DstAddresses: AddressSpecs{"0.0.0.0/0"},
			Method:       "http",
			Upstream:     "http://proxy.example.com:3128",
		},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected global tls defaults to be ignored for non-https upstream, got: %v", err)
	}
}

func TestAllowedClientAddressesSupportsCSVAndListForms(t *testing.T) {
	var cfg Config
	doc := `
listen: ":41080"
allowed_client_addresses: "127.0.0.1,10.0.0.0/24,192.168.1.10-192.168.1.20"
routing:
  default:
    method: "reject"
timeouts:
  connect: "10s"
  idle: "30s"
http:
  max_header_bytes: 65536
logging:
  level: "info"
`
	if err := yaml.Unmarshal([]byte(doc), &cfg); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid allowed_client_addresses csv forms, got: %v", err)
	}
}

func TestAllowedClientAddressesRejectsInvalidRangeOrder(t *testing.T) {
	cfg := Default()
	cfg.AllowedClientAddresses = AddressSpecs{"10.0.0.20-10.0.0.10"}
	cfg.Routing.Rules = []RouteRule{
		{
			DstPorts:     PortSpecs{443},
			DstAddresses: AddressSpecs{"0.0.0.0/0"},
			Method:       "connect",
			Upstream:     "https://proxy.example.com:8443",
		},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected invalid client address range to fail validation")
	}
}

func TestResolveTLSPrefersOverrideOverGlobal(t *testing.T) {
	global := &UpstreamTLS{MinVersion: "1.2"}
	override := &UpstreamTLS{MinVersion: "1.3"}
	got := ResolveTLS(global, override)
	if got != override {
		t.Fatalf("expected override tls pointer to be preferred")
	}
}
