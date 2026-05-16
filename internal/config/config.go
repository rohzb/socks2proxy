// File config.go defines runtime configuration structures, parsing, and validation.
package config

import (
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"socks2proxy/internal/addressspec"
)

// Config is the fully merged runtime service configuration.
type Config struct {
	Listen                 string       `yaml:"listen"`
	AllowedClientAddresses AddressSpecs `yaml:"allowed_client_addresses"`
	TLS                    *UpstreamTLS `yaml:"tls,omitempty"`
	Timeouts               Timeouts     `yaml:"timeouts"`
	HTTP                   HTTP         `yaml:"http"`
	Logging                Logging      `yaml:"logging"`
	Routing                Routing      `yaml:"routing"`
}

// Routing holds explicit target-port routing behavior.
type Routing struct {
	Rules   []RouteRule  `yaml:"rules"`
	Default *DefaultRule `yaml:"default,omitempty"`
}

// RouteRule defines how to handle a specific destination port.
type RouteRule struct {
	DstPorts     PortSpecs    `yaml:"dst_ports,omitempty"`
	DstPort      PortSpecs    `yaml:"dst_port,omitempty"`
	DstAddresses AddressSpecs `yaml:"dst_addresses,omitempty"`
	DstAddress   AddressSpecs `yaml:"dst_address,omitempty"`
	SourceIP     string       `yaml:"source_ip,omitempty"`
	Method       string       `yaml:"method"`
	Upstream     string       `yaml:"upstream,omitempty"`
	TLS          *UpstreamTLS `yaml:"tls,omitempty"`
}

// DefaultRule defines fallback behavior when no routing rule matches.
type DefaultRule struct {
	Method   string       `yaml:"method"`
	Upstream string       `yaml:"upstream,omitempty"`
	SourceIP string       `yaml:"source_ip,omitempty"`
	TLS      *UpstreamTLS `yaml:"tls,omitempty"`
}

// UpstreamTLS defines optional TLS behavior for HTTPS upstream proxy connections.
type UpstreamTLS struct {
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify,omitempty"`
	CACertFile         string `yaml:"ca_cert_file,omitempty"`
	MinVersion         string `yaml:"min_version,omitempty"`
}

// UpstreamEndpoint is a parsed upstream proxy endpoint.
type UpstreamEndpoint struct {
	Scheme     string
	Address    string
	ServerName string
}

// PortSpecs supports scalar/list YAML and comma/range formats.
// Examples:
// - "80"
// - "80,443,10000-10100"
// - [80, 443, "10000-10100"]
type PortSpecs []int

// AddressSpecs supports scalar/list YAML and comma formats.
// Each entry may be single IP, CIDR, or range IP-IP.
type AddressSpecs []string

// UnmarshalYAML decodes destination port specs from scalar or sequence YAML.
func (p *PortSpecs) UnmarshalYAML(node *yaml.Node) error {
	tokens, err := parseScalarOrList(node)
	if err != nil {
		return err
	}
	ports := make([]int, 0)
	for _, t := range tokens {
		for _, part := range splitCSV(t) {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			if strings.Contains(part, "-") {
				bounds := strings.Split(part, "-")
				if len(bounds) != 2 {
					return fmt.Errorf("invalid port range %q", part)
				}
				minP, err := strconv.Atoi(strings.TrimSpace(bounds[0]))
				if err != nil {
					return fmt.Errorf("invalid port range start %q: %w", bounds[0], err)
				}
				maxP, err := strconv.Atoi(strings.TrimSpace(bounds[1]))
				if err != nil {
					return fmt.Errorf("invalid port range end %q: %w", bounds[1], err)
				}
				if minP > maxP {
					return fmt.Errorf("invalid port range %q: start > end", part)
				}
				for v := minP; v <= maxP; v++ {
					ports = append(ports, v)
				}
				continue
			}
			v, err := strconv.Atoi(part)
			if err != nil {
				return fmt.Errorf("invalid port value %q: %w", part, err)
			}
			ports = append(ports, v)
		}
	}
	*p = PortSpecs(ports)
	return nil
}

// UnmarshalYAML decodes address specs from scalar or sequence YAML.
func (a *AddressSpecs) UnmarshalYAML(node *yaml.Node) error {
	tokens, err := parseScalarOrList(node)
	if err != nil {
		return err
	}
	values := make([]string, 0)
	for _, t := range tokens {
		for _, part := range splitCSV(t) {
			part = strings.TrimSpace(part)
			if part != "" {
				values = append(values, part)
			}
		}
	}
	*a = AddressSpecs(values)
	return nil
}

func parseScalarOrList(node *yaml.Node) ([]string, error) {
	switch node.Kind {
	case yaml.ScalarNode:
		return []string{node.Value}, nil
	case yaml.SequenceNode:
		out := make([]string, 0, len(node.Content))
		for _, item := range node.Content {
			if item.Kind != yaml.ScalarNode {
				return nil, errors.New("only scalar entries are supported in list")
			}
			out = append(out, item.Value)
		}
		return out, nil
	default:
		return nil, errors.New("expected scalar or list")
	}
}

func splitCSV(s string) []string {
	return strings.Split(s, ",")
}

// Timeouts holds all duration-based service timeouts.
type Timeouts struct {
	Connect Duration `yaml:"connect"`
	Idle    Duration `yaml:"idle"`
}

// HTTP holds HTTP handling related settings.
type HTTP struct {
	MaxHeaderBytes int `yaml:"max_header_bytes"`
}

// Logging holds logger configuration options.
type Logging struct {
	Level string `yaml:"level"`
}

// Duration is a YAML-friendly wrapper around time.Duration.
type Duration struct {
	time.Duration
}

// UnmarshalYAML parses a YAML duration string into a Duration value.
func (d *Duration) UnmarshalYAML(node *yaml.Node) error {
	var raw string
	if err := node.Decode(&raw); err != nil {
		return err
	}
	parsed, err := time.ParseDuration(raw)
	if err != nil {
		return err
	}
	d.Duration = parsed
	return nil
}

// MarshalYAML renders a Duration as a duration string for YAML output.
func (d Duration) MarshalYAML() (any, error) {
	return d.String(), nil
}

// Default returns the baseline service configuration.
func Default() Config {
	return Config{
		Listen:                 ":41080",
		AllowedClientAddresses: []string{"127.0.0.1/32", "::1/128"},
		Timeouts: Timeouts{
			Connect: Duration{Duration: 10 * time.Second},
			Idle:    Duration{Duration: 300 * time.Second},
		},
		HTTP: HTTP{MaxHeaderBytes: 65536},
		Logging: Logging{
			Level: "info",
		},
		Routing: Routing{
			Rules: []RouteRule{
				{DstPorts: PortSpecs{80}, DstAddresses: AddressSpecs{"0.0.0.0/0", "::/0"}, Method: "http"},
				{DstPorts: PortSpecs{443}, DstAddresses: AddressSpecs{"0.0.0.0/0", "::/0"}, Method: "connect"},
			},
		},
	}
}

// LoadFile loads YAML configuration from the provided path.
func LoadFile(path string) (Config, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	cfg := Default()
	if err := yaml.Unmarshal(bytes, &cfg); err != nil {
		return Config{}, fmt.Errorf("failed to parse config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Validate ensures required settings are present and internally consistent.
func (c Config) Validate() error {
	if c.Listen == "" {
		return errors.New("listen is required")
	}
	listenAddr, err := net.ResolveTCPAddr("tcp", c.Listen)
	if err != nil {
		return fmt.Errorf("invalid listen address %q: %w", c.Listen, err)
	}
	if listenAddr.Port < 1 || listenAddr.Port > 65535 {
		return fmt.Errorf("listen port out of range: %d", listenAddr.Port)
	}
	if len(c.AllowedClientAddresses) == 0 {
		return errors.New("allowed_client_addresses cannot be empty")
	}
	for _, spec := range c.AllowedClientAddresses {
		if err := addressspec.Validate(spec); err != nil {
			return fmt.Errorf("invalid allowed_client_addresses entry %q: %w", spec, err)
		}
	}
	if len(c.Routing.Rules) == 0 && c.Routing.Default == nil {
		return errors.New("routing must define at least one rule or routing.default")
	}
	if c.TLS != nil {
		if err := validateTLSOptions(c.TLS); err != nil {
			return fmt.Errorf("tls: %w", err)
		}
	}
	for _, rule := range c.Routing.Rules {
		ports := rulePorts(rule)
		addresses := ruleAddresses(rule)
		if len(ports) == 0 && len(addresses) == 0 {
			return errors.New("routing rule must define at least one selector: dst_port(s) or dst_address(es)")
		}
		for _, p := range ports {
			if p < 1 || p > 65535 {
				return fmt.Errorf("routing rule port out of range: %d", p)
			}
		}
		for _, spec := range addresses {
			if err := addressspec.Validate(spec); err != nil {
				return fmt.Errorf("routing rule has invalid dst_addresses entry %q: %w", spec, err)
			}
		}
		effectiveTLS := ResolveTLS(c.TLS, rule.TLS)
		if err := validateRouteMethodAndUpstream(rule.Method, rule.Upstream, effectiveTLS, rule.TLS != nil); err != nil {
			return fmt.Errorf("routing rule: %w", err)
		}
		if rule.SourceIP != "" && net.ParseIP(rule.SourceIP) == nil {
			return fmt.Errorf("routing rule: invalid source_ip %q", rule.SourceIP)
		}
	}
	if c.Routing.Default != nil {
		effectiveTLS := ResolveTLS(c.TLS, c.Routing.Default.TLS)
		if err := validateRouteMethodAndUpstream(c.Routing.Default.Method, c.Routing.Default.Upstream, effectiveTLS, c.Routing.Default.TLS != nil); err != nil {
			return fmt.Errorf("routing.default: %w", err)
		}
		if c.Routing.Default.SourceIP != "" && net.ParseIP(c.Routing.Default.SourceIP) == nil {
			return fmt.Errorf("routing.default: invalid source_ip %q", c.Routing.Default.SourceIP)
		}
	}
	if c.HTTP.MaxHeaderBytes <= 0 {
		return errors.New("http.max_header_bytes must be > 0")
	}
	if !isAllowedLogLevel(c.Logging.Level) {
		return fmt.Errorf("unsupported logging.level %q (allowed: debug, info, warn, error)", c.Logging.Level)
	}
	if c.Timeouts.Connect.Duration <= 0 || c.Timeouts.Idle.Duration <= 0 {
		return errors.New("timeouts.connect and timeouts.idle must be > 0")
	}
	return nil
}

// AllowedPorts returns the list of target ports declared in routing rules.
func (c Config) AllowedPorts() []int {
	ports := make([]int, 0, len(c.Routing.Rules))
	for _, rule := range c.Routing.Rules {
		ports = append(ports, rulePorts(rule)...)
	}
	return ports
}

func rulePorts(rule RouteRule) []int {
	merged := make([]int, 0, len(rule.DstPorts)+len(rule.DstPort))
	merged = append(merged, rule.DstPorts...)
	merged = append(merged, rule.DstPort...)
	return merged
}

func ruleAddresses(rule RouteRule) []string {
	merged := make([]string, 0, len(rule.DstAddresses)+len(rule.DstAddress))
	merged = append(merged, rule.DstAddresses...)
	merged = append(merged, rule.DstAddress...)
	return merged
}

func validateHostPort(v string) error {
	host, port, err := net.SplitHostPort(v)
	if err != nil {
		return err
	}
	if host == "" {
		return errors.New("host is empty")
	}
	p, err := strconv.Atoi(port)
	if err != nil {
		return err
	}
	if p < 1 || p > 65535 {
		return fmt.Errorf("port out of range: %d", p)
	}
	return nil
}

func validateRouteMethodAndUpstream(method string, upstream string, tlsOpt *UpstreamTLS, explicitTLS bool) error {
	switch method {
	case "http", "connect":
		if upstream == "" {
			return fmt.Errorf("method=%q requires upstream", method)
		}
		endpoint, err := ParseUpstreamEndpoint(upstream)
		if err != nil {
			return fmt.Errorf("invalid upstream %q: %w", upstream, err)
		}
		if hasTLSOptions(tlsOpt) {
			if endpoint.Scheme != "https" {
				if explicitTLS {
					return errors.New("tls options are only allowed with https upstream")
				}
				return nil
			}
			if err := validateTLSOptions(tlsOpt); err != nil {
				return err
			}
		}
	case "direct", "reject":
		if upstream != "" {
			return fmt.Errorf("method=%q must not define upstream", method)
		}
		if explicitTLS && hasTLSOptions(tlsOpt) {
			return fmt.Errorf("method=%q must not define tls options", method)
		}
	default:
		return fmt.Errorf("unsupported method %q", method)
	}
	return nil
}

func validateTLSOptions(tlsOpt *UpstreamTLS) error {
	if tlsOpt == nil {
		return nil
	}
	if tlsOpt.CACertFile != "" {
		pemData, err := os.ReadFile(tlsOpt.CACertFile)
		if err != nil {
			return fmt.Errorf("failed to read tls.ca_cert_file %q: %w", tlsOpt.CACertFile, err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pemData) {
			return fmt.Errorf("failed to parse PEM certificates from tls.ca_cert_file %q", tlsOpt.CACertFile)
		}
	}
	if tlsOpt.MinVersion != "" && !isAllowedTLSMinVersion(tlsOpt.MinVersion) {
		return fmt.Errorf("unsupported tls.min_version %q (allowed: 1.2, 1.3)", tlsOpt.MinVersion)
	}
	return nil
}

func hasTLSOptions(tlsOpt *UpstreamTLS) bool {
	if tlsOpt == nil {
		return false
	}
	return tlsOpt.InsecureSkipVerify || tlsOpt.CACertFile != "" || tlsOpt.MinVersion != ""
}

// ResolveTLS resolves effective TLS settings with per-rule/default override.
//
// If override is present, it fully replaces global defaults.
func ResolveTLS(global *UpstreamTLS, override *UpstreamTLS) *UpstreamTLS {
	if override != nil {
		return override
	}
	if global != nil {
		return global
	}
	return nil
}

// ParseUpstreamEndpoint parses upstream in URL form.
// Supported URL schemes are http and https.
func ParseUpstreamEndpoint(raw string) (UpstreamEndpoint, error) {
	if raw == "" {
		return UpstreamEndpoint{}, errors.New("upstream is empty")
	}
	if !strings.Contains(raw, "://") {
		return UpstreamEndpoint{}, errors.New("upstream must include scheme (http:// or https://)")
	}

	u, err := url.Parse(raw)
	if err != nil {
		return UpstreamEndpoint{}, err
	}
	switch u.Scheme {
	case "http", "https":
	default:
		return UpstreamEndpoint{}, fmt.Errorf("unsupported upstream scheme %q", u.Scheme)
	}
	if u.Host == "" {
		return UpstreamEndpoint{}, errors.New("upstream host is empty")
	}
	if u.Path != "" && u.Path != "/" {
		return UpstreamEndpoint{}, errors.New("upstream URL must not include path")
	}
	if u.RawQuery != "" || u.Fragment != "" || u.User != nil {
		return UpstreamEndpoint{}, errors.New("upstream URL must not include query, fragment, or userinfo")
	}
	if err := validateHostPort(u.Host); err != nil {
		return UpstreamEndpoint{}, err
	}
	host, _, _ := net.SplitHostPort(u.Host)
	return UpstreamEndpoint{
		Scheme:     u.Scheme,
		Address:    u.Host,
		ServerName: host,
	}, nil
}

// HasNonRejectDefault reports whether routing.default exists and is not reject.
func (c Config) HasNonRejectDefault() bool {
	return c.Routing.Default != nil && c.Routing.Default.Method != "reject"
}

func isAllowedLogLevel(v string) bool {
	switch v {
	case "debug", "info", "warn", "error":
		return true
	default:
		return false
	}
}

func isAllowedTLSMinVersion(v string) bool {
	switch v {
	case "1.2", "1.3":
		return true
	default:
		return false
	}
}
