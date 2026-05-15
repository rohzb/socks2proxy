// Package addressspec provides validation and matching for address selectors.
package addressspec

import (
	"errors"
	"net"
	"strings"
)

// Validate checks whether spec is one of:
// - single IP
// - CIDR
// - IP range in form "IP-IP"
func Validate(spec string) error {
	if ip := net.ParseIP(spec); ip != nil {
		return nil
	}
	if _, _, err := net.ParseCIDR(spec); err == nil {
		return nil
	}
	parts := strings.Split(spec, "-")
	if len(parts) != 2 {
		return errors.New("must be single IP, CIDR, or range IP-IP")
	}
	start := net.ParseIP(strings.TrimSpace(parts[0]))
	end := net.ParseIP(strings.TrimSpace(parts[1]))
	if start == nil || end == nil {
		return errors.New("invalid IP range bounds")
	}
	start4 := start.To4()
	end4 := end.To4()
	if (start4 == nil) != (end4 == nil) {
		return errors.New("IP range bounds must be same family")
	}
	if start4 != nil {
		if bytesCompare(start4, end4) > 0 {
			return errors.New("IP range start must be <= end")
		}
		return nil
	}
	s16 := start.To16()
	e16 := end.To16()
	if s16 == nil || e16 == nil {
		return errors.New("invalid IP range bounds")
	}
	if bytesCompare(s16, e16) > 0 {
		return errors.New("IP range start must be <= end")
	}
	return nil
}

// Matches reports whether ip matches spec.
func Matches(spec string, ip net.IP) bool {
	if single := net.ParseIP(spec); single != nil {
		return single.Equal(ip)
	}
	if _, network, err := net.ParseCIDR(spec); err == nil {
		return network.Contains(ip)
	}
	parts := strings.Split(spec, "-")
	if len(parts) != 2 {
		return false
	}
	start := net.ParseIP(strings.TrimSpace(parts[0]))
	end := net.ParseIP(strings.TrimSpace(parts[1]))
	if start == nil || end == nil {
		return false
	}
	ip4 := ip.To4()
	start4 := start.To4()
	end4 := end.To4()
	if ip4 != nil && start4 != nil && end4 != nil {
		return bytesCompare(ip4, start4) >= 0 && bytesCompare(ip4, end4) <= 0
	}
	ip16 := ip.To16()
	start16 := start.To16()
	end16 := end.To16()
	if ip16 == nil || start16 == nil || end16 == nil {
		return false
	}
	return bytesCompare(ip16, start16) >= 0 && bytesCompare(ip16, end16) <= 0
}

func bytesCompare(a, b []byte) int {
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return 1
	}
	return 0
}
