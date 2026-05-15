// File acl.go contains CIDR and port allowlist implementations.
package acl

import (
	"fmt"
	"net"

	"socks2proxy/internal/addressspec"
)

// AddressAllowlist checks whether a client IP is allowed by configured specs.
type AddressAllowlist struct {
	specs []string
}

// NewAddressAllowlist builds a client IP allowlist using address selectors.
func NewAddressAllowlist(specs []string) (*AddressAllowlist, error) {
	list := &AddressAllowlist{specs: make([]string, 0, len(specs))}
	for _, spec := range specs {
		if err := addressspec.Validate(spec); err != nil {
			return nil, fmt.Errorf("invalid allowlist address spec %q: %w", spec, err)
		}
		list.specs = append(list.specs, spec)
	}
	return list, nil
}

// Contains reports whether the given IP is allowed.
func (c *AddressAllowlist) Contains(ip net.IP) bool {
	if c == nil || len(c.specs) == 0 {
		return false
	}
	for _, spec := range c.specs {
		if addressspec.Matches(spec, ip) {
			return true
		}
	}
	return false
}

// PortAllowlist checks whether a destination port is permitted.
type PortAllowlist struct {
	allowed  map[int]struct{}
	allowAll bool
}

// NewPortAllowlist builds a destination port allowlist.
func NewPortAllowlist(ports []int) *PortAllowlist {
	m := make(map[int]struct{}, len(ports))
	for _, p := range ports {
		m[p] = struct{}{}
	}
	return &PortAllowlist{allowed: m}
}

// NewAllowAllPortAllowlist returns a port ACL that allows any port.
func NewAllowAllPortAllowlist() *PortAllowlist {
	return &PortAllowlist{allowAll: true}
}

// Allows reports whether a destination port is allowed.
func (p *PortAllowlist) Allows(port int) bool {
	if p == nil {
		return false
	}
	if p.allowAll {
		return true
	}
	_, ok := p.allowed[port]
	return ok
}
