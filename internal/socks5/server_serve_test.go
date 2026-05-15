package socks5

import (
	"errors"
	"net"
	"testing"
	"time"

	"socks2proxy/internal/acl"
	"socks2proxy/internal/logging"
	"socks2proxy/internal/proxy"
)

type fakeAddr string

func (a fakeAddr) Network() string { return "tcp" }
func (a fakeAddr) String() string  { return string(a) }

type fakeListener struct{}

func (f *fakeListener) Accept() (net.Conn, error) { return nil, errors.New("boom") }
func (f *fakeListener) Close() error              { return nil }
func (f *fakeListener) Addr() net.Addr            { return fakeAddr("127.0.0.1:0") }

func TestServeReturnsListenError(t *testing.T) {
	orig := listenFunc
	listenFunc = func(network, address string) (net.Listener, error) { return nil, errors.New("listen-fail") }
	t.Cleanup(func() { listenFunc = orig })

	clientACL, err := acl.NewAddressAllowlist([]string{"127.0.0.1/32"})
	if err != nil {
		t.Fatalf("acl: %v", err)
	}
	s := &Server{
		ListenAddr:  "127.0.0.1:0",
		ClientACL:   clientACL,
		PortACL:     acl.NewAllowAllPortAllowlist(),
		Router:      &proxy.Router{DefaultRule: proxy.Rule{Method: proxy.MethodReject}},
		Logger:      logging.New("error"),
		IdleTimeout: time.Second,
	}
	if err := s.Serve(); err == nil {
		t.Fatalf("expected listen error")
	}
}

func TestServeReturnsAcceptError(t *testing.T) {
	orig := listenFunc
	listenFunc = func(network, address string) (net.Listener, error) { return &fakeListener{}, nil }
	t.Cleanup(func() { listenFunc = orig })

	clientACL, err := acl.NewAddressAllowlist([]string{"127.0.0.1/32"})
	if err != nil {
		t.Fatalf("acl: %v", err)
	}
	s := &Server{
		ListenAddr:  "127.0.0.1:0",
		ClientACL:   clientACL,
		PortACL:     acl.NewAllowAllPortAllowlist(),
		Router:      &proxy.Router{DefaultRule: proxy.Rule{Method: proxy.MethodReject}},
		Logger:      logging.New("error"),
		IdleTimeout: time.Second,
	}
	if err := s.Serve(); err == nil {
		t.Fatalf("expected serve to return accept error")
	}
}
