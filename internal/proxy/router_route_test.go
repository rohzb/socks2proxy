package proxy

import (
	"net"
	"strings"
	"testing"
	"time"
)

func TestRouterRouteRejectAndUnsupported(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	r := &Router{DefaultRule: Rule{Method: MethodReject}}
	if err := r.Route(c1, "10.0.0.1", 80); err == nil {
		t.Fatalf("expected reject route error")
	}

	r = &Router{
		ConnectTimeout: 50 * time.Millisecond,
		IdleTimeout:    50 * time.Millisecond,
		Rules: []Rule{{
			DstPorts:     map[int]struct{}{80: {}},
			DstAddresses: []string{"0.0.0.0/0"},
			Method:       Method("weird"),
			Upstream:     "http://127.0.0.1:1",
		}},
		DefaultRule: Rule{Method: MethodReject},
	}
	if err := r.Route(c1, "10.0.0.1", 80); err == nil {
		t.Fatalf("expected unsupported method path to fail")
	}
}

func TestRouterRouteInvalidUpstreamAndDialFailure(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	r := &Router{
		ConnectTimeout: 50 * time.Millisecond,
		IdleTimeout:    50 * time.Millisecond,
		Rules: []Rule{{
			DstPorts:     map[int]struct{}{80: {}},
			DstAddresses: []string{"0.0.0.0/0"},
			Method:       MethodHTTP,
			Upstream:     "bad-upstream",
		}},
		DefaultRule: Rule{Method: MethodReject},
	}
	err := r.Route(c1, "10.0.0.1", 80)
	if err == nil || !strings.Contains(err.Error(), "parse upstream proxy endpoint") {
		t.Fatalf("expected parse upstream error, got: %v", err)
	}

	r.Rules[0].Upstream = "http://127.0.0.1:1"
	err = r.Route(c1, "10.0.0.1", 80)
	if err == nil || !strings.Contains(err.Error(), "dial upstream proxy") {
		t.Fatalf("expected dial upstream error, got: %v", err)
	}
}
