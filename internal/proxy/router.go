// File router.go selects and dispatches per-port proxy forwarding behavior.
package proxy

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"socks2proxy/internal/addressspec"
	"socks2proxy/internal/config"
)

// Method is the routing action selected for a matched rule.
type Method string

const (
	// MethodHTTP routes traffic using upstream HTTP proxying semantics.
	MethodHTTP Method = "http"
	// MethodConnect routes traffic using upstream CONNECT tunneling.
	MethodConnect Method = "connect"
	// MethodDirect routes traffic by dialing targets directly.
	MethodDirect Method = "direct"
	// MethodReject rejects traffic without forwarding.
	MethodReject Method = "reject"
)

// Rule defines one routing rule for destination port/address matching.
type Rule struct {
	DstPorts     map[int]struct{}
	DstAddresses []string
	Method       Method
	Upstream     string
	TLS          *config.UpstreamTLS
}

// Router selects routing rules and dispatches traffic handling.
type Router struct {
	ConnectTimeout time.Duration
	IdleTimeout    time.Duration
	MaxHeaderBytes int
	Rules          []Rule
	DefaultRule    Rule
}

// RulesFromConfig converts validated config rules to runtime proxy rules.
func RulesFromConfig(cfgRouting config.Routing, globalTLS *config.UpstreamTLS) []Rule {
	cfgRules := cfgRouting.Rules
	rules := make([]Rule, 0, len(cfgRules))
	for _, r := range cfgRules {
		pm := make(map[int]struct{})
		for _, p := range r.DstPorts {
			pm[p] = struct{}{}
		}
		for _, p := range r.DstPort {
			pm[p] = struct{}{}
		}
		addresses := make([]string, 0, len(r.DstAddresses)+len(r.DstAddress))
		addresses = append(addresses, r.DstAddresses...)
		addresses = append(addresses, r.DstAddress...)
		rules = append(rules, Rule{
			DstPorts:     pm,
			DstAddresses: addresses,
			Method:       Method(r.Method),
			Upstream:     r.Upstream,
			TLS:          config.ResolveTLS(globalTLS, r.TLS),
		})
	}
	return rules
}

// DefaultRuleFromConfig converts routing.default to runtime default behavior.
func DefaultRuleFromConfig(cfgRouting config.Routing, globalTLS *config.UpstreamTLS) Rule {
	if cfgRouting.Default == nil {
		return Rule{Method: MethodReject}
	}
	return Rule{
		Method:   Method(cfgRouting.Default.Method),
		Upstream: cfgRouting.Default.Upstream,
		TLS:      config.ResolveTLS(globalTLS, cfgRouting.Default.TLS),
	}
}

// Route forwards a target flow using the configured per-port routing rule.
func (r *Router) Route(client net.Conn, targetHost string, targetPort int) error {
	rule := r.DefaultRule
	for _, candidate := range r.Rules {
		if !candidate.matchesPort(targetPort) {
			continue
		}
		if !candidate.matchesAddress(targetHost) {
			continue
		}
		rule = candidate
		break
	}

	if rule.Method == MethodReject {
		return errors.New("target rejected by routing rule")
	}
	if rule.Method == MethodDirect {
		return HandleDirect(client, targetHost, targetPort, r.ConnectTimeout, r.IdleTimeout)
	}

	dialer := net.Dialer{Timeout: r.ConnectTimeout}
	ctx, cancel := context.WithTimeout(context.Background(), r.ConnectTimeout)
	defer cancel()

	upstreamEndpoint, err := config.ParseUpstreamEndpoint(rule.Upstream)
	if err != nil {
		return fmt.Errorf("parse upstream proxy endpoint: %w", err)
	}

	upstream, err := dialUpstream(ctx, dialer, upstreamEndpoint, rule.TLS)
	if err != nil {
		return fmt.Errorf("dial upstream proxy: %w", err)
	}
	defer upstream.Close()

	_ = upstream.SetDeadline(time.Now().Add(r.IdleTimeout))
	_ = client.SetDeadline(time.Now().Add(r.IdleTimeout))

	switch rule.Method {
	case MethodHTTP:
		return HandleHTTP(client, upstream, targetHost, r.MaxHeaderBytes)
	case MethodConnect:
		return HandleConnect(client, upstream, targetHost, targetPort)
	default:
		return fmt.Errorf("unsupported routing method %q", rule.Method)
	}
}

func (r Rule) matchesPort(port int) bool {
	_, ok := r.DstPorts[port]
	return ok
}

func (r Rule) matchesAddress(host string) bool {
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	for _, spec := range r.DstAddresses {
		if addressspec.Matches(spec, ip) {
			return true
		}
	}
	return false
}

func dialUpstream(ctx context.Context, dialer net.Dialer, endpoint config.UpstreamEndpoint, tlsOpt *config.UpstreamTLS) (net.Conn, error) {
	switch endpoint.Scheme {
	case "http":
		return dialer.DialContext(ctx, "tcp", endpoint.Address)
	case "https":
		minTLSVersion := uint16(tls.VersionTLS12)
		if tlsOpt != nil {
			switch tlsOpt.MinVersion {
			case "1.3":
				minTLSVersion = tls.VersionTLS13
			case "1.2", "":
				minTLSVersion = tls.VersionTLS12
			}
		}
		conf := &tls.Config{
			MinVersion: minTLSVersion,
			ServerName: endpoint.ServerName,
		}
		if tlsOpt != nil {
			conf.InsecureSkipVerify = tlsOpt.InsecureSkipVerify
			if tlsOpt.CACertFile != "" {
				pemData, err := os.ReadFile(tlsOpt.CACertFile)
				if err != nil {
					return nil, fmt.Errorf("read tls.ca_cert_file %q: %w", tlsOpt.CACertFile, err)
				}
				pool := x509.NewCertPool()
				if !pool.AppendCertsFromPEM(pemData) {
					return nil, fmt.Errorf("parse PEM certificates from tls.ca_cert_file %q", tlsOpt.CACertFile)
				}
				conf.RootCAs = pool
			}
		}
		tlsDialer := &tls.Dialer{
			NetDialer: &dialer,
			Config:    conf,
		}
		return tlsDialer.DialContext(ctx, "tcp", endpoint.Address)
	default:
		return nil, fmt.Errorf("unsupported upstream scheme %q", endpoint.Scheme)
	}
}

// newBufferedReader returns a buffered reader for a network connection.
func newBufferedReader(conn net.Conn) *bufio.Reader {
	return bufio.NewReader(conn)
}

// ParsePort parses and validates a TCP port number from string form.
func ParsePort(portStr string) (int, error) {
	p, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, fmt.Errorf("invalid port %q: %w", portStr, err)
	}
	if p < 1 || p > 65535 {
		return 0, fmt.Errorf("port out of range: %d", p)
	}
	return p, nil
}
